package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
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

// WriteTo writes key to writer.
func (key *PrivateKey) WriteTo(w io.Writer) error {
	b, err := x509.MarshalECPrivateKey((*ecdsa.PrivateKey)(key))
	if err != nil {
		return err
	}

	pemBlock := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: b,
	}

	if err := pem.Encode(w, &pemBlock); err != nil {
		return err
	}

	return nil
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

// Base64 encodes public key to base64.
func (pub *PublicKey) Base64() string {
	buf := pub.Bytes()
	return base64.StdEncoding.EncodeToString(buf)
}

// BytesToPubkey decodes public key from bytes.
func BytesToPubkey(pub []byte) *PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	return &PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

// Base64ToPubkey decodes public key from base64.
func Base64ToPubkey(s string) (*PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	key := BytesToPubkey(buf)
	if key == nil {
		return nil, errors.New("pubkey is invalid")
	}

	return key, nil
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

// StringToPubkey decode public key from base64 to common.PublicKey
// NOTE: deprecated
func StringToPubkey(pub string) (*PublicKey, error) {
	bb, err := base64.StdEncoding.DecodeString(pub)
	if err != nil {
		return nil, err
	}

	key := BytesToPubkey(bb)
	if key == nil {
		return nil, errors.New("pubkey is invalid")
	}

	return key, nil
}
