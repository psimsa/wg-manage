package psk

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
)

type Psk struct{}

func (p Psk) ShortCommand() string {
	return "psk"
}

func (p Psk) LongCommand() string {
	return "psk"
}

func (p Psk) Run() {
	cmd := flag.NewFlagSet("psk", flag.ExitOnError)
	mode := cmd.String("mode", "", "PSK mode: single (one global key), per-peer (unique key per peer), remove (clear all), status (default)")
	configFile := cmd.String("config", "config.yaml", "Config file name")
	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	switch *mode {
	case "", "status":
		printStatus(&cfg)
	case "single":
		cfg.PresharedKey = strPtr(wg.GetPresharedKey())
		for i := range cfg.Peers {
			cfg.Peers[i].PresharedKey = nil
		}
		models.SaveYaml(cfg, *configFile)
		fmt.Println("Generated a single global preshared key.")
	case "per-peer":
		cfg.PresharedKey = nil
		for i := range cfg.Peers {
			cfg.Peers[i].PresharedKey = strPtr(wg.GetPresharedKey())
		}
		models.SaveYaml(cfg, *configFile)
		fmt.Printf("Generated a unique preshared key for each of %d peers.\n", len(cfg.Peers))
	case "remove":
		cfg.PresharedKey = nil
		for i := range cfg.Peers {
			cfg.Peers[i].PresharedKey = nil
		}
		models.SaveYaml(cfg, *configFile)
		fmt.Println("Removed all preshared keys.")
	default:
		fmt.Println("Unknown mode. Use one of: status, single, per-peer, remove")
		os.Exit(1)
	}
}

func printStatus(cfg *models.Configuration) {
	global := cfg.PresharedKey != nil && *cfg.PresharedKey != ""
	fmt.Printf("global preshared key: %v\n", global)
	for i := range cfg.Peers {
		p := &cfg.Peers[i]
		own := p.PresharedKey != nil && *p.PresharedKey != ""
		source := "inherited global"
		if own {
			source = "own per-peer key"
		} else if !global {
			source = "none"
		}
		fmt.Printf("  %-20s %s\n", p.Name, source)
	}
}

func strPtr(s string) *string { return &s }

func (p Psk) PrintHelp() {
	fmt.Println("[psk] [-mode {status|single|per-peer|remove}] -config {config.yaml}")
	fmt.Println("\tManage preshared keys: status (default), single global key, unique per-peer keys, or remove all")
	fmt.Println("\tExample: wg-manage psk -mode per-peer")
}
