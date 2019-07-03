package election

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// TODO implement&test coinRound
//const coinRound = 10 // every 10th round is a round with pseudorandom votes

type Election struct {
	// election params
	frameToDecide uint32

	nodes         []ElectionNode
	totalStake    uint64 // the sum of stakes (n)
	superMajority uint64 // the quorum (should be 2/3n + 1)

	// election state
	decidedRoots map[hash.Peer]voteValue // decided roots at "frameToDecide"
	votes        map[voteId]voteValue

	// external world
	stronglySee RootStronglySeeRootFn

	logger.Instance
}

type ElectionNode struct {
	Nodeid      hash.Peer
	StakeAmount uint64
}

// specifies a slot {nodeid, frame}. Normal nodes can have only one root with this pair.
// Due to a fork, different roots may occupy the same slot
type RootSlot struct {
	Frame  uint32
	Nodeid hash.Peer
}

// @return hash of root B, if root A strongly sees root B.
// Due to a fork, there may be many roots B with the same slot,
// but strongly seen may be only one of them (if no more than 1/3n are Byzantine), with a specific hash.
type RootStronglySeeRootFn func(a hash.Event, b RootSlot) *hash.Event

type voteId struct {
	fromRoot  hash.Event
	forNodeid hash.Peer
}
type voteValue struct {
	decided  bool
	yes      bool
	seenRoot hash.Event
}

type ElectionRes struct {
	DecidedFrame     uint32
	DecidedSfWitness hash.Event
}

func NewElection(
	nodes []ElectionNode,
	totalStake uint64,
	superMajority uint64,
	frameToDecide uint32,
	stronglySeeFn RootStronglySeeRootFn,
) *Election {
	return &Election{
		nodes:         nodes,
		totalStake:    totalStake,
		superMajority: superMajority,
		frameToDecide: frameToDecide,
		decidedRoots:  make(map[hash.Peer]voteValue),
		votes:         make(map[voteId]voteValue),
		stronglySee:   stronglySeeFn,
		Instance:      logger.MakeInstance(),
	}
}

// erase the current election state, prepare for new election frame
func (el *Election) ResetElection(frameToDecide uint32) {
	el.frameToDecide = frameToDecide
	el.votes = make(map[voteId]voteValue)
	el.decidedRoots = make(map[hash.Peer]voteValue)
}

// return root slots which are not within el.decidedRoots
func (el *Election) notDecidedRoots() []hash.Peer {
	notDecidedRoots := make([]hash.Peer, 0, len(el.nodes))

	for _, node := range el.nodes {
		if _, ok := el.decidedRoots[node.Nodeid]; !ok {
			notDecidedRoots = append(notDecidedRoots, node.Nodeid)
		}
	}
	if len(notDecidedRoots)+len(el.decidedRoots) != len(el.nodes) { // sanity check
		el.Fatal("Mismatch of roots")
	}
	return notDecidedRoots
}

type weightedRoot struct {
	root        hash.Event
	stakeAmount uint64
}

// @return all the roots which are strongly seen by the specified root at the specified frame
func (el *Election) stronglySeenRoots(root hash.Event, frame uint32) []weightedRoot {
	seenRoots := make([]weightedRoot, 0, len(el.nodes))
	for _, node := range el.nodes {
		slot := RootSlot{
			Frame:  frame,
			Nodeid: node.Nodeid,
		}
		seenRoot := el.stronglySee(root, slot)
		if seenRoot != nil {
			seenRoots = append(seenRoots, weightedRoot{
				root:        *seenRoot,
				stakeAmount: node.StakeAmount,
			})
		}
	}
	return seenRoots
}

type GetRootsFn func(slot RootSlot) []hash.Event

// The function is similar to ProcessRoot, but it fully re-processes the current voting.
// This routine should be called after node startup, and after each decided frame.
func (el *Election) ProcessKnownRoots(maxKnownFrame uint32, getRootsFn GetRootsFn) (*ElectionRes, error) {
	// iterate all the roots from lowest frame to highest, call ProcessRootVotes for each
	for frame := el.frameToDecide + 1; frame <= maxKnownFrame; frame++ {
		for _, node := range el.nodes {
			slot := RootSlot{
				Frame:  frame,
				Nodeid: node.Nodeid,
			}
			roots := getRootsFn(slot)
			// if there's more than 1 root, then all of them are forks. it's fine
			for _, root := range roots {
				decided, err := el.ProcessRoot(root, slot)
				if decided != nil || err != nil {
					return decided, err
				}
			}
		}
	}

	return nil, nil
}
