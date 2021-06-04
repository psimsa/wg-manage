package bootstrap

import (
	"flag"
	"fmt"
	"os"

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
	configFile := cmd.String("output", "config.yaml", "Output file name")

	cmd.Parse(os.Args[2:])
	cfg := models.Configuration{
		Peers: make([]models.Peer, 3),
	}
	priv, pub := wg.GetKeyPair()
	server := models.NewPeer("Server", priv, pub)
	server.Endpoint = endpoint
	server.Address = make([]string, 1)
	server.Address[0] = "10.0.2.1/32"
	server.AllowedIps = make([]string, 1)
	server.AllowedIps[0] = "0.0.0.0/0"
	server.PostUp = []string{
		"sysctl -q -w net.ipv4.ip_forward=1",
		"iptables -A FORWARD -i wg0 -j ACCEPT",
		"iptables -A FORWARD -o wg0 -j ACCEPT",
		"iptables -t nat -A POSTROUTING -o wg0 -j MASQUERADE",
		"iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
	}
	server.PostDown = []string{
		"sysctl -q -w net.ipv4.ip_forward=0",
		"iptables -D FORWARD -i wg0 -j ACCEPT",
		"iptables -D FORWARD -o wg0 -j ACCEPT",
		"iptables -t nat -D POSTROUTING -o wg0 -j MASQUERADE",
		"iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
	}
	cfg.Peers[0] = *server

	priv, pub = wg.GetKeyPair()
	pc := models.NewPeer("My Laptop", priv, pub)
	pc.Address = make([]string, 1)
	pc.Address[0] = "10.0.2.2/32"
	pc.AllowedIps = make([]string, 1)
	pc.AllowedIps[0] = "10.0.2.2/32"
	if *pkl == true {
		pc.PersistentKeepalive = new(int)
		*pc.PersistentKeepalive = 21
	}
	cfg.Peers[1] = *pc

	priv, pub = wg.GetKeyPair()
	phone := models.NewPeer("My Phone", priv, pub)
	phone.Address = make([]string, 1)
	phone.Address[0] = "10.0.2.3/32"
	phone.AllowedIps = make([]string, 1)
	phone.AllowedIps[0] = "10.0.2.3/32"
	cfg.Peers[2] = *phone

	models.SaveYaml(cfg, *configFile)

}
func (b Bootstrap) PrintHelp() {
	fmt.Println("[bootstrap | b] -endpoint {some.server.somewhere:51820} -persistent {false} -output {config.yaml}")
	fmt.Println("\tCreates a simple, ready to run wireguard network configuration with one server and two clients, 'kill-switch' (all client traffic goes through wireguard, including external) and NAT.")
	fmt.Println("\tAssumes 10.0.2.0/24 CIDR, wg0 and eth0 on server.")
	fmt.Println("\tExample: wg-manage b -endpoint myhome.someddnsprovider.com")
}
