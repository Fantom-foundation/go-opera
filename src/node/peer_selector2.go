package node

import (
	"math"
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

// GetFlagTableFn declares flag table function signature
type GetFlagTableFn func() (map[string]int64, error)

// SmartPeerSelector provides selection based on FlagTable of a randomly chosen undermined event
type SmartPeerSelector struct {
	peers        *peers.Peers
	localAddr    string
	last         string
	GetFlagTable GetFlagTableFn
	pals         map[string]bool
}

// SmartPeerSelectorCreationFnArgs specifies which additional arguments are required to create a SmartPeerSelector
type SmartPeerSelectorCreationFnArgs struct {
	GetFlagTable GetFlagTableFn
	LocalAddr    string
}

// NewSmartPeerSelector creates a new smart peer selection struct
func NewSmartPeerSelector(participants *peers.Peers, args SmartPeerSelectorCreationFnArgs) *SmartPeerSelector {

	return &SmartPeerSelector{
		localAddr:    args.LocalAddr,
		peers:        participants,
		GetFlagTable: args.GetFlagTable,
		pals:         make(map[string]bool),
	}
}

// NewSmartPeerSelectorWrapper implements SelectorCreationFn to allow dynamic creation of SmartPeerSelector ie NewNode
func NewSmartPeerSelectorWrapper(participants *peers.Peers, args interface{}) PeerSelector {
	return NewSmartPeerSelector(participants, args.(SmartPeerSelectorCreationFnArgs))
}

// Peers returns all known peers
func (ps *SmartPeerSelector) Peers() *peers.Peers {
	return ps.peers
}

// UpdateLast sets the last peer communicated with (avoid double talk)
func (ps *SmartPeerSelector) UpdateLast(peer string) {
	// We need an exclusive access to ps.last for writing;
	// let use peers' lock instead of adding additional lock.
	// ps.last is accessed for read under peers' lock
	ps.peers.Lock()
	defer ps.peers.Unlock()

	ps.last = peer
}

// Next returns the next peer based on the flag table cost function selection
func (ps *SmartPeerSelector) Next() *peers.Peer {
	flagTable, err := ps.GetFlagTable()
	if err != nil {
		flagTable = nil
	}

	ps.peers.Lock()
	defer ps.peers.Unlock()

	sortedSrc := ps.peers.ToPeerByUsedSlice()
	n := int(2*len(sortedSrc)/3 + 1)
	if n < len(sortedSrc) {
		sortedSrc = sortedSrc[0:n]
	}
	selected := make([]*peers.Peer, len(sortedSrc))
	sCount := 0
	flagged := make([]*peers.Peer, len(sortedSrc))
	fCount := 0
	minUsedIdx := 0
	minUsedVal := int64(math.MaxInt64)
	var lastused []*peers.Peer

	for _, p := range sortedSrc {
		if p.NetAddr == ps.localAddr {
			continue
		}
		if p.NetAddr == ps.last || p.PubKeyHex == ps.last {
			lastused = append(lastused, p)
			continue
		}
		// skip peers we are alredy engaged with
		if _, ok := ps.pals[p.NetAddr]; ok {
			continue
		}

		if f, ok := flagTable[p.PubKeyHex]; ok && f == 1 {
			flagged[fCount] = p
			fCount += 1
			continue
		}

		if p.Used < minUsedVal {
			minUsedVal = p.Used
			minUsedIdx = sCount
		}
		selected[sCount] = p
		sCount += 1
	}

	selected = selected[minUsedIdx:sCount]
	if len(selected) < 1 {
		selected = flagged[0:fCount]
	}
	if len(selected) < 1 {
		selected = lastused
	}
	if len(selected) == 1 {
		selected[0].Used++
		return selected[0]
	}
	if len(selected) < 1 {
		return nil
	}

	i := rand.Intn(len(selected))
	selected[i].Used++
	return selected[i]
}

// Indicate we are in communication with a peer
// so it would be excluded from next peer selection
func (ps *SmartPeerSelector)Engage(peer string) {
	ps.peers.Lock()
	defer ps.peers.Unlock()
	ps.pals[peer] = true
}

// Indicate we are not in communication with a peer
// so it could be selected as a next peer
func (ps *SmartPeerSelector)Dismiss(peer string) {
	ps.peers.Lock()
	defer ps.peers.Unlock()
	if _, ok := ps.pals[peer]; ok {
		delete(ps.pals, peer)
	}
}
