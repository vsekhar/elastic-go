// Package lib is used to test static analysis across library boundaries.
package lib

var RemoteVar int

func init() {
	RemoteVar = 42
}
