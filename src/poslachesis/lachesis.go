package lachesis

import (
	"crypto/ecdsa"

	"github.com/dgraph-io/badger"
	"google.golang.org/grpc"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posnode"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

// Lachesis is a lachesis node implementation.
type Lachesis struct {
	Consensus *posposet.Poset
	Node      *posnode.Node

	consensusStore *posposet.Store
	nodeStore      *posnode.Store
}

// New makes lachesis node.
// It does not start any process.
func New(db *badger.DB, host string, key *ecdsa.PrivateKey, opts ...grpc.DialOption) *Lachesis {
	dbNode, dbPoset := makeStorages(db)

	c := posposet.New(dbPoset, dbNode)
	n := posnode.New(host, key, dbNode, c, nil, opts...)

	return &Lachesis{
		Consensus: c,
		Node:      n,

		consensusStore: dbPoset,
		nodeStore:      dbNode,
	}
}

// Start inits and starts whole lachesis node.
func (l *Lachesis) Start(genesis map[hash.Peer]uint64) {
	l.init(genesis)

	l.Consensus.Start()
	l.Node.Start()
}

// Stop stops whole lachesis node.
func (l *Lachesis) Stop() {
	l.Node.Stop()
	l.Consensus.Stop()
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
