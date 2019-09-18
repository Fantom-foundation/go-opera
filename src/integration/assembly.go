package integration

import (
	"github.com/ethereum/go-ethereum/cmd/utils"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

// MakeEngine makes consensus engine from config.
func MakeEngine(dataDir string, gossipCfg *gossip.Config) (*poset.Poset, *gossip.Store) {
	makeDb := dbProducer(dataDir)
	gdb, cdb := makeStorages(makeDb)

	// write genesis

	genesisAtropos, genesisState, err := gdb.ApplyGenesis(&gossipCfg.Net)
	if err != nil {
		utils.Fatalf("Failed to write EVM genesis state: %v", err)
	}

	err = cdb.ApplyGenesis(&gossipCfg.Net.Genesis, genesisAtropos, genesisState)
	if err != nil {
		utils.Fatalf("Failed to write Poset genesis state: %v", err)
	}

	// create consensus
	engine := poset.New(gossipCfg.Net.Dag, cdb, gdb)

	return engine, gdb
}
