//go:build js && wasm

package main

import (
	"strings"
	"syscall/js"
	"testing"
	"time"
)

func waitAPI(t *testing.T) {
	t.Helper()
	select {
	case <-apiReady:
	case <-time.After(10 * time.Second):
		t.Fatal("wasm api did not become ready")
	}
}

var apiReady = make(chan struct{})

func init() {
	js.Global().Set("goExit", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		close(apiReady)
		return nil
	}))
}

func call(name string, args ...interface{}) map[string]interface{} {
	fn := js.Global().Get("wgManage").Get(name)
	res := fn.Invoke(args...)
	out := map[string]interface{}{}
	out["ok"] = res.Get("ok").Bool()
	if res.Get("error").Truthy() {
		out["error"] = res.Get("error").String()
	}
	if res.Get("result").Truthy() {
		out["result"] = res.Get("result")
	}
	return out
}

func TestBootstrapAndGenerate(t *testing.T) {
	waitAPI(t)
	b := call("bootstrap", "vpn.example.com:51820")
	if !b["ok"].(bool) {
		t.Fatalf("bootstrap failed: %v", b["error"])
	}
	yaml := b["result"].(js.Value).String()
	if !strings.Contains(yaml, "settings:") || !strings.Contains(yaml, "10.0.2.1/32") {
		t.Errorf("bootstrap yaml unexpected:\n%s", yaml)
	}

	peers := call("peerNames", yaml)
	if !peers["ok"].(bool) {
		t.Fatalf("peerNames failed: %v", peers["error"])
	}
	arr := peers["result"].(js.Value)
	if arr.Length() != 3 {
		t.Errorf("expected 3 peers, got %d", arr.Length())
	}

	conf := call("generateConf", yaml, "Server")
	if !conf["ok"].(bool) {
		t.Fatalf("generateConf failed: %v", conf["error"])
	}
	confStr := conf["result"].(js.Value).String()
	for _, want := range []string{"[Interface]", "PrivateKey=", "ListenPort=51820", "[Peer]", "AllowedIPs=10.0.2.2/32"} {
		if !strings.Contains(confStr, want) {
			t.Errorf("server conf missing %q:\n%s", want, confStr)
		}
	}

	laptop := call("generateConf", yaml, "My Laptop")
	laptopStr := laptop["result"].(js.Value).String()
	if !strings.Contains(laptopStr, "Endpoint=vpn.example.com:51820") || !strings.Contains(laptopStr, "AllowedIPs=0.0.0.0/0") {
		t.Errorf("laptop conf missing endpoint/killswitch route:\n%s", laptopStr)
	}
}

func TestAddRemovePeer(t *testing.T) {
	waitAPI(t)
	b := call("bootstrap", "vpn.example.com:51820")
	yaml := b["result"].(js.Value).String()

	added := call("addPeer", yaml, "Tablet", "client")
	if !added["ok"].(bool) {
		t.Fatalf("addPeer failed: %v", added["error"])
	}
	yaml2 := added["result"].(js.Value).String()
	if !strings.Contains(yaml2, "- 10.0.2.4/32") || !strings.Contains(yaml2, "name: Tablet") {
		t.Errorf("addPeer did not auto-allocate:\n%s", yaml2)
	}

	removed := call("removePeer", yaml2, "Tablet")
	if !removed["ok"].(bool) {
		t.Fatalf("removePeer failed: %v", removed["error"])
	}
	if strings.Contains(removed["result"].(js.Value).String(), "Tablet") {
		t.Error("Tablet should be gone")
	}

	bad := call("removePeer", yaml2, "Nope")
	if bad["ok"].(bool) {
		t.Error("removing unknown peer should fail")
	}
}

func TestVizAndParse(t *testing.T) {
	waitAPI(t)
	b := call("bootstrap", "vpn.example.com:51820")
	yaml := b["result"].(js.Value).String()

	v := call("viz", yaml)
	if !v["ok"].(bool) {
		t.Fatalf("viz failed: %v", v["error"])
	}
	if !strings.Contains(v["result"].(js.Value).String(), "[SERVER] Server") {
		t.Errorf("viz output unexpected:\n%s", v["result"].(js.Value).String())
	}

	bad := call("parseYaml", "not: [valid")
	if bad["ok"].(bool) {
		t.Error("invalid yaml should fail")
	}
}
