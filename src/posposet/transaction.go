package posposet

/*
"Fantom Proof of Stake FIP-2" implementation here.

"Special Purpose Vehicle" - a special smart contract acting as an internal
market-maker for FTG tokens, managing the collection of transaction fees
and the payment of all rewards.
*/

import (
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// isEventValid validates event according to frame state.
func (p *Poset) isEventValid(e *Event, f *Frame) bool {
	// NOTE: issue
	//  a) if e.txns change f.Balances we will need to reconsensus all (but we need );
	//  b) if e.txns dont change f.Balances we will get invalid sequence for valid events;
	//db := p.store.StateDB(frame.Balances)
	// TODO: solution. How about b) + fine of node's invalid txns later (at applyTransactions())?
	return true
}

// applyTransactions execs ordered txns on state.
// TODO: fine of invalid txns
// TODO: transaction fees
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

// applyRewards calcs block rewards.
func applyRewards(db *state.DB, ordered Events) {
	// TODO: implement it
}
