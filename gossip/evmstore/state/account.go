package state

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash // merkle root of the storage trie
	CodeHash []byte
}

func newAccount() *Account {
	account := new(Account)
	account.Nonce = 0
	account.Balance = big.NewInt(0)
	account.Root = emptyRoot
	account.CodeHash = emptyCodeHash
	return account
}

func (a *Account) Copy(image *Account) {
	a.Nonce = image.Nonce
	a.Balance = big.NewInt(0).Set(image.Balance)
	copy(a.Root[:], image.Root[:])
	copy(a.CodeHash[:], image.CodeHash[:])
	//a.Incarnation = image.Incarnation
}

func (a *Account) IsEmptyCodeHash() bool {
	return IsEmptyCodeHash(a.CodeHash)
}

func IsEmptyCodeHash(codeHash []byte) bool {
	return len(codeHash) == 0 && bytes.Equal(codeHash, emptyCodeHash)
}

func (a *Account) IsEmptyRoot() bool {
	return a.Root == emptyRoot || a.Root == common.Hash{}
}
