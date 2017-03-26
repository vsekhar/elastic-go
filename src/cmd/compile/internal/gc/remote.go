// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cmd_go_bootstrap

package gc

import (
	"log"

	"golang.org/x/tools/go/loader"
)

// set by command line flag -remote
var flag_remote bool

// Identifies pointers that leak addresses across goroutines:
//  - Pointers sent/received over channels
//  - Pointers passed as arguments into go function calls
//  - Pointers passed as method receivers into go function calls
//  - Pointers passed as free variables into closures
//
// Then identifies using points-to analysis all allocation sites to which each
// of those pointers may point and tags them for remote allocation:
//  - Explicit allocation via var
//  - Explicit allocation via new
//  - Variables whose address is taken (&var) and corresponding allocation site
//  - Allocation of function arguments
func analyzeRemote(args []string) {
	var conf loader.Config
	rest, err := conf.FromArgs(args, false) // no tests
	if err != nil {
		log.Fatalf("remote:analyse: %v", err)
	}
	if len(rest) != 0 {
		log.Fatalf("remote:analyze: unused args %v", rest)
	}
	_, err = conf.Load()
}

func globalRemote(all []*Node) {
	// 1. build call graph
	// 2. Populate list of goroutine "root" functions
	// 3. Find all OPROC calls and add callees to root functions
	// 4. Assign number to each goroutine root (array index?)
	// 5. Go through call graph and tag functions with goroutine root numbers
	// 6. Go through call graph and tag PEXTERN vars with goroutine root numbers
	// 7. Go through PEXTERN vars, relable those with multiple goroutine root numbers as PAUTOREMOTE
}

func escapesRemote(all []*Node) {
	// TODO(vsekhar): use visitBottomUp and other helpers from esc.go
	// visitBottomUp calls analyze() with a set of functions that only call
	// each other or previously analyze'd functions. Useful?

	if Debug_remote > 0 {
		dumplist("remote: ", Nodes{&all})
	}
	globalRemote(all)
}
