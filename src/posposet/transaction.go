package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

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
				log.Warnf("Cannot send %d from %s to %s: balance is insufficient, skipped", tx.Amount, sender.String(), receiver.String())
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
