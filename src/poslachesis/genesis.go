package lachesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// Net describes lachesis net.
type Net struct {
	Name    string
	Genesis map[hash.Peer]inter.Stake
}

// FakeNet generates fake net with n-nodes genesis.
func FakeNet(n int) (*Net, []*crypto.PrivateKey) {
	genesis := make(map[hash.Peer]inter.Stake, n)
	keys := make([]*crypto.PrivateKey, n)
	for i := 0; i < n; i++ {
		keys[i] = crypto.GenerateFakeKey(i)
		id := cryptoaddr.AddressOf(keys[i].Public())
		genesis[id] = 1000000000
	}

	return &Net{
		Name:    "fake",
		Genesis: genesis,
	}, keys
}

// MainNet returns builtin genesis keys of mainnet.
func MainNet() *Net {
	return &Net{
		Name: "main",
		Genesis: map[hash.Peer]inter.Stake{
			// TODO: fill with official keys and balances.
		},
	}
}

// TestNet returns builtin genesis keys of testnet.
func TestNet() *Net {
	return &Net{
		Name: "test",
		Genesis: map[hash.Peer]inter.Stake{
			// TODO: fill with official keys and balances.
		},
	}
}
