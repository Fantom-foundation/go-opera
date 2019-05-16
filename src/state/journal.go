package state

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// journalEntry is a modification entry in the state change journal that can be
// reverted on demand.
type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(*DB)

	// dirtied returns the address modified by this journal entry.
	dirtied() *hash.Peer
}

// journal contains the list of state modifications applied since the last state
// commit. These are tracked to be able to be reverted in case of an execution
// exception or revertal request.
type journal struct {
	entries []journalEntry    // Current changes tracked by the journal
	dirties map[hash.Peer]int // Dirty accounts and the number of changes
}

// newJournal create a new initialized journal.
func newJournal() *journal {
	return &journal{
		dirties: make(map[hash.Peer]int),
	}
}

// append inserts a new modification entry to the end of the change journal.
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if addr := entry.dirtied(); addr != nil {
		j.dirties[*addr]++
	}
}

// revert undoes a batch of journalled modifications along with any reverted
// dirty handling too.
func (j *journal) revert(statedb *DB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		// Undo the changes made by the operation
		j.entries[i].revert(statedb)

		// Drop any dirty tracking induced by the change
		if addr := j.entries[i].dirtied(); addr != nil {
			if j.dirties[*addr]--; j.dirties[*addr] == 0 {
				delete(j.dirties, *addr)
			}
		}
	}
	j.entries = j.entries[:snapshot]
}

// dirty explicitly sets an address to dirty, even if the change entries would
// otherwise suggest it as clean. This method is an ugly hack to handle the RIPEMD
// precompile consensus exception.
func (j *journal) dirty(addr hash.Peer) {
	j.dirties[addr]++
}

// length returns the current number of entries in the journal.
func (j *journal) length() int {
	return len(j.entries)
}

type (
	// Changes to the account trie.
	createObjectChange struct {
		account *hash.Peer
	}
	resetObjectChange struct {
		prev *stateObject
	}
	suicideChange struct {
		account  *hash.Peer
		prev     bool // whether account had already suicide
		prevData Account
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *hash.Peer
		prev    uint64
	}
	storageChange struct {
		account       *hash.Peer
		key, prevalue hash.Hash
	}
	delegationChange struct {
		account *hash.Peer
		addr    hash.Peer
		amount  int64
		until   uint64
	}
	expirationChange struct {
		account           *hash.Peer
		prevDelegatedTo   uint64
		prevDelegatedFrom uint64
		deleted           [2]map[string]map[uint64]uint64
	}

	// Changes to other state values.
	addPreimageChange struct {
		hash hash.Hash
	}
	touchChange struct {
		account *hash.Peer
	}
)

func (ch createObjectChange) revert(s *DB) {
	delete(s.stateObjects, *ch.account)
	delete(s.stateObjectsDirty, *ch.account)
}

func (ch createObjectChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch resetObjectChange) revert(s *DB) {
	s.setStateObject(ch.prev)
}

func (ch resetObjectChange) dirtied() *hash.Peer {
	return nil
}

func (ch suicideChange) revert(s *DB) {
	obj := s.getStateObject(*ch.account)
	if obj != nil {
		obj.suicided = ch.prev
		obj.data = ch.prevData
	}
}

func (ch suicideChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch touchChange) revert(s *DB) {
}

func (ch touchChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch balanceChange) revert(s *DB) {
	s.getStateObject(*ch.account).data.Balance = ch.prev
}

func (ch balanceChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch storageChange) revert(s *DB) {
	s.getStateObject(*ch.account).setState(ch.key, ch.prevalue)
}

func (ch storageChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch delegationChange) revert(s *DB) {
	s.getStateObject(*ch.account).delegateTo(ch.addr, ch.amount, ch.until, true)
}

func (ch delegationChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch expirationChange) revert(s *DB) {
	s.getStateObject(*ch.account).addDelegations(ch.deleted)
}

func (ch expirationChange) dirtied() *hash.Peer {
	return ch.account
}

func (ch addPreimageChange) revert(s *DB) {
	delete(s.preimages, ch.hash)
}

func (ch addPreimageChange) dirtied() *hash.Peer {
	return nil
}
