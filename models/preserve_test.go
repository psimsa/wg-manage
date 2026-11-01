package models

import (
	"strings"
	"testing"
)

const customKeysYaml = `projectmeta: my homelab
settings:
  ipPool: 10.0.2.0/24
  myCustomSetting: hello
presharedKey: PSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSKPSK=
peers:
- name: Server
  location: Berlin
  owner: alice
  tags:
  - infra
  - vpn
  interfaceSection:
    privateKey: PRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRI=
    listenPort: 51820
    customIfKey: ifval
  interfaceSectionWgQuick:
    address:
    - 10.0.2.1/32
    customWgQuickKey: wqval
  peerSection:
    publicKey: PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU=
    allowedIps:
    - 0.0.0.0/0
    endpoint: vpn.example.com:51820
    customPeerKey: pkval
- name: My Laptop
  description: work laptop
  interfaceSection:
    privateKey: PRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRI2=
  interfaceSectionWgQuick:
    address:
    - 10.0.2.2/32
  peerSection:
    publicKey: PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU2=
    allowedIps:
    - 10.0.2.2/32
`

func TestCustomKeysPreservedOnRoundTrip(t *testing.T) {
	cfg := Configuration{}
	if err := unmarshalYaml([]byte(customKeysYaml), &cfg); err != nil {
		t.Fatal(err)
	}
	out := string(GetYaml(cfg))
	for _, want := range []string{
		"projectmeta: my homelab",
		"myCustomSetting: hello",
		"location: Berlin",
		"owner: alice",
		"- infra",
		"- vpn",
		"customIfKey: ifval",
		"customWgQuickKey: wqval",
		"customPeerKey: pkval",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("custom key %q lost in round trip:\n%s", want, out)
		}
	}

	cfg2 := Configuration{}
	if err := unmarshalYaml([]byte(out), &cfg2); err != nil {
		t.Fatal(err)
	}
	out2 := string(GetYaml(cfg2))
	if out != out2 {
		t.Errorf("second round trip not stable:\n--- first ---\n%s\n--- second ---\n%s", out, out2)
	}
}

func TestCustomKeysSurviveMutation(t *testing.T) {
	cfg := Configuration{}
	if err := unmarshalYaml([]byte(customKeysYaml), &cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Peers = RemovePeerByName(cfg.Peers, "My Laptop")
	cfg.Peers = append(cfg.Peers, *NewPeer("New", "priv", "pub"))
	out := string(GetYaml(cfg))
	if !strings.Contains(out, "location: Berlin") || !strings.Contains(out, "projectmeta: my homelab") {
		t.Errorf("custom keys lost after remove+add:\n%s", out)
	}
	if strings.Contains(out, "work laptop") {
		t.Errorf("removed peer's keys should be gone:\n%s", out)
	}
}

func TestKnownKeysNotDuplicatedIntoExtra(t *testing.T) {
	cfg := Configuration{}
	if err := unmarshalYaml([]byte(customKeysYaml), &cfg); err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.Extra["peers"]; ok {
		t.Error("'peers' must not land in Extra")
	}
	if _, ok := cfg.Extra["projectmeta"]; !ok {
		t.Error("'projectmeta' should land in Extra")
	}
	if len(cfg.Peers) != 2 || cfg.Peers[0].Name != "Server" {
		t.Errorf("known fields broken: %+v", cfg.Peers)
	}
	out := string(GetYaml(cfg))
	if strings.Count(out, "name: Server") != 1 {
		t.Errorf("'name' duplicated:\n%s", out)
	}
}
