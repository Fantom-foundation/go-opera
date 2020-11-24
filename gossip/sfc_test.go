package gossip

// SFC contracts
// NOTE: assumed that opera-sfc repo is in the same dir than go-opera repo

// sfc proxy
//1go:generate bash -c "cd ../../opera-sfc && git checkout main && docker run --rm -v $(pwd):/src -v $(pwd)/../go-opera/gossip/contract:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src/contracts --overwrite /src/contracts/sfc/Migrations.sol"
//1go:generate mkdir -p ./contract/sfcproxy
//1go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/Migrations.bin --abi=./contract/solc/Migrations.abi --pkg=sfcproxy --type=Contract --out=contract/sfcproxy/contract.go

// main (genesis)
//go:generate bash -c "NPM_CONFIG_PREFIX=~ cd ../../opera-sfc && git checkout main && docker run --rm -v $(pwd)/../go-opera/gossip/contract/solc:/src/build/contracts -v $(pwd):/src -w /src node:10.23.0 bash -c 'npm install -g truffle@v5.1.4 && npm install && truffle compile --all'"
//go:generate bash -c "cd ../../opera-sfc && docker run --rm -v $(pwd):/src -v $(pwd)/../go-opera/gossip/contract:/dst ethereum/solc:0.5.12 @openzeppelin/contracts/math=/src/node_modules/@openzeppelin/contracts/math --optimize --optimize-runs=2000 --bin --abi --allow-paths /src --overwrite -o /dst/solc/ /src/contracts/sfc/SFC.sol"
//go:generate mkdir -p ./contract/sfc100
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/SFC.bin --abi=./contract/solc/SFC.abi --pkg=sfc100 --type=Contract --out=contract/sfc100/contract.go

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/sfc100"
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

	var (
		sfc10 *sfc100.Contract
	)

	_ = true &&

		t.Run("Genesis v1.0.0", func(t *testing.T) {
			// nothing to do
		}) &&

		t.Run("Some transfers I", func(t *testing.T) {
			cicleTransfers(t, env, 1)
		}) &&

		t.Run("Upgrade to v1.0.0", func(t *testing.T) {
			require := require.New(t)

			r := env.ApplyBlock(nextEpoch,
				env.Contract(1, utils.ToFtm(0), sfc100.ContractBin),
			)
			newImpl := r[0].ContractAddress

			admin := env.Payer(1)
			tx, err := sfcProxy.ContractTransactor.UpgradeTo(admin, newImpl)
			require.NoError(err)
			env.ApplyBlock(sameEpoch, tx)

			impl, err := sfcProxy.Implementation(env.ReadOnly())
			require.NoError(err)
			require.Equal(newImpl, impl, "SFC-proxy: implementation address")

			sfc11, err = sfc110.NewContract(sfc.ContractAddress, env)
			require.NoError(err)

			epoch, err := sfc11.ContractCaller.CurrentEpoch(env.ReadOnly())
			require.NoError(err)
			require.Equal(0, epoch.Cmp(big.NewInt(2)), "current epoch")
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
