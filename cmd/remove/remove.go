package remove

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/models"
)

type Remove struct{}

func (r Remove) ShortCommand() string {
	return "r"
}

func (r Remove) LongCommand() string {
	return "remove"
}

func (r Remove) Run() {
	cmd := flag.NewFlagSet("remove", flag.ExitOnError)
	name := cmd.String("name", "", "Name of the peer to remove")
	key := cmd.String("key", "", "Public key of the peer to remove")
	configFile := cmd.String("config", "config.yaml", "Config file name")

	cmd.Parse(os.Args[2:])
	cfg := models.LoadYaml(*configFile)
	var newPeers []models.Peer
	if *name != "" {
		newPeers = models.RemovePeerByName(cfg.Peers, *name)
	} else if *key != "" {
		newPeers = models.RemovePeerByPublicKey(cfg.Peers, *key)
	} else {
		fmt.Println("Must specify either name or public key")
		os.Exit(1)
	}
	if len(cfg.Peers) != len(newPeers) {
		cfg.Peers = newPeers
		if err := models.SaveYaml(cfg, *configFile); err != nil {
			fmt.Printf("could not write config: %v\n", err)
			os.Exit(1)
		}
	} else {

		fmt.Println("No match found.")
	}
}
func (r Remove) PrintHelp() {
	fmt.Println("[remove | r] [-name {} | -key {}] -config {config.yaml}")
	fmt.Println("\tRemoves a record from the config file by name or public key")
	fmt.Println("\tExample: wg-manage r -name MyHomeComputer")
}
