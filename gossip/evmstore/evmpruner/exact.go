package evmpruner

import (
	"errors"
	"io"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type exactSetStore struct {
	db kvdb.Store
}

func NewLevelDBSet(name string) (*exactSetStore, io.Closer, error) {
	db, err := leveldb.New(name, 256*opt.MiB, 0, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return &exactSetStore{db}, db, nil
}

func (set *exactSetStore) Put(key []byte, _ []byte) error {
	// If the key length is not 32bytes, ensure it's contract code
	// entry with new scheme.
	if len(key) != common.HashLength {
		isCode, codeKey := rawdb.IsCodeKey(key)
		if !isCode {
			return errors.New("invalid entry")
		}
		return set.db.Put(codeKey, []byte{})
	}
	return set.db.Put(key, []byte{})
}

func (set *exactSetStore) Delete(key []byte) error { panic("not supported") }

func (set *exactSetStore) Contain(key []byte) (bool, error) {
	return set.db.Has(key)
}

func (set *exactSetStore) Commit(filename, tempname string) error {
	// No need in manual writing
	return nil
}
