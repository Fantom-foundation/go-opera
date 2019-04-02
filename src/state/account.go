package state

//go:generate protoc --go_out=plugins=grpc:./ account.proto

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Account is the PoS representation of accounts.
// These objects are stored in the main account trie.

// Root converts bytes to hash.
func (a *Account) Root() (h hash.Hash) {
	h.SetBytes(a.RawRoot)
	return
}

// SetRoot converts hash to bytes.
func (a *Account) SetRoot(h hash.Hash) {
	a.RawRoot = h.Bytes()
}
