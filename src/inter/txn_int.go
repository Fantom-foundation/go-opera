package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

// InternalTransaction is for stake transfer.
type InternalTransaction struct {
	Nonce      idx.Txn
	Amount     Stake
	Receiver   hash.Peer
	UntilBlock idx.Block
}

// ToWire converts to wire.
func (tx *InternalTransaction) ToWire() *wire.InternalTransaction {
	if tx == nil {
		return nil
	}
	return &wire.InternalTransaction{
		Nonce:      uint64(tx.Nonce),
		Amount:     uint64(tx.Amount),
		Receiver:   tx.Receiver.Hex(),
		UntilBlock: uint64(tx.UntilBlock),
	}
}

// WireToInternalTransaction converts from wire.
func WireToInternalTransaction(w *wire.InternalTransaction) *InternalTransaction {
	if w == nil {
		return nil
	}
	return &InternalTransaction{
		Nonce:      idx.Txn(w.Nonce),
		Amount:     Stake(w.Amount),
		Receiver:   hash.HexToPeer(w.Receiver),
		UntilBlock: idx.Block(w.UntilBlock),
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

/*
 * Utils:
 */

// TransactionHashOf calcs hash of transaction.
func TransactionHashOf(sender hash.Peer, nonce idx.Txn) hash.Transaction {
	buf := append(sender.Bytes(), nonce.Bytes()...)
	return hash.Transaction(hash.Of(buf))
}
