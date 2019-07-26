package posnode

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// SetTxnsEvent stores txn-to-event index.
func (s *Store) SetTxnsEvent(e hash.Event, sender hash.Peer, txns ...*inter.InternalTransaction) {
	batch := s.table.Txn2Event.NewBatch()
	defer batch.Reset()

	for _, txn := range txns {
		idx := inter.TransactionHashOf(sender, txn.Nonce)
		if err := batch.Put(idx.Bytes(), e.Bytes()); err != nil {
			s.Fatal(err)
		}
	}

	if err := batch.Write(); err != nil {
		s.Fatal(err)
	}
}

// GetTxnsEvent returns event includes the specified txn.
func (s *Store) GetTxnsEvent(idx hash.Transaction) *inter.Event {
	buf, err := s.table.Txn2Event.Get(idx.Bytes())
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	e := hash.BytesToEvent(buf)

	return s.GetEvent(e)
}
