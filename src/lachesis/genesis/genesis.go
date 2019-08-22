package genesis

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
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
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
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
func FakeGenesis(n int) (Genesis, []hash.Peer, []*crypto.PrivateKey) {
	balances := make(map[hash.Peer]pos.Stake, n)
	keys := make([]*crypto.PrivateKey, n)
	ids := make([]hash.Peer, n)
	for i := 0; i < n; i++ {
		keys[i] = crypto.GenerateFakeKey(i)
		ids[i] = cryptoaddr.AddressOf(keys[i].Public())
		balances[ids[i]] = 1000000000
	}

	return Genesis{
		//Alloc:  balances,
		Time: genesisTestTime,
	}, ids, keys
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
