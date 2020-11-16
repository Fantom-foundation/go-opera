package gossip

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	eth "github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc/eventmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/evmmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sealmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sfcmodule"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

const (
	maxEpochDuration = time.Hour
	sameEpoch        = maxEpochDuration / 1000
	nextEpoch        = maxEpochDuration
)

type testEnv struct {
	network   opera.Config
	blockProc BlockProc
	store     *Store

	lastBlock     idx.Block
	lastBlockTime time.Time
}

func newTestEnv() *testEnv {
	network := opera.MainNetConfig()
	network.Dag.MaxEpochDuration = maxEpochDuration

	dbs := flushable.NewSyncedPool(
		memorydb.NewProducer(""))

	env := &testEnv{
		network: network,
		blockProc: BlockProc{
			SealerModule:        sealmodule.New(network),
			TxListenerModule:    sfcmodule.NewSfcTxListenerModule(network),
			GenesisTxTransactor: sfcmodule.NewSfcTxGenesisTransactor(network),
			PreTxTransactor:     sfcmodule.NewSfcTxPreTransactor(network),
			PostTxTransactor:    sfcmodule.NewSfcTxTransactor(network),
			EventsModule:        eventmodule.New(network),
			EVMModule:           evmmodule.New(network),
		},
		store: NewStore(dbs, LiteStoreConfig()),
	}
	_, _, err := env.store.ApplyGenesis(env.blockProc, &network)
	if err != nil {
		panic(err)
	}

	return env
}

func (env *testEnv) consensusCallbackBeginBlockFn() lachesis.BeginBlockFn {
	return consensusCallbackBeginBlockFn(
		env.network,
		env.store,
		env.blockProc,
		false, nil, nil,
	)
}

func (env *testEnv) ApplyBlock(spent time.Duration, txs ...*eth.Transaction) eth.Receipts {
	env.lastBlock++
	env.lastBlockTime = env.lastBlockTime.Add(spent)

	eventBuilder := inter.MutableEventPayload{}
	eventBuilder.SetTxs(txs)
	// TODO: fill the event

	event := eventBuilder.Build()
	env.store.SetEvent(event)

	beginBlock := env.consensusCallbackBeginBlockFn()
	process := beginBlock(&lachesis.Block{
		Atropos: event.ID(),
	})

	process.ApplyEvent(event)
	_ = process.EndBlock()

	return nil
}

func TestEnv(t *testing.T) {
	_ = newTestEnv()
}
