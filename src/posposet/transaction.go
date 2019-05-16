package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// TODO: validate txns
func applyTransactions(db *state.DB, ordered Events) {
	for _, e := range ordered {
		sender := e.Creator
		for _, tx := range e.InternalTransactions {
			receiver := tx.Receiver

			if db.FreeBalance(sender) < tx.Amount {
				log.Warnf("Cann't send %d from %s to %s: balance is not enough, skipped", tx.Amount, sender.String(), receiver.String())
				continue
			}

			if !db.Exist(receiver) {
				db.CreateAccount(receiver)
			}

			if tx.UntilBlock == 0 {
				db.Transfer(sender, receiver, tx.Amount)
			} else {
				db.Delegate(sender, receiver, tx.Amount, tx.UntilBlock)
			}
		}
	}
}
