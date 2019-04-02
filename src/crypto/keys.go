package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// GenerateKey creates new private key.
func GenerateKey() common.PrivateKey {
	key, err := GenerateECDSAKey()
	if err != nil {
		panic(err)
	}
	return common.PrivateKey{
		PrivateKey: *key,
	}
}

// GenerateECDSAKey generate ECDSA Key
func GenerateECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
