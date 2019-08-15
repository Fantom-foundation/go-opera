package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// ApplyBlock execs ordered txns on state.
// TODO: replace with EVM transactions
func (s *Service) ApplyBlock(block *inter.Block, stateHash hash.Hash, members pos.Members) (hash.Hash, pos.Members) {
	db := s.store.StateDB(stateHash)
	for _, id := range block.Events {
		e := s.store.GetEvent(id)
		sender := e.Creator
		for _, tx := range e.InternalTransactions {
			receiver := tx.Receiver

			if db.FreeBalance(sender) < tx.Amount {
				s.Warnf("cannot send %d from %s to %s: balance is insufficient, skipped", tx.Amount, sender.String(), receiver.String())
				continue
			}

			if !db.Exist(receiver) {
				db.CreateAccount(receiver)
			}

			if tx.UntilBlock == 0 {
				s.Infof("transfer %d from %s to %s", tx.Amount, sender.String(), receiver.String())
				db.Transfer(sender, receiver, tx.Amount)
			} else {
				s.Infof("delegate %d from %s to %s for %d", tx.Amount, sender.String(), receiver.String(), tx.UntilBlock)
				db.Delegate(sender, receiver, tx.Amount, tx.UntilBlock)
			}

			members.Add(sender, db.VoteBalance(sender))
			members.Add(receiver, db.VoteBalance(receiver))
		}
	}
	stateHash, err := db.Commit(true)
	if err != nil {
		panic(err)
	}

	return stateHash, members
}
