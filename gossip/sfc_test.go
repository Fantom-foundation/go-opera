package gossip

// SFC contracts
// NOTE: assumed that opera-sfc repo is in the same dir than go-opera repo

// sfc proxy
//go:generate bash -c "cd ../../opera-sfc && git checkout main && docker run --rm -v $(pwd):/src -v $(pwd)/../go-opera/gossip/contract:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src/contracts --overwrite /src/contracts/sfc/Migrations.sol"
//go:generate mkdir -p ./contract/sfcproxy
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/Migrations.bin --abi=./contract/solc/Migrations.abi --pkg=sfcproxy --type=Contract --out=contract/sfcproxy/contract.go

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/sfcproxy"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/Fantom-foundation/go-opera/opera/params"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestSFC(t *testing.T) {
	logger.SetTestMode(t)

	env := newTestEnv()
	defer env.Close()

	_, err := sfcproxy.NewContract(sfc.ContractAddress, env)
	require.NoError(t, err)

	_ = true &&

		t.Run("Genesis v1.0.0", func(t *testing.T) {
			// nothing to do
		}) &&

		t.Run("Some transfers I", func(t *testing.T) {
			cicleTransfers(t, env, 1)
		})

}

func cicleTransfers(t *testing.T, env *testEnv, count uint64) {
	require := require.New(t)

	balances := make([]*big.Int, 3)
	for i := range balances {
		balances[i] = env.State().GetBalance(env.Address(i + 1))
	}

	var gasUsed uint64
	for i := uint64(0); i < count; i++ {
		rr := env.ApplyBlock(sameEpoch,
			env.Transfer(1, 2, utils.ToFtm(100)),
		)
		gasUsed += rr[0].GasUsed // per 1 account
		env.ApplyBlock(sameEpoch,
			env.Transfer(2, 3, utils.ToFtm(100)),
		)
		env.ApplyBlock(sameEpoch,
			env.Transfer(3, 1, utils.ToFtm(100)),
		)
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
