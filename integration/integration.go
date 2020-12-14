package integration

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/simulations/adapters"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

// NewIntegration creates gossip service for the integration test
func NewIntegration(ctx *adapters.ServiceContext, genesis opera.Genesis, stack *node.Node) *gossip.Service {
	gossipCfg := gossip.FakeConfig(len(genesis.State.Validators))
	cfg := Configs{
		Opera:         gossipCfg,
		OperaStore:    gossip.DefaultStoreConfig(),
		Lachesis:      abft.DefaultConfig(),
		LachesisStore: abft.DefaultStoreConfig(),
		VectorClock:   vecmt.DefaultConfig(),
	}

	dbs := gossip.NewSyncedPool(DBProducer(ctx.Config.DataDir))
	engine, dagIndex, gdb, blockProc := MakeEngine(dbs, cfg, genesis)

	valKeystore := valkeystore.NewDefaultMemKeystore()

	pubKey := validator.PubKey{
		Raw:  crypto.FromECDSAPub(&ctx.Config.PrivateKey.PublicKey),
		Type: "secp256k1",
	}

	// unlock the key
	_ = valKeystore.Add(pubKey, crypto.FromECDSA(ctx.Config.PrivateKey), validator.FakePassword)
	_ = valKeystore.Unlock(pubKey, validator.FakePassword)
	signer := valkeystore.NewSigner(valKeystore)

	// find a genesis validator which corresponds to the key
	for _, v := range genesis.State.Validators {
		if v.PubKey.String() == pubKey.String() {
			gossipCfg.Emitter.Validator.ID = v.ID
			gossipCfg.Emitter.Validator.PubKey = v.PubKey
		}
	}

	gossipCfg.Emitter.EmitIntervals.Max = 3 * time.Second
	gossipCfg.Emitter.EmitIntervals.DoublesignProtection = 0

	svc, err := gossip.NewService(stack, gossipCfg, genesis.Rules, gdb, signer, blockProc, engine, dagIndex)
	if err != nil {
		panic(err)
	}
	err = engine.Bootstrap(svc.GetConsensusCallbacks())
	if err != nil {
		return nil
	}

	return svc
}
