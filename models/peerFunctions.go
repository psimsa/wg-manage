package models

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"gopkg.in/yaml.v2"

	"github.com/ofcoursedude/wg-manage/utils"
)

func (p *Peer) GetInterface(w io.Writer) {
	fmt.Fprintln(w, "[Interface]")
	utils.WriteItemIfAny(p.PrivateKey, "PrivateKey", &w)
	utils.WriteItemIfAny(p.ListenPort, "ListenPort", &w)
	utils.WriteItemIfAny(p.FwMark, "FwMark", &w)
	utils.WriteItemIfAny(p.InterfaceSectionWgQuick.SaveConfig, "SaveConfig", &w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.Address, "Address", &w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.DNS, "DNS", &w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PreUp, "PreUp", &w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PostUp, "PostUp", &w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PreDown, "PreDown", &w)
	utils.WriteItemsIfAny(p.InterfaceSectionWgQuick.PostDown, "PostDown", &w)

	fmt.Fprintln(w)
}

func (p *Peer) GetPeer(w io.Writer, cfg *Configuration, parent *Peer) {
	fmt.Fprintln(w, "[Peer]")
	fmt.Fprintf(w, "# %s", p.Name)
	fmt.Fprintln(w, "")
	utils.WriteItemIfAny(p.PublicKey, "PublicKey", &w)
	utils.WriteItemIfAny(cfg.PresharedKey, "PresharedKey", &w)
	utils.WriteItemsIfAny(p.PeerSection.AllowedIps, "AllowedIPs", &w)
	utils.WriteItemIfAny(p.PeerSection.Endpoint, "Endpoint", &w)
	utils.WriteItemIfAny(parent.PeerSection.PersistentKeepalive, "PersistentKeepalive", &w)
	utils.WriteItemIfAny(p.PeerSectionWgQuick.MTU, "MTU", &w)
	utils.WriteItemIfAny(p.PeerSectionWgQuick.Table, "Table", &w)
	overrides := parent.PeerOverrides[p.PublicKey]
	for _, item := range overrides {
		fmt.Fprintln(w, item)
	}

	fmt.Fprintln(w)
}

func SaveYaml(configuration Configuration, configFile string) {
	data := GetYaml(configuration)
	utils.SaveToFile(configFile, data)
}

func GetYaml(configuration Configuration) []byte {
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
	utils.HandleError(err, "could not serialize to yaml")
	return data
}

func LoadYaml(configFile string) Configuration {
	cfg := Configuration{}
	data, err := ioutil.ReadFile(configFile)
	utils.HandleError(err, "could not open yaml file")
	err = yaml.Unmarshal(data, &cfg)
	utils.HandleError(err, "could not parse yaml file")
	return cfg
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
