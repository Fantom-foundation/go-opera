package poset

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
	"github.com/Fantom-foundation/go-lachesis/src/pos"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// InmemStore struct
type InmemStore struct {
	cacheSize              int
	participants           *peers.Peers
	eventCache             *lru.Cache           // hash => Event
	roundCreatedCache      *lru.Cache           // round number => RoundCreated
	roundReceivedCache     *lru.Cache           // round received number => RoundReceived
	blockCache             *lru.Cache           // index => Block
	frameCache             *lru.Cache           // round received => Frame
	consensusCache         *common.RollingIndex // consensus index => hash
	totConsensusEvents     int64
	repertoireByPubKey     map[string]*peers.Peer
	repertoireByID         map[uint64]*peers.Peer
	participantEventsCache *ParticipantEventsCache // pubkey => Events
	rootsByParticipant     map[string]Root         // [participant] => Root
	rootsBySelfParent      map[EventHash]Root      // [Root.SelfParent.Hash] => Root
	lastRound              int64
	lastConsensusEvents    map[string]EventHash // [participant] => hex() of last consensus event
	lastBlock              int64

	lastRoundLocker          sync.RWMutex
	lastBlockLocker          sync.RWMutex
	totConsensusEventsLocker sync.RWMutex

	states    state.Database
	stateRoot common.Hash
}

// NewInmemStore constructor
func NewInmemStore(participants *peers.Peers, cacheSize int, posConf *pos.Config) *InmemStore {
	rootsByParticipant := make(map[string]Root)

	participants.RLock()
	for pk, pid := range participants.ByPubKey {
		root := NewBaseRoot(pid.ID)
		rootsByParticipant[pk] = root
	}
	participants.RUnlock()

	eventCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.eventCache:", err)
		os.Exit(31)
	}
	roundCreatedCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.roundCreatedCache:", err)
		os.Exit(32)
	}
	roundReceivedCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.roundReceivedCache:", err)
		os.Exit(35)
	}
	blockCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.blockCache:", err)
		os.Exit(33)
	}
	frameCache, err := lru.New(cacheSize)
	if err != nil {
		fmt.Println("Unable to init InmemStore.frameCache:", err)
		os.Exit(34)
	}

	store := &InmemStore{
		cacheSize:              cacheSize,
		participants:           participants,
		eventCache:             eventCache,
		roundCreatedCache:      roundCreatedCache,
		roundReceivedCache:     roundReceivedCache,
		blockCache:             blockCache,
		frameCache:             frameCache,
		consensusCache:         common.NewRollingIndex("ConsensusCache", cacheSize),
		repertoireByPubKey:     make(map[string]*peers.Peer),
		repertoireByID:         make(map[uint64]*peers.Peer),
		participantEventsCache: NewParticipantEventsCache(cacheSize, participants),
		rootsByParticipant:     rootsByParticipant,
		lastRound:              -1,
		lastBlock:              -1,
		lastConsensusEvents:    map[string]EventHash{},
		states: state.NewDatabase(
			kvdb.NewTable(
				kvdb.NewMemDatabase(), statePrefix)),
	}

	participants.OnNewPeer(func(peer *peers.Peer) {
		root := NewBaseRoot(peer.ID)
		store.rootsByParticipant[peer.PubKeyHex] = root
		store.repertoireByPubKey[peer.PubKeyHex] = peer
		store.repertoireByID[peer.ID] = peer
		store.rootsBySelfParent = nil
		if _, err := store.RootsBySelfParent(); err != nil {
			panic(err)
		}
		old := store.participantEventsCache
		store.participantEventsCache = NewParticipantEventsCache(cacheSize, participants)
		store.participantEventsCache.Import(old)
	})

	store.setPeers(0, participants)
	if err = store.setLeafEvents(store.rootsByParticipant); err != nil {
		panic(err)
	}

	// TODO: replace with real genesis
	store.stateRoot, err = pos.FakeGenesis(participants, posConf, store.states)
	if err != nil {
		fmt.Println("Unable to init genesis state:", err)
		os.Exit(36)
	}

	return store
}

/*
 * Store interface implementation:
 */

// TopologicalEvents returns event in topological order.
func (s *InmemStore) TopologicalEvents() ([]Event, error) {
	// NOTE: it's used for bootstrap only, so is not implemented
	return nil, nil
}

// CacheSize size of cache
func (s *InmemStore) CacheSize() int {
	return s.cacheSize
}

// Participants returns participants
func (s *InmemStore) Participants() (*peers.Peers, error) {
	return s.participants, nil
}

func (s *InmemStore) setPeers(round int64, participants *peers.Peers) {
	// Extend ParticipantEventsCache and Roots with new peers
	participants.RLock()
	defer participants.RUnlock()
	for _, peer := range participants.ByID {
		s.repertoireByPubKey[peer.PubKeyHex] = peer
		s.repertoireByID[peer.ID] = peer
	}
}

// RepertoireByPubKey retrieves cached PubKey map of peers
func (s *InmemStore) RepertoireByPubKey() map[string]*peers.Peer {
	return s.repertoireByPubKey
}

// RepertoireByID retrieve cached ID map of peers
func (s *InmemStore) RepertoireByID() map[uint64]*peers.Peer {
	return s.repertoireByID
}

// RootsBySelfParent TODO
func (s *InmemStore) RootsBySelfParent() (map[EventHash]Root, error) {
	if s.rootsBySelfParent == nil {
		s.rootsBySelfParent = make(map[EventHash]Root)
		for _, root := range s.rootsByParticipant {
			var hash EventHash
			hash.Set(root.SelfParent.Hash)
			s.rootsBySelfParent[hash] = root
		}
	}
	return s.rootsBySelfParent, nil
}

// GetEventBlock gets specific event block by hash
func (s *InmemStore) GetEventBlock(hash EventHash) (Event, error) {
	res, ok := s.eventCache.Get(hash)
	if !ok {
		return Event{}, common.NewStoreErr("EventCache", common.KeyNotFound, hash.String())
	}

	return res.(Event), nil
}

// SetEvent set event for event block
func (s *InmemStore) SetEvent(event Event) error {
	eventHash := event.Hash()
	_, err := s.GetEventBlock(eventHash)
	if err != nil && !common.Is(err, common.KeyNotFound) {
		return err
	}
	if common.Is(err, common.KeyNotFound) {
		if err := s.addParticipantEvent(event.GetCreator(), eventHash, event.Index()); err != nil {
			return err
		}
	}

	// fmt.Println("Adding event to cache", event.Hex())
	s.eventCache.Add(eventHash, event)

	return nil
}

func (s *InmemStore) addParticipantEvent(participant string, hash EventHash, index int64) error {
	return s.participantEventsCache.Set(participant, hash, index)
}

// ParticipantEvents events for the participant
func (s *InmemStore) ParticipantEvents(participant string, skip int64) (EventHashes, error) {
	return s.participantEventsCache.Get(participant, skip)
}

// ParticipantEvent specific event
func (s *InmemStore) ParticipantEvent(participant string, index int64) (hash EventHash, err error) {
	hash, err = s.participantEventsCache.GetItem(participant, index)
	if err == nil {
		return
	}

	root, ok := s.rootsByParticipant[participant]
	if !ok {
		err = common.NewStoreErr("InmemStore.Roots", common.NoRoot, participant)
		return
	}

	if root.SelfParent.Index == index {
		hash.Set(root.SelfParent.Hash)
		err = nil
	}
	return
}

// LastEventFrom participant
func (s *InmemStore) LastEventFrom(participant string) (last EventHash, isRoot bool, err error) {
	// try to get the last event from this participant
	last, err = s.participantEventsCache.GetLast(participant)
	if err == nil || !common.Is(err, common.Empty) {
		return
	}
	// if there is none, grab the root
	if root, ok := s.rootsByParticipant[participant]; ok {
		last.Set(root.SelfParent.Hash)
		isRoot = true
		err = nil
	} else {
		err = common.NewStoreErr("InmemStore.Roots", common.NoRoot, participant)
	}
	return
}

// LastConsensusEventFrom participant
func (s *InmemStore) LastConsensusEventFrom(participant string) (last EventHash, isRoot bool, err error) {
	// try to get the last consensus event from this participant
	last, ok := s.lastConsensusEvents[participant]
	if ok {
		return
	}
	// if there is none, grab the root
	root, ok := s.rootsByParticipant[participant]
	if ok {
		last.Set(root.SelfParent.Hash)
		isRoot = true
	} else {
		err = common.NewStoreErr("InmemStore.Roots", common.NoRoot, participant)
	}

	return
}

// KnownEvents returns all known events
func (s *InmemStore) KnownEvents() map[uint64]int64 {
	known := s.participantEventsCache.Known()
	s.participants.RLock()
	defer s.participants.RUnlock()
	for p, pid := range s.participants.ByPubKey {
		if known[pid.ID] == -1 {
			root, ok := s.rootsByParticipant[p]
			if ok {
				known[pid.ID] = root.SelfParent.Index
			}
		}
	}
	return known
}

// ConsensusEvents returns all consensus events
func (s *InmemStore) ConsensusEvents() EventHashes {
	lastWindow, _ := s.consensusCache.GetLastWindow()
	res := make(EventHashes, len(lastWindow))
	for i, item := range lastWindow {
		res[i] = item.(EventHash)
	}
	return res
}

// ConsensusEventsCount returns count of all consnesus events
func (s *InmemStore) ConsensusEventsCount() int64 {
	s.totConsensusEventsLocker.RLock()
	defer s.totConsensusEventsLocker.RUnlock()
	return s.totConsensusEvents
}

// AddConsensusEvent to store
func (s *InmemStore) AddConsensusEvent(event Event) error {
	s.totConsensusEventsLocker.Lock()
	defer s.totConsensusEventsLocker.Unlock()
	err := s.consensusCache.Set(event.Hash(), s.totConsensusEvents)
	if err != nil {
		return err
	}
	s.totConsensusEvents++
	s.lastConsensusEvents[event.GetCreator()] = event.Hash()
	return nil
}

// GetRoundCreated retrieves created round by ID
func (s *InmemStore) GetRoundCreated(r int64) (RoundCreated, error) {
	res, ok := s.roundCreatedCache.Get(r)
	if !ok {
		return *NewRoundCreated(), common.NewStoreErr("RoundCreatedCache", common.KeyNotFound, strconv.FormatInt(r, 10))
	}
	return res.(RoundCreated), nil
}

// SetRoundCreated stores created round by ID
func (s *InmemStore) SetRoundCreated(r int64, round RoundCreated) error {
	s.lastRoundLocker.Lock()
	defer s.lastRoundLocker.Unlock()
	s.roundCreatedCache.Add(r, round)
	if r > s.lastRound {
		s.lastRound = r
	}
	return nil
}

// GetRoundReceived gets received round by ID
func (s *InmemStore) GetRoundReceived(r int64) (RoundReceived, error) {
	res, ok := s.roundReceivedCache.Get(r)
	if !ok {
		return *NewRoundReceived(), common.NewStoreErr("RoundReceivedCache", common.KeyNotFound, strconv.FormatInt(r, 10))
	}
	return res.(RoundReceived), nil
}

// SetRoundReceived stores received round by ID
func (s *InmemStore) SetRoundReceived(r int64, round RoundReceived) error {
	s.lastRoundLocker.Lock()
	defer s.lastRoundLocker.Unlock()
	s.roundReceivedCache.Add(r, round)
	if r > s.lastRound {
		s.lastRound = r
	}
	return nil
}

// LastRound getter
func (s *InmemStore) LastRound() int64 {
	s.lastRoundLocker.RLock()
	defer s.lastRoundLocker.RUnlock()
	return s.lastRound
}

// RoundClothos all clothos for the specified round
func (s *InmemStore) RoundClothos(r int64) EventHashes {
	round, err := s.GetRoundCreated(r)
	if err != nil {
		return EventHashes{}
	}
	return round.Clotho()
}

// RoundEvents returns events for the round
func (s *InmemStore) RoundEvents(r int64) int {
	round, err := s.GetRoundCreated(r)
	if err != nil {
		return 0
	}
	return len(round.Message.Events)
}

// GetRoot for participant
func (s *InmemStore) GetRoot(participant string) (Root, error) {
	res, ok := s.rootsByParticipant[participant]
	if !ok {
		return Root{}, common.NewStoreErr("RootCache", common.KeyNotFound, participant)
	}
	return res, nil
}

// GetBlock for index
func (s *InmemStore) GetBlock(index int64) (Block, error) {
	res, ok := s.blockCache.Get(index)
	if !ok {
		return Block{}, common.NewStoreErr("BlockCache", common.KeyNotFound, strconv.FormatInt(index, 10))
	}
	return res.(Block), nil
}

// SetBlock TODO
func (s *InmemStore) SetBlock(block Block) error {
	s.lastBlockLocker.Lock()
	defer s.lastBlockLocker.Unlock()
	index := block.Index()
	_, err := s.GetBlock(index)
	if err != nil && !common.Is(err, common.KeyNotFound) {
		return err
	}
	s.blockCache.Add(index, block)
	if index > s.lastBlock {
		s.lastBlock = index
	}
	return nil
}

// LastBlockIndex getter
func (s *InmemStore) LastBlockIndex() int64 {
	s.lastBlockLocker.RLock()
	defer s.lastBlockLocker.RUnlock()
	return s.lastBlock
}

// GetFrame by index
func (s *InmemStore) GetFrame(index int64) (Frame, error) {
	res, ok := s.frameCache.Get(index)
	if !ok {
		return Frame{}, common.NewStoreErr("FrameCache", common.KeyNotFound, strconv.FormatInt(index, 10))
	}
	return res.(Frame), nil
}

// SetFrame in the store
func (s *InmemStore) SetFrame(frame Frame) error {
	index := frame.Round
	_, err := s.GetFrame(index)
	if err != nil && !common.Is(err, common.KeyNotFound) {
		return err
	}
	s.frameCache.Add(index, frame)
	return nil
}

// Reset resets the store
func (s *InmemStore) Reset(roots map[string]Root) error {
	eventCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.eventCache:", errr)
		os.Exit(41)
	}
	roundCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.roundCreatedCache:", errr)
		os.Exit(42)
	}
	roundReceivedCache, errr := lru.New(s.cacheSize)
	if errr != nil {
		fmt.Println("Unable to reset InmemStore.roundReceivedCache:", errr)
		os.Exit(45)
	}
	// FIXIT: Should we recreate blockCache, frameCache and participantEventsCache here as well
	//        and reset lastConsensusEvents ?
	s.rootsByParticipant = roots
	s.rootsBySelfParent = nil
	s.eventCache = eventCache
	s.roundCreatedCache = roundCache
	s.roundReceivedCache = roundReceivedCache
	s.consensusCache = common.NewRollingIndex("ConsensusCache", s.cacheSize)
	err := s.participantEventsCache.Reset()
	s.lastRoundLocker.Lock()
	s.lastRound = -1
	s.lastRoundLocker.Unlock()
	s.lastBlockLocker.Lock()
	s.lastBlock = -1
	s.lastBlockLocker.Unlock()

	if _, err := s.RootsBySelfParent(); err != nil {
		return err
	}

	return err
}

// Close the store
func (s *InmemStore) Close() error {
	return nil
}

// NeedBootstrap for the store
func (s *InmemStore) NeedBootstrap() bool {
	return false
}

// StorePath getter
func (s *InmemStore) StorePath() string {
	return ""
}

// StateDB returns state database
func (s *InmemStore) StateDB() state.Database {
	return s.states
}

// StateRoot returns genesis state hash.
func (s *InmemStore) StateRoot() common.Hash {
	return s.stateRoot
}


func (s *InmemStore) setLeafEvents(roots map[string]Root) error {
	for participant, root := range roots {
		var creator []byte
		var selfParentHash EventHash
		selfParentHash.Set(root.SelfParent.Hash)
		if _, err := fmt.Sscanf(participant, "0x%X", &creator); err != nil {
			return err
		}
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
