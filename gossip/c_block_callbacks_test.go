package gossip

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/params"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestConsensusCallback(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const blockCount = 100

	env := newTestEnv()
	defer env.Close()

	accounts := len(env.validators)

	// save start balances
	balances := make([]*big.Int, accounts)
	for i := range balances {
		balances[i] = env.State().GetBalance(env.Address(i + 1))
	}

	for n := uint64(0); n < blockCount; n++ {
		// transfers
		txs := make([]*types.Transaction, accounts)
		for i := range txs {
			from := (i)%accounts + 1
			to := (i+1)%accounts + 1
			txs[i] = env.Transfer(from, to, utils.ToFtm(100))
		}
		tm := sameEpoch
		if n%10 == 0 {
			tm = nextEpoch
		}
		rr := env.ApplyBlock(tm, txs...)
		for i, r := range rr {
			fee := big.NewInt(0).Mul(big.NewInt(int64(r.GasUsed)), params.MinGasPrice)
			balances[i] = big.NewInt(0).Sub(balances[i], fee)
		}

		// some acts to check data race
		bs := env.store.GetBlockState()
		require.LessOrEqual(n, uint64(bs.LastBlock))
	}

	// check balances
	for i := range balances {
		require.Equal(
			balances[i],
			env.State().GetBalance(env.Address(i+1)),
			fmt.Sprintf("account%d", i),
		)
	}

}
