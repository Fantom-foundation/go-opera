package gossip

// SFC contracts
// NOTE: assumed that opera-sfc repo is in the same dir than go-opera repo
// sfc proxy
//go:generate bash -c "cd ../../opera-sfc && git checkout main && docker run --rm -v $(pwd):/src -v $(pwd)/../go-opera/gossip/contract:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src/contracts --overwrite /src/contracts/sfc/Migrations.sol"
//go:generate mkdir -p ./contract/sfcproxy
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/Migrations.bin --abi=./contract/solc/Migrations.abi --pkg=sfcproxy --type=Contract --out=contract/sfcproxy/contract.go
// main (genesis)
//go:generate bash -c "cd ../../opera-sfc && git checkout main && docker run --rm -v $(pwd)/../go-opera/gossip/contract/solc:/src/build/contracts -v $(pwd):/src -w /src node:10.23.0 bash -c 'NPM_CONFIG_PREFIX=~ npm install'"
//go:generate bash -c "docker run --rm -v $(pwd)/../../opera-sfc:/src -v $(pwd)/contract:/dst ethereum/solc:0.5.12 @openzeppelin/contracts/math=/src/node_modules/@openzeppelin/contracts/math --optimize --optimize-runs=2000 --bin --abi --allow-paths /src --overwrite -o /dst/solc/ /src/contracts/sfc/SFC.sol"
//go:generate mkdir -p ./contract/sfc100
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/SFC.bin --abi=./contract/solc/SFC.abi --pkg=sfc100 --type=Contract --out=contract/sfc100/contract.go

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/sfc100"
	"github.com/Fantom-foundation/go-opera/gossip/contract/sfcproxy"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestSFC(t *testing.T) {
	logger.SetTestMode(t)
	logger.SetLevel("debug")

	env := newTestEnv()
	defer env.Close()

	sfcProxy, err := sfcproxy.NewContract(sfc.ContractAddress, env)
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
			tx, err := sfcProxy.ContractTransactor.Upgrade(admin, newImpl)
			require.NoError(err)
			env.ApplyBlock(sameEpoch, tx)

			sfc10, err = sfc100.NewContract(sfc.ContractAddress, env)
			require.NoError(err)

			epoch, err := sfc10.ContractCaller.CurrentEpoch(env.ReadOnly())
			require.NoError(err)
			require.Equal(0, epoch.Cmp(big.NewInt(3)), "current epoch %s", epoch.String())
		})

}

func cicleTransfers(t *testing.T, env *testEnv, count uint64) {
	require := require.New(t)
	accounts := len(env.validators)

	// save start balances
	balances := make([]*big.Int, accounts)
	for i := range balances {
		balances[i] = env.State().GetBalance(env.Address(i + 1))
	}

	for i := uint64(0); i < count; i++ {
		// transfers
		txs := make([]*types.Transaction, accounts)
		for i := range txs {
			from := (i)%accounts + 1
			to := (i+1)%accounts + 1
			txs[i] = env.Transfer(from, to, utils.ToFtm(100))
		}

		rr := env.ApplyBlock(sameEpoch, txs...)
		for i, r := range rr {
			fee := big.NewInt(0).Mul(new(big.Int).SetUint64(r.GasUsed), txs[i].GasPrice())
			balances[i] = big.NewInt(0).Sub(balances[i], fee)
		}
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
