package valkeystore

import (
	"errors"
	"os"
	"path"

	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

var (
	NotFound = errors.New("key is not found")
)

type FileKeystore struct {
	enc *encryption.Keystore
	dir string
}

func NewFileKeystore(dir string, enc *encryption.Keystore) *FileKeystore {
	return &FileKeystore{
		enc: enc,
		dir: dir,
	}
}

func (f *FileKeystore) Has(pubkey validator.PubKey) bool {
	return fileExists(f.PathOf(pubkey))
}

func (f *FileKeystore) Add(pubkey validator.PubKey, key []byte, auth string) error {
	return f.enc.StoreKey(f.PathOf(pubkey), pubkey, key, auth)
}

func (f *FileKeystore) Get(pubkey validator.PubKey, auth string) (*encryption.PrivateKey, error) {
	if !f.Has(pubkey) {
		return nil, NotFound
	}
	return f.enc.ReadKey(pubkey, f.PathOf(pubkey), auth)
}

func (f *FileKeystore) PathOf(pubkey validator.PubKey) string {
	return path.Join(f.dir, pubkey.String())
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
