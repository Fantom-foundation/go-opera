package internal

import (
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/poslachesis"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func NewIntegration(cfg *adapters.NodeConfig, net *lachesis.Net) *gossip.Service {
	makeDb := dbProducer(cfg.DataDir)
	gdb, cdb := makeStorages(makeDb)

	err := cdb.ApplyGenesis(net.Genesis, 0)
	if err != nil {
		panic(err)
	}

	c := posposet.New(cdb, gdb)

	return gossip.NewService(gdb, c)
}
