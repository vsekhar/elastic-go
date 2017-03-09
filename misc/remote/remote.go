// Tests -buildmode=remote

package main

import (
	"os"
)

var remoteVar = 42
var localVar = 43

func main() {
	if os.Getenv("GOREMOTE") == "" {
		panic("no GOREMOTE specified")
	}

	ch := make(chan struct{})
	go func() {
		remoteVar = remoteVar + 1
		ch <- struct{}{}
	}()
	<-ch
	localVar = remoteVar + 3
	// remoteVar == 43, remote allocation
	// localVar == 46, local allocation

	// TODO: check allocation status
}
