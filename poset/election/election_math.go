package election

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// ProcessRoot calculates Atropos votes only for the new root.
// If this root observes that the current election is decided, then return decided Atropos
func (el *Election) ProcessRoot(newRoot RootAndSlot) (*Res, error) {
	res, err := el.chooseAtropos()
	if err != nil || res != nil {
		return res, err
	}

	if newRoot.Slot.Frame <= el.frameToDecide {
		// too old root, out of interest for current election
		return nil, nil
	}
	round := newRoot.Slot.Frame - el.frameToDecide
	if round == 0 {
		return nil, nil
	}

	notDecidedRoots := el.notDecidedRoots()

	var observedRoots []RootAndSlot
	var observedRootsMap map[idx.StakerID]RootAndSlot
	if round == 1 {
		observedRootsMap = el.observedRootsMap(newRoot.ID, newRoot.Slot.Frame-1)
	} else {
		observedRoots = el.observedRoots(newRoot.ID, newRoot.Slot.Frame-1)
	}

	for _, validatorSubject := range notDecidedRoots {
		vote := voteValue{}

		if round == 1 {
			// in initial round, vote "yes" if observe the subject
			observedRoot, ok := observedRootsMap[validatorSubject]
			vote.yes = ok
			vote.decided = false
			if ok {
				vote.observedRoot = observedRoot.ID
			}
		} else {
			var (
				yesVotes = el.validators.NewCounter()
				noVotes  = el.validators.NewCounter()
				allVotes = el.validators.NewCounter()
			)

			// calc number of "yes" and "no", weighted by validator's stake
			var subjectHash *hash.Event
			for _, observedRoot := range observedRoots {
				vid := voteID{
					forValidator: validatorSubject,
					fromRoot:     observedRoot.ID,
				}

				if vote, ok := el.votes[vid]; ok {
					if vote.yes && subjectHash != nil && *subjectHash != vote.observedRoot {
						return nil, fmt.Errorf("forkless caused by 2 fork roots => more than 1/3W are Byzantine (%s != %s, election frame=%d, validator=%d)",
							subjectHash.String(), vote.observedRoot.String(), el.frameToDecide, validatorSubject)
					}

					if vote.yes {
						subjectHash = &vote.observedRoot
						yesVotes.Count(observedRoot.Slot.Validator)
					} else {
						noVotes.Count(observedRoot.Slot.Validator)
					}
					if !allVotes.Count(observedRoot.Slot.Validator) {
						// it shouldn't be possible to get here, because we've taken 1 root from every node above
						return nil, fmt.Errorf("forkless caused by 2 fork roots => more than 1/3W are Byzantine (%s, election frame=%d, validator=%d)",
							subjectHash.String(), el.frameToDecide, validatorSubject)
					}
				} else {
					el.Log.Crit("Every root must vote for every not decided subject. Possibly roots are processed out of order",
						"root", newRoot.ID.String())
				}
			}
			// sanity checks
			if !allVotes.HasQuorum() {
				el.Log.Crit("Root must be forkless caused by at least 2/3W of prev roots. Possibly roots are processed out of order",
					"root", newRoot.ID.String(),
					"votes", allVotes.Sum())
			}

			// vote as majority of votes
			vote.yes = yesVotes.Sum() >= noVotes.Sum()
			if vote.yes && subjectHash != nil {
				vote.observedRoot = *subjectHash
			}

			// If supermajority is observed, then the final decision may be made.
			// It's guaranteed to be final and consistent unless more than 1/3W are Byzantine.
			vote.decided = yesVotes.HasQuorum() || noVotes.HasQuorum()
			if vote.decided {
				el.decidedRoots[validatorSubject] = vote
			}
		}
		// save vote for next rounds
		vid := voteID{
			fromRoot:     newRoot.ID,
			forValidator: validatorSubject,
		}
		el.votes[vid] = vote
	}

	// check if election is decided
	return el.chooseAtropos()
}
