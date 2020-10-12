package valkeystore

import (
	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type RawKeystoreI interface {
	Has(pubkey validator.PubKey) bool
	Add(pubkey validator.PubKey, key []byte, auth string) error
	Get(pubkey validator.PubKey, auth string) (*encryption.PrivateKey, error)
}

type KeystoreI interface {
	RawKeystoreI
	Unlock(pubkey validator.PubKey, auth string) error
	Unlocked(pubkey validator.PubKey) bool
	GetUnlocked(pubkey validator.PubKey) (*encryption.PrivateKey, error)
}
