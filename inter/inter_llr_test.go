package inter

import (
	"encoding/hex"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/stretchr/testify/require"
)

var test_block_votes = LlrBlockVotes{
	Start: 9000,
	Epoch: 342,
	Votes: []hash.Hash{
		hash.Of(nil),
	},
}
var test_signed_event_locator = SignedEventLocator{
	Locator: EventLocator{
		BaseHash:    hash.Of(nil),
		NetForkID:   1,
		Epoch:       42,
		Seq:         9_000_000,
		Lamport:     142,
		Creator:     242,
		PayloadHash: hash.Of(nil),
	},
	Sig: [64]byte{},
}
var test_txs_and_misbehaviour_proofs_hash = hash.Of(nil)
var test_epoch_vote_hash = hash.Of(nil)
var test_block_votes_hash = hash.Of(nil)

var test_llr_signed_block_votes = LlrSignedBlockVotes{
	Signed:                       test_signed_event_locator,
	TxsAndMisbehaviourProofsHash: test_txs_and_misbehaviour_proofs_hash,
	EpochVoteHash:                test_epoch_vote_hash,
	Val:                          test_block_votes,
}
var test_llr_signed_epoch_vote = LlrSignedEpochVote{
	Signed:                       test_signed_event_locator,
	TxsAndMisbehaviourProofsHash: test_txs_and_misbehaviour_proofs_hash,
	BlockVotesHash:               test_block_votes_hash,
	Val: LlrEpochVote{
		Epoch: test_signed_event_locator.Locator.Epoch,
		Vote:  test_epoch_vote_hash,
	},
}

func TestLlrBlockVotesHashing(t *testing.T) {
	require.Equal(t, "7db20066868c07236a5f5641eac6a1ddb5cd1dc1602252a3672758056b4562a8", hex.EncodeToString(test_block_votes.Hash().Bytes()))
}

func TestLlrSignedBlockVotesHashing(t *testing.T) {
	require.Equal(t, "26b5c58d72b48f2ce40f5e18df4dda66f644839d00611b9a7812e8f9f708f143", hex.EncodeToString(test_llr_signed_block_votes.CalcPayloadHash().Bytes()))
}

func TestLlrSignedEpochVoteHashing(t *testing.T) {
	require.Equal(t, "e6c5f9d9e45f04b286c8361d8b5851f7ee16538f408ad86afa4d2b2612a10a75", hex.EncodeToString(test_llr_signed_epoch_vote.CalcPayloadHash().Bytes()))
}
