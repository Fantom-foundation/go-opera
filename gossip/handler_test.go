package gossip

import (
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/eventmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/evmmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sealmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sfcmodule"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestPM(t *testing.T) {
	require := require.New(t)

	genStore := makegenesis.FakeGenesisStore(genesisStakers, utils.ToFtm(genesisBalance), utils.ToFtm(genesisStake))
	genesis := genStore.GetGenesis()
	network := genesis.Rules

	config := DefaultConfig()
	store := NewMemStore()
	blockProc := BlockProc{
		SealerModule:        sealmodule.New(network),
		TxListenerModule:    sfcmodule.NewSfcTxListenerModule(network),
		GenesisTxTransactor: sfcmodule.NewSfcTxGenesisTransactor(genesis),
		PreTxTransactor:     sfcmodule.NewSfcTxPreTransactor(network),
		PostTxTransactor:    sfcmodule.NewSfcTxTransactor(network),
		EventsModule:        eventmodule.New(network),
		EVMModule:           evmmodule.New(network),
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
	checkers := makeCheckers(config.HeavyCheck, network, &heavyCheckReader, &gasPowerCheckReader, store)
	processEvent := func(e *inter.EventPayload) error {
		return nil
	}

	txpool := evmcore.NewTxPool(config.TxPool, network.EvmChainConfig(), &EvmStateReader{
		ServiceFeed: feed,
		store:       store,
	})

	pm, err := NewProtocolManager(
		config,
		network,
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
