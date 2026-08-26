package models

import (
	"fmt"
	"io"
	"os"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/ofcoursedude/wg-manage/utils"
)

func (p *Peer) GetInterface(w io.Writer) {
	fmt.Fprintln(w, "[Interface]")
	utils.WriteItemIfAny(p.PrivateKey, "PrivateKey", w)
	utils.WriteItemIfAny(p.ListenPort, "ListenPort", w)
	utils.WriteItemIfAny(p.FwMark, "FwMark", w)
	utils.WriteItemIfAny(p.InterfaceSectionWgQuick.SaveConfig, "SaveConfig", w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.Address, "Address", w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.DNS, "DNS", w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PreUp, "PreUp", w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PostUp, "PostUp", w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PreDown, "PreDown", w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PostDown, "PostDown", w)

	fmt.Fprintln(w)
}

func (p *Peer) GetPeer(w io.Writer, cfg *Configuration, parent *Peer) {
	fmt.Fprintln(w, "[Peer]")
	fmt.Fprintf(w, "# %s", p.Name)
	fmt.Fprintln(w, "")
	utils.WriteItemIfAny(p.PublicKey, "PublicKey", w)
	utils.WriteItemIfAny(cfg.PresharedKey, "PresharedKey", w)
	utils.WriteItemsIfAny(p.PeerSection.AllowedIps, "AllowedIPs", w)
	utils.WriteItemIfAny(p.PeerSection.Endpoint, "Endpoint", w)
	utils.WriteItemIfAny(parent.PeerSection.PersistentKeepalive, "PersistentKeepalive", w)
	utils.WriteItemIfAny(p.PeerSectionWgQuick.MTU, "MTU", w)
	utils.WriteItemIfAny(p.PeerSectionWgQuick.Table, "Table", w)
	overrides := parent.PeerOverrides[p.PublicKey]
	for _, item := range overrides {
		fmt.Fprintln(w, item)
	}

	fmt.Fprintln(w)
}

func SaveYaml(configuration Configuration, configFile string) error {
	data, err := GetYaml(configuration)
	if err != nil {
		return err
	}
	return utils.SaveToFile(configFile, data)
}

func GetYaml(configuration Configuration) ([]byte, error) {
	sort.Slice(configuration.Peers, func(i, j int) bool {
		if len(configuration.Peers[i].Address) == 0 {
			return false
		}
		if len(configuration.Peers[j].Address) == 0 {
			return true
		}
		return configuration.Peers[i].Address[0] < configuration.Peers[j].Address[0]
	})

	data, err := yaml.Marshal(configuration)
	if err != nil {
		return nil, fmt.Errorf("could not serialize yaml: %w", err)
	}
	return data, nil
}

func LoadYaml(configFile string) Configuration {
	cfg, err := LoadYamlFile(configFile)
	if err != nil {
		fmt.Printf("could not load yaml: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func LoadYamlFile(configFile string) (Configuration, error) {
	cfg := Configuration{}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, fmt.Errorf("could not read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("could not parse config: %w", err)
	}
	return cfg, nil
}

func RemovePeerByName(peers []Peer, name string) []Peer {
	for i, peer := range peers {
		if peer.Name == name {
			peers[i] = peers[len(peers)-1]
			return peers[:len(peers)-1]
		}
	}
	return peers
}

func RemovePeerByPublicKey(peers []Peer, publicKey string) []Peer {
	for i, peer := range peers {
		if peer.PublicKey == publicKey {
			peers[i] = peers[len(peers)-1]
			return peers[:len(peers)-1]
		}
	}
	return peers
}
