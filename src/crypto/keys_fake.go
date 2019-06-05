package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/rand"
)

// GenerateFakeKey creates fake private key.
func GenerateFakeKey(n int) *PrivateKey {
	reader := rand.New(rand.NewSource(int64(n)))
	key, err := ecdsa.GenerateKey(elliptic.P256(), reader)
	if err != nil {
		panic(err)
	}
	return (*PrivateKey)(key)
}
