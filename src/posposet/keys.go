package posposet

import (
	"crypto/ecdsa"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

type (
	// PrivateKey is a private key wrapper.
	PrivateKey struct {
		ecdsa.PrivateKey
	}

	// PublicKey is a public key wrapper.
	PublicKey struct {
		ecdsa.PublicKey
	}
)

// GenerateKey creates new private key.
func GenerateKey() PrivateKey {
	key, err := crypto.GenerateECDSAKey()
	if err != nil {
		panic(err)
	}
	return PrivateKey{*key}
}

// PublicKey returns public part of key.
func (key *PrivateKey) PublicKey() PublicKey {
	return PublicKey{key.PrivateKey.PublicKey}
}

// Bytes returns public key bytes.
func (pk *PublicKey) Bytes() []byte {
	return crypto.FromECDSAPub(&pk.PublicKey)
}
