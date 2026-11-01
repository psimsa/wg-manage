package models

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v2"
)

func ParseYaml(data []byte) (Configuration, error) {
	cfg := Configuration{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("could not parse yaml: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

func (cfg *Configuration) ToYaml() ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not serialize to yaml: %w", err)
	}
	return data, nil
}

func (cfg *Configuration) RenderPeerConf(name string) (string, error) {
	for i := range cfg.Peers {
		if cfg.Peers[i].Name != name {
			continue
		}
		peer := cfg.Peers[i]
		var buf bytes.Buffer
		peer.GetInterface(&buf)
		for _, peer2 := range cfg.Peers {
			if peer2.PublicKey == peer.PublicKey || (peer2.Endpoint == nil && peer.Endpoint == nil) {
				continue
			}
			peer2.GetPeer(&buf, cfg, &peer)
		}
		return buf.String(), nil
	}
	return "", fmt.Errorf("no peer named %q", name)
}

func (cfg *Configuration) RenderAll() map[string]string {
	out := map[string]string{}
	for _, peer := range cfg.Peers {
		if conf, err := cfg.RenderPeerConf(peer.Name); err == nil {
			out[peer.Name] = conf
		}
	}
	return out
}
