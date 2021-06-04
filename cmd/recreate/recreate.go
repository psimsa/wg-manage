package recreate

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
)

type Recreate struct {
}

func (r Recreate) PrintHelp() {
	fmt.Println("[recreate | rc] -config {config.yaml}")
	fmt.Println("\t(*DANGEROUS*) Recreates all private and public keys (*DANGEROUS*)")
	fmt.Println("\t(outputs to stdout for safety reasons)")
}

func (r Recreate) Run() {
	cmd := flag.NewFlagSet("recreate", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	oldnew := make([]struct {
		old string
		new string
	}, len(cfg.Peers))

	for i := range cfg.Peers {
		priv, pub := wg.GetKeyPair()
		oldPub := cfg.Peers[i].PublicKey
		oldnew[i] = struct {
			old string
			new string
		}{old: oldPub, new: pub}
		cfg.Peers[i].PrivateKey = priv
		cfg.Peers[i].PublicKey = pub
	}

	formatted := string(models.GetYaml(cfg))
	for _, pair := range oldnew {
		formatted = strings.ReplaceAll(formatted, pair.old, pair.new)
	}

	fmt.Println(formatted)
}

func (r Recreate) ShortCommand() string {
	return "rc"
}

func (r Recreate) LongCommand() string {
	return "recreate"
}
