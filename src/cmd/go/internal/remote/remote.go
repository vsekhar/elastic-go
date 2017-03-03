// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package remote

import (
	"net"
	"os"
	"os/exec"

	"cmd/go/internal/base"
	pb "runtime/remote/api"

	"google.golang.org/grpc"
)

var CmdRemote = &base.Command{
	UsageLine: "remote [binary]",
	Short:     "run a remote binary",
	Long: `
Remote runs a remote binary using the resources of the local machine.

If a binary is specified, remote starts a passthrough remote runtime and
executes the binary with it.

If a binary is not specified, remote attempts to connect to a remote runtime
specified by the GOREMOTE environment variable and waits for a job to be sent.
`,
}

func init() {
	CmdRemote.Run = runRemote
}

func runRemote(cmd *base.Command, args []string) {
	switch len(args) {
	case 0:
		runClient(cmd)
	case 1:
		runPassthrough(cmd, args[0])
	default:
		base.Fatalf("go remote: invalid arguments")
	}
}

func runClient(cmd *base.Command) {
	base.Fatalf("go remote: client not implemented")
}

func runPassthrough(cmd *base.Command, binpath string) {
	srv := grpc.NewServer()
	pb.RegisterRemoteRuntimeServer(srv, newServer())
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		base.Fatalf("go remote: failed to listen - %v", err)
	}
	_, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		base.Fatalf("go remote: bad listen port - %v", err)
	}
	os.Setenv("GOREMOTE", "localhost:"+port)

	ccmd := exec.Command(binpath)
	ccmd.Stdin = os.Stdin
	ccmd.Stdout = os.Stdout
	ccmd.Stderr = os.Stderr
	err = ccmd.Run()
	if err != nil {
		base.Fatalf("go remote: %s", err)
	}
}
