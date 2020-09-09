package integration

import (
	"time"

	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/opera"
)

// NewIntegration creates gossip service for the integration test
func NewIntegration(ctx *adapters.ServiceContext, network opera.Config) *gossip.Service {
	gossipCfg := gossip.DefaultConfig(network)

	engine, dagIndex, _, gdb := MakeEngine(ctx.Config.DataDir, &gossipCfg)

	coinbase := SetAccountKey(
		ctx.NodeContext.AccountManager,
		ctx.Config.PrivateKey,
		"fakepassword",
	)

	gossipCfg.Emitter.Validator = coinbase.Address
	gossipCfg.Emitter.EmitIntervals.Max = 3 * time.Second
	gossipCfg.Emitter.EmitIntervals.SelfForkProtection = 0

	svc, err := gossip.NewService(ctx.NodeContext, &gossipCfg, gdb, engine, dagIndex)
	if err != nil {
		panic(err)
	}
	err = engine.Bootstrap(svc.GetConsensusCallbacks())
	if err != nil {
		return nil
	}

	return svc
}
