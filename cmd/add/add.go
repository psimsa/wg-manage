package add

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
)

type Add struct{}

func (a Add) ShortCommand() string {
	return "a"
}

func (a Add) LongCommand() string {
	return "add"
}

func (a Add) Run() {
	cmd := flag.NewFlagSet("add", flag.ExitOnError)
	name := cmd.String("name", "peer-1", "Name of the peer")
	ip := cmd.String("ip", "", "IP address of the new peer")
	endpoint := cmd.String("endpoint", "", "Endpoint, can be empty")
	configFile := cmd.String("config", "config.yaml", "Config file name")
	pkl := cmd.Bool("persistent", false, "Whether persistent keep alive should be set for one client")
	ar := cmd.String("add-routing", "", "Whether to add routing for the new peer")

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
		listenPort := strings.Split(*endpoint, ":")[1]
		intVar, err := strconv.Atoi(listenPort)
		if err != nil {
			fmt.Println("Error parsing listen port")
			os.Exit(1)
		}
		fmt.Println(peer.Endpoint)
		peer.ListenPort = &intVar
	}
	if *pkl {
		peer.PersistentKeepalive = new(int)
		*peer.PersistentKeepalive = 21
	}
	if *ar != "" {
		peer.PostUp = append(peer.PostUp,
			"sysctl -q -w net.ipv4.ip_forward=1",
			"iptables -A FORWARD -i wg0 -j ACCEPT",
			"iptables -A FORWARD -o wg0 -j ACCEPT",
			"iptables -t nat -A POSTROUTING -o wg0 -j MASQUERADE",
			fmt.Sprintf("iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", *ar),
		)
		peer.PostDown = append(peer.PostDown,
			"sysctl -q -w net.ipv4.ip_forward=0",
			"iptables -D FORWARD -i wg0 -j ACCEPT",
			"iptables -D FORWARD -o wg0 -j ACCEPT",
			"iptables -t nat -D POSTROUTING -o wg0 -j MASQUERADE",
			fmt.Sprintf("iptables -t nat -D POSTROUTING -o %s -j MASQUERADE", *ar),
		)
	}

	cfg.Peers = append(cfg.Peers, peer)

	models.SaveYaml(cfg, *configFile)
}

func (a Add) PrintHelp() {
	fmt.Println("[add | a] -name {peer-1} -ip {} -endpoint {} -persistent {false} -add-routing {false} -config {config.yaml}")
	fmt.Println("\tAdd a new record to the yaml file")
	fmt.Println("\tExample: wg-manage a -name MyHomeComputer -ip 10.0.2.10")
}
