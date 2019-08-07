package posposet

/*
"Fantom Proof of Stake FIP-2" implementation here.

"Special Purpose Vehicle" - a special smart contract acting as an internal
market-maker for FTG tokens, managing the collection of transaction fees
and the payment of all rewards.
*/

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// isEventValid validates event according to frame state.
func (p *Poset) isEventValid(e *Event, fInfo *FrameInfo) bool {
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
func (p *Poset) applyTransactions(db *state.DB, ordered inter.Events, members internal.Members) {
	for _, e := range ordered {
		sender := e.Creator
		for _, tx := range e.InternalTransactions {
			receiver := tx.Receiver

			if db.FreeBalance(sender) < tx.Amount {
				p.Warnf("cannot send %d from %s to %s: balance is insufficient, skipped", tx.Amount, sender.String(), receiver.String())
				continue
			}

			if !db.Exist(receiver) {
				db.CreateAccount(receiver)
			}

			if tx.UntilBlock == 0 {
				p.Infof("transfer %d from %s to %s", tx.Amount, sender.String(), receiver.String())
				db.Transfer(sender, receiver, tx.Amount)
			} else {
				p.Infof("delegate %d from %s to %s for %d", tx.Amount, sender.String(), receiver.String(), tx.UntilBlock)
				db.Delegate(sender, receiver, tx.Amount, tx.UntilBlock)
			}

			members.Add(sender, db.VoteBalance(sender))
			members.Add(receiver, db.VoteBalance(receiver))
		}
	}
}

// applyRewards calcs block rewards.
func (p *Poset) applyRewards(db *state.DB, ordered inter.Events, members internal.Members) {
	// TODO: implement it
}
