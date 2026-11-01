package initialize

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/wg"

	"github.com/ofcoursedude/wg-manage/models"
)

type Initialize struct{}

func (i Initialize) ShortCommand() string {
	return "i"
}

func (i Initialize) LongCommand() string {
	return "init"
}

func (i Initialize) Run() {

	cmd := flag.NewFlagSet("init", flag.ExitOnError)
	peerCount := cmd.Int("peers", 2, "Peer count - typically at least 2")
	configFile := cmd.String("output", "config.yaml", "Output file name")
	simple := cmd.Bool("simple", false, "Whether to create only basic structure")
	preshared := cmd.String("preshared", "single", "Preshared key mode: none, single (one global key), per-peer (unique key per peer)")
	settings := cmd.Bool("settings", true, "Whether to include the default settings section (ip pool, defaults)")

	cmd.Parse(os.Args[2:])

	cfg := models.Configuration{
		Peers: make([]models.Peer, *peerCount),
	}
	if *settings {
		cfg.Settings = models.DefaultSettings()
	}
	switch *preshared {
	case "none", "false":
	case "single", "true":
		presharedKey := wg.GetPresharedKey()
		cfg.PresharedKey = &presharedKey
	case "per-peer":
	default:
		fmt.Println("Error: -preshared must be one of: none, single, per-peer")
		os.Exit(1)
	}

	for i := range cfg.Peers {
		var peer models.Peer
		if *simple {
			peer = models.Peer{}
			peer.Name = "<add name>"
		} else {
			peer = *models.SamplePeer()
			peer.Name = fmt.Sprintf("%s-%d", peer.Name, i+1)
		}

		priv, pub := wg.GetKeyPair()
		peer.PrivateKey = priv
		peer.PublicKey = pub
		if *preshared == "per-peer" {
			psk := wg.GetPresharedKey()
			peer.PresharedKey = &psk
		}
		cfg.Peers[i] = peer
	}

	models.SaveYaml(cfg, *configFile)
}

func (i Initialize) PrintHelp() {
	fmt.Println("[init | i] -clients {1} -servers {1} -output {config.yaml}")
	fmt.Println("\tInitialize basic configuration with specified number of clients and servers")
	fmt.Println("\tRequires further editing of the resulting yaml file with endpoint information.")
	fmt.Println("\tExample: wg-manage i -clients 10 -servers 2")
}
