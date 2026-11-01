package models

import (
	"fmt"
	"net/netip"
	"sort"
)

func parsePrefix(s string) (netip.Prefix, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("%q is not a valid CIDR: %w", s, err)
	}
	return netip.PrefixFrom(p.Addr(), p.Bits()), nil
}

func (cfg *Configuration) PoolPrefix() (netip.Prefix, error) {
	if cfg.Settings == nil || cfg.Settings.IPPool == "" {
		return netip.Prefix{}, fmt.Errorf("no settings.ipPool configured")
	}
	return parsePrefix(cfg.Settings.IPPool)
}

func (cfg *Configuration) SubnetPrefix(role string) (netip.Prefix, bool, error) {
	if cfg.Settings == nil {
		return netip.Prefix{}, false, nil
	}
	sub, ok := cfg.Settings.Subnets[role]
	if !ok || sub.CIDR == "" {
		return netip.Prefix{}, false, nil
	}
	p, err := parsePrefix(sub.CIDR)
	return p, true, err
}

func PeerAddresses(p *Peer) []netip.Prefix {
	var out []netip.Prefix
	for _, a := range p.Address {
		if prefix, err := parsePrefix(a); err == nil {
			out = append(out, prefix)
		}
	}
	return out
}

func (cfg *Configuration) UsedAddresses() (map[netip.Addr]string, error) {
	used := make(map[netip.Addr]string)
	for i := range cfg.Peers {
		for _, prefix := range PeerAddresses(&cfg.Peers[i]) {
			addr := prefix.Addr()
			if _, dup := used[addr]; dup {
				return nil, fmt.Errorf("address %s is used by both %q and %q", addr, used[addr], cfg.Peers[i].Name)
			}
			used[addr] = cfg.Peers[i].Name
		}
	}
	return used, nil
}

func AllocateAddress(rangePrefix netip.Prefix, used map[netip.Addr]string) (netip.Addr, error) {
	if !rangePrefix.IsValid() {
		return netip.Addr{}, fmt.Errorf("invalid address range")
	}
	first := rangePrefix.Addr()
	last, err := lastAddr(rangePrefix)
	if err != nil {
		return netip.Addr{}, err
	}
	addr := first.Next()
	for addr.IsValid() && compareAddr(addr, last) <= 0 {
		if rangePrefix.Addr().Is4() && isBoundaryAddr(addr, rangePrefix) {
			addr = addr.Next()
			continue
		}
		if _, taken := used[addr]; !taken {
			return addr, nil
		}
		addr = addr.Next()
	}
	return netip.Addr{}, fmt.Errorf("address range %s is exhausted", rangePrefix)
}

func isBoundaryAddr(addr netip.Addr, p netip.Prefix) bool {
	if !addr.Is4() {
		return false
	}
	last, err := lastAddr(p)
	if err != nil {
		return false
	}
	return addr == p.Addr() || addr == last
}

func lastAddr(p netip.Prefix) (netip.Addr, error) {
	bits := 32
	if p.Addr().Is6() {
		bits = 128
	}
	hostBits := bits - p.Bits()
	if hostBits >= bits {
		return netip.Addr{}, fmt.Errorf("prefix %s too large", p)
	}
	if hostBits > 64 {
		return netip.Addr{}, fmt.Errorf("prefix %s too large to enumerate", p)
	}
	var mask uint64 = ^uint64(0) >> (64 - hostBits)
	base := p.Addr()
	if base.Is4() {
		b := base.As4()
		v := uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
		v |= mask
		return netip.AddrFrom4([4]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}), nil
	}
	b16 := base.As16()
	var lo uint64
	for i := 8; i < 16; i++ {
		lo = lo<<8 | uint64(b16[i])
	}
	lo |= mask
	var out [16]byte
	copy(out[:8], b16[:8])
	for i := 0; i < 8; i++ {
		out[8+i] = byte(lo >> (56 - 8*i))
	}
	return netip.AddrFrom16(out), nil
}

func compareAddr(a, b netip.Addr) int {
	ab, bb := a.As16(), b.As16()
	for i := 0; i < 16; i++ {
		if ab[i] != bb[i] {
			if ab[i] < bb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func hostPrefixBits(addr netip.Addr) int {
	if addr.Is4() {
		return 32
	}
	return 128
}

func (cfg *Configuration) AllocateForPeer(p *Peer) error {
	pool, err := cfg.PoolPrefix()
	if err != nil {
		return err
	}
	role := "clients"
	if p.IsServer() {
		role = "servers"
	}
	rangePrefix := pool
	if sub, ok, err := cfg.SubnetPrefix(role); err != nil {
		return err
	} else if ok {
		rangePrefix = sub
	}
	used, err := cfg.UsedAddresses()
	if err != nil {
		return err
	}
	addr, err := AllocateAddress(rangePrefix, used)
	if err != nil {
		return fmt.Errorf("cannot allocate address for %q from %s subnet: %w", p.Name, role, err)
	}
	entry := fmt.Sprintf("%s/%d", addr, hostPrefixBits(addr))
	p.Address = []string{entry}
	if len(p.AllowedIps) == 0 {
		p.AllowedIps = []string{entry}
	}
	return nil
}

func (cfg *Configuration) ValidateAddressInPool(addrStr string) error {
	pool, err := cfg.PoolPrefix()
	if err != nil {
		return nil
	}
	prefix, err := parsePrefix(addrStr)
	if err != nil {
		return err
	}
	if !pool.Contains(prefix.Addr()) {
		return fmt.Errorf("address %s is outside ipPool %s", addrStr, pool)
	}
	return nil
}

type PoolEntry struct {
	Address netip.Addr
	Peer    string
	Free    bool
}

func (cfg *Configuration) PoolStatus() ([]PoolEntry, netip.Prefix, error) {
	pool, err := cfg.PoolPrefix()
	if err != nil {
		return nil, pool, err
	}
	used, err := cfg.UsedAddresses()
	if err != nil {
		return nil, pool, err
	}
	bits := 32
	if pool.Addr().Is6() {
		bits = 128
	}
	if bits-pool.Bits() > 16 {
		return nil, pool, fmt.Errorf("ipPool %s too large to list (%d addresses); use subnets", pool, 1<<(bits-pool.Bits()))
	}
	first := pool.Addr().Next()
	last, err := lastAddr(pool)
	if err != nil {
		return nil, pool, err
	}
	var entries []PoolEntry
	addr := first
	for addr.IsValid() && compareAddr(addr, last) <= 0 {
		if !(pool.Addr().Is4() && isBoundaryAddr(addr, pool)) {
			name, taken := used[addr]
			entries = append(entries, PoolEntry{Address: addr, Peer: name, Free: !taken})
		}
		addr = addr.Next()
	}
	sort.Slice(entries, func(i, j int) bool { return compareAddr(entries[i].Address, entries[j].Address) < 0 })
	return entries, pool, nil
}
