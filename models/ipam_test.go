package models

import (
	"net/netip"
	"testing"
)

func mustPrefix(t *testing.T, s string) netip.Prefix {
	t.Helper()
	p, err := netip.ParsePrefix(s)
	if err != nil {
		t.Fatal(err)
	}
	return netip.PrefixFrom(p.Addr(), p.Bits())
}

func TestLastAddr(t *testing.T) {
	cases := map[string]string{
		"10.0.2.0/24":  "10.0.2.255",
		"10.0.2.0/26":  "10.0.2.63",
		"10.0.2.64/26": "10.0.2.127",
		"10.0.0.0/30":  "10.0.0.3",
		"fd00::/126":   "fd00::3",
		"fd00::/64":    "fd00::ffff:ffff:ffff:ffff",
	}
	for in, want := range cases {
		last, err := lastAddr(mustPrefix(t, in))
		if err != nil {
			t.Errorf("%s: %v", in, err)
			continue
		}
		if last.String() != want {
			t.Errorf("%s: last = %s, want %s", in, last, want)
		}
	}
}

func TestAllocateAddress(t *testing.T) {
	pool := mustPrefix(t, "10.0.2.64/26")
	addr, err := AllocateAddress(pool, nil)
	if err != nil || addr.String() != "10.0.2.65" {
		t.Errorf("first alloc = %v, %v; want 10.0.2.65", addr, err)
	}
	used := map[netip.Addr]string{
		netip.MustParseAddr("10.0.2.65"): "a",
		netip.MustParseAddr("10.0.2.66"): "b",
	}
	addr, err = AllocateAddress(pool, used)
	if err != nil || addr.String() != "10.0.2.67" {
		t.Errorf("second alloc = %v, %v; want 10.0.2.67", addr, err)
	}
}

func TestAllocateAddressSkipsBoundaries(t *testing.T) {
	pool := mustPrefix(t, "10.0.0.0/30")
	addr, err := AllocateAddress(pool, nil)
	if err != nil || addr.String() != "10.0.0.1" {
		t.Errorf("/30 first alloc = %v, %v; want 10.0.0.1 (skip network)", addr, err)
	}
	used := map[netip.Addr]string{
		netip.MustParseAddr("10.0.0.1"): "a",
		netip.MustParseAddr("10.0.0.2"): "b",
	}
	if _, err := AllocateAddress(pool, used); err == nil {
		t.Error("/30 with .1 and .2 used should be exhausted (broadcast skipped)")
	}
}

func TestAllocateAddressExhaustion(t *testing.T) {
	pool := mustPrefix(t, "10.0.2.64/30")
	used := map[netip.Addr]string{
		netip.MustParseAddr("10.0.2.65"): "a",
		netip.MustParseAddr("10.0.2.66"): "b",
	}
	if _, err := AllocateAddress(pool, used); err == nil {
		t.Error("expected exhaustion error for /30 with both hosts used")
	}
}

func TestAllocateAddressIPv6(t *testing.T) {
	pool := mustPrefix(t, "fd00::/126")
	addr, err := AllocateAddress(pool, nil)
	if err != nil || addr.String() != "fd00::1" {
		t.Errorf("v6 first alloc = %v, %v; want fd00::1", addr, err)
	}
}

func testCfgWithPeers(addrs ...string) *Configuration {
	cfg := &Configuration{Settings: DefaultSettings()}
	for i, a := range addrs {
		p := NewPeer(string(rune('A'+i)), "priv", "pub")
		p.Address = []string{a}
		p.AllowedIps = []string{a}
		cfg.Peers = append(cfg.Peers, *p)
	}
	return cfg
}

func TestAllocateForPeerClient(t *testing.T) {
	cfg := testCfgWithPeers("10.0.2.2/32", "10.0.2.3/32")
	p := NewPeer("New", "priv", "pub")
	if err := cfg.AllocateForPeer(p); err != nil {
		t.Fatal(err)
	}
	if p.Address[0] != "10.0.2.1/32" {
		t.Errorf("client got %s, want 10.0.2.1/32 (first free in clients range)", p.Address[0])
	}
	if p.AllowedIps[0] != "10.0.2.1/32" {
		t.Errorf("allowedIps = %v", p.AllowedIps)
	}
}

func TestAllocateForPeerServer(t *testing.T) {
	cfg := testCfgWithPeers("10.0.2.2/32")
	ep := "vpn.example.com:51820"
	p := NewPeer("Srv", "priv", "pub")
	p.Endpoint = &ep
	if err := cfg.AllocateForPeer(p); err != nil {
		t.Fatal(err)
	}
	if p.Address[0] != "10.0.2.1/32" {
		t.Errorf("server got %s, want 10.0.2.1/32 (servers subnet)", p.Address[0])
	}
}

func TestAllocateForPeerNoSettings(t *testing.T) {
	cfg := &Configuration{}
	p := NewPeer("X", "priv", "pub")
	if err := cfg.AllocateForPeer(p); err == nil {
		t.Error("expected error when no ipPool configured")
	}
}

func TestUsedAddressesConflict(t *testing.T) {
	cfg := testCfgWithPeers("10.0.2.2/32", "10.0.2.2/32")
	if _, err := cfg.UsedAddresses(); err == nil {
		t.Error("expected duplicate-address error")
	}
}

func TestValidateAddressInPool(t *testing.T) {
	cfg := &Configuration{Settings: DefaultSettings()}
	if err := cfg.ValidateAddressInPool("10.0.2.100/32"); err != nil {
		t.Errorf("in-pool address rejected: %v", err)
	}
	if err := cfg.ValidateAddressInPool("192.168.1.5/32"); err == nil {
		t.Error("out-of-pool address accepted")
	}
	cfgNoPool := &Configuration{}
	if err := cfgNoPool.ValidateAddressInPool("192.168.1.5/32"); err != nil {
		t.Errorf("no pool configured should skip validation: %v", err)
	}
}

func TestPoolStatus(t *testing.T) {
	cfg := testCfgWithPeers("10.0.2.2/32")
	entries, pool, err := cfg.PoolStatus()
	if err != nil {
		t.Fatal(err)
	}
	if pool.String() != "10.0.2.0/24" {
		t.Errorf("pool = %s", pool)
	}
	free := 0
	usedSeen := false
	for _, e := range entries {
		if e.Address.String() == "10.0.2.2" {
			if e.Free || e.Peer != "A" {
				t.Errorf("10.0.2.2 should be used by A: %+v", e)
			}
			usedSeen = true
		}
		if e.Free {
			free++
		}
		if e.Address.String() == "10.0.2.0" || e.Address.String() == "10.0.2.255" {
			t.Errorf("boundary address %s should not be listed", e.Address)
		}
	}
	if !usedSeen {
		t.Error("10.0.2.2 not found in pool status")
	}
	if free != 253 {
		t.Errorf("free addresses = %d, want 253", free)
	}
}
