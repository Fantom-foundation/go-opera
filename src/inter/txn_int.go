package inter

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// InternalTransaction is for stake transfer.
type InternalTransaction struct {
	Nonce      idx.Txn
	Amount     pos.Stake
	Receiver   hash.Peer
	UntilBlock idx.Block
}

/*
 * Utils:
 */

// TransactionHashOf calcs hash of transaction.
func TransactionHashOf(sender hash.Peer, nonce idx.Txn) hash.Transaction {
	buf := append(sender.Bytes(), nonce.Bytes()...)
	return hash.Transaction(hash.Of(buf))
}
