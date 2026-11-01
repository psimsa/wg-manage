package tui

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/utils"
)

func generateConfigs(cfg *models.Configuration, outputDir string, pngEnabled bool) error {
	path, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	for _, peer := range cfg.Peers {
		filename := fmt.Sprintf("%s/%s.conf", path, peer.Name)
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		peer.GetInterface(file)
		for _, peer2 := range cfg.Peers {
			if peer2.PublicKey == peer.PublicKey || (peer2.Endpoint == nil && peer.Endpoint == nil) {
				continue
			}
			peer2.GetPeer(file, cfg, &peer)
		}
		file.Close()
		if pngEnabled {
			if err := writeQrCode(filename); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeQrCode(file string) error {
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
	ext := filepath.Ext(file)
	pngFileName := file[:len(file)-len(ext)] + ".png"
	pngFile, err := os.Create(pngFileName)
	if err != nil {
		return err
	}
	defer pngFile.Close()
	return png.Encode(pngFile, qrCode)
}

var _ = utils.HandleError
