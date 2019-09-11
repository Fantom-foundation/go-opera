package internal

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

func NewIntegration(ctx *adapters.ServiceContext, network lachesis.Config) *gossip.Service {
	makeDb := dbProducer(ctx.Config.DataDir)
	gdb, cdb := makeStorages(makeDb)

	genesisFiWitness, genesisState, err := gdb.ApplyGenesis(&network)
	if err != nil {
		utils.Fatalf("Failed to write EVM genesis state: %v", err)
	}
	err = cdb.ApplyGenesis(&network.Genesis, genesisFiWitness, genesisState)
	if err != nil {
		utils.Fatalf("Failed to write Poset genesis state: %v", err)
	}

	c := poset.New(network.Dag, cdb, gdb)

	config := gossip.DefaultConfig(network)

	svc, err := gossip.NewService(ctx.NodeContext, config, gdb, c)
	if err != nil {
		panic(err)
	}

	return svc
}
