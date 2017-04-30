// Package remote is added by the compiler when compiling a program with
// -buildmode=remote. It is not intended to be imported or used directly.

// +build !cmd_go_bootstrap

package remote

// TODO: ensure this package is added by compiler, cannot import from package
// runtime due to cyclic dependency

import (
	"os"

	"google.golang.org/grpc"

	pb "internal/remote/api"
)

var conn *grpc.ClientConn
var client pb.RemoteRuntimeClient

func init() {
	println("internal/remote added")
}

func initRemote() {
	// connect to remote runtime, panic if failed
	r := os.Getenv("GOREMOTE")
	if r == "" {
		panic("remote binary requires remote runtime specified in environment variable GOREMOTE")
	}
	conn, err := grpc.Dial(r, grpc.WithInsecure())
	if err != nil {
		panic("failed to connect to remote runtime: " + err.Error())
	}
	client = pb.NewRemoteRuntimeClient(conn)
}
