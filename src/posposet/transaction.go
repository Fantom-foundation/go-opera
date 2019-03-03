package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

type (
	InternalTransaction struct {
		Amount   uint64
		Receiver common.Address
	}
)

/*
 * Poset's methods:
 */

func (p *Poset) applyTransactions(balances common.Hash, ordered Events) common.Hash {
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
