package wg

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/ofcoursedude/wg-manage/utils"
)

func GetKeyPair() (string, string) {

	key, err := wgtypes.GeneratePrivateKey()
	utils.HandleError(err, "cannot generate wireguard key")
	return key.String(), key.PublicKey().String()
}
func GetPresharedKey() string {
	key, err := wgtypes.GenerateKey()
	utils.HandleError(err, "Cannot generate preshared key")
	return key.String()
}
