// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cmd_go_bootstrap

package gc

import (
	"bytes"
	"flag"
	"go/types"
	"log"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// set by command line flag -remote
var flag_remote bool

func logPkg(pkg *ssa.Package) {
	buf := new(bytes.Buffer)
	ssa.WritePackage(buf, pkg)
	log.Printf("remote: analyze: logPkg: \n%s", buf)
	for n, m := range pkg.Members {
		if f, ok := m.(*ssa.Function); ok {
			buf.Reset()
			ssa.WriteFunction(buf, f)
			log.Printf("  Function (%s): \n%s", n, buf)
		}
	}
}

// visitNodes calls f with each node starting with root and proceeding down
// callgraph edges depth-first. (see also callgraph.GraphVisitEdges)
// visitNodes visits the root and all downstream nodes in depth-first order.
// The node function is called for each edge in postorder.  If it
// returns non-nil, visitation stops and visitNodes returns that value.
// (see also callgraph.GraphVisitEdges)
func visitNodes(root *callgraph.Node, node func(*callgraph.Node) error) error {
	seen := make(map[*callgraph.Node]bool)
	var visit func(*callgraph.Node) error
	visit = func(n *callgraph.Node) error {
		if !seen[n] {
			seen[n] = true
			for _, e := range n.Out {
				if err := visit(e.Callee); err != nil {
					return err
				}
			}
			if err := node(n); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(root); err != nil {
		return err
	}
	return nil
}

// addQueries adds points-to queries to the provided ptaConf from the values
// provided in vars. Pointer values are added to ptrs from which results can
// be read after the analysis is run.
func addQueries(ptaConf *pointer.Config, ptrs map[*pointer.Pointer]struct{}, vars map[ssa.Value]struct{}) {
	for v := range vars {
		// Query pointer-like objects directly
		if pointer.CanPoint(v.Type()) {
			ptr, err := ptaConf.AddExtendedQuery(v, "x")
			if err != nil {
				panic(err)
			}
			ptrs[ptr] = struct{}{}
		}

		if p, ok := v.Type().Underlying().(*types.Pointer); ok {
			e := p.Elem()

			// handlePointee adds queries related to the pointee (if any). It is
			// expressed as function in order to recurse into sub-types as needed.
			var handlePointee func(types.Type)
			handlePointee = func(t types.Type) {
				if pointer.CanPoint(t) {
					ptr, err := ptaConf.AddExtendedQuery(v, "*x")
					if err != nil {
						panic(err)
					}
					ptrs[ptr] = struct{}{}
				}

				switch x := t.(type) {
				case *types.Named:
					handlePointee(x.Underlying())
				case *types.Struct:
					for i := 0; i < x.NumFields(); i++ {
						f := x.Field(i)
						if pointer.CanPoint(f.Type()) {
							ptr, err := ptaConf.AddExtendedQuery(v, "x."+f.Name())
							if err != nil {
								panic(err)
							}
							ptrs[ptr] = struct{}{}
						}
					}
				}
			}
			handlePointee(e)
		}
	}
}

// escapesRemote is called on each invocation of cmd/compile, but is a no-op
// on each invocation except the one during which the main package is being
// compiled.
func escapesRemote() {
	if Debug_remote > 0 {
		log.Printf("remote: escapesRemote invoked with: %v", flag.Args())
	}

	// Load the requested files from args
	var conf loader.Config
	rest, err := conf.FromArgs(flag.Args(), false) // no tests
	if err != nil {
		log.Fatalf("remote: analyze: %v", err)
	}
	if len(rest) != 0 {
		log.Fatalf("remote: analyze: unused args %v", rest)
	}
	iprog, err := conf.Load()
	if err != nil {
		log.Fatalf("remote: analyze: %v", err)
	}

	// Determine if there is a main, otherwise return
	var mainPkg *types.Package
	for p, _ := range iprog.AllPackages {
		if p.Name() == "main" {
			mainPkg = p
		}
	}
	if mainPkg == nil {
		if Debug_remote > 0 {
			log.Print("remote: main package not found in this invocation")
		}
		return
	}


	if Debug_remote > 0 {
		for p, _ := range iprog.AllPackages {
			log.Printf("remote: analyzing %s", p.Name())
		}
	}

	prog := ssautil.CreateProgram(iprog, 0)
	mainPkgSSA := prog.Package(mainPkg)
	prog.Build()
	config := &pointer.Config{
		Mains:          []*ssa.Package{mainPkgSSA},
		BuildCallGraph: true,
	}
	ptrres, err := pointer.Analyze(config)
	if err != nil {
		log.Fatalf("remote: analyze: %v", err)
	}
	cg := ptrres.CallGraph

	// Find all go call sites (gosites) and their possible callees.
	gosites := make(map[ssa.CallInstruction][]*callgraph.Edge)
	err = callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		if _, ok := edge.Site.(*ssa.Go); ok {
			gosites[edge.Site] = append(gosites[edge.Site], edge)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("remote: analyze: %v", err)
	}

	if Debug_remote > 0 {
		for site, edges := range gosites {
			for _, edge := range edges {
				log.Printf("remote: analyze: found gosite of %v at %v", edge.Callee, iprog.Fset.Position(site.Pos()))
			}
		}
	}

	// asyncFuncs are any functions that may run asynchronously to the 'root'
	// goroutine. They are all functions downstream of the gosites found above.
	asyncFuncs := make(map[*callgraph.Node]struct{})
	for _, edges := range gosites {
		for _, edge := range edges {
			err := visitNodes(edge.Callee, func(n *callgraph.Node) error {
				asyncFuncs[n] = struct{}{}
				return nil
			})
			if err != nil {
				log.Fatalf("remote: analyze: %v", err)
			}
		}
	}

	if Debug_remote > 0 {
		for af, _ := range asyncFuncs {
			log.Printf("remote: analyze: found asyncFunc %v at %v", af, iprog.Fset.Position(af.Func.Pos()))
		}
	}

	// Any references to a global among the asyncFuncs results in that
	// global receiving remote allocation.
	remoteGlobals := make(map[*ssa.Global][]ssa.Instruction, 0)
	operands := make([]*ssa.Value, 0)
	scanBlock := func(blk *ssa.BasicBlock) {
		if blk == nil {
			return
		}
		for _, inst := range blk.Instrs {
			operands = inst.Operands(operands[:0])
			for _, o := range operands {
				if g, ok := (*o).(*ssa.Global); ok {
					remoteGlobals[g] = append(remoteGlobals[g], inst)
				}
			}
		}
	}
	var scanFunc func(*ssa.Function)
	scanFunc = func(f *ssa.Function) {
		for _, blk := range f.Blocks {
			scanBlock(blk)
		}
		scanBlock(f.Recover)
		for _, af := range f.AnonFuncs {
			scanFunc(af)
		}
	}
	for n, _ := range asyncFuncs {
		scanFunc(n.Func)
	}

	if Debug_remote > 0 {
		for rg, _ := range remoteGlobals {
			log.Printf("remote: analyze: found remoteGlobal %v at %v", rg, iprog.Fset.Position(rg.Pos()))
		}
	}

	// Gather pointer-like vars from gosites in bridgeVars
	bridgeVars := make(map[ssa.Value]struct{})
	for _, edges := range gosites {
		for _, edge := range edges {
			for _, p := range edge.Callee.Func.Params {
				if pointer.CanPoint(p.Type()) {
					bridgeVars[p] = struct{}{}
				}
			}
			for _, fv := range edge.Callee.Func.FreeVars {
				if pointer.CanPoint(fv.Type()) {
					bridgeVars[fv] = struct{}{}
				}
			}
		}
	}

	if Debug_remote > 0 {
		for v, _ := range bridgeVars {
			log.Printf("remote: analyze: found bridgeVar %v at %v", v, iprog.Fset.Position(v.Pos()))
		}
	}

	// Expand bridgeVars to their points-to sets
	ptaConf := &pointer.Config{
		Mains:          []*ssa.Package{mainPkgSSA},
		BuildCallGraph: false,
	}
	ptrs := make(map[*pointer.Pointer]struct{})
	addQueries(ptaConf, ptrs, bridgeVars)
	remoteVars := make(map[ssa.Value]struct{})
	numRemoteVars := 0
	rounds := 0
	for {
		rounds++
		addQueries(ptaConf, ptrs, remoteVars)
		_, err := pointer.Analyze(ptaConf)
		if err != nil {
			log.Fatalf("remote: analyze: %v", err)
		}
		for ptr := range ptrs {
			for _, l := range ptr.PointsTo().Labels() {
				remoteVars[l.Value()] = struct{}{}
			}
		}

		// Nothing new? Then stop here.
		if len(remoteVars) == numRemoteVars {
			break
		}

		// Prepare for next attempt
		ptrs = make(map[*pointer.Pointer]struct{})
		ptaConf = &pointer.Config{
			Mains:          []*ssa.Package{mainPkgSSA},
			BuildCallGraph: false,
		}
		numRemoteVars = len(remoteVars)
	}
	if Debug_remote > 0 {
		log.Printf("Analysis took %d rounds", rounds)
	}

	if Debug_remote > 0 {
		for v, _ := range remoteVars {
			log.Printf("remote: analyze: found remoteVar %v declared at %v", v, iprog.Fset.Position(v.Pos()))
		}
	}
}
