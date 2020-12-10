package valkeystore

import (
	"errors"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

var (
	ErrAlreadyUnlocked = errors.New("already unlocked")
	ErrLocked          = errors.New("key is locked")
)

type CachedKeystore struct {
	backend RawKeystoreI
	cache   map[string]*encryption.PrivateKey
}

func NewCachedKeystore(backend RawKeystoreI) *CachedKeystore {
	return &CachedKeystore{
		backend: backend,
		cache:   make(map[string]*encryption.PrivateKey),
	}
}

func (c *CachedKeystore) Unlocked(pubkey validatorpk.PubKey) bool {
	_, ok := c.cache[c.idxOf(pubkey)]
	return ok
}

func (c *CachedKeystore) Has(pubkey validatorpk.PubKey) bool {
	if c.Unlocked(pubkey) {
		return true
	}
	return c.backend.Has(pubkey)
}

func (c *CachedKeystore) Unlock(pubkey validatorpk.PubKey, auth string) error {
	if c.Unlocked(pubkey) {
		return ErrAlreadyUnlocked
	}
	key, err := c.backend.Get(pubkey, auth)
	if err != nil {
		return err
	}
	c.cache[c.idxOf(pubkey)] = key
	return nil
}

func (c *CachedKeystore) GetUnlocked(pubkey validatorpk.PubKey) (*encryption.PrivateKey, error) {
	if !c.Unlocked(pubkey) {
		return nil, ErrLocked
	}
	return c.cache[c.idxOf(pubkey)], nil
}

func (c *CachedKeystore) idxOf(pubkey validatorpk.PubKey) string {
	return string(pubkey.Bytes())
}

func (c *CachedKeystore) Add(pubkey validatorpk.PubKey, key []byte, auth string) error {
	return c.backend.Add(pubkey, key, auth)
}

func (c *CachedKeystore) Get(pubkey validatorpk.PubKey, auth string) (*encryption.PrivateKey, error) {
	return c.backend.Get(pubkey, auth)
}
