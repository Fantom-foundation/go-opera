package genesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

type Config struct {
	Name      string
	NetworkId uint64
	Balances  map[hash.Peer]pos.Stake
	StateHash hash.Hash
	Time      inter.Timestamp
}

// FakeNet generates fake net with n-nodes genesis.
func FakeNet(n int) (*Config, []hash.Peer, []*crypto.PrivateKey) {
	balances := make(map[hash.Peer]pos.Stake, n)
	keys := make([]*crypto.PrivateKey, n)
	ids := make([]hash.Peer, n)
	for i := 0; i < n; i++ {
		keys[i] = crypto.GenerateFakeKey(i)
		ids[i] = cryptoaddr.AddressOf(keys[i].Public())
		balances[ids[i]] = 1000000000
	}

	return &Config{
		Name:      "fake",
		NetworkId: 3,
		Balances:  balances,
	}, ids, keys
}

// FakeNet generates fake net with n-nodes genesis.
func EmptyFakeNet() (*Config) {
	return &Config{
		Name:      "fake",
		NetworkId: 3,
	}
}

// MainNet returns builtin genesis keys of mainnet.
func MainNet() *Config {
	return &Config{
		Name:      "main",
		NetworkId: 1,
		Balances: map[hash.Peer]pos.Stake{
			// TODO: fill with official keys and balances.
		},
	}
}

// TestNet returns builtin genesis keys of testnet.
func TestNet() *Config {
	return &Config{
		Name:      "test",
		NetworkId: 2,
		Balances: map[hash.Peer]pos.Stake{
			// TODO: fill with official keys and balances.
		},
	}
}
