package inter

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type EventsDoublesign struct {
	Pair [2]SignedEventLocator
}

type BlockVoteDoublesign struct {
	Block idx.Block
	Pair  [2]LlrSignedBlockVotes
}

func (p BlockVoteDoublesign) GetVote(i int) hash.Hash {
	return p.Pair[i].Votes[p.Block-p.Pair[i].Start]
}

type WrongBlockVote struct {
	Block idx.Block
	Votes LlrSignedBlockVotes
}

func (p WrongBlockVote) GetVote() hash.Hash {
	return p.Votes.Votes[p.Block-p.Votes.Start]
}

type EpochVoteDoublesign struct {
	Pair [2]LlrSignedEpochVote
}

type WrongEpochVote struct {
	Votes LlrSignedEpochVote
}

type MisbehaviourProof struct {
	EventsDoublesign *EventsDoublesign `rlp:"nil"`

	BlockVoteDoublesign *BlockVoteDoublesign `rlp:"nil"`

	WrongBlockVote *WrongBlockVote `rlp:"nil"`

	EpochVoteDoublesign *EpochVoteDoublesign `rlp:"nil"`

	WrongEpochVote *WrongEpochVote `rlp:"nil"`
}
