package internal

import (
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/inter/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

func NewIntegration(cfg *adapters.NodeConfig, gen *genesis.Config) *gossip.Service {
	makeDb := dbProducer(cfg.DataDir)
	gdb, cdb := makeStorages(makeDb)

	err := cdb.ApplyGenesis(gen)
	if err != nil {
		panic(err)
	}

	c := posposet.New(cdb, gdb)

	g, err := gossip.NewService(&gossip.DefaultConfig, new(event.TypeMux), gdb, c)
	if err != nil {
		panic(err)
	}

	return g
}
