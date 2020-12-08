package gossip

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discv5"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/ethapi"
	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/parentscheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/filters"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
	"github.com/Fantom-foundation/go-opera/gossip/occuredtxs"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/Fantom-foundation/go-opera/vecmt"
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

func (f *ServiceFeed) SubscribeNewEmitted(ch chan<- *inter.EventPayload) notify.Subscription {
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

type BlockProc struct {
	SealerModule        blockproc.SealerModule
	TxListenerModule    blockproc.TxListenerModule
	GenesisTxTransactor blockproc.TxTransactor
	PreTxTransactor     blockproc.TxTransactor
	PostTxTransactor    blockproc.TxTransactor
	EventsModule        blockproc.ConfirmedEventsModule
	EVMModule           blockproc.EVM
}

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config Config
	net    opera.Rules

	wg   sync.WaitGroup
	done chan struct{}

	// server
	p2pServer *p2p.Server
	Name      string
	Topic     discv5.Topic

	serverPool *serverPool

	accountManager *accounts.Manager

	// application
	store               *Store
	engine              lachesis.Consensus
	dagIndexer          *vecmt.Index
	engineMu            *sync.RWMutex
	emitter             *emitter.Emitter
	txpool              *evmcore.TxPool
	occurredTxs         *occuredtxs.Buffer
	heavyCheckReader    HeavyCheckReader
	gasPowerCheckReader GasPowerCheckReader
	checkers            *eventcheck.Checkers
	uniqueEventIDs      uniqueID

	blockProcWg      sync.WaitGroup
	blockProcTasks   *workers.Workers
	blockProcModules BlockProc

	feed ServiceFeed

	// application protocol
	pm *ProtocolManager

	EthAPI        *EthAPIBackend
	netRPCService *ethapi.PublicNetAPI

	stopped bool

	logger.Instance
}

func NewService(stack *node.Node, config Config, net opera.Rules, store *Store, signer valkeystore.SignerI, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = stack.ResolvePath(config.TxPool.Journal)
	}

	svc, err := newService(config, net, store, signer, blockProc, engine, dagIndexer)
	if err != nil {
		return nil, err
	}

	svc.p2pServer = stack.Server()
	svc.accountManager = stack.AccountManager()
	// Create the net API service
	svc.netRPCService = ethapi.NewPublicNetAPI(svc.p2pServer, svc.net.NetworkID)

	return svc, nil
}

func newService(config Config, net opera.Rules, store *Store, signer valkeystore.SignerI, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index) (*Service, error) {
	svc := &Service{
		config:           config,
		net:              net,
		done:             make(chan struct{}),
		Name:             fmt.Sprintf("Node-%d", rand.Int()),
		store:            store,
		engine:           engine,
		blockProcModules: blockProc,
		dagIndexer:       dagIndexer,
		engineMu:         new(sync.RWMutex),
		occurredTxs:      occuredtxs.New(txsRingBufferSize, types.NewEIP155Signer(net.EvmChainConfig().ChainID)),
		uniqueEventIDs:   uniqueID{new(big.Int)},
		Instance:         logger.MakeInstance(),
	}

	svc.blockProcTasks = workers.New(&svc.wg, svc.done, 1)

	// create server pool
	trustedNodes := []string{}
	svc.serverPool = newServerPool(store.async.table.Peers, svc.done, &svc.wg, trustedNodes)

	// create tx pool
	stateReader := svc.GetEvmStateReader()
	svc.txpool = evmcore.NewTxPool(config.TxPool, net.EvmChainConfig(), stateReader)

	// create checkers
	svc.heavyCheckReader.Addrs.Store(NewEpochPubKeys(svc.store, svc.store.GetEpoch()))                                                  // read pub keys of current epoch from disk
	svc.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(svc.store, svc.store.GetValidators(), svc.store.GetEpoch(), &svc.net.Economy)) // read gaspower check data from disk
	svc.checkers = makeCheckers(config.HeavyCheck, svc.net, &svc.heavyCheckReader, &svc.gasPowerCheckReader, svc.store)

	// create protocol manager
	var err error
	svc.pm, err = NewProtocolManager(config, net, &svc.feed, svc.txpool, svc.engineMu, svc.checkers, store, svc.processEvent, svc.serverPool)
	if err != nil {
		return nil, err
	}

	// create API backend
	svc.EthAPI = &EthAPIBackend{config.ExtRPCEnabled, svc, stateReader, nil}
	svc.EthAPI.gpo = gasprice.NewOracle(svc.EthAPI, svc.config.GPO)

	// load epoch DB
	svc.store.loadEpochStore(svc.store.GetEpoch())
	es := svc.store.getEpochStore(svc.store.GetEpoch())
	svc.dagIndexer.Reset(svc.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
		return svc.store.GetEvent(id)
	})

	svc.emitter = svc.makeEmitter(signer)

	return svc, nil
}

// makeCheckers builds event checkers
func makeCheckers(heavyCheckCfg heavycheck.Config, net opera.Rules, heavyCheckReader *HeavyCheckReader, gasPowerCheckReader *GasPowerCheckReader, store *Store) *eventcheck.Checkers {
	// create signatures checker
	heavyCheck := heavycheck.New(heavyCheckCfg, heavyCheckReader, types.NewEIP155Signer(net.EvmChainConfig().ChainID))

	// create gaspower checker
	gaspowerCheck := gaspowercheck.New(gasPowerCheckReader)

	return &eventcheck.Checkers{
		Basiccheck:    basiccheck.New(&net.Dag),
		Epochcheck:    epochcheck.New(store),
		Parentscheck:  parentscheck.New(),
		Heavycheck:    heavyCheck,
		Gaspowercheck: gaspowerCheck,
	}
}

func (s *Service) makeEmitter(signer valkeystore.SignerI) *emitter.Emitter {
	// randomize event time to decrease peak load, and increase chance of catching double instances of validator
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	emitterCfg := s.config.Emitter // copy data
	emitterCfg.EmitIntervals = *emitterCfg.EmitIntervals.RandomizeEmitTime(r)

	return emitter.NewEmitter(s.net, emitterCfg,
		emitter.EmitterWorld{
			Store:       s.store,
			EngineMu:    s.engineMu,
			Txpool:      s.txpool,
			Signer:      signer,
			OccurredTxs: s.occurredTxs,
			Check: func(emitted *inter.EventPayload, parents inter.Events) error {
				// sanity check
				return s.checkers.Validate(emitted, parents.Interfaces())
			},
			Process: func(emitted *inter.EventPayload) error {
				err := s.processEvent(emitted)
				if err != nil {
					s.Log.Crit("Self-event connection failed", "err", err.Error())
				}

				s.feed.newEmittedEvent.Send(emitted) // PM listens and will broadcast it
				if err != nil {
					s.Log.Crit("Failed to post self-event", "err", err.Error())
				}
				return nil
			},
			Build:    s.buildEvent,
			DagIndex: s.dagIndexer,
			IsSynced: func() bool {
				return atomic.LoadUint32(&s.pm.synced) != 0
			},
			PeersNum: func() int {
				return s.pm.peers.Len()
			},
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
func (s *Service) Start() error {
	var genesis hash.Event
	genesis = s.store.GetBlock(0).Atropos
	s.Topic = discv5.Topic("opera@" + genesis.Hex())

	if s.p2pServer.DiscV5 != nil {
		go func(topic discv5.Topic) {
			s.Log.Info("Starting topic registration")
			defer s.Log.Info("Terminated topic registration")

			s.p2pServer.DiscV5.RegisterTopic(topic, s.done)
		}(s.Topic)
	}

	s.blockProcTasks.Start(1)

	s.pm.Start(s.p2pServer.MaxPeers)

	s.serverPool.start(s.p2pServer, s.Topic)

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
	s.stopped = true

	return s.store.Commit(nil, true)
}

// AccountManager return node's account manager
func (s *Service) AccountManager() *accounts.Manager {
	return s.accountManager
}
