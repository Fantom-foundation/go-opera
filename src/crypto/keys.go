package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// GenerateKey creates new private key.
func GenerateKey() *common.PrivateKey {
	key, err := GenerateECDSAKey()
	if err != nil {
		panic(err)
	}
	return (*common.PrivateKey)(key)
}

// GenerateECDSAKey generate ECDSA Key
func GenerateECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// ReadPemToKey reads PEM from reader and parses key.
func ReadPemToKey(r io.Reader) (*common.PrivateKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return PemToKey(data)
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
	block, err := keyToPemBlock(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(block), nil
}

func keyToPemBlock(key *common.PrivateKey) (*pem.Block, error) {
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

// WriteKeyTo writes key to writer in PEM.
func WriteKeyTo(w io.Writer, key *common.PrivateKey) error {
	block, err := keyToPemBlock(key)
	if err != nil {
		return err
	}

	return pem.Encode(w, block)
}
