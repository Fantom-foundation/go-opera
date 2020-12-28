package gossip

import (
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

func TestPM(t *testing.T) {
	require := require.New(t)

	done := make(chan struct{})
	wg := new(sync.WaitGroup)

	config := DefaultConfig()
	net := opera.TestNetRules()
	store := NewMemStore()
	// TODO: store.ApplyGenesis()

	var (
		heavyCheckReader    HeavyCheckReader
		gasPowerCheckReader GasPowerCheckReader
		// TODO: init
	)

	mu := new(sync.RWMutex)
	feed := new(ServiceFeed)
	serverPool := newServerPool(store.async.table.Peers, done, wg, []string{})
	checkers := makeCheckers(config.HeavyCheck, net, &heavyCheckReader, &gasPowerCheckReader, store)
	processEvent := func(e *inter.EventPayload) error {
		return nil
	}

	txpool := evmcore.NewTxPool(config.TxPool, net.EvmChainConfig(), &EvmStateReader{
		ServiceFeed: feed,
		store:       store,
	})

	pm, err := NewProtocolManager(
		config,
		net,
		feed,
		txpool,
		mu,
		checkers,
		store,
		processEvent,
		serverPool)
	require.NoError(err)

	pm.Start(3)
	defer pm.Stop()

	protocol := pm.makeProtocol(lachesis62)

	peer1 := p2p.NewPeer(enode.RandomID(enode.ID{}, 1), "fake-node-1", []p2p.Cap{})
	err = protocol.Run(peer1, new(fuzzMsgReadWriter))
	require.NoError(err)

	close(done)
	wg.Wait()
}

type fuzzMsgReadWriter struct {
}

func (rw *fuzzMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	return p2p.Msg{}, nil
}

func (rw *fuzzMsgReadWriter) WriteMsg(p2p.Msg) error {
	return nil
}
