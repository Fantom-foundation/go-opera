package inter

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalcMisbehaviourProofsHash(t *testing.T) {
	v := []MisbehaviourProof{
		MisbehaviourProof{
			EventsDoublesign: &EventsDoublesign{
				Pair: [2]SignedEventLocator{
					test_signed_event_locator,
					test_signed_event_locator,
				},
			},
		},
		MisbehaviourProof{
			BlockVoteDoublesign: &BlockVoteDoublesign{
				Block: test_block_votes.Start,
				Pair: [2]LlrSignedBlockVotes{
					test_llr_signed_block_votes,
					test_llr_signed_block_votes,
				},
			},
		},
		MisbehaviourProof{
			WrongBlockVote: &WrongBlockVote{
				Block: test_block_votes.Start,
				Pals: [2]LlrSignedBlockVotes{
					test_llr_signed_block_votes,
					test_llr_signed_block_votes,
				},
				WrongEpoch: true,
			},
		},
		MisbehaviourProof{
			EpochVoteDoublesign: &EpochVoteDoublesign{
				Pair: [2]LlrSignedEpochVote{
					test_llr_signed_epoch_vote,
					test_llr_signed_epoch_vote,
				},
			},
		},
		MisbehaviourProof{
			WrongEpochVote: &WrongEpochVote{
				Pals: [2]LlrSignedEpochVote{
					test_llr_signed_epoch_vote,
					test_llr_signed_epoch_vote,
				},
			},
		},
	}

	require.Equal(t, "85834ef7fc1d75d65832b1f9b45b43c4f5677811bb2d384208553d32ab49def1", hex.EncodeToString(CalcMisbehaviourProofsHash(v).Bytes()))
}
