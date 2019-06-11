package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
)

type (
	// PrivateKey is a private key wrapper.
	PrivateKey ecdsa.PrivateKey

	// PublicKey is a public key wrapper.
	PublicKey ecdsa.PublicKey
)

// GenerateKey creates new private key.
func GenerateKey() (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return (*PrivateKey)(key), nil
}

// Public returns public part of key.
func (key *PrivateKey) Public() *PublicKey {
	return (*PublicKey)(&key.PublicKey)
}

// Sign signs with key.
func (key *PrivateKey) Sign(hash []byte) (r, s *big.Int, err error) {
	return ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(key), hash)
}

// WriteTo writes key to writer in PEM.
func (key *PrivateKey) WriteTo(w io.Writer) error {
	block, err := keyToPemBlock(key)
	if err != nil {
		return err
	}

	return pem.Encode(w, block)
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

// BytesToPubKey decodes public key from bytes.
func BytesToPubKey(pub []byte) *PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pub)
	return &PublicKey{Curve: elliptic.P256(), X: x, Y: y}
}

// Base64ToPubKey decodes public key from base64.
func Base64ToPubKey(s string) (*PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	key := BytesToPubKey(buf)
	if key == nil {
		return nil, errors.New("pubkey is invalid")
	}

	return key, nil
}

// ReadPemToKey reads PEM from reader and parses key.
func ReadPemToKey(r io.Reader) (*PrivateKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return PemToKey(data)
}

// PemToKey parses key from PEM.
func PemToKey(b []byte) (*PrivateKey, error) {
	if len(b) == 0 {
		return nil, nil
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("error decoding PEM block from data")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	return (*PrivateKey)(key), err
}

// KeyToPem encodes key to PEM.
func KeyToPem(key *PrivateKey) ([]byte, error) {
	block, err := keyToPemBlock(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(block), nil
}

func keyToPemBlock(key *PrivateKey) (*pem.Block, error) {
	b, err := x509.MarshalECPrivateKey((*ecdsa.PrivateKey)(key))
	if err != nil {
		return nil, err
	}

	block := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}

	return &block, nil
}
