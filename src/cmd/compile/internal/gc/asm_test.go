// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"bytes"
	"fmt"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestAssembly checks to make sure the assembly generated for
// functions contains certain expected instructions.
func TestAssembly(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	if runtime.GOOS == "windows" {
		// TODO: remove if we can get "go tool compile -S" to work on windows.
		t.Skipf("skipping test: recursive windows compile not working")
	}
	dir, err := ioutil.TempDir("", "TestAssembly")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	t.Run("platform", func(t *testing.T) {
		for _, ats := range allAsmTests {
			ats := ats
			t.Run(ats.os+"/"+ats.arch, func(tt *testing.T) {
				tt.Parallel()

				asm := ats.compileToAsm(tt, dir)

				for i, at := range ats.tests {
					fa := funcAsm(asm, i)

					at.verifyAsm(tt, fa)
				}
			})
		}
	})
}

// funcAsm returns the assembly listing for f{funcIndex}
func funcAsm(asm string, funcIndex int) string {
	if i := strings.Index(asm, fmt.Sprintf("TEXT\t\"\".f%d(SB)", funcIndex)); i >= 0 {
		asm = asm[i:]
	}

	if i := strings.Index(asm, fmt.Sprintf("TEXT\t\"\".f%d(SB)", funcIndex+1)); i >= 0 {
		asm = asm[:i+1]
	}

	return asm
}

type asmTest struct {
	// function to compile, must be named fX,
	// where X is this test's index in asmTests.tests.
	function string
	// regexps that must match the generated assembly
	regexps []string
}

func (at asmTest) verifyAsm(t *testing.T, fa string) {
	for _, r := range at.regexps {
		if b, err := regexp.MatchString(r, fa); !b || err != nil {
			t.Errorf("expected:%s\ngo:%s\nasm:%s\n", r, at.function, fa)
		}
	}
}

type asmTests struct {
	arch    string
	os      string
	imports []string
	tests   []*asmTest
}

func (ats *asmTests) generateCode() []byte {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "package main")
	for _, s := range ats.imports {
		fmt.Fprintf(&buf, "import %q\n", s)
	}

	for _, t := range ats.tests {
		fmt.Fprintln(&buf, t.function)
	}

	return buf.Bytes()
}

// compile compiles the package pkg for architecture arch and
// returns the generated assembly.  dir is a scratch directory.
func (ats *asmTests) compileToAsm(t *testing.T, dir string) string {
	// create test directory
	testDir := filepath.Join(dir, fmt.Sprintf("%s_%s", ats.arch, ats.os))
	err := os.Mkdir(testDir, 0700)
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}

	// Create source.
	src := filepath.Join(testDir, "test.go")
	err = ioutil.WriteFile(src, ats.generateCode(), 0600)
	if err != nil {
		t.Fatalf("error writing code: %v", err)
	}

	// First, install any dependencies we need.  This builds the required export data
	// for any packages that are imported.
	for _, i := range ats.imports {
		out := filepath.Join(testDir, i+".a")

		if s := ats.runGo(t, "build", "-o", out, "-gcflags=-dolinkobj=false", i); s != "" {
			t.Fatalf("Stdout = %s\nWant empty", s)
		}
	}

	// Now, compile the individual file for which we want to see the generated assembly.
	asm := ats.runGo(t, "tool", "compile", "-I", testDir, "-S", "-o", filepath.Join(testDir, "out.o"), src)

	// Get rid of code for "".init. Also gets rid of type algorithms & other junk.
	if i := strings.Index(asm, "\n\"\".init "); i >= 0 {
		asm = asm[:i+1]
	}

	return asm
}

// runGo runs go command with the given args and returns stdout string.
// go is run with GOARCH and GOOS set as ats.arch and ats.os respectively
func (ats *asmTests) runGo(t *testing.T, args ...string) string {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(testenv.GoToolPath(t), args...)
	cmd.Env = append(os.Environ(), "GOARCH="+ats.arch, "GOOS="+ats.os)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("error running cmd: %v", err)
	}

	if s := stderr.String(); s != "" {
		t.Fatalf("Stderr = %s\nWant empty", s)
	}

	return stdout.String()
}

var allAsmTests = []*asmTests{
	{
		arch:    "amd64",
		os:      "linux",
		imports: []string{"encoding/binary"},
		tests:   linuxAMD64Tests,
	},
	{
		arch:    "386",
		os:      "linux",
		imports: []string{"encoding/binary"},
		tests:   linux386Tests,
	},
	{
		arch:    "s390x",
		os:      "linux",
		imports: []string{"encoding/binary"},
		tests:   linuxS390XTests,
	},
	{
		arch:  "arm",
		os:    "linux",
		tests: linuxARMTests,
	},
	{
		arch:  "arm64",
		os:    "linux",
		tests: linuxARM64Tests,
	},
}

var linuxAMD64Tests = []*asmTest{
	{
		`
		func f0(x int) int {
			return x * 64
		}
		`,
		[]string{"\tSHLQ\t\\$6,"},
	},
	{
		`
		func f1(x int) int {
			return x * 96
		}
		`,
		[]string{"\tSHLQ\t\\$5,", "\tLEAQ\t\\(.*\\)\\(.*\\*2\\),"},
	},
	// Load-combining tests.
	{
		`
		func f2(b []byte) uint64 {
			return binary.LittleEndian.Uint64(b)
		}
		`,
		[]string{"\tMOVQ\t\\(.*\\),"},
	},
	{
		`
		func f3(b []byte, i int) uint64 {
			return binary.LittleEndian.Uint64(b[i:])
		}
		`,
		[]string{"\tMOVQ\t\\(.*\\)\\(.*\\*1\\),"},
	},
	{
		`
		func f4(b []byte) uint32 {
			return binary.LittleEndian.Uint32(b)
		}
		`,
		[]string{"\tMOVL\t\\(.*\\),"},
	},
	{
		`
		func f5(b []byte, i int) uint32 {
			return binary.LittleEndian.Uint32(b[i:])
		}
		`,
		[]string{"\tMOVL\t\\(.*\\)\\(.*\\*1\\),"},
	},
	{
		`
		func f6(b []byte) uint64 {
			return binary.BigEndian.Uint64(b)
		}
		`,
		[]string{"\tBSWAPQ\t"},
	},
	{
		`
		func f7(b []byte, i int) uint64 {
			return binary.BigEndian.Uint64(b[i:])
		}
		`,
		[]string{"\tBSWAPQ\t"},
	},
	{
		`
		func f8(b []byte, v uint64) {
			binary.BigEndian.PutUint64(b, v)
		}
		`,
		[]string{"\tBSWAPQ\t"},
	},
	{
		`
		func f9(b []byte, i int, v uint64) {
			binary.BigEndian.PutUint64(b[i:], v)
		}
		`,
		[]string{"\tBSWAPQ\t"},
	},
	{
		`
		func f10(b []byte) uint32 {
			return binary.BigEndian.Uint32(b)
		}
		`,
		[]string{"\tBSWAPL\t"},
	},
	{
		`
		func f11(b []byte, i int) uint32 {
			return binary.BigEndian.Uint32(b[i:])
		}
		`,
		[]string{"\tBSWAPL\t"},
	},
	{
		`
		func f12(b []byte, v uint32) {
			binary.BigEndian.PutUint32(b, v)
		}
		`,
		[]string{"\tBSWAPL\t"},
	},
	{
		`
		func f13(b []byte, i int, v uint32) {
			binary.BigEndian.PutUint32(b[i:], v)
		}
		`,
		[]string{"\tBSWAPL\t"},
	},
	{
		`
		func f14(b []byte) uint16 {
			return binary.BigEndian.Uint16(b)
		}
		`,
		[]string{"\tROLW\t\\$8,"},
	},
	{
		`
		func f15(b []byte, i int) uint16 {
			return binary.BigEndian.Uint16(b[i:])
		}
		`,
		[]string{"\tROLW\t\\$8,"},
	},
	{
		`
		func f16(b []byte, v uint16) {
			binary.BigEndian.PutUint16(b, v)
		}
		`,
		[]string{"\tROLW\t\\$8,"},
	},
	{
		`
		func f17(b []byte, i int, v uint16) {
			binary.BigEndian.PutUint16(b[i:], v)
		}
		`,
		[]string{"\tROLW\t\\$8,"},
	},
	// Structure zeroing.  See issue #18370.
	{
		`
		type T1 struct {
			a, b, c int
		}
		func f18(t *T1) {
			*t = T1{}
		}
		`,
		[]string{"\tMOVQ\t\\$0, \\(.*\\)", "\tMOVQ\t\\$0, 8\\(.*\\)", "\tMOVQ\t\\$0, 16\\(.*\\)"},
	},
	// TODO: add a test for *t = T{3,4,5} when we fix that.
	// Also test struct containing pointers (this was special because of write barriers).
	{
		`
		type T2 struct {
			a, b, c *int
		}
		func f19(t *T2) {
			*t = T2{}
		}
		`,
		[]string{"\tMOVQ\t\\$0, \\(.*\\)", "\tMOVQ\t\\$0, 8\\(.*\\)", "\tMOVQ\t\\$0, 16\\(.*\\)", "\tCALL\truntime\\.writebarrierptr\\(SB\\)"},
	},
	// Rotate tests
	{
		`
		func f20(x uint64) uint64 {
			return x<<7 | x>>57
		}
		`,
		[]string{"\tROLQ\t[$]7,"},
	},
	{
		`
		func f21(x uint64) uint64 {
			return x<<7 + x>>57
		}
		`,
		[]string{"\tROLQ\t[$]7,"},
	},
	{
		`
		func f22(x uint64) uint64 {
			return x<<7 ^ x>>57
		}
		`,
		[]string{"\tROLQ\t[$]7,"},
	},
	{
		`
		func f23(x uint32) uint32 {
			return x<<7 + x>>25
		}
		`,
		[]string{"\tROLL\t[$]7,"},
	},
	{
		`
		func f24(x uint32) uint32 {
			return x<<7 | x>>25
		}
		`,
		[]string{"\tROLL\t[$]7,"},
	},
	{
		`
		func f25(x uint32) uint32 {
			return x<<7 ^ x>>25
		}
		`,
		[]string{"\tROLL\t[$]7,"},
	},
	{
		`
		func f26(x uint16) uint16 {
			return x<<7 + x>>9
		}
		`,
		[]string{"\tROLW\t[$]7,"},
	},
	{
		`
		func f27(x uint16) uint16 {
			return x<<7 | x>>9
		}
		`,
		[]string{"\tROLW\t[$]7,"},
	},
	{
		`
		func f28(x uint16) uint16 {
			return x<<7 ^ x>>9
		}
		`,
		[]string{"\tROLW\t[$]7,"},
	},
	{
		`
		func f29(x uint8) uint8 {
			return x<<7 + x>>1
		}
		`,
		[]string{"\tROLB\t[$]7,"},
	},
	{
		`
		func f30(x uint8) uint8 {
			return x<<7 | x>>1
		}
		`,
		[]string{"\tROLB\t[$]7,"},
	},
	{
		`
		func f31(x uint8) uint8 {
			return x<<7 ^ x>>1
		}
		`,
		[]string{"\tROLB\t[$]7,"},
	},
	// Rotate after inlining (see issue 18254).
	{
		`
		func g(x uint32, k uint) uint32 {
			return x<<k | x>>(32-k)
		}
		func f32(x uint32) uint32 {
			return g(x, 7)
		}
		`,
		[]string{"\tROLL\t[$]7,"},
	},
	{
		`
		func f33(m map[int]int) int {
			return m[5]
		}
		`,
		[]string{"\tMOVQ\t[$]5,"},
	},
	// Direct use of constants in fast map access calls. Issue 19015.
	{
		`
		func f34(m map[int]int) bool {
			_, ok := m[5]
			return ok
		}
		`,
		[]string{"\tMOVQ\t[$]5,"},
	},
	{
		`
		func f35(m map[string]int) int {
			return m["abc"]
		}
		`,
		[]string{"\"abc\""},
	},
	{
		`
		func f36(m map[string]int) bool {
			_, ok := m["abc"]
			return ok
		}
		`,
		[]string{"\"abc\""},
	},
	// Bit test ops on amd64, issue 18943.
	{
		`
		func f37(a, b uint64) int {
			if a&(1<<(b&63)) != 0 {
				return 1
			}
			return -1
		}
		`,
		[]string{"\tBTQ\t"},
	},
	{
		`
		func f38(a, b uint64) bool {
			return a&(1<<(b&63)) != 0
		}
		`,
		[]string{"\tBTQ\t"},
	},
	{
		`
		func f39(a uint64) int {
			if a&(1<<60) != 0 {
				return 1
			}
			return -1
		}
		`,
		[]string{"\tBTQ\t\\$60"},
	},
	{
		`
		func f40(a uint64) bool {
			return a&(1<<60) != 0
		}
		`,
		[]string{"\tBTQ\t\\$60"},
	},
}

var linux386Tests = []*asmTest{
	{
		`
		func f0(b []byte) uint32 {
			return binary.LittleEndian.Uint32(b)
		}
		`,
		[]string{"\tMOVL\t\\(.*\\),"},
	},
	{
		`
		func f1(b []byte, i int) uint32 {
			return binary.LittleEndian.Uint32(b[i:])
		}
		`,
		[]string{"\tMOVL\t\\(.*\\)\\(.*\\*1\\),"},
	},
}

var linuxS390XTests = []*asmTest{
	{
		`
		func f0(b []byte) uint32 {
			return binary.LittleEndian.Uint32(b)
		}
		`,
		[]string{"\tMOVWBR\t\\(.*\\),"},
	},
	{
		`
		func f1(b []byte, i int) uint32 {
			return binary.LittleEndian.Uint32(b[i:])
		}
		`,
		[]string{"\tMOVWBR\t\\(.*\\)\\(.*\\*1\\),"},
	},
	{
		`
		func f2(b []byte) uint64 {
			return binary.LittleEndian.Uint64(b)
		}
		`,
		[]string{"\tMOVDBR\t\\(.*\\),"},
	},
	{
		`
		func f3(b []byte, i int) uint64 {
			return binary.LittleEndian.Uint64(b[i:])
		}
		`,
		[]string{"\tMOVDBR\t\\(.*\\)\\(.*\\*1\\),"},
	},
	{
		`
		func f4(b []byte) uint32 {
			return binary.BigEndian.Uint32(b)
		}
		`,
		[]string{"\tMOVWZ\t\\(.*\\),"},
	},
	{
		`
		func f5(b []byte, i int) uint32 {
			return binary.BigEndian.Uint32(b[i:])
		}
		`,
		[]string{"\tMOVWZ\t\\(.*\\)\\(.*\\*1\\),"},
	},
	{
		`
		func f6(b []byte) uint64 {
			return binary.BigEndian.Uint64(b)
		}
		`,
		[]string{"\tMOVD\t\\(.*\\),"},
	},
	{
		`
		func f7(b []byte, i int) uint64 {
			return binary.BigEndian.Uint64(b[i:])
		}
		`,
		[]string{"\tMOVD\t\\(.*\\)\\(.*\\*1\\),"},
	},
	{
		`
		func f8(x uint64) uint64 {
			return x<<7 + x>>57
		}
		`,
		[]string{"\tRLLG\t[$]7,"},
	},
	{
		`
		func f9(x uint64) uint64 {
			return x<<7 | x>>57
		}
		`,
		[]string{"\tRLLG\t[$]7,"},
	},
	{
		`
		func f10(x uint64) uint64 {
			return x<<7 ^ x>>57
		}
		`,
		[]string{"\tRLLG\t[$]7,"},
	},
	{
		`
		func f11(x uint32) uint32 {
			return x<<7 + x>>25
		}
		`,
		[]string{"\tRLL\t[$]7,"},
	},
	{
		`
		func f12(x uint32) uint32 {
			return x<<7 | x>>25
		}
		`,
		[]string{"\tRLL\t[$]7,"},
	},
	{
		`
		func f13(x uint32) uint32 {
			return x<<7 ^ x>>25
		}
		`,
		[]string{"\tRLL\t[$]7,"},
	},
	// Fused multiply-add/sub instructions.
	{
		`
		func f14(x, y, z float64) float64 {
			return x * y + z
		}
		`,
		[]string{"\tFMADD\t"},
	},
	{
		`
		func f15(x, y, z float64) float64 {
			return x * y - z
		}
		`,
		[]string{"\tFMSUB\t"},
	},
	{
		`
		func f16(x, y, z float32) float32 {
			return x * y + z
		}
		`,
		[]string{"\tFMADDS\t"},
	},
	{
		`
		func f17(x, y, z float32) float32 {
			return x * y - z
		}
		`,
		[]string{"\tFMSUBS\t"},
	},
}

var linuxARMTests = []*asmTest{
	{
		`
		func f0(x uint32) uint32 {
			return x<<7 + x>>25
		}
		`,
		[]string{"\tMOVW\tR[0-9]+@>25,"},
	},
	{
		`
		func f1(x uint32) uint32 {
			return x<<7 | x>>25
		}
		`,
		[]string{"\tMOVW\tR[0-9]+@>25,"},
	},
	{
		`
		func f2(x uint32) uint32 {
			return x<<7 ^ x>>25
		}
		`,
		[]string{"\tMOVW\tR[0-9]+@>25,"},
	},
}

var linuxARM64Tests = []*asmTest{
	{
		`
		func f0(x uint64) uint64 {
			return x<<7 + x>>57
		}
		`,
		[]string{"\tROR\t[$]57,"},
	},
	{
		`
		func f1(x uint64) uint64 {
			return x<<7 | x>>57
		}
		`,
		[]string{"\tROR\t[$]57,"},
	},
	{
		`
		func f2(x uint64) uint64 {
			return x<<7 ^ x>>57
		}
		`,
		[]string{"\tROR\t[$]57,"},
	},
	{
		`
		func f3(x uint32) uint32 {
			return x<<7 + x>>25
		}
		`,
		[]string{"\tRORW\t[$]25,"},
	},
	{
		`
		func f4(x uint32) uint32 {
			return x<<7 | x>>25
		}
		`,
		[]string{"\tRORW\t[$]25,"},
	},
	{
		`
		func f5(x uint32) uint32 {
			return x<<7 ^ x>>25
		}
		`,
		[]string{"\tRORW\t[$]25,"},
	},
}

// TestLineNumber checks to make sure the generated assembly has line numbers
// see issue #16214
func TestLineNumber(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir, err := ioutil.TempDir("", "TestLineNumber")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "x.go")
	err = ioutil.WriteFile(src, []byte(issue16214src), 0644)
	if err != nil {
		t.Fatalf("could not write file: %v", err)
	}

	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-S", "-o", filepath.Join(dir, "out.o"), src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fail to run go tool compile: %v", err)
	}

	if strings.Contains(string(out), "unknown line number") {
		t.Errorf("line number missing in assembly:\n%s", out)
	}
}

var issue16214src = `
package main

func Mod32(x uint32) uint32 {
	return x % 3 // frontend rewrites it as HMUL with 2863311531, the LITERAL node has unknown Pos
}
`
