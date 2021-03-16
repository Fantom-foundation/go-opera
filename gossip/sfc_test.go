package gossip

// compile SFC with truffle
//go:generate bash -c "cd ../../opera-sfc && git checkout 2aba3feb7ef932e85f6d1127fd07f451b83d7266 && docker run --rm -v $(pwd)/../go-opera/gossip/contract/solc:/src/build/contracts -v $(pwd):/src -w /src node:10.5.0 bash -c 'export NPM_CONFIG_PREFIX=~; npm install --no-save; npm install --no-save truffle@5.1.4; npm run build'"
//go:generate bash -c "cd ./contract/solc && for f in *.json; do jq -j .bytecode $DOLLAR{f} > $DOLLAR{f%.json}.bin; jq -j .deployedBytecode $DOLLAR{f} > $DOLLAR{f%.json}.bin-runtime; jq -c .abi $DOLLAR{f} > $DOLLAR{f%.json}.abi; done"
// wrap LegacySfcWrapper with golang
//go:generate mkdir -p ./contract/sfc100
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/LegacySfcWrapper.bin --abi=./contract/solc/LegacySfcWrapper.abi --pkg=sfc100 --type=Contract --out=contract/sfc100/contract.go
//go:generate bash -c "(echo -ne '\nvar ContractBinRuntime = \"'; cat contract/solc/LegacySfcWrapper.bin-runtime; echo '\"') >> contract/sfc100/contract.go"
// wrap NetworkInitializer with golang
//go:generate mkdir -p ./contract/netinit100
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/NetworkInitializer.bin --abi=./contract/solc/NetworkInitializer.abi --pkg=netinit100 --type=Contract --out=contract/netinit100/contract.go
//go:generate bash -c "(echo -ne '\nvar ContractBinRuntime = \"'; cat contract/solc/NetworkInitializer.bin-runtime; echo '\"') >> contract/netinit100/contract.go"
// wrap NodeDriver with golang
//go:generate mkdir -p ./contract/driver100
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/NodeDriver.bin --abi=./contract/solc/NodeDriver.abi --pkg=driver100 --type=Contract --out=contract/driver100/contract.go
//go:generate bash -c "(echo -ne '\nvar ContractBinRuntime = \"'; cat contract/solc/NodeDriver.bin-runtime; echo '\"') >> contract/driver100/contract.go"
// wrap NodeDriverAuth with golang
//go:generate mkdir -p ./contract/driverauth100
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/NodeDriverAuth.bin --abi=./contract/solc/NodeDriverAuth.abi --pkg=driverauth100 --type=Contract --out=contract/driverauth100/contract.go
//go:generate bash -c "(echo -ne '\nvar ContractBinRuntime = \"'; cat contract/solc/NodeDriverAuth.bin-runtime; echo '\"') >> contract/driverauth100/contract.go"

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/driver100"
	"github.com/Fantom-foundation/go-opera/gossip/contract/driverauth100"
	"github.com/Fantom-foundation/go-opera/gossip/contract/netinit100"
	"github.com/Fantom-foundation/go-opera/gossip/contract/sfc100"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/opera/genesis/driver"
	"github.com/Fantom-foundation/go-opera/opera/genesis/driverauth"
	"github.com/Fantom-foundation/go-opera/opera/genesis/netinit"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/Fantom-foundation/go-opera/utils"
)

func TestSFC(t *testing.T) {
	logger.SetTestMode(t)
	logger.SetLevel("debug")

	env := newTestEnv()
	defer env.Close()

	/*
		sfcProxy, err := sfcproxy.NewContract(sfc.ContractAddress, env)
		require.NoError(t, err)

		var (
			sfc10 *sfc100.Contract
		)
	*/

	_ = true &&

		t.Run("Genesis SFC", func(t *testing.T) {
			require := require.New(t)

			exp := sfc.GetContractBin()
			got, err := env.CodeAt(nil, sfc.ContractAddress, nil)
			require.NoError(err)
			require.Equal(exp, got, "genesis SFC contract")
			require.Equal(exp, hexutil.MustDecode(sfc100.ContractBinRuntime), "genesis SFC contract version")
		}) &&

		t.Run("Network initializer", func(t *testing.T) {
			require := require.New(t)

			exp := netinit.GetContractBin()
			got, err := env.CodeAt(nil, netinit.ContractAddress, nil)
			require.NoError(err)
			require.NotEmpty(exp, "genesis NetworkInitializer contract")
			require.Empty(got, "genesis NetworkInitializer should be destructed")
			require.Equal(exp, hexutil.MustDecode(netinit100.ContractBinRuntime), "genesis NetworkInitializer contract version")
		}) &&

		t.Run("Genesis Driver", func(t *testing.T) {
			require := require.New(t)

			exp := driver.GetContractBin()
			got, err := env.CodeAt(nil, driver.ContractAddress, nil)
			require.NoError(err)
			require.Equal(exp, got, "genesis Driver contract")
			require.Equal(exp, hexutil.MustDecode(driver100.ContractBinRuntime), "genesis Driver contract version")
		}) &&

		t.Run("Genesis DriverAuth", func(t *testing.T) {
			require := require.New(t)

			exp := driverauth.GetContractBin()
			got, err := env.CodeAt(nil, driverauth.ContractAddress, nil)
			require.NoError(err)
			require.Equal(exp, got, "genesis DriverAuth contract")
			require.Equal(exp, hexutil.MustDecode(driverauth100.ContractBinRuntime), "genesis DriverAuth contract version")
		}) &&

		t.Run("Some transfers I", func(t *testing.T) {
			cicleTransfers(t, env, 1)
		}) /* &&

		// TODO: up to v1.0.0
		t.Run("Upgrade to develop", func(t *testing.T) {
			require := require.New(t)
			admin := 1

			rr := env.ApplyBlock(nextEpoch,
				env.Contract(admin, utils.ToFtm(0), sfc100.ContractBin),
			)
			newImpl := rr[0].ContractAddress

			tx, err := sfcProxy.ContractTransactor.Upgrade(env.Payer(admin), newImpl)
			require.NoError(err)
			env.ApplyBlock(sameEpoch, tx)

			sfc10, err = sfc100.NewContract(sfc.ContractAddress, env)
			require.NoError(err)

			exp := hexutil.MustDecode(sfc100.ContractBinRuntime)
			got, err := env.CodeAt(nil, newImpl, nil)
			require.NoError(err)
			require.Equal(exp, got, "new SFC contract")

			epoch, err := sfc10.ContractCaller.CurrentEpoch(env.ReadOnly())
			require.NoError(err)
			require.Equal(0, epoch.Cmp(big.NewInt(3)), "current epoch %s", epoch.String())
		})*/
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
