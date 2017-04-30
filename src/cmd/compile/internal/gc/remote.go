// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cmd_go_bootstrap

package gc

import (
	"bytes"
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

func escapesRemote(args []string, all []*Node) {
	// Compile SSA
	var conf loader.Config
	rest, err := conf.FromArgs(args, false) // no tests
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
	fset := iprog.Fset
	if Debug_remote > 0 {
		for p, _ := range iprog.AllPackages {
			log.Printf("remote: analyze: package %s", p.Name())
		}
	}
	prog := ssautil.CreateProgram(iprog, 0)
	mainPkg := prog.Package(iprog.Created[0].Pkg)
	prog.Build()

	if Debug_remote > 0 {
		logPkg(mainPkg)
	}

	// Build call graph
	config := &pointer.Config{
		Mains:          []*ssa.Package{mainPkg},
		BuildCallGraph: true,
	}
	ptrres, err := pointer.Analyze(config)
	if err != nil {
		log.Fatalf("remote:analyze: %v", err)
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
		log.Fatalf("remote:analyze: %v", err)
	}

	if Debug_remote > 0 {
		for site, edges := range gosites {
			for _, edge := range edges {
				log.Printf("remote: analyze: gosite of %v at %v", edge.Callee, fset.Position(site.Pos()))
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
			log.Printf("remote: analyze: asyncFunc %v at %v", af, fset.Position(af.Func.Pos()))
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
			log.Printf("remote: analyze: remoteGlobal %v at %v", rg, fset.Position(rg.Pos()))
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
			log.Printf("remote: analyze: bridgeVar %v at %v", v, fset.Position(v.Pos()))
			for _, r := range *v.Referrers() {
				log.Printf("                       ref %v at %v", r, fset.Position(r.Pos()))
			}
		}
	}

	// Expand bridgeVars to their points-to sets
	config = &pointer.Config{
		Mains:          []*ssa.Package{mainPkg},
		BuildCallGraph: false,
	}
	for v, _ := range bridgeVars {
		config.AddQuery(v)
	}
	ptrres, err = pointer.Analyze(config)
	if err != nil {
		log.Fatalf("remote:analyze: %v", err)
	}
	remoteVars := make(map[ssa.Value][]*pointer.Label)
	for _, ptr := range ptrres.Queries {
		for _, l := range ptr.PointsTo().Labels() {
			remoteVars[l.Value()] = append(remoteVars[l.Value()], l)
		}
	}

	if Debug_remote > 0 {
		for v, _ := range remoteVars {
			log.Printf("remote: analyze: remoteVar %v declared at %v", v, fset.Position(v.Pos()))
			for _, r := range *v.Referrers() {
				log.Printf("                       ref %v at %v", r, fset.Position(r.Pos()))
			}
		}
	}

	// TODO: map remoteGlobals and remoteVars to declarations in parse tree,
	// rewrite:
	//  - type --> uint64
	//  - declaration/allocation --> remote api Alloc()
	//  - access --> remote api Get/Set() and static cast

	// parse tree output:
	if Debug_remote > 0 {
		dumplist("remote: ", Nodes{&all})
	}
}
