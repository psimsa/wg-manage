package viz

import (
	"strings"
	"testing"

	"github.com/ofcoursedude/wg-manage/models"
)

func testCfg() *models.Configuration {
	cfg := &models.Configuration{Settings: models.DefaultSettings()}
	server := models.NewPeer("Server", "privS", "PUBSERVERKEY0000000000000000000000000000=")
	ep := "vpn.example.com:51820"
	server.Endpoint = &ep
	server.Address = []string{"10.0.2.1/32"}
	server.AllowedIps = []string{"0.0.0.0/0"}
	cfg.ApplyDefaultsToPeer(server)
	laptop := models.NewPeer("Laptop", "privL", "PUBLAPTOPKEY00000000000000000000000000=")
	laptop.Address = []string{"10.0.2.2/32"}
	laptop.AllowedIps = []string{"10.0.2.2/32"}
	laptop.PeerOverrides = map[string][]string{
		"PUBSERVERKEY0000000000000000000000000000=": {"AllowedIPs=192.168.0.0/24"},
	}
	cfg.Peers = []models.Peer{*server, *laptop}
	return cfg
}

func TestRenderASCII(t *testing.T) {
	out, err := Render(testCfg(), "ascii")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"WireGuard network",
		"pool 10.0.2.0/24",
		"[SERVER] Server  10.0.2.1/32  [PUBSERVE]  vpn.example.com:51820",
		"[client] Laptop  10.0.2.2/32  [PUBLAPTO]",
		"(+1 override lines)",
		"free addresses:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("ascii output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderDOT(t *testing.T) {
	out, err := Render(testCfg(), "dot")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"digraph wg {",
		`"Server" [label=`,
		"fillcolor=lightblue",
		`"Laptop" -> "Server";`,
		"}",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dot output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderMermaid(t *testing.T) {
	out, err := Render(testCfg(), "mermaid")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"graph TD", "Server[\"", "Laptop --> Server", "classDef server"} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderUnknownFormat(t *testing.T) {
	if _, err := Render(testCfg(), "yaml"); err == nil {
		t.Error("expected error for unknown format")
	}
}

func TestRenderNoServers(t *testing.T) {
	cfg := &models.Configuration{}
	c := models.NewPeer("Only", "p", "PUBONLY00000000000000000000000000000000=")
	cfg.Peers = []models.Peer{*c}
	out, err := Render(cfg, "ascii")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "no servers") || !strings.Contains(out, "[client] Only") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestSanitizeID(t *testing.T) {
	if got := sanitizeID("My Laptop 2"); got != "My_Laptop_2" {
		t.Errorf("sanitizeID = %q", got)
	}
}
