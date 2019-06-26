package election

import (
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

// may be used in tests to match election state
func (el *Election) DebugStateHash() hash.Hash {
	hasher := sha3.New256()
	for vid, vote := range el.votes {
		hasher.Write(vid.fromRoot.Bytes())
		hasher.Write(vote.seenRoot.Bytes())
	}
	for nodeid, vote := range el.decidedRoots {
		hasher.Write(nodeid.Bytes())
		hasher.Write(vote.seenRoot.Bytes())
	}
	return hash.FromBytes(hasher.Sum(nil))
}
