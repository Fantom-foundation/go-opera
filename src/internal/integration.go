package internal

import (
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func NewIntegration(cfg *adapters.NodeConfig) *gossip.Service {
	makeDb := dbProducer(cfg.DataDir)
	gdb, cdb := makeStorages(makeDb)

	c := posposet.New(cdb, gdb)

	return gossip.NewService(gdb, c)
}
