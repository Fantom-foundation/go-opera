package gossip

import (
	"fmt"
	"math/big"
	"testing"

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

	// valid tx
	_, tx1ok, cBallot, err := ballot.DeployBallot(env.Pay(1), env, proposals)
	require.NoError(err)
	require.NotNil(cBallot)
	require.NotNil(tx1ok)

	// invalid tx
	tx2reverted, err := cBallot.Vote(env.Pay(2), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx2reverted)

	// valid tx
	tx3ok, err := cBallot.GiveRightToVote(env.Pay(1), env.Address(3))
	require.NoError(err)
	require.NotNil(tx3ok)

	// invalid tx
	tx4skipped, err := cBallot.Vote(withLowGas(env.Pay(2)), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx4skipped)

	//  valid tx
	tx5ok, err := cBallot.Vote(env.Pay(3), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx5ok)

	receipts, err := env.ApplyTxs(nextEpoch,
		tx1ok,
		tx2reverted,
		tx3ok,
		tx4skipped,
		tx5ok,
	)
	require.NoError(err)

	for _, r := range receipts {
		fmt.Printf(">>>>>>>>> tx[%s] status %d\n", r.TxHash.String(), r.Status)
	}

	require.Len(receipts, 0)
}
