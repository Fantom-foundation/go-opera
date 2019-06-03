package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// ReadPrivateKey reads private key from reader.
func ReadPrivateKey(r io.Reader) (*common.PrivateKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("error decoding PEM block from data")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return (*common.PrivateKey)(key), nil
}

// GeneratePrivateKey generates new private key.
func GeneratePrivateKey() (*common.PrivateKey, error) {
	key, err := GenerateECDSAKey()
	if err != nil {
		return nil, err
	}

	return (*common.PrivateKey)(key), nil
}
