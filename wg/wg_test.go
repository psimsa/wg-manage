package wg

import (
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestGetKeyPair(t *testing.T) {
	priv, pub := GetKeyPair()
	key, err := wgtypes.ParseKey(priv)
	if err != nil {
		t.Fatalf("private key not parseable: %v", err)
	}
	if key.PublicKey().String() != pub {
		t.Errorf("public key %q does not match private key (derived %q)", pub, key.PublicKey().String())
	}
}

func TestGetKeyPairUnique(t *testing.T) {
	a, _ := GetKeyPair()
	b, _ := GetKeyPair()
	if a == b {
		t.Error("two generated private keys are identical")
	}
}

func TestGetPresharedKey(t *testing.T) {
	psk := GetPresharedKey()
	if _, err := wgtypes.ParseKey(psk); err != nil {
		t.Fatalf("preshared key not parseable: %v", err)
	}
}
