package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/src/utils"
)

type (
	// PrivateKey is a private key wrapper.
	PrivateKey ecdsa.PrivateKey

	// PublicKey is a public key wrapper.
	PublicKey ecdsa.PublicKey
)

// GenerateKey creates new private key.
func GenerateKey() (*PrivateKey, error) {
	key, err := ecdsa.GenerateKey(S256(), rand.Reader)
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
func (key *PrivateKey) Sign(hash []byte) ([]byte, error) {
	return Sign(hash, key)
}

// SignRaw signs with key. Unrecoverable.
// NOTE: deprecated.
func (key *PrivateKey) SignRaw(hash []byte) (r, s *big.Int, err error) {
	return ecdsa.Sign(rand.Reader, (*ecdsa.PrivateKey)(key), hash)
}

// WriteTo writes key to writer in ETH format.
func (key *PrivateKey) WriteTo(w io.Writer) error {
	_, err := w.Write(key.Bytes())
	return err
}

// Bytes (ETH format) exports a private key into a binary dump.
func (key *PrivateKey) Bytes() []byte {
	if key == nil {
		return nil
	}
	return utils.PaddedBigBytes(key.D, key.Params().BitSize/8)
}

// VerifyRaw verifies the signatures.
// NOTE: deprecated.
func (pub *PublicKey) VerifyRaw(hash []byte, r, s *big.Int) bool {
	return ecdsa.Verify((*ecdsa.PublicKey)(pub), hash, r, s)
}

// Bytes encodes public key to bytes.
func (pub *PublicKey) Bytes() []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
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
	x, y := elliptic.Unmarshal(S256(), pub)
	return &PublicKey{Curve: S256(), X: x, Y: y}
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

// ReadKey reads from reader and parses key.
func ReadKey(r io.Reader) (*PrivateKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return new(PrivateKey).SetBytes(data, true)
}

// SetBytes (ETH format) creates a private key with the given D value. The strict parameter
// controls whether the key's length should be enforced at the curve size or
// it can also accept legacy encodings (0 prefixes).
func (key *PrivateKey) SetBytes(d []byte, strict bool) (*PrivateKey, error) {
	key.PublicKey.Curve = S256()
	if strict && 8*len(d) != key.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", key.Params().BitSize)
	}
	key.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if key.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if key.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	key.PublicKey.X, key.PublicKey.Y = key.PublicKey.Curve.ScalarBaseMult(d)
	if key.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return key, nil
}
