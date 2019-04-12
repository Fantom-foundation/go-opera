package lachesis

import (
	"crypto/ecdsa"

	"github.com/dgraph-io/badger"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

// Lachesis is a lachesis node implementation.
type Lachesis struct {
	host           string
	conf           *Config
	node           *posnode.Node
	nodeStore      *posnode.Store
	consensus      *posposet.Poset
	consensusStore *posposet.Store

	service
}

// New makes lachesis node.
// It does not start any process.
func New(db *badger.DB, host string, key *ecdsa.PrivateKey, conf Config, opts ...grpc.DialOption) *Lachesis {
	return makeLachesis(db, host, key, &conf, nil, opts...)
}

func makeLachesis(db *badger.DB, host string, key *ecdsa.PrivateKey, conf *Config, listen network.ListenFunc, opts ...grpc.DialOption) *Lachesis {
	ndb, cdb := makeStorages(db)

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
	}
}

// Start inits and starts whole lachesis node.
func (l *Lachesis) Start(genesis map[hash.Peer]uint64) {
	l.init(genesis)

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

func (l *Lachesis) init(genesis map[hash.Peer]uint64) {
	if err := l.consensusStore.ApplyGenesis(genesis); err != nil {
		panic(err)
	}
}

/*
 * Utils:
 */

func makeStorages(db *badger.DB) (*posnode.Store, *posposet.Store) {
	var (
		posetKVdb kvdb.Database
		nodeKVdb  kvdb.Database
	)
	if db == nil {
		posetKVdb = kvdb.NewMemDatabase()
		nodeKVdb = kvdb.NewMemDatabase()
	} else {
		db := kvdb.NewBadgerDatabase(db)
		posetKVdb = kvdb.NewTable(db, "p_")
		nodeKVdb = kvdb.NewTable(db, "n_")
	}

	return posnode.NewStore(nodeKVdb),
		posposet.NewStore(posetKVdb)
}
