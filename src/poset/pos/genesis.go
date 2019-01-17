package pos

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/poset/state"
)

// FakeGenesis is a stub
func FakeGenesis(participants *peers.Peers, conf *Config, db state.Database) {
	if conf == nil {
		conf = DefaultConfig()
	}

	balance := conf.TotalSupply / uint64(participants.Len())

	statedb, _ := state.New(common.Hash{}, db)

	for _, p := range participants.ToPeerSlice() {
		statedb.AddBalance(p.Address(), balance)
	}
	_, _ = statedb.Commit(true)
}
