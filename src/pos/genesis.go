package pos

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// FakeGenesis is a stub
func FakeGenesis(participants *peers.Peers, conf *Config, db state.Database) (hash.Hash, error) {
	if conf == nil {
		conf = DefaultConfig()
	}

	balance := conf.TotalSupply / uint64(participants.Len())

	statedb, _ := state.New(hash.Hash{}, db)

	for _, p := range participants.ToPeerSlice() {
		addr := hash.Address(p.Address())
		statedb.AddBalance(addr, balance)
	}
	return statedb.Commit(true)
}
