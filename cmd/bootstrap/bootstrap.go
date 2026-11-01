package bootstrap

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
)

type Bootstrap struct{}

func (b Bootstrap) ShortCommand() string {
	return "b"
}

func (b Bootstrap) LongCommand() string {
	return "bootstrap"
}

func (b Bootstrap) Run() {
	cmd := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	endpoint := cmd.String("endpoint", "some.server.somewhere:51820", "The new wireguard server endpoint")
	pkl := cmd.Bool("persistent", false, "Whether persistent keep alive should be set for one client")
	port := cmd.Int("port", 0, "Listen port for the server (default: from endpoint, else settings.serverDefaults.listenPort, else 51820)")
	preshared := cmd.String("preshared", "none", "Preshared key mode: none, single (one global key), per-peer (unique key per peer)")
	configFile := cmd.String("output", "config.yaml", "Output file name")

	cmd.Parse(os.Args[2:])
	settings := models.DefaultSettings()
	if *port != 0 {
		if *port < 1 || *port > 65535 {
			fmt.Println("Error: -port must be between 1 and 65535")
			os.Exit(1)
		}
		settings.ServerDefaults.ListenPort = port
	}
	cfg := models.Configuration{
		Settings: settings,
		Peers:    make([]models.Peer, 3),
	}
	egress := settings.ServerDefaults.EgressInterface
	priv, pub := wg.GetKeyPair()
	server := models.NewPeer("Server", priv, pub)
	server.Endpoint = endpoint
	if err := cfg.AllocateForPeer(server); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	server.AllowedIps = make([]string, 1)
	server.AllowedIps[0] = "0.0.0.0/0"
	server.PostUp = []string{
		"sysctl -q -w net.ipv4.ip_forward=1",
		"iptables -A FORWARD -i wg0 -j ACCEPT",
		"iptables -A FORWARD -o wg0 -j ACCEPT",
		"iptables -t nat -A POSTROUTING -o wg0 -j MASQUERADE",
		fmt.Sprintf("iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", egress),
	}
	server.PostDown = []string{
		"sysctl -q -w net.ipv4.ip_forward=0",
		"iptables -D FORWARD -i wg0 -j ACCEPT",
		"iptables -D FORWARD -o wg0 -j ACCEPT",
		"iptables -t nat -D POSTROUTING -o wg0 -j MASQUERADE",
		fmt.Sprintf("iptables -t nat -D POSTROUTING -o %s -j MASQUERADE", egress),
	}
	if *port != 0 {
		server.ListenPort = port
		if parts := strings.Split(*endpoint, ":"); len(parts) == 2 {
			*endpoint = parts[0] + fmt.Sprintf(":%d", *port)
			server.Endpoint = endpoint
		} else {
			*endpoint = fmt.Sprintf("%s:%d", *endpoint, *port)
			server.Endpoint = endpoint
		}
	}
	cfg.ApplyDefaultsToPeer(server)
	cfg.Peers[0] = *server

	priv, pub = wg.GetKeyPair()
	pc := models.NewPeer("My Laptop", priv, pub)
	if err := cfg.AllocateForPeer(pc); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if *pkl == true {
		pc.PersistentKeepalive = new(int)
		*pc.PersistentKeepalive = 21
	}
	cfg.ApplyDefaultsToPeer(pc)
	cfg.Peers[1] = *pc

	priv, pub = wg.GetKeyPair()
	phone := models.NewPeer("My Phone", priv, pub)
	if err := cfg.AllocateForPeer(phone); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	cfg.ApplyDefaultsToPeer(phone)
	cfg.Peers[2] = *phone

	switch *preshared {
	case "none":
	case "single":
		cfg.PresharedKey = strPtr(wg.GetPresharedKey())
	case "per-peer":
		for i := range cfg.Peers {
			cfg.Peers[i].PresharedKey = strPtr(wg.GetPresharedKey())
		}
	default:
		fmt.Println("Error: -preshared must be one of: none, single, per-peer")
		os.Exit(1)
	}

	models.SaveYaml(cfg, *configFile)

}

func strPtr(s string) *string { return &s }
func (b Bootstrap) PrintHelp() {
	fmt.Println("[bootstrap | b] -endpoint {some.server.somewhere:51820} -persistent {false} -port {0=auto} -preshared {none|single|per-peer} -output {config.yaml}")
	fmt.Println("\tCreates a simple, ready to run wireguard network configuration with one server and two clients, 'kill-switch' (all client traffic goes through wireguard, including external) and NAT.")
	fmt.Println("\tAssumes 10.0.2.0/24 CIDR, wg0 and eth0 on server.")
	fmt.Println("\tExample: wg-manage b -endpoint myhome.someddnsprovider.com")
}
