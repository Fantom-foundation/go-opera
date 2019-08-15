package internal

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// FakeNet generates fake net with n-nodes genesis.
func FakeNet(n int) (*genesis.Config, []*crypto.PrivateKey) {
	balances := make(map[hash.Peer]pos.Stake, n)
	keys := make([]*crypto.PrivateKey, n)
	for i := 0; i < n; i++ {
		keys[i] = crypto.GenerateFakeKey(i)
		id := cryptoaddr.AddressOf(keys[i].Public())
		balances[id] = 1000000000
	}

	return &genesis.Config{
		Balances: balances,
		Time:     genesisTestTime,
	}, keys
}

// MainNet returns builtin genesis keys of mainnet.
func MainNet() *genesis.Config {
	return &genesis.Config{
		Balances: map[hash.Peer]pos.Stake{
			// TODO: fill with official keys and balances.
		},
		Time: genesisTestTime,
	}
}

// TestNet returns builtin genesis keys of testnet.
func TestNet() *genesis.Config {
	return &genesis.Config{
		Balances: map[hash.Peer]pos.Stake{
			// TODO: fill with official keys and balances.
		},
		Time: genesisTestTime,
	}
}
