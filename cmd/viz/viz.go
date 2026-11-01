package viz

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/viz"
)

type Viz struct{}

func (v Viz) ShortCommand() string {
	return "v"
}

func (v Viz) LongCommand() string {
	return "viz"
}

func (v Viz) Run() {
	cmd := flag.NewFlagSet("viz", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")
	format := cmd.String("format", "ascii", "Output format: ascii, dot (graphviz), mermaid")
	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	out, err := viz.Render(&cfg, *format)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	fmt.Print(out)
}

func (v Viz) PrintHelp() {
	fmt.Println("[viz | v] -config {config.yaml} -format {ascii|dot|mermaid}")
	fmt.Println("\tVisualizes the configured wireguard network (servers, clients, overrides, free addresses)")
	fmt.Println("\tExample: wg-manage v -format dot | dot -Tsvg > network.svg")
}
