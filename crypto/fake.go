package crypto

import (
	"crypto/ecdsa"
	"math/rand"

	"github.com/ethereum/go-ethereum/crypto"
)

// FakeKey gets n-th fake private key.
func FakeKey(n int) *ecdsa.PrivateKey {
	reader := rand.New(rand.NewSource(int64(n)))

	key, err := ecdsa.GenerateKey(crypto.S256(), reader)
	if err != nil {
		panic(err)
	}

	return key
}
