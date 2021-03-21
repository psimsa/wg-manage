package format

import (
	"flag"
	"fmt"
	"github.com/ofcoursedude/wg-manage/models"
	"os"
)

type Format struct {
}

func (f Format) PrintHelp() {
	fmt.Println("[format | f] -input {config.yaml}")
	fmt.Println("\tFormats config file to stdout")
}

func (f Format) Run() {
	cmd := flag.NewFlagSet("format", flag.ExitOnError)
	configFile := cmd.String("input", "config.yaml", "Input file to format")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	formatted := models.GetYaml(cfg)
	fmt.Println(string(formatted))
}

func (f Format) ShortCommand() string {
	return "f"
}

func (f Format) LongCommand() string {
	return "format"
}
