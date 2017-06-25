#!/usr/bin/env bash
# Copyright 2016 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

function cleanup() {
    rm -f a.out
}
trap cleanup EXIT

# go build -buildmode=remote -o a.out -gcflags="-d remote"
go build -o a.out
OUTPUT=$(go remote ./a.out 2>&1)
EXPECTED="var1: 1
var2: 4
var3: 5
lib RemoteVar: 42"
if [ "$OUTPUT" != "$EXPECTED" ]; then
  echo invalid output:
  echo $OUTPUT
  exit 1
fi
