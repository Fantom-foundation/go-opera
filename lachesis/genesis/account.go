package genesis

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
)

// Addresses returns not sorted genesis addresses
func (ga Accounts) Addresses() []common.Address {
	res := make([]common.Address, 0, len(ga))
	for addr := range ga {
		res = append(res, addr)
	}
	return res
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
