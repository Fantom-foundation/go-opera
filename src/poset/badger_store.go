package poset

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/pos"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

const (
	participantPrefix   = "participant"
	rootSuffix          = "root"
	roundCreatedPrefix  = "roundCreated"
	roundReceivedPrefix = "roundReceived"
	topoPrefix          = "topo"
	blockPrefix         = "block"
	framePrefix         = "frame"
	statePrefix         = "state"
)

// BadgerStore struct for badger config data
type BadgerStore struct {
	participants *peers.Peers
	inmemStore   *InmemStore
	db           *badger.DB
	path         string
	needBoostrap bool

	states    state.Database
	stateRoot common.Hash
}

// NewBadgerStore creates a brand new Store with a new database
func NewBadgerStore(participants *peers.Peers, cacheSize int, path string, posConf *pos.Config) (*BadgerStore, error) {
	inmemStore := NewInmemStore(participants, cacheSize, posConf)
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false
	handle, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	store := &BadgerStore{
		participants: participants,
		inmemStore:   inmemStore,
		db:           handle,
		path:         path,
		states: state.NewDatabase(
			kvdb.NewTable(
				kvdb.NewBadgerDatabase(
					handle), statePrefix)),
	}
	if err := store.dbSetParticipants(participants); err != nil {
		return nil, err
	}
	if err := store.dbSetRoots(inmemStore.rootsByParticipant); err != nil {
		return nil, err
	}

	// TODO: replace with real genesis
	store.stateRoot, err = pos.FakeGenesis(participants, posConf, store.states)
	if err != nil {
		return nil, err
	}

	return store, nil
}

// LoadBadgerStore creates a Store from an existing database
func LoadBadgerStore(cacheSize int, path string) (*BadgerStore, error) {

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false
	handle, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	store := &BadgerStore{
		db:           handle,
		path:         path,
		needBoostrap: true,
		states: state.NewDatabase(
			kvdb.NewTable(
				kvdb.NewBadgerDatabase(
					handle), statePrefix)),
	}

	participants, err := store.dbGetParticipants()
	if err != nil {
		return nil, err
	}

	inmemStore := NewInmemStore(participants, cacheSize, nil)

	// read roots from db and put them in InmemStore
	roots := make(map[string]Root)
	for p := range participants.ByPubKey {
		root, err := store.dbGetRoot(p)
		if err != nil {
			return nil, err
		}
		roots[p] = root
	}

	if err := inmemStore.Reset(roots); err != nil {
		return nil, err
	}

	store.participants = participants
	store.inmemStore = inmemStore

	return store, nil
}

// LoadOrCreateBadgerStore load or create a new badger store
func LoadOrCreateBadgerStore(participants *peers.Peers, cacheSize int, path string, posConf *pos.Config) (*BadgerStore, error) {
	store, err := LoadBadgerStore(cacheSize, path)

	if err != nil {
		fmt.Println("Could not load store - creating new")
		store, err = NewBadgerStore(participants, cacheSize, path, posConf)

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

// ==============================================================================
// Keys

func topologicalEventKey(index int64) []byte {
	return []byte(fmt.Sprintf("%s_%09d", topoPrefix, index))
}

func participantKey(participant string) []byte {
	return []byte(fmt.Sprintf("%s_%s", participantPrefix, participant))
}

func participantEventKey(participant string, index int64) []byte {
	return []byte(fmt.Sprintf("%s__event_%09d", participant, index))
}

func participantRootKey(participant string) []byte {
	return []byte(fmt.Sprintf("%s_%s", participant, rootSuffix))
}

func roundCreatedKey(index int64) []byte {
	return []byte(fmt.Sprintf("%s_%09d", roundCreatedPrefix, index))
}

func roundReceivedKey(index int64) []byte {
	return []byte(fmt.Sprintf("%s_%09d", roundReceivedPrefix, index))
}
func blockKey(index int64) []byte {
	return []byte(fmt.Sprintf("%s_%09d", blockPrefix, index))
}

func frameKey(index int64) []byte {
	return []byte(fmt.Sprintf("%s_%09d", framePrefix, index))
}

/*
 * Store interface implementation:
 */

// TopologicalEvents returns event in topological order.
func (s *BadgerStore) TopologicalEvents() ([]Event, error) {
	var res []Event
	var evKey string
	t := int64(0)
	err := s.db.View(func(txn *badger.Txn) error {
		key := topologicalEventKey(t)
		item, errr := txn.Get(key)
		for errr == nil {
			errrr := item.Value(func(v []byte) error {
				evKey = string(v)
				return nil
			})
			if errrr != nil {
				break
			}

			eventItem, err := txn.Get([]byte(evKey))
			if err != nil {
				return err
			}
			err = eventItem.Value(func(eventBytes []byte) error {
				event := &Event{
					roundReceived:    RoundNIL,
					round:            RoundNIL,
					lamportTimestamp: LamportTimestampNIL,
				}

				if err := event.ProtoUnmarshal(eventBytes); err != nil {
					return err
				}
				res = append(res, *event)
				return nil
			})
			if err != nil {
				return err
			}

			t++
			key = topologicalEventKey(t)
			item, errr = txn.Get(key)
		}

		if !isDBKeyNotFound(errr) {
			return errr
		}

		return nil
	})

	return res, err
}

// CacheSize returns the cache size for the store
func (s *BadgerStore) CacheSize() int {
	return s.inmemStore.CacheSize()
}

// Participants returns all participants in the store
func (s *BadgerStore) Participants() (*peers.Peers, error) {
	return s.participants, nil
}

// RepertoireByPubKey gets PubKey map of peers
func (s *BadgerStore) RepertoireByPubKey() map[string]*peers.Peer {
	return s.inmemStore.RepertoireByPubKey()
}

// RepertoireByID gets ID map of peers
func (s *BadgerStore) RepertoireByID() map[uint64]*peers.Peer {
	return s.inmemStore.RepertoireByID()
}

// RootsBySelfParent returns the roots for the self parent
func (s *BadgerStore) RootsBySelfParent() (map[EventHash]Root, error) {
	return s.inmemStore.RootsBySelfParent()
}

// GetEventBlock get specific event block by hash
func (s *BadgerStore) GetEventBlock(hash EventHash) (event Event, err error) {
	// try to get it from cache
	event, err = s.inmemStore.GetEventBlock(hash)
	// if not in cache, try to get it from db
	if err != nil {
		event, err = s.dbGetEventBlock(hash)
	}
	return event, mapError(err, "Event", hash.String())
}

// SetEvent set a specific event
func (s *BadgerStore) SetEvent(event Event) error {
	// try to add it to the cache
	if err := s.inmemStore.SetEvent(event); err != nil {
		return err
	}
	// try to add it to the db
	return s.dbSetEvents([]Event{event})
}

// ParticipantEvents return all participant events
func (s *BadgerStore) ParticipantEvents(participant string, skip int64) (EventHashes, error) {
	res, err := s.inmemStore.ParticipantEvents(participant, skip)
	if err != nil {
		res, err = s.dbParticipantEvents(participant, skip)
	}
	return res, err
}

// ParticipantEvent get specific participant event
func (s *BadgerStore) ParticipantEvent(participant string, index int64) (EventHash, error) {
	result, err := s.inmemStore.ParticipantEvent(participant, index)
	if err != nil {
		result, err = s.dbParticipantEvent(participant, index)
	}
	return result, mapError(err, "ParticipantEvent", string(participantEventKey(participant, index)))
}

// LastEventFrom returns the last event for a particpant
func (s *BadgerStore) LastEventFrom(participant string) (last EventHash, isRoot bool, err error) {
	return s.inmemStore.LastEventFrom(participant)
}

// LastConsensusEventFrom returns the last consensus events for a participant
func (s *BadgerStore) LastConsensusEventFrom(participant string) (last EventHash, isRoot bool, err error) {
	return s.inmemStore.LastConsensusEventFrom(participant)
}

// KnownEvents returns all known events
func (s *BadgerStore) KnownEvents() map[uint64]int64 {
	known := make(map[uint64]int64)
	s.participants.RLock()
	defer s.participants.RUnlock()
	for p, pid := range s.participants.ByPubKey {
		index := int64(-1)
		last, isRoot, err := s.LastEventFrom(p)
		if err == nil {
			if isRoot {
				root, err := s.GetRoot(p)
				if err != nil {
					index = root.SelfParent.Index
				}
			} else {
				lastEvent, err := s.GetEventBlock(last)
				if err == nil {
					index = lastEvent.Index()
				}
			}

		}
		known[pid.ID] = index
	}
	return known
}

// ConsensusEvents returns all consensus events
func (s *BadgerStore) ConsensusEvents() EventHashes {
	return s.inmemStore.ConsensusEvents()
}

// ConsensusEventsCount returns the count for all known consensus events
func (s *BadgerStore) ConsensusEventsCount() int64 {
	return s.inmemStore.ConsensusEventsCount()
}

// AddConsensusEvent adds a consensus event to the store
func (s *BadgerStore) AddConsensusEvent(event Event) error {
	return s.inmemStore.AddConsensusEvent(event)
}

// GetRoundCreated gets the created round info for a given index
func (s *BadgerStore) GetRoundCreated(r int64) (RoundCreated, error) {
	res, err := s.inmemStore.GetRoundCreated(r)
	if err != nil {
		res, err = s.dbGetRoundCreated(r)
	}
	return res, mapError(err, "RoundCreated", string(roundCreatedKey(r)))
}

// SetRoundCreated sets the created round info for a given index
func (s *BadgerStore) SetRoundCreated(r int64, round RoundCreated) error {
	if err := s.inmemStore.SetRoundCreated(r, round); err != nil {
		return err
	}
	return s.dbSetRoundCreated(r, round)
}

// GetRoundReceived gets the received round for a given index
func (s *BadgerStore) GetRoundReceived(r int64) (RoundReceived, error) {
	res, err := s.inmemStore.GetRoundReceived(r)
	if err != nil {
		res, err = s.dbGetRoundReceived(r)
	}
	return res, mapError(err, "RoundReceived", string(roundReceivedKey(r)))
}

// SetRoundReceived sets the received round info for a given index
func (s *BadgerStore) SetRoundReceived(r int64, round RoundReceived) error {
	if err := s.inmemStore.SetRoundReceived(r, round); err != nil {
		return err
	}
	return s.dbSetRoundReceived(r, round)
}

// LastRound returns the last round for the store
func (s *BadgerStore) LastRound() int64 {
	return s.inmemStore.LastRound()
}

// RoundClothos returns all clothos for a round
func (s *BadgerStore) RoundClothos(r int64) EventHashes {
	round, err := s.GetRoundCreated(r)
	if err != nil {
		return EventHashes{}
	}
	return round.Clotho()
}

// RoundEvents returns all events for a round
func (s *BadgerStore) RoundEvents(r int64) int {
	round, err := s.GetRoundCreated(r)
	if err != nil {
		return 0
	}
	return len(round.Message.Events)
}

// GetRoot returns the root for a participant
func (s *BadgerStore) GetRoot(participant string) (Root, error) {
	root, err := s.inmemStore.GetRoot(participant)
	if err != nil {
		root, err = s.dbGetRoot(participant)
	}
	return root, mapError(err, "Root", string(participantRootKey(participant)))
}

// GetBlock returns the block for a given index
func (s *BadgerStore) GetBlock(rr int64) (Block, error) {
	res, err := s.inmemStore.GetBlock(rr)
	if err != nil {
		res, err = s.dbGetBlock(rr)
	}
	return res, mapError(err, "Block", string(blockKey(rr)))
}

// SetBlock add a block
func (s *BadgerStore) SetBlock(block Block) error {
	if err := s.inmemStore.SetBlock(block); err != nil {
		return err
	}
	return s.dbSetBlock(block)
}

// LastBlockIndex returns the last block index (height)
func (s *BadgerStore) LastBlockIndex() int64 {
	return s.inmemStore.LastBlockIndex()
}

// GetFrame returns a specific frame for the index
func (s *BadgerStore) GetFrame(rr int64) (Frame, error) {
	res, err := s.inmemStore.GetFrame(rr)
	if err != nil {
		res, err = s.dbGetFrame(rr)
	}
	return res, mapError(err, "Frame", string(frameKey(rr)))
}

// SetFrame add a frame
func (s *BadgerStore) SetFrame(frame Frame) error {
	if err := s.inmemStore.SetFrame(frame); err != nil {
		return err
	}
	return s.dbSetFrame(frame)
}

// Reset all roots
func (s *BadgerStore) Reset(roots map[string]Root) error {
	return s.inmemStore.Reset(roots)
}

// Close badger
func (s *BadgerStore) Close() error {
	if err := s.inmemStore.Close(); err != nil {
		return err
	}
	return s.db.Close()
}

// NeedBoostrap checks if bootstrapping is required
func (s *BadgerStore) NeedBoostrap() bool {
	return s.needBoostrap
}

// StorePath returns the path to the file on disk
func (s *BadgerStore) StorePath() string {
	return s.path
}

// StateDB returns state database
func (s *BadgerStore) StateDB() state.Database {
	return s.states
}

// StateRoot returns genesis state hash.
func (s *BadgerStore) StateRoot() common.Hash {
	return s.stateRoot
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// DB Methods

func (s *BadgerStore) dbGetEventBlock(hash EventHash) (Event, error) {
	var eventBytes []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash.Bytes())
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			eventBytes = val
			return nil
		})
		return err
	})

	if err != nil {
		return Event{}, err
	}

	event := new(Event)
	if err := event.ProtoUnmarshal(eventBytes); err != nil {
		return Event{}, err
	}

	return *event, nil
}

func (s *BadgerStore) dbSetEvents(events []Event) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()

	for _, event := range events {
		eventHash := event.Hash()
		val, err := event.ProtoMarshal()
		if err != nil {
			return err
		}
		// check if it already exists
		existent := false
		val2, err := tx.Get(eventHash.Bytes())
		if err != nil && !isDBKeyNotFound(err) {
			return err
		}
		if val2 != nil {
			existent = true
		}

		// insert [event hash] => [event bytes]
		if err := tx.Set(eventHash.Bytes(), val); err != nil {
			return err
		}

		if !existent {
			// insert [topo_index] => [event hash]
			topoKey := topologicalEventKey(event.Message.TopologicalIndex)
			if err := tx.Set(topoKey, eventHash.Bytes()); err != nil {
				return err
			}
			// insert [participant_index] => [event hash]
			peKey := participantEventKey(event.GetCreator(), event.Index())
			if err := tx.Set(peKey, eventHash.Bytes()); err != nil {
				return err
			}
		}
	}
	return tx.Commit(nil)
}

func (s *BadgerStore) dbParticipantEvents(participant string, skip int64) (res EventHashes, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		i := skip + 1
		key := participantEventKey(participant, i)
		item, errr := txn.Get(key)
		for errr == nil {
			errrr := item.Value(func(v []byte) error {
				var hash EventHash
				hash.Set(v)
				res = append(res, hash)
				return nil
			})
			if errrr != nil {
				break
			}

			i++
			key = participantEventKey(participant, i)
			item, errr = txn.Get(key)
		}

		if !isDBKeyNotFound(errr) {
			return errr
		}

		return nil
	})
	return
}

func (s *BadgerStore) dbParticipantEvent(participant string, index int64) (hash EventHash, err error) {
	key := participantEventKey(participant, index)

	err = s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			hash.Set(val)
			return nil
		})
		return err
	})

	return
}

func (s *BadgerStore) dbSetRoots(roots map[string]Root) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()
	for participant, root := range roots {
		val, err := root.ProtoMarshal()
		if err != nil {
			return err
		}
		key := participantRootKey(participant)
		// fmt.Println("Setting root", participant, "->", key)
		// insert [participant_root] => [root bytes]
		if err := tx.Set(key, val); err != nil {
			return err
		}
	}
	return tx.Commit(nil)
}

func (s *BadgerStore) dbSetRootEvents(roots map[string]Root) error {
	for participant, root := range roots {
		var creator []byte
		var selfParentHash EventHash
		selfParentHash.Set(root.SelfParent.Hash)
		fmt.Sscanf(participant, "0x%X", &creator)
		body := EventBody{
			Creator: creator,
			Index:   root.SelfParent.Index,
			Parents: EventHashes{EventHash{}, EventHash{}}.Bytes(), // make([][]byte, 2),
		}
		event := Event{
			Message: &EventMessage{
				Hash:             root.SelfParent.Hash,
				CreatorID:        root.SelfParent.CreatorID,
				TopologicalIndex: -1,
				Body:             &body,
				FlagTable:        FlagTable{selfParentHash: 1}.Marshal(),
				ClothoProof:      [][]byte{root.SelfParent.Hash},
			},
			lamportTimestamp: 0,
			round:            0,
			roundReceived:    0, /*RoundNIL*/
		}
		if err := s.SetEvent(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *BadgerStore) dbGetRoot(participant string) (Root, error) {
	var rootBytes []byte
	key := participantRootKey(participant)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			rootBytes = val
			return nil
		})
		return err
	})

	if err != nil {
		return Root{}, err
	}

	root := new(Root)
	if err := root.ProtoUnmarshal(rootBytes); err != nil {
		return Root{}, err
	}

	return *root, nil
}

func (s *BadgerStore) dbGetRoundCreated(index int64) (RoundCreated, error) {
	var roundBytes []byte
	key := roundCreatedKey(index)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			roundBytes = val
			return nil
		})
		return err
	})

	if err != nil {
		return *NewRoundCreated(), err
	}

	roundInfo := new(RoundCreated)
	if err := roundInfo.ProtoUnmarshal(roundBytes); err != nil {
		return *NewRoundCreated(), err
	}
	// In the current design, Queued field must be re-calculated every time for
	// each round. When retrieving a round info from a database, this field
	// should be ignored.
	roundInfo.Message.Queued = false

	return *roundInfo, nil
}

func (s *BadgerStore) dbSetRoundCreated(index int64, round RoundCreated) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()

	key := roundCreatedKey(index)
	val, err := round.ProtoMarshal()
	if err != nil {
		return err
	}

	// insert [round_index] => [round bytes]
	if err := tx.Set(key, val); err != nil {
		return err
	}

	return tx.Commit(nil)
}

func (s *BadgerStore) dbGetRoundReceived(index int64) (RoundReceived, error) {
	var roundBytes []byte
	key := roundReceivedKey(index)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			roundBytes = val
			return nil
		})
		return err
	})

	if err != nil {
		return *NewRoundReceived(), err
	}

	roundInfo := new(RoundReceived)
	if err := roundInfo.ProtoUnmarshal(roundBytes); err != nil {
		return *NewRoundReceived(), err
	}

	return *roundInfo, nil
}

func (s *BadgerStore) dbSetRoundReceived(index int64, round RoundReceived) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()

	key := roundReceivedKey(index)
	val, err := round.ProtoMarshal()
	if err != nil {
		return err
	}

	// insert [round_index] => [round bytes]
	if err := tx.Set(key, val); err != nil {
		return err
	}

	return tx.Commit(nil)
}

func (s *BadgerStore) dbGetParticipants() (*peers.Peers, error) {
	res := peers.NewPeers()

	err := s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(participantPrefix)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := string(item.Key())

			pubKey := k[len(participantPrefix)+1:]

			res.AddPeer(peers.NewPeer(pubKey, ""))
		}

		return nil
	})

	return res, err
}

func (s *BadgerStore) dbSetParticipants(participants *peers.Peers) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()

	participants.RLock()
	defer participants.RUnlock()
	for participant, id := range participants.ByPubKey {
		key := participantKey(participant)
		val := []byte(fmt.Sprint(id.ID))
		// insert [participant_participant] => [id]
		if err := tx.Set(key, val); err != nil {
			return err
		}
	}
	return tx.Commit(nil)
}

func (s *BadgerStore) dbGetBlock(index int64) (Block, error) {
	var blockBytes []byte
	key := blockKey(index)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			blockBytes = val
			return nil
		})
		return err
	})

	if err != nil {
		return Block{}, err
	}

	block := new(Block)
	if err := block.ProtoUnmarshal(blockBytes); err != nil {
		return Block{}, err
	}

	return *block, nil
}

func (s *BadgerStore) dbSetBlock(block Block) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()

	key := blockKey(block.Index())
	val, err := block.ProtoMarshal()
	if err != nil {
		return err
	}

	// insert [index] => [block bytes]
	if err := tx.Set(key, val); err != nil {
		return err
	}

	return tx.Commit(nil)
}

func (s *BadgerStore) dbGetFrame(index int64) (Frame, error) {
	var frameBytes []byte
	key := frameKey(index)
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			frameBytes = val
			return nil
		})
		return err
	})

	if err != nil {
		return Frame{}, err
	}

	frame := new(Frame)
	if err := frame.ProtoUnmarshal(frameBytes); err != nil {
		return Frame{}, err
	}

	return *frame, nil
}

func (s *BadgerStore) dbSetFrame(frame Frame) error {
	tx := s.db.NewTransaction(true)
	defer tx.Discard()

	key := frameKey(frame.Round)
	val, err := frame.ProtoMarshal()
	if err != nil {
		return err
	}

	// insert [index] => [block bytes]
	if err := tx.Set(key, val); err != nil {
		return err
	}

	return tx.Commit(nil)
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func isDBKeyNotFound(err error) bool {
	return err == badger.ErrKeyNotFound
}

func mapError(err error, name, key string) error {
	if err != nil {
		if isDBKeyNotFound(err) {
			return common.NewStoreErr(name, common.KeyNotFound, key)
		}
	}
	return err
}
