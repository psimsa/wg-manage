package format

import (
	"flag"
	"fmt"
	"os"

	"github.com/ofcoursedude/wg-manage/models"
)

type Format struct{}

func (f Format) PrintHelp() {
	fmt.Println("[format | f] -input {config.yaml}")
	fmt.Println("\tFormats config file to stdout")
}

func (f Format) Run() {
	cmd := flag.NewFlagSet("format", flag.ExitOnError)
	configFile := cmd.String("input", "config.yaml", "Input file to format")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	formatted, err := models.GetYaml(cfg)
	if err != nil {
		fmt.Printf("could not format config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(formatted))

}

func (f Format) ShortCommand() string {
	return "f"
}

func (f Format) LongCommand() string {
	return "format"
}
