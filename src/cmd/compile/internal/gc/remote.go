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
	log.Printf("remote:analyze Package: \n%s", buf)
	for n, m := range pkg.Members {
		if f, ok := m.(*ssa.Function); ok {
			buf.Reset()
			ssa.WriteFunction(buf, f)
			log.Printf("  Function (%s): \n%s", n, buf)
		}
	}
}

func escapesRemote(args []string, all []*Node) {
	// Compile SSA
	var conf loader.Config
	rest, err := conf.FromArgs(args, false) // no tests
	if err != nil {
		log.Fatalf("remote:analyze: %v", err)
	}
	if len(rest) != 0 {
		log.Fatalf("remote:analyze: unused args %v", rest)
	}
	iprog, err := conf.Load()
	if err != nil {
		log.Fatalf("remote:analyze: %v", err)
	}
	fset := iprog.Fset
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
	r1, err := pointer.Analyze(config)
	cg := r1.CallGraph

	// Find all go call sites (gosites) and their possible callees.
	gosites := make(map[ssa.CallInstruction][]*callgraph.Edge)
	callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		if _, ok := edge.Site.(*ssa.Go); ok {
			gosites[edge.Site] = append(gosites[edge.Site], edge)
		}
		return nil
	})

	if Debug_remote > 0 {
		for site, edges := range gosites {
			for _, edge := range edges {
				log.Printf("remote: analyse: 'go' call site at %v --> %v", fset.Position(site.Pos()), edge.Callee)
			}
		}
	}

	// asyncFuncs are any functions that may run asynchronously to the 'root'
	// goroutine. They are all functions downstream of the gosites found above.
	asyncFuncs := make(map[*callgraph.Node]struct{})
	var visit func(*callgraph.Node)
	visit = func(n *callgraph.Node) {
		if _, ok := asyncFuncs[n]; !ok {
			asyncFuncs[n] = struct{}{}
			for _, next := range n.Out {
				visit(next.Callee)
			}
		}
	}
	for _, edges := range gosites {
		for _, edge := range edges {
			visit(edge.Callee)
		}
	}

	if Debug_remote > 0 {
		for af, _ := range asyncFuncs {
			log.Printf("remote: analyse: async func %v", af)
		}
	}

	// Any references to a global among the asyncFuncs results in that
	// global receiving remote allocation.
	remoteGlobals := make(map[*ssa.Global]struct{}, 0)
	operands := make([]*ssa.Value, 0)
	scanBlock := func(blk *ssa.BasicBlock) {
		if blk == nil {
			return
		}
		for _, inst := range blk.Instrs {
			operands = inst.Operands(operands[:0])
			for _, o := range operands {
				if g, ok := (*o).(*ssa.Global); ok {
					remoteGlobals[g] = struct{}{}
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
			log.Printf("remote: analyse: global for remote allocation: %v", rg)
		}
	}

	// In SSA form, instructions have Operands (*ssa.Value's) and ssa.Value's
	// have Referrers (*ssa.Instructions). However globals (named functions and
	// global variables) are not populated with Referrers, so we need to build
	// a list of referrers for globals here.
	// referrersToGlobals := make(map[*ssa.Global][]ssa.Instruction)
	// visited := make(map[*callgraph.Node]struct{})

	// Using call graph:
	//  - Find all pointer-like vars transiting `go` function calls
	//    (as arguments, method receivers or free variables)
	//  - Find all pointer-like global vars accessed across `go` function calls
	//  - Find all vars sent/received over channels

	// Pointer analysis #2: query above pointer-like vars for allocation sites

	// Map to xtop syntax tree
	// Re-write allocation sites for remote allocation
	// Re-write access sites

	// Resources
	// ---------

	// Call graph edge traversal:
	// result, err := pointer.Analyze(config)
	// var edges []string
	// callgraph.GraphVisitEdges(result.CallGraph, func(edge *callgraph.Edge) error {
	// 	caller := edge.Caller.Func
	// 	if caller.Pkg == mainPkg {
	// 		edges = append(edges, fmt.Sprint(caller, " --> ", edge.Callee.Func))
	// 	}
	// 	return nil
	// })

	// parse tree output:
	// if Debug_remote > 0 {
	// 	dumplist("remote: ", Nodes{&all})
	// }
}
