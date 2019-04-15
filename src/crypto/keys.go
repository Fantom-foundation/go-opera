package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// GenerateKey creates new private key.
func GenerateKey() *common.PrivateKey {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	return (*common.PrivateKey)(key)
}

// GenerateECDSAKey generate ECDSA Key
func GenerateECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// PemToKey parses key from PEM.
func PemToKey(b []byte) (*common.PrivateKey, error) {
	if len(b) == 0 {
		return nil, nil
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("error decoding PEM block from data")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	return (*common.PrivateKey)(key), err
}

// KeyToPem encodes key to PEM.
func KeyToPem(key *common.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey((*ecdsa.PrivateKey)(key))
	if err != nil {
		return nil, err
	}

	block := &pem.Block{Type: "PRIVATE KEY", Bytes: b}

	return pem.EncodeToMemory(block), nil
}
