package genesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

type Config struct {
	NetworkID uint64
	Balances  map[hash.Peer]pos.Stake
	StateHash hash.Hash
	Time      inter.Timestamp
}
