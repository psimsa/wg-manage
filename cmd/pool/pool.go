package pool

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/models"
)

type Pool struct{}

func (p Pool) ShortCommand() string {
	return "p"
}

func (p Pool) LongCommand() string {
	return "pool"
}

func (p Pool) Run() {
	cmd := flag.NewFlagSet("pool", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")
	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	entries, pool, err := cfg.PoolStatus()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Printf("IP pool: %s\n", pool)
	for name, sub := range cfg.Settings.Subnets {
		fmt.Printf("  subnet %-8s %s\n", name+":", sub.CIDR)
	}
	fmt.Println()
	free := 0
	for _, e := range entries {
		if e.Free {
			free++
			continue
		}
		fmt.Printf("  %-18s %s\n", e.Address, e.Peer)
	}
	fmt.Printf("\n%d of %d addresses used, %d free\n", len(entries)-free, len(entries), free)
}

func (p Pool) PrintHelp() {
	fmt.Println("[pool | p] -config {config.yaml}")
	fmt.Println("\tShows the IP pool, subnets, used addresses (with owning peer) and free addresses")
	fmt.Println("\tExample: wg-manage p")
}
