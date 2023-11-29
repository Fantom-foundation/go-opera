package gossip

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/ballot"
	"github.com/Fantom-foundation/go-opera/logger"
)

func TestTxIndexing(t *testing.T) {
	logger.SetTestMode(t)
	logger.SetLevel("debug")
	require := require.New(t)

	env := newTestEnv(2, 3)
	defer env.Close()

	proposals := [][32]byte{
		ballotOption("Option 1"),
		ballotOption("Option 2"),
		ballotOption("Option 3"),
	}

	// preparing
	_, tx1pre, cBallot, err := ballot.DeployBallot(env.Pay(1), env, proposals)
	require.NoError(err)
	require.NotNil(cBallot)
	require.NotNil(tx1pre)
	tx2pre, err := cBallot.GiveRightToVote(env.Pay(1), env.Address(3))
	require.NoError(err)
	require.NotNil(tx2pre)
	receipts, err := env.BlockTxs(nextEpoch,
		tx1pre,
		tx2pre,
	)
	require.NoError(err)
	require.Len(receipts, 2)
	for i, r := range receipts {
		require.Equal(types.ReceiptStatusSuccessful, r.Status, i)
	}
	// invalid tx
	tx1reverted, err := cBallot.Vote(env.Pay(2), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx1reverted)
	//  valid tx
	tx2ok, err := cBallot.Vote(env.Pay(3), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx2ok)
	// skipped tx
	_, tx3skipped, _, err := ballot.DeployBallot(withLowGas(env.Pay(1)), env, proposals)
	require.NoError(err)
	require.NotNil(tx3skipped)

	receipts, err = env.BlockTxs(nextEpoch,
		tx1reverted,
		tx2ok,
		tx3skipped,
	)
	require.NoError(err)
	require.Len(receipts, 3)
	var block *big.Int
	for i, r := range receipts {
		if block == nil {
			block = r.BlockNumber
		}
		require.Equal(block.Uint64(), r.BlockNumber.Uint64(), i)
	}

}
