#!/usr/bin/env bash
# Copyright 2016 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

# TODO: compile a remote program, then run it with the `go remote` command

function cleanup() {
    rm -f a.out
}
trap cleanup EXIT

go build -buildmode=remote -o a.out remote.go
# go remote ./a.out
./a.out
