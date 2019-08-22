package gossip

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config Config

	wg   sync.WaitGroup
	done chan struct{}

	// server
	Name   string
	Topics []discv5.Topic

	serverPool *serverPool

	// my identity
	me         hash.Peer
	privateKey *crypto.PrivateKey

	// application
	store    *Store
	engine   Consensus
	engineMu *sync.RWMutex
	emitter  *Emitter

	mux *event.TypeMux

	// application protocol
	pm *ProtocolManager

	logger.Instance
}

func NewService(config Config, mux *event.TypeMux, store *Store, engine Consensus) (*Service, error) {
	svc := &Service{
		config: config,

		done: make(chan struct{}),

		Name: fmt.Sprintf("Node-%d", rand.Int()),

		store: store,

		engineMu: new(sync.RWMutex),

		mux: mux,

		Instance: logger.MakeInstance(),
	}

	engine = &HookedEngine{
		engine:       engine,
		processEvent: svc.processEvent,
	}
	svc.engine = engine

	engine.Bootstrap(svc.ApplyBlock)

	trustedNodes := []string{}

	svc.serverPool = newServerPool(store.table.Peers, svc.done, &svc.wg, trustedNodes)

	var err error
	svc.pm, err = NewProtocolManager(&config, svc.mux, &dummyTxPool{}, svc.engineMu, store, engine)

	return svc, err
}

func (s *Service) processEvent(realEngine Consensus, e *inter.Event) error {
	// s.engineMu is locked here

	if s.store.HasEvent(e.Hash()) { // sanity check
		s.store.Fatalf("ProcessEvent: event is already processed %s", e.Hash().String())
	}

	oldEpoch := realEngine.CurrentSuperFrameN()

	s.store.SetEvent(e)
	if realEngine != nil {
		err := realEngine.ProcessEvent(e)
		if err != nil { // TODO make it possible to write only on success
			s.store.DeleteEvent(e.Epoch, e.Hash())
			return err
		}
	}
	// set member's last event. we don't care about forks, because this index is used only for emitter
	s.store.SetLastEvent(e.Epoch, e.Creator, e.Hash())

	// track events with no descendants, i.e. heads
	for _, parent := range e.Parents {
		if s.store.IsHead(e.Epoch, parent) {
			s.store.DelHead(e.Epoch, parent)
		}
	}
	s.store.AddHead(e.Epoch, e.Hash())

	s.packs_onNewEvent(e)

	newEpoch := realEngine.CurrentSuperFrameN()
	if newEpoch != oldEpoch {
		s.packs_onNewEpoch(oldEpoch, newEpoch)
		s.mux.Post(newEpoch)
		s.mux.Post(s.store.GetPacksNumOrDefault(newEpoch))
	}

	return nil
}

func (s *Service) makeEmitter() *Emitter {
	return NewEmitter(&s.config, s.me, s.privateKey, s.engineMu, s.store, s.engine, func(emitted *inter.Event) {
		// s.engineMu is locked here

		err := s.engine.ProcessEvent(emitted)
		if err != nil {
			s.Fatalf("Self-event connection failed: %s", err.Error())
		}

		err = s.pm.mux.Post(emitted) // PM listens and will broadcast it
		if err != nil {
			s.Fatalf("Failed to post self-event: %s", err.Error())
		}
	},
	)
}

// Protocols returns protocols the service can communicate on.
func (s *Service) Protocols() []p2p.Protocol {
	protos := make([]p2p.Protocol, len(ProtocolVersions))
	for i, vsn := range ProtocolVersions {
		protos[i] = s.pm.makeProtocol(vsn)
		protos[i].Attributes = []enr.Entry{s.currentEnr()}
	}
	return protos
}

// APIs returns api methods the service wants to expose on rpc channels.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{}
}

// Start method invoked when the node is ready to start the service.
func (s *Service) Start(srv *p2p.Server) error {

	var genesis hash.Hash
	genesis = s.engine.GetGenesisHash()
	s.Topics = []discv5.Topic{
		discv5.Topic("lachesis@" + genesis.Hex()),
	}

	if srv.DiscV5 != nil {
		for _, topic := range s.Topics {
			topic := topic
			go func() {
				s.Info("Starting topic registration")
				defer s.Info("Terminated topic registration")

				srv.DiscV5.RegisterTopic(topic, s.done)
			}()
		}
	}
	s.privateKey = (*crypto.PrivateKey)(srv.PrivateKey)
	s.me = cryptoaddr.AddressOf(s.privateKey.Public())

	s.pm.Start(srv.MaxPeers)

	s.emitter = s.makeEmitter()

	return nil
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	fmt.Println("Service stopping...")
	s.pm.Stop()
	s.wg.Wait()
	return nil
}
