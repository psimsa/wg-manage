package regenerate

import (
	"flag"
	"fmt"
	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
	"os"
)

type Recreate struct {
}

func (r Recreate) PrintHelp() {
	fmt.Println("[rc | recreate] -config {config.yaml}")
	fmt.Println("\t(*DANGEROUS*) Recreates all private and public keys (*DANGEROUS*)")
	fmt.Println("\t(outputs to stdout for safety reasons)")
}

func (r Recreate) Run() {
	cmd := flag.NewFlagSet("recreate", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	for i := range cfg.Peers {
		priv, pub := wg.GetKeyPair()
		cfg.Peers[i].PrivateKey = priv
		cfg.Peers[i].PublicKey = pub
	}

	formatted := models.GetYaml(cfg)
	fmt.Println(string(formatted))
}

func (r Recreate) ShortCommand() string {
	return "rc"
}

func (r Recreate) LongCommand() string {
	return "recreate"
}
