package posnode

import (
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// SetNonce stores nonce for the specified node.
func (s *Store) SetNonce(index uint64, creator hash.Peer) {
	if err := s.Nonce.Put(creator.Bytes(), intToBytes(index)); err != nil {
		panic(err)
	}
}

// GetNonce returns nonce for the specified node.
func (s *Store) GetNonce(creator hash.Peer) uint64 {
	buf, err := s.Nonce.Get(creator.Bytes())
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return 0
	}

	return bytesToInt(buf)
}

// IncreaseNonce increases nonce value for the specified node depending on the number of transactions.
func (s *Store) IncreaseNonce(creator hash.Peer, txs []*inter.InternalTransaction) (fromIndex uint64, toIndex uint64) {
	txCount := uint64(len(txs))
	fromIndex = s.GetNonce(creator)
	toIndex = fromIndex + txCount
	if txCount < 1 {
		return
	}

	s.SetNonce(toIndex, creator)
	return
}

// SetNonceEvent accepts transaction and stores the event
// for the specified node, depending on the nonce value and creator.
func (s *Store) SetNonceEvent(index uint64, e *inter.Event) {
	if e == nil {
		return
	}
	key := append(e.Creator.Bytes(), intToBytes(index)...)
	s.set(s.NonceEvents, key, e.ToWire())
}

// SetBatchNonceEvent accepts multiple transactions and batch stores the event
// for the specified node, depending on the nonce values and creator.
func (s *Store) SetBatchNonceEvent(fromIndex uint64, e *inter.Event) {
	txs := e.InternalTransactions
	if e == nil || len(txs) < 1 {
		return
	}

	batch := s.NonceEvents.NewBatch()
	defer batch.Reset()

	w := e.ToWire()
	for i, tx := range txs {
		if tx == nil {
			continue
		}

		var pbf proto.Buffer
		if err := pbf.Marshal(w); err != nil {
			panic(err)
		}
		key := append(e.Creator.Bytes(), intToBytes(fromIndex+uint64(i))...)
		if err := batch.Put(key, pbf.Bytes()); err != nil {
			panic(err)
		}
	}

	if err := batch.Write(); err != nil {
		panic(err)
	}
}

// GetNonceEvent returns the event
// for the specified node, depending on the nonce value and creator.
func (s *Store) GetNonceEvent(index uint64, creator hash.Peer) (e *inter.Event) {
	w := s.GetNonceWireEvent(index, creator)
	if w != nil {
		e = inter.WireToEvent(w)
	}
	return
}

// GetNonceEvent returns the wire event
// for the specified node, depending on the nonce value and creator.
// As a result, this function returns a serialized protobuf message.
func (s *Store) GetNonceWireEvent(index uint64, creator hash.Peer) (w *wire.Event) {
	key := append(creator.Bytes(), intToBytes(index)...)
	if s.has(s.NonceEvents, key) {
		w = s.get(s.NonceEvents, key, &wire.Event{}).(*wire.Event)
	}
	return
}

// SetNonceTx accepts the transaction and stores it,
// as a key are used the nonce value and creator.
func (s *Store) SetNonceTx(index uint64, creator hash.Peer, tx *inter.InternalTransaction) {
	if tx == nil {
		return
	}
	key := append(creator.Bytes(), intToBytes(index)...)
	s.set(s.NonceTxs, key, tx.ToWire())
}

// SetBatchNonceTx accepts multiple transactions and batch stores them,
// as a key are used the nonce values and creator.
func (s *Store) SetBatchNonceTx(fromIndex uint64, creator hash.Peer, txs []*inter.InternalTransaction) {
	if len(txs) < 1 {
		return
	}

	batch := s.NonceTxs.NewBatch()
	defer batch.Reset()

	for i, tx := range txs {
		if tx == nil {
			continue
		}
		var pbf proto.Buffer
		w := tx.ToWire()
		if err := pbf.Marshal(w); err != nil {
			panic(err)
		}
		key := append(creator.Bytes(), intToBytes(fromIndex+uint64(i))...)
		if err := batch.Put(key, pbf.Bytes()); err != nil {
			panic(err)
		}
	}

	if err := batch.Write(); err != nil {
		panic(err)
	}
}

// GetNonceTx returns the transaction
// for the specified node, depending on the nonce value and creator.
func (s *Store) GetNonceTx(index uint64, creator hash.Peer) (tx *inter.InternalTransaction) {
	w := s.GetNonceWireTx(index, creator)
	if w != nil {
		tx = inter.WireToInternalTransaction(w)
	}
	return
}

// GetNonceWireTx returns the wire transaction
// for the specified node, depending on the nonce value and creator.
// As a result, this function returns a serialized protobuf message.
func (s *Store) GetNonceWireTx(index uint64, creator hash.Peer) (tx *wire.InternalTransaction) {
	key := append(creator.Bytes(), intToBytes(index)...)
	if s.has(s.NonceTxs, key) {
		tx = s.get(s.NonceTxs, key, &wire.InternalTransaction{}).(*wire.InternalTransaction)
	}
	return
}
