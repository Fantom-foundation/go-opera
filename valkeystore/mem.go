package valkeystore

import (
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type MemKeystore struct {
	mem map[string]*encryption.PrivateKey
}

func NewMemKeystore() *MemKeystore {
	return &MemKeystore{
		mem: make(map[string]*encryption.PrivateKey),
	}
}

func (m *MemKeystore) Has(pubkey validator.PubKey) bool {
	_, ok := m.mem[m.idxOf(pubkey)]
	return ok
}

func (m *MemKeystore) Add(pubkey validator.PubKey, key []byte, _ string) error {
	if pubkey.Type != "secp256k1" {
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
	return nil
}

func (m *MemKeystore) Get(pubkey validator.PubKey, auth string) (*encryption.PrivateKey, error) {
	if !m.Has(pubkey) {
		return nil, NotFound
	}
	return m.mem[m.idxOf(pubkey)], nil
}

func (m *MemKeystore) idxOf(pubkey validator.PubKey) string {
	return string(pubkey.Raw) + pubkey.Type
}
