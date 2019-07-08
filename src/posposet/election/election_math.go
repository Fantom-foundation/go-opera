package election

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// calculate SfWitness votes only for the new root.
// If this root sees that the current election is decided, then @return decided SfWitness
func (el *Election) ProcessRoot(newRoot RootAndSlot) (*ElectionRes, error) {
	if len(el.decidedRoots) == len(el.members) {
		// current election is already decided
		return el.chooseSfWitness()
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
			// in initial round, vote "yes" if subject is strongly seen
			seenRoot := el.stronglySee(newRoot.Root, memberSubject, el.frameToDecide)
			vote.yes = seenRoot != nil
			vote.decided = false
			if seenRoot != nil {
				vote.seenRoot = *seenRoot
			}
		} else if round > 1 {
			sSeenRoots := el.stronglySeenRoots(newRoot.Root, newRoot.Slot.Frame-1)

			var (
				yesVotes = el.members.NewCounter()
				noVotes  = el.members.NewCounter()
				allVotes = el.members.NewCounter()
			)

			// calc number of "yes" and "no", weighted by member's stake
			var subjectHash *hash.Event
			for _, sSeenRoot := range sSeenRoots {
				vid := voteId{
					forMember: memberSubject,
					fromRoot:  sSeenRoot.Root,
				}

				if vote, ok := el.votes[vid]; ok {
					if vote.yes && subjectHash != nil && *subjectHash != vote.seenRoot {
						return nil, fmt.Errorf("2 fork roots are strongly seen => more than 1/3n are Byzantine (%s != %s, election frame=%d, member=%s)",
							subjectHash.String(), vote.seenRoot.String(), el.frameToDecide, memberSubject.String())
					}

					if vote.yes {
						subjectHash = &vote.seenRoot
						yesVotes.Count(sSeenRoot.Slot.Addr)
					} else {
						noVotes.Count(sSeenRoot.Slot.Addr)
					}
					if !allVotes.Count(sSeenRoot.Slot.Addr) {
						// it shouldn't be possible to get here, because we've taken 1 root from every node above
						return nil, fmt.Errorf("2 fork roots are strongly seen => more than 1/3n are Byzantine (%s, election frame=%d, member=%s)",
							subjectHash.String(), el.frameToDecide, memberSubject.String())
					}
				} else {
					el.Fatal("Every root must vote for every not decided subject. Possibly roots are processed out of order, root=", newRoot.Root.String())
				}
			}
			// sanity checks
			if !allVotes.HasQuorum() {
				el.Fatal("Root must see at least 2/3n of prev roots. Possibly roots are processed out of order, root=", newRoot.Root.String(), " ", allVotes.Sum())
			}

			// vote as majority of votes
			vote.yes = yesVotes.Sum() >= noVotes.Sum()
			if vote.yes && subjectHash != nil {
				vote.seenRoot = *subjectHash
			}

			// If supermajority is seen, then the final decision may be made.
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
		return el.chooseSfWitness()
	}

	return nil, nil
}
