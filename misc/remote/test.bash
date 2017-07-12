#!/usr/bin/env bash
# Copyright 2016 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

function cleanup() {
    rm -f a.out
}
trap cleanup EXIT

go build -v -buildmode=remote -o a.out -gcflags="-d remote"
OUTPUT=$(go remote ./a.out 2>&1)
EXPECTED="internal/remote added
var1: 1
var2: 4
var3: 5
lib RemoteVar: 42"
if [ "$OUTPUT" != "$EXPECTED" ]; then
  echo invalid output:
  echo "$OUTPUT"
  echo expected:
  echo "$EXPECTED"
  exit 1
fi
