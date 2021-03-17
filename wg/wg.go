package wg

import (
	"github.com/ofcoursedude/wg-manage/oscustom"
	"os/exec"
	"strings"

	"github.com/ofcoursedude/wg-manage/utils"
)

func GetKeyPair() (string, string) {
	privCmd := exec.Command("wg", "genkey")
	privCmd.Wait()
	privOutput, err := privCmd.Output()
	utils.HandleError(err, "Cannot generate private key")

	privKey := strings.Trim(string(privOutput), oscustom.NewLine)

	pubCmd := exec.Command("wg", "pubkey")
	pubCmdStdInWriter, err := pubCmd.StdinPipe()
	utils.HandleError(err, "Cannot write private key to stdin")

	pubCmdStdInWriter.Write(privOutput)
	pubCmdStdInWriter.Close()
	pubCmd.Wait()
	pubOutput, err := pubCmd.Output()
	utils.HandleError(err, "Cannot generate public key")

	pubKey := strings.Trim(string(pubOutput), oscustom.NewLine)
	return privKey, pubKey
}

func GetPresharedKey() string {
	preshCmd := exec.Command("wg", "genpsk")
	preshCmd.Wait()
	preshOutput, err := preshCmd.Output()
	utils.HandleError(err, "Cannot generate preshared key")
	preshKey := strings.Trim(string(preshOutput), oscustom.NewLine)
	return preshKey
}
