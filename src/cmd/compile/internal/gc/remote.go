// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cmd_go_bootstrap

package gc

import (
	"log"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// set by command line flag -remote
var flag_remote bool

// Traverses program and finds all ssa.Value's that are the roots to `go`
// keyword function calls
func findGoRoots(prog *ssa.Program) []*ssa.Value {
	// TBD
	return nil
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
	prog := ssautil.CreateProgram(iprog, 0)

	// Find all function expressions at `go` call sites
	// TBD

	// Pointer analysis #1: query for all possible functions called at `go`
	// call sites; also generate callgraph
	mainPkg := prog.Package(iprog.Created[0].Pkg)
	prog.Build()
	config := &pointer.Config{
		Mains:          []*ssa.Package{mainPkg},
		BuildCallGraph: true,
	}
	r1, err := pointer.Analyze(config)
	cg := r1.CallGraph
	_ = cg

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
