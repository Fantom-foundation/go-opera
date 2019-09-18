package internal

import (
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

func NewIntegration(ctx *adapters.ServiceContext, network lachesis.Config) *gossip.Service {
	gossipCfg := gossip.DefaultConfig(network)

	engine, gdb := MakeEngine(ctx.Config.DataDir, &gossipCfg)

	svc, err := gossip.NewService(ctx.NodeContext, gossipCfg, gdb, engine)
	if err != nil {
		panic(err)
	}

	return svc
}
