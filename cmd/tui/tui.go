package tui

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/tui"
)

type Tui struct{}

func (t Tui) ShortCommand() string {
	return "t"
}

func (t Tui) LongCommand() string {
	return "tui"
}

func (t Tui) Run() {
	cmd := flag.NewFlagSet("tui", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")
	cmd.Parse(os.Args[2:])

	if err := tui.Run(*configFile); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func (t Tui) PrintHelp() {
	fmt.Println("[tui | t] -config {config.yaml}")
	fmt.Println("\tInteractive terminal UI: manage peers, routing, overrides and settings without editing yaml")
	fmt.Println("\tExample: wg-manage tui")
}
