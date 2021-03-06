package valkeystore

import (
	"errors"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type MemKeystore struct {
	mem  map[string]*encryption.PrivateKey
	auth map[string]string
}

func NewMemKeystore() *MemKeystore {
	return &MemKeystore{
		mem:  make(map[string]*encryption.PrivateKey),
		auth: make(map[string]string),
	}
}

func (m *MemKeystore) Has(pubkey validatorpk.PubKey) bool {
	_, ok := m.mem[m.idxOf(pubkey)]
	return ok
}

func (m *MemKeystore) Add(pubkey validatorpk.PubKey, key []byte, auth string) error {
	if m.Has(pubkey) {
		return ErrAlreadyExists
	}
	if pubkey.Type != validatorpk.Types.Secp256k1 {
		return encryption.ErrNotSupportedType
	}
	decoded, err := crypto.ToECDSA(key)
	if err != nil {
		return err
	}
	m.mem[m.idxOf(pubkey)] = &encryption.PrivateKey{
		Type:    pubkey.Type,
		Bytes:   key,
		Decoded: decoded,
	}
	m.auth[m.idxOf(pubkey)] = auth
	return nil
}

func (m *MemKeystore) Get(pubkey validatorpk.PubKey, auth string) (*encryption.PrivateKey, error) {
	if !m.Has(pubkey) {
		return nil, ErrNotFound
	}
	if m.auth[m.idxOf(pubkey)] != auth {
		return nil, errors.New("could not decrypt key with given password")
	}
	return m.mem[m.idxOf(pubkey)], nil
}

func (m *MemKeystore) idxOf(pubkey validatorpk.PubKey) string {
	return string(pubkey.Bytes())
}
