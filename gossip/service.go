package gossip

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/ethapi"
	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/parentscheck"
	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/gossip/filters"
	"github.com/Fantom-foundation/go-lachesis/gossip/gasprice"
	"github.com/Fantom-foundation/go-lachesis/gossip/occuredtxs"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/params"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

const (
	txsRingBufferSize = 20000 // Maximum number of stored hashes of included but not confirmed txs
)

type ServiceFeed struct {
	scope notify.SubscriptionScope

	newEpoch        notify.Feed
	newPack         notify.Feed
	newEmittedEvent notify.Feed
	newBlock        notify.Feed
	newTxs          notify.Feed
	newLogs         notify.Feed
}

func (f *ServiceFeed) SubscribeNewEpoch(ch chan<- idx.Epoch) notify.Subscription {
	return f.scope.Track(f.newEpoch.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewPack(ch chan<- idx.Pack) notify.Subscription {
	return f.scope.Track(f.newPack.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewEmitted(ch chan<- *inter.Event) notify.Subscription {
	return f.scope.Track(f.newEmittedEvent.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewBlock(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return f.scope.Track(f.newBlock.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewTxs(ch chan<- core.NewTxsEvent) notify.Subscription {
	return f.scope.Track(f.newTxs.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewLogs(ch chan<- []*types.Log) notify.Subscription {
	return f.scope.Track(f.newLogs.Subscribe(ch))
}

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config *Config

	wg   sync.WaitGroup
	done chan struct{}

	// server
	Name  string
	Topic discv5.Topic

	serverPool *serverPool

	// application
	node                *node.ServiceContext
	store               *Store
	app                 *app.Store
	engine              Consensus
	engineMu            *sync.RWMutex
	emitter             *Emitter
	txpool              *evmcore.TxPool
	occurredTxs         *occuredtxs.Buffer
	heavyCheckReader    HeavyCheckReader
	gasPowerCheckReader GasPowerCheckReader
	checkers            *eventcheck.Checkers

	// global variables. TODO refactor to pass them as arguments if possible
	blockParticipated map[idx.StakerID]bool // validators who participated in last block
	currentEvent      hash.Event            // current event which is being processed

	feed ServiceFeed

	// application protocol
	pm *ProtocolManager

	EthAPI        *EthAPIBackend
	netRPCService *ethapi.PublicNetAPI

	logger.Instance
}

func NewService(ctx *node.ServiceContext, config *Config, store *Store, engine Consensus, app *app.Store) (*Service, error) {
	svc := &Service{
		config: config,

		done: make(chan struct{}),

		Name: fmt.Sprintf("Node-%d", rand.Int()),

		node:  ctx,
		store: store,
		app:   app,

		engineMu:          new(sync.RWMutex),
		occurredTxs:       occuredtxs.New(txsRingBufferSize, types.NewEIP155Signer(config.Net.EvmChainConfig().ChainID)),
		blockParticipated: make(map[idx.StakerID]bool),

		Instance: logger.MakeInstance(),
	}

	// wrap engine
	svc.engine = &HookedEngine{
		engine:       engine,
		processEvent: svc.processEvent,
	}
	svc.engine.Bootstrap(inter.ConsensusCallbacks{
		ApplyBlock:              svc.applyBlock,
		SelectValidatorsGroup:   svc.selectValidatorsGroup,
		OnEventConfirmed:        svc.onEventConfirmed,
		IsEventAllowedIntoBlock: svc.isEventAllowedIntoBlock,
	})

	// create server pool
	trustedNodes := []string{}
	svc.serverPool = newServerPool(store.table.Peers, svc.done, &svc.wg, trustedNodes)

	// create tx pool
	stateReader := svc.GetEvmStateReader()
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	svc.txpool = evmcore.NewTxPool(config.TxPool, config.Net.EvmChainConfig(), stateReader)

	// create checkers
	svc.heavyCheckReader.Addrs.Store(ReadEpochPubKeys(svc.app, svc.engine.GetEpoch()))                                                                     // read pub keys of current epoch from disk
	svc.gasPowerCheckReader.Ctx.Store(ReadGasPowerContext(svc.store, svc.app, svc.engine.GetValidators(), svc.engine.GetEpoch(), &svc.config.Net.Economy)) // read gaspower check data from disk
	svc.checkers = makeCheckers(&svc.config.Net, &svc.heavyCheckReader, &svc.gasPowerCheckReader, svc.engine, svc.store)

	// create protocol manager
	var err error
	svc.pm, err = NewProtocolManager(config, &svc.feed, svc.txpool, svc.engineMu, svc.checkers, store, svc.engine, svc.serverPool)

	// create API backend
	svc.EthAPI = &EthAPIBackend{config.ExtRPCEnabled, svc, stateReader, nil}
	svc.EthAPI.gpo = gasprice.NewOracle(svc.EthAPI, svc.config.GPO)

	return svc, err
}

// makeCheckers builds event checkers
func makeCheckers(net *lachesis.Config, heavyCheckReader *HeavyCheckReader, gasPowerCheckReader *GasPowerCheckReader, engine Consensus, store *Store) *eventcheck.Checkers {
	// create signatures checker
	ledgerID := net.EvmChainConfig().ChainID
	heavyCheck := heavycheck.NewDefault(&net.Dag, heavyCheckReader, types.NewEIP155Signer(ledgerID))

	// create gaspower checker
	gaspowerCheck := gaspowercheck.New(gasPowerCheckReader)

	return &eventcheck.Checkers{
		Basiccheck:    basiccheck.New(&net.Dag),
		Epochcheck:    epochcheck.New(&net.Dag, engine),
		Parentscheck:  parentscheck.New(&net.Dag),
		Heavycheck:    heavyCheck,
		Gaspowercheck: gaspowerCheck,
	}
}

func (s *Service) makeEmitter() *Emitter {
	// randomize event time to decrease peak load, and increase chance of catching double instances of validator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	emitterCfg := s.config.Emitter // copy data
	emitterCfg.EmitIntervals = *emitterCfg.EmitIntervals.RandomizeEmitTime(r)

	return NewEmitter(&s.config.Net, &emitterCfg,
		EmitterWorld{
			Am:          s.AccountManager(),
			Engine:      s.engine,
			EngineMu:    s.engineMu,
			Store:       s.store,
			App:         s.app,
			Txpool:      s.txpool,
			OccurredTxs: s.occurredTxs,
			OnEmitted: func(emitted *inter.Event) {
				// s.engineMu is locked here

				err := s.engine.ProcessEvent(emitted)
				if err != nil {
					s.Log.Crit("Self-event connection failed", "err", err.Error())
				}

				s.feed.newEmittedEvent.Send(emitted) // PM listens and will broadcast it
				if err != nil {
					s.Log.Crit("Failed to post self-event", "err", err.Error())
				}
			},
			IsSynced: func() bool {
				return atomic.LoadUint32(&s.pm.synced) != 0
			},
			PeersNum: func() int {
				return s.pm.peers.Len()
			},
			AddVersion: func(e *inter.Event) *inter.Event {
				// serialization version
				e.Version = 0
				// node version
				if e.Seq <= 1 && len(s.config.Emitter.VersionToPublish) > 0 {
					version := []byte("v-" + s.config.Emitter.VersionToPublish)
					if len(version) <= params.MaxExtraData {
						e.Extra = version
					}
				}

				return e
			},
			Checkers: s.checkers,
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
	apis := ethapi.GetAPIs(s.EthAPI)

	apis = append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.EthAPI),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)

	return apis
}

// Start method invoked when the node is ready to start the service.
func (s *Service) Start(srv *p2p.Server) error {
	// Start the RPC service
	s.netRPCService = ethapi.NewPublicNetAPI(srv, s.config.Net.NetworkID)

	var genesis common.Hash
	genesis = s.engine.GetGenesisHash()
	s.Topic = discv5.Topic("lachesis@" + genesis.Hex())

	if srv.DiscV5 != nil {
		go func(topic discv5.Topic) {
			s.Log.Info("Starting topic registration")
			defer s.Log.Info("Terminated topic registration")

			srv.DiscV5.RegisterTopic(topic, s.done)
		}(s.Topic)
	}

	s.pm.Start(srv.MaxPeers)

	s.serverPool.start(srv, s.Topic)

	s.emitter = s.makeEmitter()
	s.emitter.SetValidator(s.config.Emitter.Validator)
	s.emitter.StartEventEmission()

	return nil
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	close(s.done)
	s.emitter.StopEventEmission()
	s.pm.Stop()
	s.wg.Wait()
	s.feed.scope.Close()

	// flush the state at exit, after all the routines stopped
	s.engineMu.Lock()
	defer s.engineMu.Unlock()

	err := s.app.Commit(nil, true)
	if err != nil {
		return err
	}
	return s.store.Commit(nil, true)
}

// AccountManager return node's account manager
func (s *Service) AccountManager() *accounts.Manager {
	return s.node.AccountManager
}
