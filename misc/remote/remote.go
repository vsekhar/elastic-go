// Tests -buildmode=remote

package main

var failed = false

func main() {
	if failed {
		panic("failed")
	}
}
