// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// No-op implementations of functions in remote.go for use during toolchain
// bootstrapping.

// +build cmd_go_bootstrap

package gc

// set by command line flag -remote
var flag_remote bool

func analyzeRemote([]string) {}
func globalRemote([]*Node)   {}
func escapesRemote([]*Node)  {}
