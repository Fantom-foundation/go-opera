//+build gofuzz

package gossip

import (
	"sync"
	"testing"

	_ "github.com/dvyukov/go-fuzz/go-fuzz-defs"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/drivermodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/eventmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/evmmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sealmodule"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils"
)

const (
	fuzzHot      int = 1  // if the fuzzer should increase priority of the given input during subsequent fuzzing;
	fuzzCold     int = -1 // if the input must not be added to corpus even if gives new coverage;
	fuzzNoMatter int = 0  // otherwise.
)

func FuzzPM(data []byte) int {
	return fuzzHot
}

func TestPM(t *testing.T) {
	const (
		genesisStakers = 3
		genesisBalance = 1e18
		genesisStake   = 2 * 4e6
	)

	require := require.New(t)

	genStore := makegenesis.FakeGenesisStore(genesisStakers, utils.ToFtm(genesisBalance), utils.ToFtm(genesisStake))
	genesis := genStore.GetGenesis()
	network := genesis.Rules

	config := DefaultConfig()
	store := NewMemStore()
	blockProc := BlockProc{
		SealerModule:        sealmodule.New(),
		TxListenerModule:    drivermodule.NewDriverTxListenerModule(),
		GenesisTxTransactor: drivermodule.NewDriverTxGenesisTransactor(genesis),
		PreTxTransactor:     drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor:    drivermodule.NewDriverTxTransactor(),
		EventsModule:        eventmodule.New(),
		EVMModule:           evmmodule.New(),
	}
	_, err := store.ApplyGenesis(blockProc, genesis)
	require.NoError(err)

	var (
		heavyCheckReader    HeavyCheckReader
		gasPowerCheckReader GasPowerCheckReader
		// TODO: init
	)

	mu := new(sync.RWMutex)
	feed := new(ServiceFeed)
	checkers := makeCheckers(config.HeavyCheck, network.EvmChainConfig().ChainID, &heavyCheckReader, &gasPowerCheckReader, store)
	processEvent := func(e *inter.EventPayload) error {
		return nil
	}

	txpool := evmcore.NewTxPool(config.TxPool, network.EvmChainConfig(), &EvmStateReader{
		ServiceFeed: feed,
		store:       store,
	})

	pm, err := NewProtocolManager(
		config,
		feed,
		txpool,
		mu,
		checkers,
		store,
		processEvent,
		nil)
	require.NoError(err)

	pm.Start(3)
	defer pm.Stop()

	p := p2p.NewPeer(enode.RandomID(enode.ID{}, 1), "fake-node-1", []p2p.Cap{})
	peer1 := pm.newPeer(lachesis62, p, new(fuzzMsgReadWriter))
	err = pm.handle(peer1)
	require.NoError(err)

}

type fuzzMsgReadWriter struct {
}

func (rw *fuzzMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	return p2p.Msg{}, nil
}

func (rw *fuzzMsgReadWriter) WriteMsg(p2p.Msg) error {
	return nil
}
