package gossip

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/dnsdisc"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/ethapi"
	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/gaspowercheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/parentscheck"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/drivermodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/eventmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/evmmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/sealmodule"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc/verwatcher"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/gossip/filters"
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils/gsignercache"
	"github.com/Fantom-foundation/go-opera/utils/wgmutex"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/Fantom-foundation/go-opera/vecmt"
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

func DefaultBlockProc(g opera.Genesis) BlockProc {
	return BlockProc{
		SealerModule:        sealmodule.New(),
		TxListenerModule:    drivermodule.NewDriverTxListenerModule(),
		GenesisTxTransactor: drivermodule.NewDriverTxGenesisTransactor(g),
		PreTxTransactor:     drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor:    drivermodule.NewDriverTxTransactor(),
		EventsModule:        eventmodule.New(),
		EVMModule:           evmmodule.New(),
	}
}

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config Config

	wg   sync.WaitGroup
	done chan struct{}

	// server
	p2pServer *p2p.Server
	Name      string

	accountManager *accounts.Manager

	// application
	store               *Store
	engine              lachesis.Consensus
	dagIndexer          *vecmt.Index
	engineMu            *sync.RWMutex
	emitter             *emitter.Emitter
	txpool              *evmcore.TxPool
	heavyCheckReader    HeavyCheckReader
	gasPowerCheckReader GasPowerCheckReader
	checkers            *eventcheck.Checkers
	uniqueEventIDs      uniqueID

	// version watcher
	verWatcher *verwatcher.VerWarcher

	blockProcWg        sync.WaitGroup
	blockProcTasks     *workers.Workers
	blockProcTasksDone chan struct{}
	blockProcModules   BlockProc

	blockBusyFlag uint32
	eventBusyFlag uint32

	feed ServiceFeed

	gpo *gasprice.Oracle

	// application protocol
	pm *ProtocolManager

	dialCandidates enode.Iterator

	EthAPI        *EthAPIBackend
	netRPCService *ethapi.PublicNetAPI

	stopped bool

	logger.Instance
}

func NewService(stack *node.Node, config Config, store *Store, signer valkeystore.SignerI, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = stack.ResolvePath(config.TxPool.Journal)
	}

	svc, err := newService(config, store, signer, blockProc, engine, dagIndexer)
	if err != nil {
		return nil, err
	}

	svc.p2pServer = stack.Server()
	svc.accountManager = stack.AccountManager()
	// Create the net API service
	svc.netRPCService = ethapi.NewPublicNetAPI(svc.p2pServer, store.GetRules().NetworkID)

	return svc, nil
}

func newService(config Config, store *Store, signer valkeystore.SignerI, blockProc BlockProc, engine lachesis.Consensus, dagIndexer *vecmt.Index) (*Service, error) {
	svc := &Service{
		config:             config,
		done:               make(chan struct{}),
		blockProcTasksDone: make(chan struct{}),
		Name:               fmt.Sprintf("Node-%d", rand.Int()),
		store:              store,
		engine:             engine,
		blockProcModules:   blockProc,
		dagIndexer:         dagIndexer,
		engineMu:           new(sync.RWMutex),
		uniqueEventIDs:     uniqueID{new(big.Int)},
		Instance:           logger.MakeInstance(),
	}

	svc.blockProcTasks = workers.New(new(sync.WaitGroup), svc.blockProcTasksDone, 1)

	// load epoch DB
	svc.store.loadEpochStore(svc.store.GetEpoch())
	es := svc.store.getEpochStore(svc.store.GetEpoch())
	svc.dagIndexer.Reset(svc.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
		return svc.store.GetEvent(id)
	})
	// load caches for mutable values to avoid race condition
	svc.store.GetBlockEpochState()
	svc.store.GetHighestLamport()

	// create GPO
	svc.gpo = gasprice.NewOracle(&GPOBackend{store}, svc.config.GPO)

	// create checkers
	net := store.GetRules()
	svc.heavyCheckReader.Addrs.Store(NewEpochPubKeys(svc.store, svc.store.GetEpoch()))                                             // read pub keys of current epoch from disk
	svc.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(svc.store, svc.store.GetValidators(), svc.store.GetEpoch(), net.Economy)) // read gaspower check data from disk
	svc.checkers = makeCheckers(config.HeavyCheck, net.EvmChainConfig().ChainID, &svc.heavyCheckReader, &svc.gasPowerCheckReader, svc.store)

	// create tx pool
	stateReader := svc.GetEvmStateReader()
	svc.txpool = evmcore.NewTxPool(config.TxPool, net.EvmChainConfig(), stateReader)

	// init dialCandidates
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})
	var err error
	svc.dialCandidates, err = dnsclient.NewIterator()

	// create protocol manager
	svc.pm, err = newHandler(handlerConfig{config, &svc.feed, svc.txpool, svc.engineMu, svc.checkers, store, svc.processEvent})
	if err != nil {
		return nil, err
	}

	// create API backend
	svc.EthAPI = &EthAPIBackend{config.ExtRPCEnabled, svc, stateReader, config.AllowUnprotectedTxs}

	svc.emitter = svc.makeEmitter(signer)

	svc.verWatcher = verwatcher.New(config.VersionWatcher, verwatcher.NewStore(store.table.NetworkVersion))

	return svc, nil
}

// makeCheckers builds event checkers
func makeCheckers(heavyCheckCfg heavycheck.Config, chainID *big.Int, heavyCheckReader *HeavyCheckReader, gasPowerCheckReader *GasPowerCheckReader, store *Store) *eventcheck.Checkers {
	// create signatures checker
	heavyCheck := heavycheck.New(heavyCheckCfg, heavyCheckReader, gsignercache.Wrap(types.NewEIP2930Signer(chainID)))

	// create gaspower checker
	gaspowerCheck := gaspowercheck.New(gasPowerCheckReader)

	return &eventcheck.Checkers{
		Basiccheck:    basiccheck.New(),
		Epochcheck:    epochcheck.New(store),
		Parentscheck:  parentscheck.New(),
		Heavycheck:    heavyCheck,
		Gaspowercheck: gaspowerCheck,
	}
}

func (s *Service) makeEmitter(signer valkeystore.SignerI) *emitter.Emitter {
	txSigner := gsignercache.Wrap(types.NewEIP2930Signer(s.store.GetRules().EvmChainConfig().ChainID))

	return emitter.NewEmitter(s.config.Emitter, emitter.World{
		External: &emitterWorld{
			s:       s,
			Store:   s.store,
			WgMutex: wgmutex.New(s.engineMu, &s.blockProcWg),
		},
		TxPool:   s.txpool,
		Signer:   signer,
		TxSigner: txSigner,
	})
}

// MakeProtocols constructs the P2P protocol definitions for `opera`.
func MakeProtocols(svc *Service, backend *ProtocolManager, network uint64, disc enode.Iterator) []p2p.Protocol {
	protocols := make([]p2p.Protocol, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		version := version // Closure

		protocols[i] = p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  protocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := NewPeer(int(version), p, rw, backend.config.Protocol.PeerCache)

				select {
				case backend.newPeerCh <- peer:
					backend.wg.Add(1)
					defer backend.wg.Done()
					err := backend.handle(peer)
					return err
				case <-backend.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return backend.NodeInfo()
			},
			PeerInfo: func(id enode.ID) interface{} {
				if p := backend.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
			Attributes:     []enr.Entry{currentENREntry(svc)},
			DialCandidates: disc,
		}
	}
	return protocols
}

// Protocols returns protocols the service can communicate on.
func (s *Service) Protocols() []p2p.Protocol {
	return MakeProtocols(s, s.pm, s.store.GetRules().NetworkID, s.dialCandidates)
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
	err := s.store.Init()
	if err != nil {
		return err
	}

	StartENRUpdater(s, s.p2pServer.LocalNode())

	s.blockProcTasks.Start(1)

	s.pm.Start(s.p2pServer.MaxPeers)

	s.emitter.Start()

	s.verWatcher.Start()

	return nil
}

// WaitBlockEnd waits until parallel block processing is complete (if any)
func (s *Service) WaitBlockEnd() {
	s.blockProcWg.Wait()
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	defer log.Info("Fantom service stopped")
	s.verWatcher.Stop()
	close(s.done)
	s.emitter.Stop()
	s.pm.Stop()
	s.wg.Wait()
	s.feed.scope.Close()

	// flush the state at exit, after all the routines stopped
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	s.stopped = true

	s.blockProcWg.Wait()
	close(s.blockProcTasksDone)
	return s.store.Commit()
}

// AccountManager return node's account manager
func (s *Service) AccountManager() *accounts.Manager {
	return s.accountManager
}
