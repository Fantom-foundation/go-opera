package poset

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

// Key struct
type Key struct {
	x EventHash
	y EventHash
}

// ToString converts key to string
func (k Key) ToString() string {
	return fmt.Sprintf("{%s, %s}", k.x, k.y)
}

// ParentRoundInfo struct
type ParentRoundInfo struct {
	round   int
	isRoot  bool
	Atropos int
}

// NewBaseParentRoundInfo constructor
func NewBaseParentRoundInfo() ParentRoundInfo {
	return ParentRoundInfo{
		round:  -1,
		isRoot: false,
	}
}

// ------------------------------------------------------------------------------

// ParticipantEventsCache struct
type ParticipantEventsCache struct {
	participants *peers.Peers
	rim          *common.RollingIndexMap
}

// NewParticipantEventsCache constructor
func NewParticipantEventsCache(size int, participants *peers.Peers) *ParticipantEventsCache {
	return &ParticipantEventsCache{
		participants: participants,
		rim:          common.NewRollingIndexMap("ParticipantEvents", size, participants.ToIDSlice()),
	}
}

// AddPeer adds peer to cache and rolling index map, returns error if it failed to add to map
func (pec *ParticipantEventsCache) AddPeer(peer *peers.Peer) error {
	pec.participants.AddPeer(peer)
	return pec.rim.AddKey(peer.ID)
}

func (pec *ParticipantEventsCache) participantID(participant string) (uint64, error) {
	peer, ok := pec.participants.ReadByPubKey(participant)

	if !ok {
		return peers.PeerNIL, common.NewStoreErr("ParticipantEvents", common.UnknownParticipant, participant)
	}

	return peer.ID, nil
}

// Get return participant events with index > skip
func (pec *ParticipantEventsCache) Get(participant string, skipIndex int64) (EventHashes, error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return EventHashes{}, err
	}

	pe, err := pec.rim.Get(id, skipIndex)
	if err != nil {
		return EventHashes{}, err
	}

	res := make(EventHashes, len(pe))
	for k := 0; k < len(pe); k++ {
		res[k].Set(pe[k].([]byte))
	}
	return res, nil
}

// GetItem get event for participant at index
func (pec *ParticipantEventsCache) GetItem(participant string, index int64) (hash EventHash, err error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return
	}

	item, err := pec.rim.GetItem(id, index)
	if err != nil {
		return
	}

	hash.Set(item.([]byte))
	return
}

// GetLast get last event for participant
func (pec *ParticipantEventsCache) GetLast(participant string) (hash EventHash, err error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return
	}

	last, err := pec.rim.GetLast(id)
	if err != nil {
		return
	}

	hash.Set(last.([]byte))
	return
}

// Set the event for the participant
func (pec *ParticipantEventsCache) Set(participant string, hash EventHash, index int64) error {
	id, err := pec.participantID(participant)
	if err != nil {
		return err
	}
	return pec.rim.Set(id, hash.Bytes(), index)
}

// Known returns [participant id] => lastKnownIndex
func (pec *ParticipantEventsCache) Known() map[uint64]int64 {
	return pec.rim.Known()
}

// Reset resets the event cache
func (pec *ParticipantEventsCache) Reset() error {
	return pec.rim.Reset()
}

// Import from another event cache
func (pec *ParticipantEventsCache) Import(other *ParticipantEventsCache) {
	pec.rim.Import(other.rim)
}

// ------------------------------------------------------------------------------

// ParticipantBlockSignaturesCache struct
type ParticipantBlockSignaturesCache struct {
	participants *peers.Peers
	rim          *common.RollingIndexMap
}

// NewParticipantBlockSignaturesCache constructor
func NewParticipantBlockSignaturesCache(size int, participants *peers.Peers) *ParticipantBlockSignaturesCache {
	return &ParticipantBlockSignaturesCache{
		participants: participants,
		rim:          common.NewRollingIndexMap("ParticipantBlockSignatures", size, participants.ToIDSlice()),
	}
}

func (psc *ParticipantBlockSignaturesCache) participantID(participant string) (uint64, error) {
	peer, ok := psc.participants.ReadByPubKey(participant)

	if !ok {
		return peers.PeerNIL, common.NewStoreErr("ParticipantBlockSignatures", common.UnknownParticipant, participant)
	}

	return peer.ID, nil
}

// Get return participant BlockSignatures where index > skip
func (psc *ParticipantBlockSignaturesCache) Get(participant string, skipIndex int64) ([]BlockSignature, error) {
	id, err := psc.participantID(participant)
	if err != nil {
		return []BlockSignature{}, err
	}

	ps, err := psc.rim.Get(id, skipIndex)
	if err != nil {
		return []BlockSignature{}, err
	}

	res := make([]BlockSignature, len(ps))
	for k := 0; k < len(ps); k++ {
		res[k] = ps[k].(BlockSignature)
	}
	return res, nil
}

// GetItem get block signature at index for participant
func (psc *ParticipantBlockSignaturesCache) GetItem(participant string, index int64) (BlockSignature, error) {
	id, err := psc.participantID(participant)
	if err != nil {
		return BlockSignature{}, err
	}

	item, err := psc.rim.GetItem(id, index)
	if err != nil {
		return BlockSignature{}, err
	}
	return item.(BlockSignature), nil
}

// GetLast get last block signature for participant
func (psc *ParticipantBlockSignaturesCache) GetLast(participant string) (BlockSignature, error) {
	peer, ok := psc.participants.ReadByPubKey(participant)

	if !ok {
		return BlockSignature{}, fmt.Errorf("participant %v not found", participant)
	}

	last, err := psc.rim.GetLast(peer.ID)
	if err != nil {
		return BlockSignature{}, err
	}

	return last.(BlockSignature), nil
}

// Set sets the last block signature for the participant
func (psc *ParticipantBlockSignaturesCache) Set(participant string, sig BlockSignature) error {
	id, err := psc.participantID(participant)
	if err != nil {
		return err
	}

	return psc.rim.Set(id, sig, sig.Index)
}

// Known returns [participant id] => last BlockSignature Index
func (psc *ParticipantBlockSignaturesCache) Known() map[uint64]int64 {
	return psc.rim.Known()
}

// Reset resets the block signature cache
func (psc *ParticipantBlockSignaturesCache) Reset() error {
	return psc.rim.Reset()
}
