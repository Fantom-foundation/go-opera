package election

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"golang.org/x/crypto/sha3"
)

// may be used in tests to match election state
func (el *Election) DebugStateHash() common.Hash {
	hasher := sha3.New256()
	for vid, vote := range el.votes {
		hasher.Write(vid.fromRoot.Bytes())
		hasher.Write(vote.seenRoot.Bytes())
	}
	for slot, vote := range el.decidedRoots {
		hasher.Write(slot.Nodeid.Bytes())
		hasher.Write(vote.seenRoot.Bytes())
	}
	return common.BytesToHash(hasher.Sum(nil))
}
