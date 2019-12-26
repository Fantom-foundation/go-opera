package genesis

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

// Accounts specifies the initial state that is part of the genesis block.
type (
	Accounts map[common.Address]Account

	// Account is an account in the state of the genesis block.
	Account struct {
		Code       []byte                      `json:"code,omitempty"`
		Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
		Balance    *big.Int                    `json:"balance" gencodec:"required"`
		Nonce      uint64                      `json:"nonce,omitempty"`
		PrivateKey *ecdsa.PrivateKey           `toml:"-"`
	}
	storageElement struct {
		Key   common.Hash
		Value common.Hash
	}
	account struct {
		Code    []byte
		Storage []storageElement
		Balance *big.Int
		Nonce   uint64
	}
	accountAndAddr struct {
		Acc  account
		Addr common.Address
	}

	VAccounts struct {
		Accounts         Accounts
		Validators       pos.GValidators
		SfcContractAdmin common.Address
	}
)

// Cheaters is a slice type for storing cheaters list.
type accountsArray []accountAndAddr

// Len returns the length of s.
func (s accountsArray) Len() int { return len(s) }

// Less compares elements
func (s accountsArray) Less(i, j int) bool {
	return bytes.Compare(s[i].Addr.Bytes(), s[j].Addr.Bytes()) < 0
}

// Swap swaps the i'th and the j'th element in s.
func (s accountsArray) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s accountsArray) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

func (a Account) sorted() account {
	sortedStorage := make([]storageElement, 0, len(a.Storage))
	for key, val := range a.Storage {
		sortedStorage = append(sortedStorage, storageElement{
			Key:   key,
			Value: val,
		})
	}
	sort.Slice(sortedStorage, func(i, j int) bool {
		return bytes.Compare(sortedStorage[i].Key.Bytes(), sortedStorage[j].Key.Bytes()) < 0
	})

	return account{
		Code:    a.Code,
		Storage: sortedStorage,
		Balance: a.Balance,
		Nonce:   a.Nonce,
	}
}

// Addresses returns not sorted genesis addresses
func (ga Accounts) Addresses() []common.Address {
	res := make([]common.Address, 0, len(ga))
	for addr := range ga {
		res = append(res, addr)
	}
	return res
}

// sortedAccounts returns sorted genesis accounts
func (ga Accounts) sortedAccounts() accountsArray {
	res := make(accountsArray, 0, len(ga))
	for addr, acc := range ga {
		res = append(res, accountAndAddr{
			Acc:  acc.sorted(),
			Addr: addr,
		})
	}

	sort.Sort(res)

	return res
}

// Hash returns accounts hash
func (ga Accounts) Hash() common.Hash {
	return types.DeriveSha(ga.sortedAccounts())
}

func (ga *Accounts) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]Account)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(Accounts)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

func (ga Accounts) Add(gb Accounts) {
	for addr, acc := range gb {
		ga[addr] = acc
	}
}
