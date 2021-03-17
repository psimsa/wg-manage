package models

import (
	"github.com/ofcoursedude/wg-manage/utils"
)

func NewPeer(name string, privateKey string, publicKey string) *Peer {
	return &Peer{
		Name:        name,
		Description: nil,
		InterfaceSection: InterfaceSection{
			PrivateKey: privateKey,
			ListenPort: nil,
			FwMark:     nil,
		},
		InterfaceSectionWgQuick: InterfaceSectionWgQuick{
			Address:    nil,
			DNS:        nil,
			SaveConfig: nil,
			PreUp:      nil,
			PostUp:     nil,
			PreDown:    nil,
			PostDown:   nil,
		},
		PeerSection: PeerSection{
			PublicKey:           publicKey,
			AllowedIps:          nil,
			Endpoint:            nil,
			PersistentKeepalive: nil,
		},
		PeerSectionWgQuick: PeerSectionWgQuick{
			MTU:   nil,
			Table: nil,
		},
	}
}

func SamplePeer() *Peer {
	ds := utils.CreateDummyStringSlice()
	s := "<add here or remove section>"
	port := new(int)
	*port = 51820
	kal := new(int)
	*kal = 21

	p := &Peer{
		Name:        "Peer",
		Description: &s,
		InterfaceSection: InterfaceSection{
			PrivateKey: "abc",
			ListenPort: port,
			FwMark:     &s,
		},
		InterfaceSectionWgQuick: InterfaceSectionWgQuick{
			Address:    ds,
			DNS:        ds,
			SaveConfig: new(bool),
			PreUp:      ds,
			PostUp:     ds,
			PreDown:    ds,
			PostDown:   ds,
		},
		PeerSection: PeerSection{
			PublicKey:           "def",
			AllowedIps:          ds,
			Endpoint:            &s,
			PersistentKeepalive: kal,
		},
		PeerSectionWgQuick: PeerSectionWgQuick{
			MTU:   &s,
			Table: &s,
		},
		PeerOverrides: make(map[string][]string, 1),
	}
	p.PeerOverrides["some public key 1"] = make([]string, 2)
	p.PeerOverrides["some public key 1"][0] = "AllowedIPs=10.0.0.0/24"
	p.PeerOverrides["some public key 1"][1] = "AllowedIPs=192.168.0.0/24"
	return p
}
