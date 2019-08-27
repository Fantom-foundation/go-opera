package genesis

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

var (
	genesisTestTime = inter.Timestamp(1565000000 * time.Second)
)

type Genesis struct {
	Alloc     Accounts
	Time      inter.Timestamp
	ExtraData []byte
}

// Accounts specifies the initial state that is part of the genesis block.
type Accounts map[common.Address]Account

// Account is an account in the state of the genesis block.
type Account struct {
	Code       []byte                      `json:"code,omitempty"`
	Storage    map[common.Hash]common.Hash `json:"storage,omitempty"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce,omitempty"`
	PrivateKey *ecdsa.PrivateKey
}

func (ga Accounts) Addresses() []common.Address {
	res := make([]common.Address, 0, len(ga))
	for addr, _ := range ga {
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

// FakeGenesis generates fake genesis with n-nodes.
func FakeGenesis(n int) Genesis {
	accounts := make(Accounts, n)

	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accounts[addr] = Account{Balance: pos.StakeToBalance(1000000), PrivateKey: key}
	}

	return Genesis{
		Alloc: accounts,
		Time:  genesisTestTime,
	}
}

// MainNet returns builtin genesis keys of mainnet.
func MainGenesis() Genesis {
	return Genesis{
		Time: genesisTestTime,
		/*Alloc: map[hash.Peer]pos.Stake{
			// TODO: fill with official keys and balances.
		},*/
	}
}

// TestGenesis returns builtin genesis keys of testnet.
func TestGenesis() Genesis {
	return Genesis{
		Time: genesisTestTime,
		/*Alloc: map[hash.Peer]pos.Stake{
			// TODO: fill with official keys and balances.
		},*/
	}
}
