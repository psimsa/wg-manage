package wg

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GetKeyPair() (string, string) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		panic(fmt.Errorf("cannot generate wireguard key: %w", err))
	}
	return key.String(), key.PublicKey().String()
}

func GetPresharedKey() string {
	key, err := wgtypes.GenerateKey()
	if err != nil {
		panic(fmt.Errorf("cannot generate preshared key: %w", err))
	}
	return key.String()
}
