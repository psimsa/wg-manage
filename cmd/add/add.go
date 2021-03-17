package add

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
)

func Run() {
	cmd := flag.NewFlagSet("add", flag.ExitOnError)
	name := cmd.String("name", "peer-1", "Name of the peer")
	ip := cmd.String("ip", "", "IP address of the new peer")
	endpoint := cmd.String("endpoint", "", "Endpoint, can be empty")
	configFile := cmd.String("config", "config.yaml", "Config file name")

	cmd.Parse(os.Args[2:])
	cfg := models.LoadYaml(*configFile)
	peer := models.Peer{}
	priv, pub := wg.GetKeyPair()
	peer.PrivateKey = priv
	peer.PublicKey = pub
	peer.Name = *name
	if *ip != "" {
		peer.Address = make([]string, 1)
		peer.Address[0] = *ip
	}
	peer.AllowedIps = peer.Address
	if *endpoint != "" {
		peer.Endpoint = endpoint
	}

	cfg.Peers = append(cfg.Peers, peer)

	models.SaveYaml(cfg, *configFile)
}
func PrintHelp() {
	fmt.Println("[add | a] -name {peer-1} -ip {} -endpoint {} -config {config.yaml}")
	fmt.Println("\tAdd a new record to the yaml file")
	fmt.Println("\tExample: wg-manage a -name MyHomeComputer -ip 10.0.2.10")
}
