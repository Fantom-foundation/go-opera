package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
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
			Receiver: t.Receiver.Bytes(),
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
			Receiver: hash.BytesToPeer(w.Receiver),
		}
	}

	return res
}

/*
 * Poset's methods:
 */

func (p *Poset) applyTransactions(balances hash.Hash, ordered Events) hash.Hash {
	db := p.store.StateDB(balances)

	for _, e := range ordered {
		sender := e.Creator
		for _, tx := range e.InternalTransactions {
			receiver := tx.Receiver
			if db.GetBalance(sender) < tx.Amount {
				log.Warnf("Cann't send %d from %s to %s: balance is not enough, skipped", tx.Amount, sender.String(), receiver.String())
				continue
			}
			if !db.Exist(receiver) {
				db.CreateAccount(receiver)
			}
			db.SubBalance(sender, tx.Amount)
			db.AddBalance(receiver, tx.Amount)
		}
	}

	newState, err := db.Commit(true)
	if err != nil {
		panic(err)
	}

	return newState
}
