package models

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadTestConfig(t *testing.T) Configuration {
	t.Helper()
	return LoadYaml(filepath.Join("testdata", "config.yaml"))
}

func TestLoadYaml(t *testing.T) {
	cfg := loadTestConfig(t)
	if cfg.PresharedKey == nil || *cfg.PresharedKey != "PSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSK=" {
		t.Errorf("unexpected preshared key: %v", cfg.PresharedKey)
	}
	if len(cfg.Peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(cfg.Peers))
	}
	server := cfg.Peers[0]
	if server.Name != "Server" || server.ListenPort == nil || *server.ListenPort != 51820 {
		t.Errorf("unexpected server peer: %+v", server)
	}
	if server.Endpoint == nil || *server.Endpoint != "vpn.example.com:51820" {
		t.Errorf("unexpected endpoint: %v", server.Endpoint)
	}
	laptop := cfg.Peers[1]
	if laptop.Description == nil || *laptop.Description != "work laptop" {
		t.Errorf("unexpected description: %v", laptop.Description)
	}
	if laptop.PersistentKeepalive == nil || *laptop.PersistentKeepalive != 25 {
		t.Errorf("unexpected keepalive: %v", laptop.PersistentKeepalive)
	}
	overrides := laptop.PeerOverrides["PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU="]
	if len(overrides) != 1 || overrides[0] != "AllowedIPs=192.168.0.0/24" {
		t.Errorf("unexpected overrides: %v", overrides)
	}
}

func TestGetYamlRoundTrip(t *testing.T) {
	cfg := loadTestConfig(t)
	data := GetYaml(cfg)
	cfg2 := Configuration{}
	if err := unmarshalYaml(data, &cfg2); err != nil {
		t.Fatalf("re-parse failed: %v", err)
	}
	if len(cfg2.Peers) != len(cfg.Peers) {
		t.Fatalf("peer count changed after round trip: %d -> %d", len(cfg.Peers), len(cfg2.Peers))
	}
	for i := range cfg.Peers {
		if cfg.Peers[i].Name != cfg2.Peers[i].Name || cfg.Peers[i].PublicKey != cfg2.Peers[i].PublicKey {
			t.Errorf("peer %d changed: %+v -> %+v", i, cfg.Peers[i], cfg2.Peers[i])
		}
	}
}

func TestGetYamlSortsPeersWithAddressFirst(t *testing.T) {
	cfg := loadTestConfig(t)
	data := string(GetYaml(cfg))
	serverIdx := strings.Index(data, "name: Server")
	laptopIdx := strings.Index(data, "name: My Laptop")
	if serverIdx == -1 || laptopIdx == -1 || serverIdx > laptopIdx {
		t.Errorf("peers with addresses should sort before those without; got order Server@%d Laptop@%d", serverIdx, laptopIdx)
	}
}

func TestGetInterfaceGolden(t *testing.T) {
	cfg := loadTestConfig(t)
	var buf bytes.Buffer
	cfg.Peers[0].GetInterface(&buf)
	assertGolden(t, "golden/interface_server.conf", buf.String())
}

func TestGetPeerGolden(t *testing.T) {
	cfg := loadTestConfig(t)
	server := cfg.Peers[0]
	laptop := cfg.Peers[1]
	var buf bytes.Buffer
	server.GetPeer(&buf, &cfg, &laptop)
	assertGolden(t, "golden/peer_server_in_laptop.conf", buf.String())
}

func TestRemovePeerByName(t *testing.T) {
	cfg := loadTestConfig(t)
	peers := RemovePeerByName(cfg.Peers, "My Phone")
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers after removal, got %d", len(peers))
	}
	for _, p := range peers {
		if p.Name == "My Phone" {
			t.Error("My Phone still present")
		}
	}
	if got := RemovePeerByName(cfg.Peers, "no such peer"); len(got) != 3 {
		t.Error("removing unknown name should be a no-op")
	}
}

func TestRemovePeerByPublicKey(t *testing.T) {
	cfg := loadTestConfig(t)
	peers := RemovePeerByPublicKey(cfg.Peers, "PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU2=")
	if len(peers) != 2 {
		t.Fatalf("expected 2 peers after removal, got %d", len(peers))
	}
	for _, p := range peers {
		if p.PublicKey == "PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU2=" {
			t.Error("peer with that key still present")
		}
	}
}

func TestSaveAndLoadYaml(t *testing.T) {
	cfg := loadTestConfig(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "out.yaml")
	SaveYaml(cfg, file)
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	cfg2 := LoadYaml(file)
	if len(cfg2.Peers) != 3 {
		t.Errorf("expected 3 peers after save/load, got %d", len(cfg2.Peers))
	}
}

func TestSamplePeerHasAllFields(t *testing.T) {
	p := SamplePeer()
	if p.Name == "" || p.PrivateKey == "" || p.PublicKey == "" {
		t.Error("sample peer missing core fields")
	}
	if len(p.PeerOverrides) != 1 {
		t.Errorf("expected 1 override entry, got %d", len(p.PeerOverrides))
	}
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if update := os.Getenv("UPDATE_GOLDEN"); update != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s missing (run with UPDATE_GOLDEN=1): %v", name, err)
	}
	if string(want) != got {
		t.Errorf("output does not match golden %s\n--- want ---\n%s\n--- got ---\n%s", name, want, got)
	}
}
