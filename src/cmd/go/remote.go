// Prevent the go remote command from being included when building the
// bootstrap go command.

// +build !cmd_go_bootstrap

package main

import (
	"cmd/go/internal/base"
	"cmd/go/internal/remote"
)

func init() {
	base.Commands = append(base.Commands, remote.CmdRemote)
}
