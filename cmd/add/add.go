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
	ip := cmd.String("ip", "", "IP address of the new peer (auto-allocated from the pool when omitted and settings.ipPool is set)")
	endpoint := cmd.String("endpoint", "", "Endpoint, can be empty")
	server := cmd.Bool("server", false, "Whether the new peer is a server (address allocated from the servers subnet)")
	port := cmd.Int("port", 0, "Listen port for a server peer (overrides endpoint port and settings default)")
	configFile := cmd.String("config", "config.yaml", "Config file name")
	pkl := cmd.Bool("persistent", false, "Whether persistent keep alive should be set for one client")
	ar := cmd.String("add-routing", "", "Whether to add routing for the new peer")
	force := cmd.Bool("force", false, "Allow an explicit -ip outside the configured pool")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	peer := models.Peer{}
	priv, pub := wg.GetKeyPair()
	peer.PrivateKey = priv
	peer.PublicKey = pub
	peer.Name = *name
	if *endpoint != "" {
		peer.Endpoint = endpoint
	}
	if *server && *endpoint == "" {
		fmt.Println("A server peer requires -endpoint")
		os.Exit(1)
	}
	if *ip != "" {
		if !*force {
			if err := cfg.ValidateAddressInPool(*ip); err != nil {
				fmt.Println("Error:", err, "(use -force to override)")
				os.Exit(1)
			}
		}
		peer.Address = make([]string, 1)
		peer.Address[0] = *ip
		peer.AllowedIps = peer.Address
	} else if cfg.Settings != nil && cfg.Settings.IPPool != "" {
		if err := cfg.AllocateForPeer(&peer); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	} else {
		peer.AllowedIps = peer.Address
	}
	if *endpoint != "" {
		parts := strings.Split(*endpoint, ":")
		if len(parts) == 2 {
			intVar, err := strconv.Atoi(parts[1])
			if err != nil || intVar < 1 || intVar > 65535 {
				fmt.Println("Error parsing listen port")
				os.Exit(1)
			}
			peer.ListenPort = &intVar
		}
	}
	if *port != 0 {
		if *port < 1 || *port > 65535 {
			fmt.Println("Error: -port must be between 1 and 65535")
			os.Exit(1)
		}
		peer.ListenPort = port
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

	cfg.ApplyDefaultsToPeer(&peer)
	cfg.Peers = append(cfg.Peers, peer)

	models.SaveYaml(cfg, *configFile)
}

func (a Add) PrintHelp() {
	fmt.Println("[add | a] -name {peer-1} [-ip {} | auto] [-server -endpoint {}] -persistent {false} -add-routing {false} -force {false} -config {config.yaml}")
	fmt.Println("\tAdd a new record to the yaml file; without -ip the address is auto-allocated from settings.ipPool")
	fmt.Println("\tExample: wg-manage a -name MyHomeComputer (auto IP) or wg-manage a -name MyHomeComputer -ip 10.0.2.10")
}
