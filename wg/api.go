package wg

func KeyPair() (priv, pub string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errRecover(r)
		}
	}()
	priv, pub = GetKeyPair()
	return priv, pub, nil
}

func PresharedKey() (string, error) {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()
	return GetPresharedKey(), nil
}

type recoverError struct{ v interface{} }

func (e recoverError) Error() string { return "key generation failed" }

func errRecover(v interface{}) error { return recoverError{v} }
