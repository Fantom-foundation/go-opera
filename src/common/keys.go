package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
)

type (
	// PrivateKey is a private key wrapper.
	PrivateKey ecdsa.PrivateKey

	// PublicKey is a public key wrapper.
	PublicKey ecdsa.PublicKey
)

// Public returns public part of key.
func (key *PrivateKey) Public() *PublicKey {
	return (*PublicKey)(&key.PublicKey)
}

// Sign signs with key.
func (key *PrivateKey) Sign(hash []byte) (r, s *big.Int, err error) {
	return ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(key), hash)
}

// Verify verifies the signatures.
func (pub *PublicKey) Verify(hash []byte, r, s *big.Int) bool {
	return ecdsa.Verify((*ecdsa.PublicKey)(pub), hash, r, s)
}

// Bytes encodes public key to bytes.
func (pub *PublicKey) Bytes() []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}

// BytesToPubkey decodes public key from bytes.
func BytesToPubkey(pub []byte) *PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	return &PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

/*
 * Utils:
 */

// ToECDSAPub convert to ECDSA public key from bytes.
// NOTE: deprecated
func ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

// FromECDSAPub create bytes from ECDSA public key.
// NOTE: deprecated
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}
