package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// GenerateFakeKey creates fake private key.
func GenerateFakeKey(n uint64) *common.PrivateKey {
	reader := rand.New(rand.NewSource(int64(n)))
	key, err := ecdsa.GenerateKey(elliptic.P256(), reader)
	if err != nil {
		panic(err)
	}
	return (*common.PrivateKey)(key)
}
