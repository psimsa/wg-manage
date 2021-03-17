package generate

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ofcoursedude/wg-manage/utils"

	"github.com/ofcoursedude/wg-manage/models"
)

func Run() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")
	outputDir := cmd.String("output", "./output", "Output directory")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	path, err := filepath.Abs(*outputDir)
	utils.HandleError(err, "error determining output path")
	_, err = os.Stat(path)
	if err != nil {
		os.MkdirAll(path, 2147484141)
	}

	for _, peer := range cfg.Peers {
		filename := fmt.Sprintf("%s/%s.conf", path, peer.Name)
		file, err := os.Create(filename)
		utils.HandleError(err, "can not create file")
		defer file.Close()

		peer.GetInterface(file)
		for _, peer2 := range cfg.Peers {
			if peer2.PublicKey == peer.PublicKey || (peer2.Endpoint == nil && peer.Endpoint == nil) {
				continue
			}
			peer2.GetPeer(file, &cfg, &peer)
		}
	}

}
func PrintHelp() {
	fmt.Println("[generate | g] -output {./output} -config {config.yaml}")
	fmt.Println("\tGenerates config files from yaml file")
	fmt.Println("\tExample: wg-manage g -config my-wireguard.yaml")
}
