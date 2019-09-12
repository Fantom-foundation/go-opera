package election

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// ProcessRoot calculates Atropos votes only for the new root.
// If this root causes that the current election is decided, then @return decided Atropos
func (el *Election) ProcessRoot(newRoot RootAndSlot) (*ElectionRes, error) {
	if len(el.decidedRoots) == len(el.members) {
		// current election is already decided
		return el.chooseAtropos()
	}

	if newRoot.Slot.Frame <= el.frameToDecide {
		// too old root, out of interest for current election
		return nil, nil
	}
	round := newRoot.Slot.Frame - el.frameToDecide

	notDecidedRoots := el.notDecidedRoots()
	for _, memberSubject := range notDecidedRoots {
		vote := voteValue{}

		if round == 1 {
			// in initial round, vote "yes" if subject is forkless caused
			causedRoot := el.forklessCauses(newRoot.Root, memberSubject, el.frameToDecide)
			vote.yes = causedRoot != nil
			vote.decided = false
			if causedRoot != nil {
				vote.causedRoot = *causedRoot
			}
		} else if round > 1 {
			forklessCausedRoots := el.forklessCausedRoots(newRoot.Root, newRoot.Slot.Frame-1)

			var (
				yesVotes = el.members.NewCounter()
				noVotes  = el.members.NewCounter()
				allVotes = el.members.NewCounter()
			)

			// calc number of "yes" and "no", weighted by member's stake
			var subjectHash *hash.Event
			for _, forklessCausedRoot := range forklessCausedRoots {
				vid := voteId{
					forMember: memberSubject,
					fromRoot:  forklessCausedRoot.Root,
				}

				if vote, ok := el.votes[vid]; ok {
					if vote.yes && subjectHash != nil && *subjectHash != vote.causedRoot {
						return nil, fmt.Errorf("2 fork roots are forkless caused => more than 1/3n are Byzantine (%s != %s, election frame=%d, member=%s)",
							subjectHash.String(), vote.causedRoot.String(), el.frameToDecide, memberSubject.String())
					}

					if vote.yes {
						subjectHash = &vote.causedRoot
						yesVotes.Count(forklessCausedRoot.Slot.Addr)
					} else {
						noVotes.Count(forklessCausedRoot.Slot.Addr)
					}
					if !allVotes.Count(forklessCausedRoot.Slot.Addr) {
						// it shouldn't be possible to get here, because we've taken 1 root from every node above
						return nil, fmt.Errorf("2 fork roots are forkless caused => more than 1/3n are Byzantine (%s, election frame=%d, member=%s)",
							subjectHash.String(), el.frameToDecide, memberSubject.String())
					}
				} else {
					el.Log.Crit("Every root must vote for every not decided subject. Possibly roots are processed out of order",
						"root", newRoot.Root.String())
				}
			}
			// sanity checks
			if !allVotes.HasQuorum() {
				el.Log.Crit("Root must cause at least 2/3n of prev roots. Possibly roots are processed out of order",
					"root", newRoot.Root.String(),
					"voites", allVotes.Sum())
			}

			// vote as majority of votes
			vote.yes = yesVotes.Sum() >= noVotes.Sum()
			if vote.yes && subjectHash != nil {
				vote.causedRoot = *subjectHash
			}

			// If supermajority is caused, then the final decision may be made.
			// It's guaranteed to be final and consistent unless more than 1/3n are Byzantine.
			vote.decided = yesVotes.HasQuorum() || noVotes.HasQuorum()
			if vote.decided {
				el.decidedRoots[memberSubject] = vote
			}
		} else {
			continue // we shouldn't be here, we checked it above the loop
		}
		// save vote for next rounds
		vid := voteId{
			fromRoot:  newRoot.Root,
			forMember: memberSubject,
		}
		el.votes[vid] = vote
	}

	frameDecided := len(el.decidedRoots) == len(el.members)
	if frameDecided {
		return el.chooseAtropos()
	}

	return nil, nil
}
