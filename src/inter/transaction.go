package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// InternalTransaction is for stake transfer.
type InternalTransaction struct {
	Amount   uint64
	Receiver hash.Peer
}

// InternalTransactionsToWire converts to wire.
func InternalTransactionsToWire(tt []*InternalTransaction) []*wire.InternalTransaction {
	if tt == nil {
		return nil
	}
	res := make([]*wire.InternalTransaction, len(tt))
	for i, t := range tt {
		res[i] = &wire.InternalTransaction{
			Amount:   t.Amount,
			Receiver: t.Receiver.Hex(),
		}
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
		res[i] = &InternalTransaction{
			Amount:   w.Amount,
			Receiver: hash.HexToPeer(w.Receiver),
		}
	}

	return res
}
