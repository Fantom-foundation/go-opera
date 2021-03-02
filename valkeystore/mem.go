package valkeystore

import (
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
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

func (m *MemKeystore) Has(pubkey validatorpk.PubKey) bool {
	_, ok := m.mem[m.idxOf(pubkey)]
	return ok
}

func (m *MemKeystore) Add(pubkey validatorpk.PubKey, key []byte, _ string) error {
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
	return nil
}

func (m *MemKeystore) Get(pubkey validatorpk.PubKey, _ string) (*encryption.PrivateKey, error) {
	if !m.Has(pubkey) {
		return nil, ErrNotFound
	}
	return m.mem[m.idxOf(pubkey)], nil
}

func (m *MemKeystore) idxOf(pubkey validatorpk.PubKey) string {
	return string(pubkey.Bytes())
}
