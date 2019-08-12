package lachesis

import (
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

// DbProducer makes db.
type DbProducer func(name string) kvdb.Database

// Lachesis is a lachesis node implementation.
type Lachesis struct {
	host           string
	conf           *Config
	node           *posnode.Node
	nodeStore      *posnode.Store
	consensus      *posposet.Poset
	consensusStore *posposet.Store

	service

	logger.Instance
}

// New makes lachesis node.
// It does not start any process.
func New(
	newDb DbProducer,
	host string,
	key *crypto.PrivateKey,
	conf *Config,
	opts ...grpc.DialOption,
) *Lachesis {
	return makeLachesis(newDb, host, key, conf, nil, opts...)
}

func makeLachesis(
	newDb DbProducer,
	host string,
	key *crypto.PrivateKey,
	conf *Config,
	listen network.ListenFunc,
	opts ...grpc.DialOption,
) *Lachesis {
	ndb, cdb := makeStorages(newDb)

	if conf == nil {
		conf = DefaultConfig()
	}

	c := posposet.New(cdb, ndb)
	n := posnode.New(host, key, ndb, c, &conf.Node, listen, opts...)

	return &Lachesis{
		host:           host,
		conf:           conf,
		node:           n,
		nodeStore:      ndb,
		consensus:      c,
		consensusStore: cdb,

		service: service{listen, nil},

		Instance: logger.MakeInstance(),
	}
}

func (l *Lachesis) init() {
	genesis := l.conf.Net.Genesis
	err := l.consensusStore.ApplyGenesis(genesis, 0)
	if err != nil {
		l.Fatal(err)
	}
}

// Start inits and starts whole lachesis node.
func (l *Lachesis) Start() {
	l.init()

	l.consensus.Start()
	l.node.Start()
	l.serviceStart()
}

// Stop stops whole lachesis node.
func (l *Lachesis) Stop() {
	l.serviceStop()
	l.node.Stop()
	l.consensus.Stop()
}

// AddPeers suggests hosts for network discovery.
func (l *Lachesis) AddPeers(hosts ...string) {
	l.node.AddBuiltInPeers(hosts...)
}

/*
 * Utils:
 */

func makeStorages(newDb DbProducer) (*posnode.Store, *posposet.Store) {
	db := newDb("lachesis")

	p := db.NewTable([]byte("p_"))
	n := db.NewTable([]byte("n_"))

	return posnode.NewStore(n),
		posposet.NewStore(p, newDb)
}
