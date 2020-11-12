package gossip

import (
	"testing"

	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"

	"github.com/Fantom-foundation/go-opera/gossip/blockproc/eventmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/evmmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sealmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sfcmodule"
	"github.com/Fantom-foundation/go-opera/opera"
)

type testEnv struct {
	blockProc BlockProc
}

func newTestEnv() *testEnv {
	network := opera.MainNetConfig()

	env := &testEnv{
		blockProc: BlockProc{
			SealerModule:        sealmodule.New(network),
			TxListenerModule:    sfcmodule.NewSfcTxListenerModule(network),
			GenesisTxTransactor: sfcmodule.NewSfcTxGenesisTransactor(network),
			PreTxTransactor:     sfcmodule.NewSfcTxPreTransactor(network),
			PostTxTransactor:    sfcmodule.NewSfcTxTransactor(network),
			EventsModule:        eventmodule.New(network),
			EVMModule:           evmmodule.New(network),
		},
	}

	dbs := flushable.NewSyncedPool(
		memorydb.NewProducer(""))
	gdb := NewStore(dbs, LiteStoreConfig())
	_, _, err := gdb.ApplyGenesis(env.blockProc, &network)
	if err != nil {
		panic(err)
	}

	return env
}

func (env *testEnv) Block() {

}

func TestEnv(t *testing.T) {
	_ = newTestEnv()
}
