package valkeystore

import (
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type RawKeystoreI interface {
	Has(pubkey validatorpk.PubKey) bool
	Add(pubkey validatorpk.PubKey, key []byte, auth string) error
	Get(pubkey validatorpk.PubKey, auth string) (*encryption.PrivateKey, error)
}

type KeystoreI interface {
	RawKeystoreI
	Unlock(pubkey validatorpk.PubKey, auth string) error
	Unlocked(pubkey validatorpk.PubKey) bool
	GetUnlocked(pubkey validatorpk.PubKey) (*encryption.PrivateKey, error)
}
