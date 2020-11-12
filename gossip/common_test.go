package gossip

import (
	"testing"

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

	return &testEnv{
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
}

func (env *testEnv) Block() {
	
}

func TestEnv(t *testing.T) {
	_ = newTestEnv()
}
