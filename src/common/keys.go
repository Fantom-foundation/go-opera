package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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

// PublicKey returns public part of key.
func (key *PrivateKey) PublicKey() PublicKey {
	return PublicKey{key.PrivateKey.PublicKey}
}

// Bytes returns public key bytes.
func (pk *PublicKey) Bytes() []byte {
	return FromECDSAPub(&pk.PublicKey)
}

/*
 * Utils:
 */

// ToECDSAPub convert to ECDSA public key from bytes
func ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

// FromECDSAPub create bytes from ECDSA public key
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}
