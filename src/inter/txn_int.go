package inter

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// InternalTransaction is for stake transfer.
type InternalTransaction struct {
	Nonce      idx.Txn
	Amount     pos.Stake
	Receiver   common.Address
	UntilBlock idx.Block
}

/*
 * Utils:
 */

// TransactionHashOf calcs hash of transaction.
func TransactionHashOf(sender common.Address, nonce idx.Txn) hash.Transaction {
	buf := append(sender.Bytes(), nonce.Bytes()...)
	return hash.Transaction(hash.Of(buf))
}
