package nokeyiserr

import (
	"errors"
	"github.com/ethereum/go-ethereum/ethdb"
)

var (
	errNotFound = errors.New("not found")
)

type Wrapper struct {
	ethdb.KeyValueStore
}

// Wrap creates new Wrapper
func Wrap(db ethdb.KeyValueStore) *Wrapper {
	return &Wrapper{db}
}

// Get implements ETH-style Get. ETH databases expect an error if key not found, unlike Lachesis
func (w *Wrapper) Get(key []byte) ([]byte, error) {
	val, err := w.KeyValueStore.Get(key)
	if val == nil && err == nil {
		return nil, errNotFound
	}
	return val, err
}
