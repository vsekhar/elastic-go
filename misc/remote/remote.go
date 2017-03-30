// Tests -buildmode=remote

package main

import (
	"fmt"
)

var var1 int
var var2 int
var var3 int

func init() {
	var1 = 0
}

func main() {
	var1 = 1
	var2 = 1
	g(0)
	done := make(chan struct{})

	// Some gymnastics to exercise the static analysis
	type functionHolder struct {
		fs []func(int, chan struct{})
	}
	fh := new(functionHolder)
	fh.fs = append(fh.fs,
		func(int, chan struct{}) {},
		h,
		func(int, chan struct{}) {},
	)
	getFunc := func(i int) func(int, chan struct{}) {
		return fh.fs[i]
	}
	go getFunc(var1)(0, done)
	<-done

	fmt.Printf("var1: %d\n", var1)
	fmt.Printf("var2: %d\n", var2)
	fmt.Printf("var3: %d\n", var3)
}

func h(i int, done chan struct{}) {
	var2 += 1
	if i == 0 {
		j()
		done <- struct{}{}
	}
}

func j() {
	var3 += 1
	g(1)
	k()
}

func k() {
	h(1, nil) // stop loop
}

func g(i int) {
	var3 += 1
	if i == 0 {
		j()
	}
}
