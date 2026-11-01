//go:build js && wasm

package main

import "testing"

func TestMain(m *testing.M) {
	registerAPI()
	close(apiReady)
	m.Run()
	select {}
}
