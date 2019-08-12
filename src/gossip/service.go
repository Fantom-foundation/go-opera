package gossip

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/cryptoaddr"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/ordering"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

const (
	PingMsg uint64 = iota
	PongMsg
)

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config *Config

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
	s      *Store
	engine Consensus

	pushEvent         ordering.PushEventFn
	isEventDownloaded ordering.IsBufferedFn

	consensusMu *sync.RWMutex

	eventMux *event.TypeMux

	// application protocol
	Pinger *PingAPI
	pm     *ProtocolManager

	logger.Instance
}

func NewService(eventMux *event.TypeMux, config *Config, s *Store, engine Consensus) (*Service, error) {
	svc := &Service{
		config: config,

		done: make(chan struct{}),

		Name: fmt.Sprintf("Node-%d", rand.Int()),

		s:      s,
		engine: engine,
		Pinger: &PingAPI{},

		consensusMu: new(sync.RWMutex),

		eventMux: eventMux,

		Instance: logger.MakeInstance(),
	}

	trustedNodes := []string{}

	svc.serverPool = newServerPool(s.table.Peers, svc.done, &svc.wg, trustedNodes)

	// build EventBuffer
	pushInBuffer, isEventBuffered := ordering.EventBuffer(ordering.Callback{
		Process: func(e *inter.Event) error {
			log.Info("New event", "hash", e.Hash())

			svc.s.SetEvent(e)
			err := svc.engine.ProcessEvent(e)
			if err != nil {
				return err
			}

			// If the event is indeed in out own graph, announce it
			if svc.pm != nil {
				if e.Creator == svc.me {
					// if I'm a creator, then propagate aggressively
					svc.pm.BroadcastEvent(e, true)
				}
				svc.pm.BroadcastEvent(e, false)
			}
			return nil
		},

		Drop: func(e *inter.Event, peer string, err error) {
			log.Warn("Event rejected", "err", err)
			svc.s.DeleteEvent(e.Hash())
			if svc.pm != nil {
				svc.pm.removePeer(peer)
			}
		},

		Exists: func(id hash.Event) *inter.Event {
			return s.GetEvent(id)
		},
	})
	svc.pushEvent = func(e *inter.Event, peer string) {
		svc.consensusMu.Lock()
		defer svc.consensusMu.Unlock()

		pushInBuffer(e, peer)
	}
	svc.isEventDownloaded = func(id hash.Event) bool {
		svc.consensusMu.RLock()
		defer svc.consensusMu.RUnlock()

		if isEventBuffered(id) {
			return true
		}
		return s.HasEvent(id)
	}

	var err error
	svc.pm, err = NewProtocolManager(config.Dag, downloader.FullSync, config.Dag.DagID, svc.eventMux, nil, svc.pushEvent, svc.isEventDownloaded, engine)

	return svc, err
}

// Protocols returns protocols the service can communicate on.
func (s *Service) Protocols() []p2p.Protocol {

	protos := make([]p2p.Protocol, len(ProtocolVersions))
	for i, vsn := range ProtocolVersions {
		protos[i] = s.pm.makeProtocol(vsn)
		protos[i].Attributes = []enr.Entry{s.currentEnr()}
	}

	protos = append(protos, s.makePingProtocol(1))
	return protos
}

// APIs returns api methods the service wants to expose on rpc channels.
func (s *Service) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "ping",
			Version:   "1.0",
			Service:   s.Pinger,
			Public:    true,
		},
	}
}

// Start method invoked when the node is ready to start the service.
func (s *Service) Start(srv *p2p.Server) error {

	var genesis hash.Hash
	if s.engine != nil {
		genesis = s.engine.GetGenesisHash()
	}
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

	return nil
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	fmt.Println("Service stopping...")
	s.pm.Stop()
	s.wg.Wait()
	return nil
}

func (s *Service) makePingProtocol(version uint) p2p.Protocol {
	return p2p.Protocol{
		Name:    "ping",
		Version: 1,
		Length:  10,
		NodeInfo: func() interface{} {
			return struct{}{}
		},
		PeerInfo: func(id enode.ID) interface{} {
			return struct{}{}
		},
		Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
			select {
			case <-s.done:
				return p2p.DiscQuitting
			default:
				s.wg.Add(1)
				defer s.wg.Done()
				return s.handlePing(peer, rw)
			}
		},
	}
}

func (s *Service) handlePing(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	errc := make(chan error)

	go func() {
		errc <- func() error {

			msg, err := rw.ReadMsg()
			if err != nil {
				return err
			}

			if msg.Code != PingMsg {
				_ = msg.Discard()
				return errResp(ErrInvalidMsgCode, "ping expected")
			}

			var req string
			err = msg.Decode(&req)
			if err != nil {
				return err
			}

			resp := s.Pinger.Hi(req)

			size, r, err := rlp.EncodeToReader(&resp)
			if err != nil {
				return err
			}

			err = rw.WriteMsg(p2p.Msg{
				Code:    PongMsg,
				Size:    uint32(size),
				Payload: r})
			if err != nil {
				return err
			}

			return nil
		}()
	}()

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case err := <-errc:
		if err != nil {
			return err
		}
	case <-timeout.C:
		return p2p.DiscReadTimeout
	}

	return nil
}

// PingAPI:

type PingAPI struct{}

func (a *PingAPI) Hi(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
