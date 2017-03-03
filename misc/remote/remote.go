// Tests -buildmode=remote

package main

import (
	"os"
)

var failed = false

func main() {
	if os.Getenv("GOREMOTE") == "" {
		panic("no GOREMOTE specified")
	}
}
