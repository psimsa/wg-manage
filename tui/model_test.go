package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ofcoursedude/wg-manage/models"
)

func testModel(t *testing.T) Model {
	t.Helper()
	cfg := models.Configuration{Settings: models.DefaultSettings()}
	server := models.NewPeer("Server", "privS", "pubS")
	ep := "vpn.example.com:51820"
	server.Endpoint = &ep
	server.Address = []string{"10.0.2.1/32"}
	server.AllowedIps = []string{"0.0.0.0/0"}
	cfg.ApplyDefaultsToPeer(server)
	client := models.NewPeer("Laptop", "privC", "pubC")
	client.Address = []string{"10.0.2.2/32"}
	client.AllowedIps = []string{"10.0.2.2/32"}
	cfg.Peers = []models.Peer{*server, *client}
	return NewModel(&cfg, "test.yaml")
}

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func specialKey(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

func send(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		updated, _ := m.Update(msg)
		switch v := updated.(type) {
		case Model:
			m = v
		case *Model:
			m = *v
		}
	}
	return m
}

func typeText(m Model, s string) Model {
	for _, r := range s {
		m = send(m, key(string(r)))
	}
	return m
}

func clearField(m Model) Model {
	n := len(m.form[m.formFocus].value)
	for i := 0; i < n; i++ {
		m = send(m, specialKey(tea.KeyBackspace))
	}
	return m
}

func TestAddPeerFormAutoIP(t *testing.T) {
	m := testModel(t)
	m = send(m, key("a"))
	if m.screen != screenForm {
		t.Fatal("expected form screen after 'a'")
	}
	m = typeText(m, "PhoneX")
	m = send(m, specialKey(tea.KeyTab))   // role
	m = send(m, specialKey(tea.KeyTab))   // endpoint
	m = send(m, specialKey(tea.KeyTab))   // egress
	m = send(m, specialKey(tea.KeyTab))   // address (leave empty = auto)
	m = send(m, specialKey(tea.KeyEnter)) // apply
	if m.screen != screenList {
		t.Fatalf("expected back on list, screen=%d status=%s", m.screen, m.statusMsg)
	}
	if len(m.peers) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(m.peers))
	}
	added := m.peers[2]
	if added.Name != "PhoneX" {
		t.Errorf("name = %q", added.Name)
	}
	if len(added.Address) == 0 || added.Address[0] != "10.0.2.3/32" {
		t.Errorf("auto IP = %v, want 10.0.2.3/32", added.Address)
	}
	if !m.dirty {
		t.Error("model should be dirty after add")
	}
}

func TestAddServerPeerForm(t *testing.T) {
	m := testModel(t)
	m = send(m, key("a"))
	m = typeText(m, "Srv2")
	m = send(m, specialKey(tea.KeyTab))
	m = clearField(m)
	m = typeText(m, "server")
	m = send(m, specialKey(tea.KeyTab)) // endpoint
	m = typeText(m, "vpn2.example.com:443")
	m = send(m, specialKey(tea.KeyTab)) // egress
	m = send(m, specialKey(tea.KeyTab)) // address empty = auto
	m = send(m, specialKey(tea.KeyEnter))
	if len(m.peers) != 3 {
		t.Fatalf("peers = %d", len(m.peers))
	}
	added := m.peers[2]
	if !added.IsServer() {
		t.Error("expected server role")
	}
	if added.Endpoint == nil || *added.Endpoint != "vpn2.example.com:443" {
		t.Errorf("endpoint = %v", added.Endpoint)
	}
	if added.Address[0] != "10.0.2.3/32" {
		t.Errorf("server auto IP = %v, want 10.0.2.3/32 (first free in servers range; .1/.2 taken)", added.Address)
	}
	if added.ListenPort == nil || *added.ListenPort != 443 {
		t.Errorf("listenPort from endpoint = %v, want 443", added.ListenPort)
	}
}

func TestDeletePeerWithConfirm(t *testing.T) {
	m := testModel(t)
	m = send(m, key("d"))
	if m.screen != screenConfirm {
		t.Fatal("expected confirm screen")
	}
	m = send(m, key("y"))
	if len(m.peers) != 1 || m.peers[0].Name != "Laptop" {
		t.Errorf("after delete: %v", m.peers)
	}
}

func TestToggleRoutingUsesEgressInterface(t *testing.T) {
	m := testModel(t)
	m.cursor = 0
	m = send(m, key("r"))
	server := m.peers[0]
	if len(server.PostUp) == 0 {
		t.Fatal("routing not enabled")
	}
	joined := strings.Join(server.PostUp, "\n")
	if !strings.Contains(joined, "iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE") {
		t.Errorf("egress interface eth0 missing from PostUp:\n%s", joined)
	}
	if !strings.Contains(strings.Join(server.PostDown, "\n"), "iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE") {
		t.Errorf("PostDown missing egress rule")
	}
	m = send(m, key("r"))
	if len(m.peers[0].PostUp) != 0 {
		t.Error("routing should be toggled off")
	}

	m.cfg.Settings.ServerDefaults.EgressInterface = "wan0"
	m = send(m, key("r"))
	if !strings.Contains(strings.Join(m.peers[0].PostUp, "\n"), "-o wan0 -j MASQUERADE") {
		t.Errorf("custom egress wan0 missing:\n%s", strings.Join(m.peers[0].PostUp, "\n"))
	}
}

func TestOverridesForm(t *testing.T) {
	m := testModel(t)
	m.cursor = 1
	m = send(m, key("o"))
	if m.screen != screenForm || !m.formOverrides {
		t.Fatalf("expected overrides form, screen=%d overrides=%v", m.screen, m.formOverrides)
	}
	m = typeText(m, "Server")
	m = send(m, specialKey(tea.KeyTab))
	m = typeText(m, "AllowedIPs=192.168.0.0/24; AllowedIPs=10.9.0.0/16")
	m = send(m, specialKey(tea.KeyEnter))
	if m.screen != screenList {
		t.Fatalf("expected list, status=%s", m.statusMsg)
	}
	ov := m.peers[1].PeerOverrides["pubS"]
	if len(ov) != 2 || ov[0] != "AllowedIPs=192.168.0.0/24" {
		t.Errorf("overrides = %v", ov)
	}
}

func TestEditPeerForm(t *testing.T) {
	m := testModel(t)
	m.cursor = 0
	m = send(m, key("e"))
	if m.screen != screenForm {
		t.Fatal("expected edit form")
	}
	m = send(m, specialKey(tea.KeyTab)) // description
	m = send(m, specialKey(tea.KeyTab)) // role (already "server")
	m = send(m, specialKey(tea.KeyTab)) // endpoint
	m = clearField(m)
	m = typeText(m, "vpn.example.com:9999")
	for m.formFocus < len(m.form)-1 {
		m = send(m, specialKey(tea.KeyTab))
	}
	m = send(m, specialKey(tea.KeyEnter))
	if m.peers[0].Endpoint == nil || *m.peers[0].Endpoint != "vpn.example.com:9999" {
		t.Errorf("endpoint not updated: %v", m.peers[0].Endpoint)
	}
	if !m.peers[0].IsServer() {
		t.Error("peer should still be a server")
	}
}

func TestSaveWritesYaml(t *testing.T) {
	m := testModel(t)
	dir := t.TempDir()
	m.configFile = dir + "/out.yaml"
	m = send(m, key("a"))
	m = typeText(m, "Saved")
	for i := 0; i < 4; i++ {
		m = send(m, specialKey(tea.KeyTab))
	}
	m = send(m, specialKey(tea.KeyEnter))
	m = send(m, key("s"))
	cfg := models.LoadYaml(m.configFile)
	if len(cfg.Peers) != 3 {
		t.Fatalf("saved peers = %d", len(cfg.Peers))
	}
	if cfg.Peers[2].Name != "Saved" {
		t.Errorf("saved peer name = %q", cfg.Peers[2].Name)
	}
	if cfg.Settings == nil || cfg.Settings.IPPool == "" {
		t.Error("settings should be saved too")
	}
	if m.dirty {
		t.Error("dirty should be false after save")
	}
}

func TestGenerateFromTui(t *testing.T) {
	m := testModel(t)
	dir := t.TempDir()
	m.cfg.Settings.Generate.OutputDir = dir + "/out"
	png := false
	m.cfg.Settings.Generate.PNG = &png
	m = send(m, key("g"))
	if !strings.Contains(m.statusMsg, "Generated") {
		t.Fatalf("status = %q", m.statusMsg)
	}
	for _, name := range []string{"Server.conf", "Laptop.conf"} {
		if _, err := os.Stat(dir + "/out/" + name); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
}

func TestQuitWithUnsavedChangesAsks(t *testing.T) {
	m := testModel(t)
	m.dirty = true
	updated, cmd := m.Update(key("q"))
	m = updated.(Model)
	if m.screen != screenConfirm {
		t.Fatalf("expected confirm on quit with dirty state, screen=%d", m.screen)
	}
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestSettingsTabAddsDefaultSettings(t *testing.T) {
	cfg := models.Configuration{}
	server := models.NewPeer("Server", "privS", "pubS")
	ep := "vpn.example.com:51820"
	server.Endpoint = &ep
	cfg.Peers = []models.Peer{*server}
	m := NewModel(&cfg, "test.yaml")

	m = send(m, specialKey(tea.KeyTab))
	if m.tab != tabSettings {
		t.Fatal("expected settings tab")
	}
	if !strings.Contains(m.View(), "No settings section") {
		t.Fatalf("expected 'no settings' hint:\n%s", m.View())
	}
	m = send(m, key("a"))
	if m.cfg.Settings == nil {
		t.Fatal("settings should be created after 'a'")
	}
	if m.cfg.Settings.IPPool != models.DefaultIPPool {
		t.Errorf("ipPool = %q", m.cfg.Settings.IPPool)
	}
	if !m.dirty {
		t.Error("model should be dirty after adding settings")
	}
	if !strings.Contains(m.View(), "ipPool: "+models.DefaultIPPool) {
		t.Errorf("settings view should show ipPool:\n%s", m.View())
	}

	m2 := testModel(t)
	m2 = send(m2, specialKey(tea.KeyTab))
	m2 = send(m2, key("a"))
	if m2.cfg.Settings.IPPool != models.DefaultIPPool {
		t.Error("existing settings should be kept, not overwritten")
	}
	if !strings.Contains(m2.statusMsg, "already present") {
		t.Errorf("status = %q", m2.statusMsg)
	}
}

func TestTabSwitchAndView(t *testing.T) {
	m := testModel(t)
	m = send(m, specialKey(tea.KeyTab))
	if m.tab != tabSettings {
		t.Error("tab should switch to settings")
	}
	v := m.View()
	if !strings.Contains(v, "ipPool: 10.0.2.0/24") {
		t.Errorf("settings view missing ipPool:\n%s", v)
	}
	m = send(m, specialKey(tea.KeyTab))
	if m.tab != tabPeers {
		t.Error("tab should switch back to peers")
	}
	v = m.View()
	if !strings.Contains(v, "Server") || !strings.Contains(v, "server") {
		t.Errorf("peers view missing entries:\n%s", v)
	}
}
