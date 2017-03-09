// This file is used by http_test.go to test if the http transport code
// accidentally depends on server code.

package main

import (
	"net/http"
)

func init() {
	client := new(http.Client)
	_ = client.Get
	transport := new(http.Transport)
	_ = transport.RoundTrip
}

func main() {

}
