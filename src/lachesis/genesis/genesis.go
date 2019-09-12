package genesis

import (
	"time"

	eth "github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
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

// FakeGenesis generates fake genesis with n-nodes.
func FakeGenesis(n int) Genesis {
	accounts := make(Accounts, n)

	for i := 0; i < n; i++ {
		key := crypto.FakeKey(i)
		addr := eth.PubkeyToAddress(key.PublicKey)
		accounts[addr] = Account{
			Balance:    pos.StakeToBalance(1000000),
			PrivateKey: key,
		}
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
