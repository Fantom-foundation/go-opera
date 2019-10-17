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
		observe RootForklessCausesRootFn

		logger.Instance
	}

	// RootForklessCausesRootFn returns hash of root B, if root B forkless causes root A.
	// Due to a fork, there may be many roots B with the same slot,
	// but A may be forkless caused only by one of them (if no more than 1/3W are Byzantine), with a specific hash.
	RootForklessCausesRootFn func(a hash.Event, b common.Address, f idx.Frame) *hash.Event

	// Slot specifies a root slot {addr, frame}. Normal validators can have only one root with this pair.
	// Due to a fork, different roots may occupy the same slot
	Slot struct {
		Frame idx.Frame
		Addr  common.Address
	}

	// RootAndSlot specifies concrete root of slot.
	RootAndSlot struct {
		Root hash.Event
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
	forklessCausesFn RootForklessCausesRootFn,
) *Election {
	el := &Election{
		observe: forklessCausesFn,

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
	notDecidedRoots := make([]common.Address, 0, len(el.validators))

	for validator := range el.validators {
		if _, ok := el.decidedRoots[validator]; !ok {
			notDecidedRoots = append(notDecidedRoots, validator)
		}
	}
	if len(notDecidedRoots)+len(el.decidedRoots) != len(el.validators) { // sanity check
		el.Log.Crit("Mismatch of roots")
	}
	return notDecidedRoots
}

// observedRoots returns all the roots at the specified frame which do forkless cause the specified root.
func (el *Election) observedRoots(root hash.Event, frame idx.Frame) []RootAndSlot {
	observedRoots := make([]RootAndSlot, 0, len(el.validators))
	for validator := range el.validators {
		slot := Slot{
			Frame: frame,
			Addr:  validator,
		}
		observedRoot := el.observe(root, validator, frame)
		if observedRoot != nil {
			observedRoots = append(observedRoots, RootAndSlot{
				Root: *observedRoot,
				Slot: slot,
			})
		}
	}
	return observedRoots
}
