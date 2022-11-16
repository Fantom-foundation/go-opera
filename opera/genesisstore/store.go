package genesisstore

import (
	"io"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore/state"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/genesis"

	"github.com/ledgerwatch/erigon-lib/kv"
)

const (
	BlocksSection = "brs"
	EpochsSection = "ers"
	EvmSection    = "evm"
)

type FilesMap func(string) (io.Reader, error)

// Store is a node persistent storage working over a physical zip archive.
type Store struct {
	fMap    FilesMap
	head    genesis.Header
	close   func() error
	chainKV kv.RwDB
	statedb *state.StateDB

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(fMap FilesMap, head genesis.Header, close func() error, chainKV kv.RwDB, statedb *state.StateDB) *Store {
	return &Store{
		fMap:     fMap,
		head:     head,
		close:    close,
		chainKV:  chainKV,
		statedb:  statedb,
		Instance: logger.New("genesis-store"),
	}
}

// Close leaves underlying database.
func (s *Store) Close() error {
	s.fMap = nil
	return s.close()
}

func (s *Store) ChainKV() kv.RwDB {
	return s.chainKV
}

func (s *Store) StateDB() *state.StateDB {
	return s.statedb
}
