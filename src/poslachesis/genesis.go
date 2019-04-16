package lachesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// Net describes lachesis net.
type Net struct {
	Name    string
	Genesis map[hash.Peer]uint64
}

// FakeNet generates fake net with n-nodes genesis.
func FakeNet(n uint64) *Net {
	genesis := make(map[hash.Peer]uint64, n)
	for i := uint64(0); i < n; i++ {
		key := crypto.GenerateFakeKey(i)
		id := hash.PeerOfPubkey(key.Public())
		genesis[id] = 1000000000
	}

	return &Net{
		Name:    "test",
		Genesis: genesis,
	}
}

// MainNet returns built in genesis keys.
func MainNet() *Net {
	return &Net{
		Name:    "main",
		Genesis: map[hash.Peer]uint64{
			// TODO: fill with official keys and balances.
		},
	}
}
