package gossip

import (
	"math/big"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/ballot"
	"github.com/Fantom-foundation/go-opera/logger"
)

func TestTxIndexing(t *testing.T) {
	logger.SetTestMode(t)
	logger.SetLevel("debug")
	require := require.New(t)

	var (
		epoch      = idx.Epoch(256)
		validators = idx.Validator(3)
		proposals  = [][32]byte{
			ballotOption("Option 1"),
			ballotOption("Option 2"),
			ballotOption("Option 3"),
		}
	)

	env := newTestEnv(epoch, validators)
	defer env.Close()

	// Preparing:

	_, tx1pre, cBallot, err := ballot.DeployBallot(env.Pay(1), env, proposals)
	require.NoError(err)
	require.NotNil(cBallot)
	require.NotNil(tx1pre)
	tx2pre, err := cBallot.GiveRightToVote(env.Pay(1), env.Address(3))
	require.NoError(err)
	require.NotNil(tx2pre)
	receipts, err := env.ApplyTxs(sameEpoch,
		tx1pre,
		tx2pre,
	)
	require.NoError(err)
	require.Len(receipts, 2)
	for i, r := range receipts {
		require.Equal(types.ReceiptStatusSuccessful, r.Status, i)
	}

	// Testing:

	// invalid tx
	tx1reverted, err := cBallot.Vote(env.Pay(2), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx1reverted)
	//  valid tx
	tx2ok, err := cBallot.Vote(env.Pay(3), big.NewInt(0))
	require.NoError(err)
	require.NotNil(tx2ok)
	// skipped tx
	tx3skipped := tx2ok
	// valid tx
	tx4ok, err := cBallot.GiveRightToVote(env.Pay(1), env.Address(2))
	require.NoError(err)
	require.NotNil(tx1reverted)

	receipts, err = env.BlockTxs(sameEpoch,
		tx1reverted,
		tx2ok,
		tx3skipped,
		tx4ok)
	require.NoError(err)
	require.Len(receipts, 3)

	var blockN *big.Int
	for i, r := range receipts {
		if blockN == nil {
			blockN = r.BlockNumber
		}
		require.Equal(blockN.Uint64(), r.BlockNumber.Uint64(), i)

		txPos := env.store.evm.GetTxPosition(r.TxHash)
		require.NotNil(txPos)

		switch r.TxHash {
		case tx1reverted.Hash():
			require.Equal(types.ReceiptStatusFailed, r.Status, i)
			require.Equal(txPos.BlockOffset, uint32(0))
		case tx2ok.Hash():
			require.Equal(types.ReceiptStatusSuccessful, r.Status, i)
			require.Equal(txPos.BlockOffset, uint32(1))
		case tx3skipped.Hash():
			t.Fatal("skipped tx's receipt found")
		case tx4ok.Hash():
			require.Equal(types.ReceiptStatusSuccessful, r.Status, i)
			require.Equal(txPos.BlockOffset, uint32(3)) // THAT shows the effect of the fix #524
		}

		for j, l := range r.Logs {
			require.Equal(txPos.BlockOffset, l.TxIndex, j)
		}
	}
}
