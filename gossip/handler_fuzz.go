//go:build gofuzz
//+build gofuzz

package gossip

import (
	"bytes"
	"errors"
	"math/rand"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	_ "github.com/dvyukov/go-fuzz/go-fuzz-defs"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/integration/makefakegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/signers/gsignercache"
)

const (
	fuzzHot      int = 1  // if the fuzzer should increase priority of the given input during subsequent fuzzing;
	fuzzCold     int = -1 // if the input must not be added to corpus even if gives new coverage;
	fuzzNoMatter int = 0  // otherwise.
)

var (
	fuzzedHandler *handler
)

func FuzzHandler(data []byte) int {
	var err error
	if fuzzedHandler == nil {
		fuzzedHandler, err = makeFuzzedHandler()
		if err != nil {
			panic(err)
		}
	}

	msg, err := newFuzzMsg(data)
	if err != nil {
		return fuzzCold
	}
	input := &fuzzMsgReadWriter{msg}
	other := &peer{
		version: ProtocolVersion,
		Peer:    p2p.NewPeer(randomID(), "fake-node-1", []p2p.Cap{}),
		rw:      input,
	}

	err = fuzzedHandler.handleMsg(other)
	if err != nil {
		return fuzzNoMatter
	}

	return fuzzHot
}

func makeFuzzedHandler() (h *handler, err error) {
	const (
		genesisStakers = 3
		genesisBalance = 1e18
		genesisStake   = 2 * 4e6
	)

	genStore := makefakegenesis.FakeGenesisStore(genesisStakers, utils.ToFtm(genesisBalance), utils.ToFtm(genesisStake))
	genesis := genStore.Genesis()

	config := DefaultConfig(cachescale.Identity)
	store := NewMemStore()
	_, err = store.ApplyGenesis(genesis)
	if err != nil {
		return
	}

	var (
		network             = opera.FakeNetRules()
		heavyCheckReader    HeavyCheckReader
		gasPowerCheckReader GasPowerCheckReader
		// TODO: init
	)

	mu := new(sync.RWMutex)
	feed := new(ServiceFeed)
	net := store.GetRules()
	txSigner := gsignercache.Wrap(types.LatestSignerForChainID(net.EvmChainConfig().ChainID))
	checkers := makeCheckers(config.HeavyCheck, txSigner, &heavyCheckReader, &gasPowerCheckReader, store)

	txpool := evmcore.NewTxPool(evmcore.DefaultTxPoolConfig, network.EvmChainConfig(), &EvmStateReader{
		ServiceFeed: feed,
		store:       store,
	})

	h, err = newHandler(
		handlerConfig{
			config:   config,
			notifier: feed,
			txpool:   txpool,
			engineMu: mu,
			checkers: checkers,
			s:        store,
			process: processCallback{
				Event: func(event *inter.EventPayload) error {
					return nil
				},
			},
		})
	if err != nil {
		return
	}

	h.Start(3)
	return
}

func randomID() (id enode.ID) {
	for i := range id {
		id[i] = byte(rand.Intn(255))
	}
	return id
}

type fuzzMsgReadWriter struct {
	msg *p2p.Msg
}

func newFuzzMsg(data []byte) (*p2p.Msg, error) {
	if len(data) < 1 {
		return nil, errors.New("empty data")
	}

	var (
		codes = []uint64{
			HandshakeMsg,
			EvmTxsMsg,
			ProgressMsg,
			NewEventIDsMsg,
			GetEventsMsg,
			EventsMsg,
			RequestEventsStream,
			EventsStreamResponse,
		}
		code = codes[int(data[0])%len(codes)]
	)
	data = data[1:]

	return &p2p.Msg{
		Code:    code,
		Size:    uint32(len(data)),
		Payload: bytes.NewReader(data),
	}, nil
}

func (rw *fuzzMsgReadWriter) ReadMsg() (p2p.Msg, error) {
	return *rw.msg, nil
}

func (rw *fuzzMsgReadWriter) WriteMsg(p2p.Msg) error {
	return nil
}
