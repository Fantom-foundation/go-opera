package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// InternalTransaction is for stake transfer.
type InternalTransaction struct {
	Index      uint64
	Amount     uint64
	Receiver   hash.Peer
	UntilBlock uint64
}

// ToWire converts to wire.
func (tx *InternalTransaction) ToWire() *wire.InternalTransaction {
	return &wire.InternalTransaction{
		Index:      tx.Index,
		Amount:     tx.Amount,
		Receiver:   tx.Receiver.Hex(),
		UntilBlock: tx.UntilBlock,
	}
}

// WireToInternalTransaction converts from wire.
func WireToInternalTransaction(w *wire.InternalTransaction) *InternalTransaction {
	return &InternalTransaction{
		Index:      w.Index,
		Amount:     w.Amount,
		Receiver:   hash.HexToPeer(w.Receiver),
		UntilBlock: w.UntilBlock,
	}
}

// InternalTransactionsToWire converts to wire.
func InternalTransactionsToWire(tt []*InternalTransaction) []*wire.InternalTransaction {
	if tt == nil {
		return nil
	}
	res := make([]*wire.InternalTransaction, len(tt))
	for i, t := range tt {
		res[i] = t.ToWire()
	}

	return res
}

// WireToInternalTransactions converts from wire.
func WireToInternalTransactions(tt []*wire.InternalTransaction) []*InternalTransaction {
	if tt == nil {
		return nil
	}
	res := make([]*InternalTransaction, len(tt))
	for i, w := range tt {
		res[i] = WireToInternalTransaction(w)
	}

	return res
}
