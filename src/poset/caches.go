package poset

import (
	"fmt"

	cm "github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

// Key struct
type Key struct {
	x string
	y string
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
	rim          *cm.RollingIndexMap
}

// NewParticipantEventsCache constructor
func NewParticipantEventsCache(size int, participants *peers.Peers) *ParticipantEventsCache {
	return &ParticipantEventsCache{
		participants: participants,
		rim:          cm.NewRollingIndexMap("ParticipantEvents", size, participants.ToIDSlice()),
	}
}

// add peer to cache and rolling index map, returns error if it failed to add to map
func (pec *ParticipantEventsCache) AddPeer(peer *peers.Peer) error {
	pec.participants.AddPeer(peer)
	return pec.rim.AddKey(peer.ID)
}

func (pec *ParticipantEventsCache) participantID(participant string) (int64, error) {
	peer, ok := pec.participants.ByPubKey[participant]

	if !ok {
		return -1, cm.NewStoreErr("ParticipantEvents", cm.UnknownParticipant, participant)
	}

	return peer.ID, nil
}

// Get return participant events with index > skip
func (pec *ParticipantEventsCache) Get(participant string, skipIndex int64) ([]string, error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return []string{}, err
	}

	pe, err := pec.rim.Get(id, skipIndex)
	if err != nil {
		return []string{}, err
	}

	res := make([]string, len(pe))
	for k := 0; k < len(pe); k++ {
		res[k] = pe[k].(string)
	}
	return res, nil
}

// GetItem get event for participant at index
func (pec *ParticipantEventsCache) GetItem(participant string, index int64) (string, error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return "", err
	}

	item, err := pec.rim.GetItem(id, index)
	if err != nil {
		return "", err
	}
	return item.(string), nil
}

// GetLast get last event for participant
func (pec *ParticipantEventsCache) GetLast(participant string) (string, error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return "", err
	}

	last, err := pec.rim.GetLast(id)
	if err != nil {
		return "", err
	}
	return last.(string), nil
}

// GetLastConsensus get last consensus for participant
func (pec *ParticipantEventsCache) GetLastConsensus(participant string) (string, error) {
	id, err := pec.participantID(participant)
	if err != nil {
		return "", err
	}

	last, err := pec.rim.GetLast(id)
	if err != nil {
		return "", err
	}
	return last.(string), nil
}

// Set the event for the participant
func (pec *ParticipantEventsCache) Set(participant string, hash string, index int64) error {
	id, err := pec.participantID(participant)
	if err != nil {
		return err
	}
	return pec.rim.Set(id, hash, index)
}

// Known returns [participant id] => lastKnownIndex
func (pec *ParticipantEventsCache) Known() map[int64]int64 {
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
	rim          *cm.RollingIndexMap
}

// NewParticipantBlockSignaturesCache constructor
func NewParticipantBlockSignaturesCache(size int, participants *peers.Peers) *ParticipantBlockSignaturesCache {
	return &ParticipantBlockSignaturesCache{
		participants: participants,
		rim:          cm.NewRollingIndexMap("ParticipantBlockSignatures", size, participants.ToIDSlice()),
	}
}

func (psc *ParticipantBlockSignaturesCache) participantID(participant string) (int64, error) {
	peer, ok := psc.participants.ByPubKey[participant]

	if !ok {
		return -1, cm.NewStoreErr("ParticipantBlockSignatures", cm.UnknownParticipant, participant)
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
	last, err := psc.rim.GetLast(psc.participants.ByPubKey[participant].ID)

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
func (psc *ParticipantBlockSignaturesCache) Known() map[int64]int64 {
	return psc.rim.Known()
}

// Reset resets the block signature cache
func (psc *ParticipantBlockSignaturesCache) Reset() error {
	return psc.rim.Reset()
}
