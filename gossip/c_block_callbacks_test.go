package gossip

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/params"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestConsensusCallback(t *testing.T) {
	logger.SetTestMode(t)
	require := require.New(t)
	count := uint64(100)

	env := newTestEnv()
	defer env.Close()

	balances := make([]*big.Int, 3)
	for i := range balances {
		balances[i] = env.State().GetBalance(env.Address(i + 1))
	}

	var gasUsed uint64
	for i := uint64(0); i < count; i++ {
		rr := env.ApplyBlock(sameEpoch,
			env.Transfer(1, 2, utils.ToFtm(100)),
			env.Transfer(2, 3, utils.ToFtm(100)),
			env.Transfer(3, 1, utils.ToFtm(100)),
		)
		gasUsed += rr[0].GasUsed // per 1 account
	}

	gp := params.MinGasPrice
	gas := big.NewInt(0).Mul(big.NewInt(int64(gasUsed)), gp)
	for i := range balances {
		require.Equal(
			big.NewInt(0).Sub(balances[i], gas),
			env.State().GetBalance(env.Address(i+1)),
			fmt.Sprintf("account%d", i),
		)
	}

}
