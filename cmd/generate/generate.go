package generate

import (
	"flag"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ofcoursedude/wg-manage/utils"

	"github.com/ofcoursedude/wg-manage/models"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

func Run() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	configFile := cmd.String("config", "config.yaml", "Config file name")
	png := cmd.Bool("png", true, "Whether to generate QR codes as well")
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

		peer.GetInterface(file)
		for _, peer2 := range cfg.Peers {
			if peer2.PublicKey == peer.PublicKey || (peer2.Endpoint == nil && peer.Endpoint == nil) {
				continue
			}
			peer2.GetPeer(file, &cfg, &peer)
		}
		file.Close()
		if *png == true {
			getQrCode(filename)
		}
	}

}
func PrintHelp() {
	fmt.Println("[generate | g] -output {./output} -png {true} -config {config.yaml}")
	fmt.Println("\tGenerates config files from yaml file")
	fmt.Println("\tExample: wg-manage g -config my-wireguard.yaml")
}

func getQrCode(file string) {
	content, err := ioutil.ReadFile(file)
	utils.HandleError(err, "Can not open file")
	qrCode, err := qr.Encode(string(content), qr.M, qr.Auto)
	utils.HandleError(err, "can not generate qr code")

	// Scale the barcode to 200x200 pixels
	qrCode, _ = barcode.Scale(qrCode, 200, 200)
	ext := path.Ext(file)
	pngFileName := file[0:len(file)-len(ext)] + ".png"

	// create the output file
	pngFile, _ := os.Create(pngFileName)
	defer pngFile.Close()

	// encode the barcode as png
	png.Encode(pngFile, qrCode)
}
