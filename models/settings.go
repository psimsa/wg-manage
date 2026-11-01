package models

import (
	"fmt"
	"net/netip"
)

const (
	DefaultIPPool      = "10.0.2.0/24"
	DefaultServersCIDR = "10.0.2.0/26"
	DefaultClientsCIDR = "10.0.2.64/26"
	DefaultOutputDir   = "./output"
	DefaultListenPort  = 51820
	DefaultEgressIface = "eth0"
)

func DefaultSettings() *Settings {
	listenPort := DefaultListenPort
	png := true
	return &Settings{
		IPPool: DefaultIPPool,
		Subnets: map[string]Subnet{
			"servers": {CIDR: DefaultServersCIDR},
			"clients": {CIDR: DefaultClientsCIDR},
		},
		ServerDefaults: &PeerDefaults{
			ListenPort:      &listenPort,
			EgressInterface: DefaultEgressIface,
		},
		ClientDefaults: &PeerDefaults{},
		Generate:       &GenerateSettings{OutputDir: DefaultOutputDir, PNG: &png},
	}
}

func (s *Settings) Validate() error {
	if s == nil {
		return nil
	}
	if s.IPPool != "" {
		pool, err := netip.ParsePrefix(s.IPPool)
		if err != nil {
			return fmt.Errorf("settings.ipPool %q is not a valid CIDR: %w", s.IPPool, err)
		}
		pool = netip.PrefixFrom(pool.Addr(), pool.Bits())
		for name, subnet := range s.Subnets {
			sub, err := netip.ParsePrefix(subnet.CIDR)
			if err != nil {
				return fmt.Errorf("settings.subnets.%s.cidr %q is not a valid CIDR: %w", name, subnet.CIDR, err)
			}
			sub = netip.PrefixFrom(sub.Addr(), sub.Bits())
			if !pool.Contains(sub.Addr()) || sub.Bits() < pool.Bits() {
				return fmt.Errorf("settings.subnets.%s.cidr %q is not contained within ipPool %q", name, subnet.CIDR, s.IPPool)
			}
			if sub.Addr().Is4() != pool.Addr().Is4() {
				return fmt.Errorf("settings.subnets.%s.cidr %q address family does not match ipPool %q", name, subnet.CIDR, s.IPPool)
			}
		}
	}
	for name := range s.Subnets {
		if s.IPPool == "" {
			return fmt.Errorf("settings.subnets.%s is set but settings.ipPool is empty", name)
		}
	}
	if d := s.ServerDefaults; d != nil && d.ListenPort != nil && (*d.ListenPort < 1 || *d.ListenPort > 65535) {
		return fmt.Errorf("settings.serverDefaults.listenPort %d out of range 1-65535", *d.ListenPort)
	}
	if d := s.ClientDefaults; d != nil && d.PersistentKeepalive != nil && (*d.PersistentKeepalive < 0 || *d.PersistentKeepalive > 65535) {
		return fmt.Errorf("settings.clientDefaults.persistentKeepalive %d out of range 0-65535", *d.PersistentKeepalive)
	}
	return nil
}

func (p *Peer) IsServer() bool {
	return p.Endpoint != nil && *p.Endpoint != ""
}

func (cfg *Configuration) ApplyDefaultsToPeer(p *Peer) {
	if cfg.Settings == nil {
		return
	}
	d := cfg.Settings.ClientDefaults
	if p.IsServer() {
		d = cfg.Settings.ServerDefaults
	}
	if d == nil {
		return
	}
	if p.ListenPort == nil && d.ListenPort != nil {
		port := *d.ListenPort
		p.ListenPort = &port
	}
	if p.PersistentKeepalive == nil && d.PersistentKeepalive != nil {
		ka := *d.PersistentKeepalive
		p.PersistentKeepalive = &ka
	}
	if len(p.PostUp) == 0 && len(d.PostUp) > 0 {
		p.PostUp = append([]string(nil), d.PostUp...)
	}
	if len(p.PostDown) == 0 && len(d.PostDown) > 0 {
		p.PostDown = append([]string(nil), d.PostDown...)
	}
}

func (cfg *Configuration) ResolveOutputDir(flagSet bool, flagValue string) string {
	if flagSet {
		return flagValue
	}
	if cfg.Settings != nil && cfg.Settings.Generate != nil && cfg.Settings.Generate.OutputDir != "" {
		return cfg.Settings.Generate.OutputDir
	}
	return DefaultOutputDir
}

func (cfg *Configuration) ResolvePNG(flagSet bool, flagValue bool) bool {
	if flagSet {
		return flagValue
	}
	if cfg.Settings != nil && cfg.Settings.Generate != nil && cfg.Settings.Generate.PNG != nil {
		return *cfg.Settings.Generate.PNG
	}
	return true
}

func (cfg *Configuration) Validate() error {
	if err := cfg.Settings.Validate(); err != nil {
		return err
	}
	return nil
}
