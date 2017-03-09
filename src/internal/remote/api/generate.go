// Package api specifies the runtime API.
//
// Programs compiled with --buildmode=remote do not ordinarily need to import
// or use this package directly. The compiler will handle it.
//
// Programs providing an implementation of the remote runtime may import this
// package in order to implement a RemoteRuntimeServer.
package api

//go:generate protoc remoteapi.proto --go_out=plugins=grpc:.
