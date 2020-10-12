package valkeystore

import (
	"sync"

	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

type SyncedKeystore struct {
	backend KeystoreI
	mu      sync.Mutex
}

func NewSyncedKeystore(backend KeystoreI) *SyncedKeystore {
	return &SyncedKeystore{
		backend: backend,
	}
}

func (s *SyncedKeystore) Unlocked(pubkey validator.PubKey) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backend.Unlocked(pubkey)
}

func (s *SyncedKeystore) Has(pubkey validator.PubKey) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backend.Has(pubkey)
}

func (s *SyncedKeystore) Unlock(pubkey validator.PubKey, auth string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backend.Unlock(pubkey, auth)
}

func (s *SyncedKeystore) GetUnlocked(pubkey validator.PubKey) (*encryption.PrivateKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backend.GetUnlocked(pubkey)
}

func (s *SyncedKeystore) Add(pubkey validator.PubKey, key []byte, auth string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backend.Add(pubkey, key, auth)
}

func (s *SyncedKeystore) Get(pubkey validator.PubKey, auth string) (*encryption.PrivateKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backend.Get(pubkey, auth)
}
