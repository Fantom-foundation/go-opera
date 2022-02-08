package valkeystore

import (
	"errors"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/valkeystore/encryption"
)

var (
	ErrNotFound      = errors.New("key is not found")
	ErrAlreadyExists = errors.New("key already exists")
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

func (f *FileKeystore) Has(pubkey validatorpk.PubKey) bool {
	return fileExists(f.PathOf(pubkey))
}

func (f *FileKeystore) Add(pubkey validatorpk.PubKey, key []byte, auth string) error {
	if f.Has(pubkey) {
		return ErrAlreadyExists
	}
	return f.enc.StoreKey(f.PathOf(pubkey), pubkey, key, auth)
}

func (f *FileKeystore) Get(pubkey validatorpk.PubKey, auth string) (*encryption.PrivateKey, error) {
	if !f.Has(pubkey) {
		return nil, ErrNotFound
	}
	return f.enc.ReadKey(pubkey, f.PathOf(pubkey), auth)
}

func (f *FileKeystore) PathOf(pubkey validatorpk.PubKey) string {
	return path.Join(f.dir, common.Bytes2Hex(pubkey.Bytes()))
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
