package election

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// TODO implement&test coinRound
//const coinRound = 10 // every 10th round is a round with pseudorandom votes

type (
	// Election is cached data of election algorithm.
	Election struct {
		// election params
		frameToDecide idx.Frame

		validators pos.Validators

		// election state
		decidedRoots map[common.Address]voteValue // decided roots at "frameToDecide"
		votes        map[voteID]voteValue

		// external world
		observe       ForklessCauseFn
		getFrameRoots GetFrameRootsFn

		logger.Instance
	}

	// ForklessCauseFn returns true if event A is forkless caused by event B
	ForklessCauseFn func(a hash.Event, b hash.Event) bool
	// GetFrameRootsFn returns all the roots in the specified frame
	GetFrameRootsFn func(f idx.Frame) []RootAndSlot

	// Slot specifies a root slot {addr, frame}. Normal validators can have only one root with this pair.
	// Due to a fork, different roots may occupy the same slot
	Slot struct {
		Frame idx.Frame
		Addr  common.Address
	}

	// RootAndSlot specifies concrete root of slot.
	RootAndSlot struct {
		ID   hash.Event
		Slot Slot
	}
)

type voteID struct {
	fromRoot     hash.Event
	forValidator common.Address
}
type voteValue struct {
	decided      bool
	yes          bool
	observedRoot hash.Event
}

// Res defines the final election result, i.e. decided frame
type Res struct {
	Frame   idx.Frame
	Atropos hash.Event
}

// New election context
func New(
	validators pos.Validators,
	frameToDecide idx.Frame,
	forklessCauseFn ForklessCauseFn,
	getFrameRoots GetFrameRootsFn,
) *Election {
	el := &Election{
		observe:       forklessCauseFn,
		getFrameRoots: getFrameRoots,

		Instance: logger.MakeInstance(),
	}

	el.Reset(validators, frameToDecide)

	return el
}

// Reset erases the current election state, prepare for new election frame
func (el *Election) Reset(validators pos.Validators, frameToDecide idx.Frame) {
	el.validators = validators
	el.frameToDecide = frameToDecide
	el.votes = make(map[voteID]voteValue)
	el.decidedRoots = make(map[common.Address]voteValue)
}

// return root slots which are not within el.decidedRoots
func (el *Election) notDecidedRoots() []common.Address {
	notDecidedRoots := make([]common.Address, 0, el.validators.Len())

	for validator := range el.validators.Iterate() {
		if _, ok := el.decidedRoots[validator]; !ok {
			notDecidedRoots = append(notDecidedRoots, validator)
		}
	}
	if len(notDecidedRoots)+len(el.decidedRoots) != el.validators.Len() { // sanity check
		el.Log.Crit("Mismatch of roots")
	}
	return notDecidedRoots
}

// observedRoots returns all the roots at the specified frame which do forkless cause the specified root.
func (el *Election) observedRoots(root hash.Event, frame idx.Frame) []RootAndSlot {
	observedRoots := make([]RootAndSlot, 0, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRoots = append(observedRoots, frameRoot)
		}
	}
	return observedRoots
}

func (el *Election) observedRootsMap(root hash.Event, frame idx.Frame) map[common.Address]RootAndSlot {
	observedRootsMap := make(map[common.Address]RootAndSlot, el.validators.Len())

	frameRoots := el.getFrameRoots(frame)
	for _, frameRoot := range frameRoots {
		if el.observe(root, frameRoot.ID) {
			observedRootsMap[frameRoot.Slot.Addr] = frameRoot
		}
	}
	return observedRootsMap
}
