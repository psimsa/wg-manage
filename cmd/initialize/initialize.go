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
	return "initialize"
}

func (i Initialize) Run() {

	cmd := flag.NewFlagSet("init", flag.ExitOnError)
	peerCount := cmd.Int("peers", 2, "Peer count - typically at least 2")
	configFile := cmd.String("output", "config.yaml", "Output file name")
	simple := cmd.Bool("simple", false, "Whether to create only basic structure")
	preshared := cmd.Bool("preshared", true, "Whether to create preshared key")

	cmd.Parse(os.Args[2:])

	cfg := models.Configuration{
		Peers: make([]models.Peer, *peerCount),
	}
	if *preshared {
		presharedKey := wg.GetPresharedKey()
		cfg.PresharedKey = &presharedKey
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
