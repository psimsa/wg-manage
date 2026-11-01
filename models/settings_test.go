package models

import (
	"testing"
)

func TestDefaultSettingsValid(t *testing.T) {
	s := DefaultSettings()
	if err := s.Validate(); err != nil {
		t.Fatalf("default settings should be valid: %v", err)
	}
	if s.IPPool != DefaultIPPool {
		t.Errorf("ipPool = %q, want %q", s.IPPool, DefaultIPPool)
	}
	if s.ServerDefaults == nil || s.ServerDefaults.ListenPort == nil || *s.ServerDefaults.ListenPort != DefaultListenPort {
		t.Errorf("serverDefaults.listenPort not set to default: %+v", s.ServerDefaults)
	}
	if s.Generate == nil || s.Generate.OutputDir != DefaultOutputDir || s.Generate.PNG == nil || !*s.Generate.PNG {
		t.Errorf("generate defaults wrong: %+v", s.Generate)
	}
}

func TestSettingsValidate(t *testing.T) {
	cases := []struct {
		name     string
		settings *Settings
		wantErr  bool
	}{
		{"nil settings", nil, false},
		{"empty settings", &Settings{}, false},
		{"valid pool only", &Settings{IPPool: "10.0.0.0/24"}, false},
		{"invalid pool", &Settings{IPPool: "10.0.0.0/33"}, true},
		{"pool not cidr", &Settings{IPPool: "not-a-cidr"}, true},
		{"subnet outside pool", &Settings{IPPool: "10.0.2.0/24", Subnets: map[string]Subnet{"servers": {CIDR: "10.0.3.0/26"}}}, true},
		{"subnet larger than pool", &Settings{IPPool: "10.0.2.0/26", Subnets: map[string]Subnet{"clients": {CIDR: "10.0.2.0/24"}}}, true},
		{"subnet without pool", &Settings{Subnets: map[string]Subnet{"servers": {CIDR: "10.0.2.0/26"}}}, true},
		{"valid subnets", &Settings{IPPool: "10.0.2.0/24", Subnets: map[string]Subnet{"servers": {CIDR: "10.0.2.0/26"}, "clients": {CIDR: "10.0.2.64/26"}}}, false},
		{"bad subnet cidr", &Settings{IPPool: "10.0.2.0/24", Subnets: map[string]Subnet{"x": {CIDR: "nope"}}}, true},
		{"ipv6 pool", &Settings{IPPool: "fd00::/64", Subnets: map[string]Subnet{"clients": {CIDR: "fd00::/80"}}}, false},
		{"v4/v6 mismatch", &Settings{IPPool: "10.0.2.0/24", Subnets: map[string]Subnet{"x": {CIDR: "fd00::/80"}}}, true},
		{"bad listen port", &Settings{IPPool: "10.0.2.0/24", ServerDefaults: &PeerDefaults{ListenPort: intPtr(70000)}}, true},
		{"bad keepalive", &Settings{ClientDefaults: &PeerDefaults{PersistentKeepalive: intPtr(-1)}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.settings.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func intPtr(i int) *int { return &i }

func TestIsServer(t *testing.T) {
	p := &Peer{}
	if p.IsServer() {
		t.Error("peer without endpoint should not be a server")
	}
	empty := ""
	p.Endpoint = &empty
	if p.IsServer() {
		t.Error("peer with empty endpoint should not be a server")
	}
	ep := "vpn.example.com:51820"
	p.Endpoint = &ep
	if !p.IsServer() {
		t.Error("peer with endpoint should be a server")
	}
}

func TestApplyDefaultsToPeer(t *testing.T) {
	s := DefaultSettings()
	ka := 25
	s.ClientDefaults.PersistentKeepalive = &ka

	cfg := &Configuration{Settings: s}

	client := &Peer{Name: "c"}
	cfg.ApplyDefaultsToPeer(client)
	if client.PersistentKeepalive == nil || *client.PersistentKeepalive != 25 {
		t.Errorf("client keepalive not applied: %v", client.PersistentKeepalive)
	}
	if client.ListenPort != nil {
		t.Errorf("client should not get listen port, got %v", *client.ListenPort)
	}

	ep := "vpn.example.com:51820"
	server := &Peer{Name: "s", Endpoint: &ep}
	cfg.ApplyDefaultsToPeer(server)
	if server.ListenPort == nil || *server.ListenPort != DefaultListenPort {
		t.Errorf("server listen port not applied: %v", server.ListenPort)
	}

	existing := 443
	server2 := &Peer{Name: "s2", Endpoint: &ep, ListenPort: &existing}
	cfg.ApplyDefaultsToPeer(server2)
	if *server2.ListenPort != 443 {
		t.Errorf("existing listen port must not be overwritten, got %d", *server2.ListenPort)
	}

	cfgNoSettings := &Configuration{}
	p := &Peer{Name: "x"}
	cfgNoSettings.ApplyDefaultsToPeer(p)
	if p.ListenPort != nil || p.PersistentKeepalive != nil {
		t.Error("nil settings must be a no-op")
	}
}

func TestResolveOutputDirAndPNG(t *testing.T) {
	cfg := &Configuration{}
	if got := cfg.ResolveOutputDir(false, "./output"); got != DefaultOutputDir {
		t.Errorf("no settings: got %q", got)
	}
	if !cfg.ResolvePNG(false, true) {
		t.Error("no settings: png should default to true")
	}

	pngFalse := false
	s := &Settings{Generate: &GenerateSettings{OutputDir: "./gen-out", PNG: &pngFalse}}
	cfg.Settings = s
	if got := cfg.ResolveOutputDir(false, "./output"); got != "./gen-out" {
		t.Errorf("settings output dir not used: %q", got)
	}
	if cfg.ResolvePNG(false, true) {
		t.Error("settings png=false not used")
	}
	if got := cfg.ResolveOutputDir(true, "./flag-out"); got != "./flag-out" {
		t.Errorf("flag must win over settings: %q", got)
	}
	if !cfg.ResolvePNG(true, true) {
		t.Error("flag png=true must win over settings")
	}
}

func TestSettingsYamlRoundTrip(t *testing.T) {
	cfg := Configuration{Settings: DefaultSettings()}
	data := GetYaml(cfg)
	cfg2 := Configuration{}
	if err := unmarshalYaml(data, &cfg2); err != nil {
		t.Fatal(err)
	}
	if cfg2.Settings == nil || cfg2.Settings.IPPool != DefaultIPPool {
		t.Fatalf("settings lost in round trip: %s", data)
	}
	if cfg2.Settings.Subnets["servers"].CIDR != DefaultServersCIDR {
		t.Errorf("subnets lost: %s", data)
	}
	if cfg2.Settings.ServerDefaults.ListenPort == nil || *cfg2.Settings.ServerDefaults.ListenPort != DefaultListenPort {
		t.Errorf("serverDefaults lost: %s", data)
	}
	if err := cfg2.Validate(); err != nil {
		t.Errorf("round-tripped settings invalid: %v", err)
	}
}
