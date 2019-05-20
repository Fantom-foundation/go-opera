package state

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// temporary solution for delegation dict
type direction int

const (
	// TO delegation direction
	TO direction = iota
	// FROM delegation direction
	FROM
)

// Storage of entries.
type Storage map[hash.Hash]hash.Hash

func (self Storage) String() (str string) {
	for key, value := range self {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

// Copy creates a deep, independent copy of the state.
func (s Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range s {
		cpy[key] = value
	}

	return cpy
}

// stateObject represents an PoS account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	address  hash.Peer
	addrHash hash.Hash // hash of address of the account
	data     Account
	db       *DB

	// DB error.
	// State objects are used by the consensus core which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	originStorage Storage // Storage cache of original entries to dedup rewrites
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	suicided bool
	deleted  bool
}

// empty returns whether the account is considered empty.
func (s *stateObject) empty() bool {
	return s.data.Balance == 0
}

// newObject creates a state object.
func newObject(db *DB, address hash.Peer, data Account) *stateObject {
	return &stateObject{
		db:            db,
		address:       address,
		addrHash:      hash.Hash(address), //hash.Of(address.Bytes()),
		data:          data,
		originStorage: make(Storage),
		dirtyStorage:  make(Storage),
	}
}

// setError remembers the first non-nil error it is called with.
func (s *stateObject) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

func (s *stateObject) touch() {
	s.db.journal.append(touchChange{
		account: &s.address,
	})
}

func (s *stateObject) getTrie(db Database) Trie {
	if s.trie == nil {
		var err error
		s.trie, err = db.OpenStorageTrie(s.addrHash, s.data.Root())
		if err != nil {
			s.trie, _ = db.OpenStorageTrie(s.addrHash, hash.Hash{})
			s.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return s.trie
}

// GetState retrieves a value from the account storage trie.
func (s *stateObject) GetState(db Database, key hash.Hash) hash.Hash {
	// If we have a dirty value for this state entry, return it
	value, dirty := s.dirtyStorage[key]
	if dirty {
		return value
	}
	// Otherwise return the entry's original value
	return s.GetCommittedState(db, key)
}

// GetCommittedState retrieves a value from the committed account storage trie.
func (s *stateObject) GetCommittedState(db Database, key hash.Hash) hash.Hash {
	// If we have the original value cached, return that
	value, cached := s.originStorage[key]
	if cached {
		return value
	}
	// Otherwise load the value from the database
	enc, err := s.getTrie(db).TryGet(key.Bytes())
	if err != nil {
		s.setError(err)
		return hash.Hash{}
	}
	if len(enc) > 0 {
		value.SetBytes(enc)
	}
	s.originStorage[key] = value
	return value
}

// SetState updates a value in account storage.
func (s *stateObject) SetState(db Database, key, value hash.Hash) {
	// If the new value is the same as old, don't set
	prev := s.GetState(db, key)
	if prev == value {
		return
	}
	// New value is different, update and journal the change
	s.db.journal.append(storageChange{
		account:  &s.address,
		key:      key,
		prevalue: prev,
	})
	s.setState(key, value)
}

func (s *stateObject) setState(key, value hash.Hash) {
	s.dirtyStorage[key] = value
}

// updateTrie writes cached storage modifications into the object's storage trie.
func (s *stateObject) updateTrie(db Database) Trie {
	tr := s.getTrie(db)
	for key, value := range s.dirtyStorage {
		delete(s.dirtyStorage, key)

		// Skip noop changes, persist actual changes
		if value == s.originStorage[key] {
			continue
		}
		s.originStorage[key] = value

		if (value == hash.Hash{}) {
			s.setError(tr.TryDelete(key.Bytes()))
			continue
		}

		s.setError(tr.TryUpdate(key.Bytes(), value.Bytes())) // bytes.TrimLeft(value)?
	}
	return tr
}

// UpdateRoot sets the trie root to the current root hash of
func (s *stateObject) updateRoot(db Database) {
	s.updateTrie(db)
	s.data.SetRoot(s.trie.Hash())
}

// CommitTrie the storage trie of the object to db.
// This updates the trie root.
func (s *stateObject) CommitTrie(db Database) error {
	s.updateTrie(db)
	if s.dbErr != nil {
		return s.dbErr
	}
	root, err := s.trie.Commit(nil)
	if err == nil {
		s.data.SetRoot(root)
	}
	return err
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (s *stateObject) AddBalance(amount uint64) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount == 0 {
		if s.empty() {
			s.touch()
		}

		return
	}
	s.SetBalance(s.data.Balance + amount)
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (s *stateObject) SubBalance(amount uint64) {
	if amount == 0 {
		return
	}
	if s.data.Balance < amount {
		panic("balance must be positive")
	}
	s.SetBalance(s.data.Balance - amount)
}

// SetBalance sets balance amount.
func (s *stateObject) SetBalance(amount uint64) {
	s.db.journal.append(balanceChange{
		account: &s.address,
		prev:    s.data.Balance,
	})
	s.data.Balance = amount
}

// DelegateTo writes data about delegation.
func (s *stateObject) DelegateTo(addr hash.Peer, amount int64, until uint64) {
	if addr == s.address || amount == 0 || until < 1 {
		panic("Impossible delegation!")
	}
	// TODO: move this check to upper level
	if amount > 0 && uint64(amount) >= s.FreeBalance() {
		panic("Too many for delegate to!")
	}
	if amount < 0 && s.data.DelegatedFrom+uint64(-1*amount) >= 15*s.FreeBalance() {
		panic("Too many for delegate from!")
	}

	s.db.journal.append(delegationChange{
		account: &s.address,
		addr:    addr,
		amount:  amount,
		until:   until,
	})
	s.delegateTo(addr, amount, until, false)
}

func (s *stateObject) delegateTo(addr hash.Peer, amount int64, until uint64, reverse bool) {
	if s.data.DelegatingTo == nil {
		s.data.DelegatingTo = make(map[string]*Borrow)
	}
	if s.data.DelegatingFrom == nil {
		s.data.DelegatingFrom = make(map[string]*Borrow)
	}

	modify := func(delegating map[string]*Borrow, total *uint64, amount uint64) {
		rec, ok := delegating[addr.Hex()]
		if !ok {
			rec = &Borrow{
				Recs: make(map[uint64]uint64),
			}
			delegating[addr.Hex()] = rec
		}

		if !reverse {
			rec.Recs[until] += amount
			*total += amount
		} else {
			rec.Recs[until] -= amount
			*total -= amount

			if rec.Recs[until] == 0 {
				delete(rec.Recs, until)
			}
			if len(rec.Recs) < 1 {
				delete(delegating, addr.Hex())
			}
		}
	}

	if amount > 0 {
		modify(s.data.DelegatingTo, &s.data.DelegatedTo, uint64(amount))
	} else {
		modify(s.data.DelegatingFrom, &s.data.DelegatedFrom, uint64(-1*amount))
	}
}

// ExpireDelegations erases data about expired delegations.
func (s *stateObject) ExpireDelegations(now uint64) {
	s.db.journal.append(expirationChange{
		account: &s.address,
		deleted: s.removeDelegations(now),
	})
}

func (s *stateObject) removeDelegations(now uint64) (deleted [2]map[string]map[uint64]uint64) {
	modify := func(x direction, delegating map[string]*Borrow, total *uint64) {
		deleted[x] = make(map[string]map[uint64]uint64)

		for addr, rec := range delegating {
			for until, amount := range rec.Recs {
				if until > now {
					continue
				}
				*total -= amount
				delete(rec.Recs, until)

				if deleted[x][addr] == nil {
					deleted[x][addr] = make(map[uint64]uint64)
				}
				deleted[x][addr][until] = amount
			}
			if len(rec.Recs) < 1 {
				delete(delegating, addr)
			}
		}
	}

	modify(TO, s.data.DelegatingTo, &s.data.DelegatedTo)
	modify(FROM, s.data.DelegatingFrom, &s.data.DelegatedFrom)
	return
}

func (s *stateObject) addDelegations(dd [2]map[string]map[uint64]uint64) {
	modify := func(x direction, delegating map[string]*Borrow, total *uint64) {
		for addr, rec := range dd[x] {
			if delegating[addr] == nil {
				delegating[addr] = &Borrow{
					Recs: make(map[uint64]uint64),
				}
			}
			for until, amount := range rec {
				delegating[addr].Recs[until] = amount
				*total += amount
			}
		}
	}

	modify(TO, s.data.DelegatingTo, &s.data.DelegatedTo)
	modify(FROM, s.data.DelegatingFrom, &s.data.DelegatedFrom)
}

func (s *stateObject) GetDelegations() (dd [2]map[hash.Peer]uint64) {
	get := func(x direction, delegating map[string]*Borrow) {
		dd[x] = make(map[hash.Peer]uint64)

		for addr, rec := range delegating {
			h := hash.HexToPeer(addr)
			for _, amount := range rec.Recs {
				dd[x][h] += amount
			}
		}
	}

	get(TO, s.data.DelegatingTo)
	get(FROM, s.data.DelegatingFrom)
	return
}

func (s *stateObject) deepCopy(db *DB) *stateObject {
	stateObject := newObject(db, s.address, s.data)
	if s.trie != nil {
		stateObject.trie = db.db.CopyTrie(s.trie)
	}
	stateObject.dirtyStorage = s.dirtyStorage.Copy()
	stateObject.originStorage = s.originStorage.Copy()
	stateObject.suicided = s.suicided
	stateObject.deleted = s.deleted
	return stateObject
}

/*
 * Attribute accessors
 */

// Returns the address of the contract/account.
func (s *stateObject) Address() hash.Peer {
	return s.address
}

// FreeBalance returns free balance.
func (s *stateObject) FreeBalance() uint64 {
	return s.data.Balance - s.data.DelegatedTo
}

// VoteBalance returns vote balance.
func (s *stateObject) VoteBalance() uint64 {
	return s.data.Balance + s.data.DelegatedFrom - s.data.DelegatedTo
}

// Data returns data.
func (s *stateObject) Data() *Account {
	return &s.data
}
