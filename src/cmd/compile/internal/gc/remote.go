// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"fmt"
)

var flag_remote bool

func escapesRemote(all []*Node) {
	// TODO(vsekhar): use visitBottomUp and other helpers from esc.go
	// visitBottomUp calls analyze() with a set of functions that only call
	// each other or previously analyze'd functions. Useful?

	// debug logging
	if Debug['m'] != 0 {
		fmt.Printf("remote: escape analysis performed")
	}
}
