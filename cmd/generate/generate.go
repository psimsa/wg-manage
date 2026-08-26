package generate

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path"
	"path/filepath"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"github.com/ofcoursedude/wg-manage/models"
)

type Generate struct{}

func (g Generate) ShortCommand() string {
	return "g"
}

func (g Generate) LongCommand() string {
	return "generate"
}

func (g Generate) Run() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")
	png := cmd.Bool("png", true, "Whether to generate QR codes as well")
	outputDir := cmd.String("output", "./output", "Output directory")

	cmd.Parse(os.Args[2:])

	cfg := models.LoadYaml(*configFile)
	outputPath, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Printf("error determining output path: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		fmt.Printf("could not create output directory: %v\n", err)
		os.Exit(1)
	}

	for _, peer := range cfg.Peers {
		filename := filepath.Join(outputPath, fmt.Sprintf("%s.conf", peer.Name))
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("can not create file: %v\n", err)
			os.Exit(1)
		}

		peer.GetInterface(file)
		for _, peer2 := range cfg.Peers {
			if peer2.PublicKey == peer.PublicKey || (peer2.Endpoint == nil && peer.Endpoint == nil) {
				continue
			}
			peer2.GetPeer(file, &cfg, &peer)
		}
		if err := file.Close(); err != nil {
			fmt.Printf("could not close file: %v\n", err)
			os.Exit(1)
		}
		if *png {
			if err := getQrCode(filename); err != nil {
				fmt.Printf("could not generate qr code: %v\n", err)
				os.Exit(1)
			}
		}
	}

}
func (g Generate) PrintHelp() {
	fmt.Println("[generate | g] -output {./output} -png {true} -config {config.yaml}")
	fmt.Println("\tGenerates config files from yaml file")
	fmt.Println("\tExample: wg-manage g -config my-wireguard.yaml")
}

func getQrCode(file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	qrCode, err := qr.Encode(string(content), qr.M, qr.Auto)
	if err != nil {
		return err
	}

	qrCode, err = barcode.Scale(qrCode, 200, 200)
	if err != nil {
		return err
	}
	ext := path.Ext(file)
	pngFileName := file[0:len(file)-len(ext)] + ".png"

	pngFile, err := os.Create(pngFileName)
	if err != nil {
		return err
	}
	defer pngFile.Close()

	if err := png.Encode(pngFile, qrCode); err != nil {
		return err
	}

	return nil
}
