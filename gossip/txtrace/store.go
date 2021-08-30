package txtrace

import (
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
)

// Store is a transaction traces persistent storage working over physical key-value database.
type Store struct {
	mainDB kvdb.Store
	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDB kvdb.Store) *Store {
	s := &Store{
		mainDB:   mainDB,
		Instance: logger.MakeInstance(),
	}
	s.SetName("TxTrace Store")
	return s
}

// Close closes underlying database.
func (s *Store) Close() {
	_ = s.mainDB.Close()
}

// SetTx stores []byte representation of transaction traces.
func (s *Store) SetTxTrace(txID common.Hash, txTraces []byte) error {
	return s.mainDB.Put(txID.Bytes(), txTraces)
}

// GetTx returns stored transaction traces.
func (s *Store) GetTx(txID common.Hash) []byte {

	buf, err := s.mainDB.Get(txID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}
	return buf
}
