package genesisstore

import (
	"io"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
)

func BlocksSection(i int) string {
	return getSectionName("brs", i)
}

func EpochsSection(i int) string {
	return getSectionName("ers", i)
}

func EvmSection(i int) string {
	return getSectionName("evm", i)
}

type FilesMap func(string) (io.Reader, error)

// Store is a node persistent storage working over a physical zip archive.
type Store struct {
	fMap  FilesMap
	head  genesis.Header
	close func() error

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(fMap FilesMap, head genesis.Header, close func() error) *Store {
	return &Store{
		fMap:     fMap,
		head:     head,
		close:    close,
		Instance: logger.New("genesis-store"),
	}
}

// Close leaves underlying database.
func (s *Store) Close() error {
	s.fMap = nil
	return s.close()
}
