package gossip

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestConsensusCallback(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)

	const (
		rounds        = 30
		validatorsNum = 3
	)

	env := newTestEnv(2, validatorsNum)
	defer env.Close()

	// save start balances
	balances := make([]*big.Int, validatorsNum)
	for i := range balances {
		balances[i] = env.State().GetBalance(env.Address(idx.ValidatorID(i + 1)))
	}

	for n := uint64(0); n < rounds; n++ {
		// transfers
		txs := make([]*types.Transaction, 0, validatorsNum-1)
		for to, from := 0, 1; from < validatorsNum; from++ {
			transfer := env.Transfer(idx.ValidatorID(from+1), idx.ValidatorID(to+1), utils.ToFtm(100))
			txs = append(txs, transfer)
		}
		tm := sameEpoch
		if n%10 == 0 {
			tm = nextEpoch
		}
		rr, err := env.ApplyTxs(tm, txs...)
		require.NoError(err)
		// subtract fees
		for i, r := range rr {
			fee := big.NewInt(0).Mul(new(big.Int).SetUint64(r.GasUsed), txs[i].GasPrice())
			balances[i+1] = big.NewInt(0).Sub(balances[i+1], fee)
		}
		// balance movements
		balances[0].Add(balances[0], utils.ToFtm(200))
		balances[1].Sub(balances[1], utils.ToFtm(100))
		balances[2].Sub(balances[2], utils.ToFtm(100))
	}

	// check balances
	for i := range balances {
		require.Equal(
			balances[i],
			env.State().GetBalance(env.Address(idx.ValidatorID(i+1))),
			fmt.Sprintf("account%d", i),
		)
	}

}
