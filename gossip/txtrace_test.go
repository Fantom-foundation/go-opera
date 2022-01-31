package gossip

// SynthereumFactoryVersioning contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/SynthereumFactoryVersioning:/src -v $(pwd)/contract:/dst ethereum/solc:0.8.4 -o /dst/solc/ --optimize --optimize-runs=200 --bin --abi --allow-paths /src --overwrite /src/FactoryVersioning.sol"
// NOTE: you have to use abigen after github.com/ethereum/go-ethereum/pull/23940, than fix contract/SynthereumFactoryVersioning/contract.go manually
//go:generate bash -c "cd ${GOPATH}/src/github.com/ethereum/go-ethereum && go run ./cmd/abigen --bin=${PWD}/contract/solc/SynthereumDeployer.bin --abi=${PWD}/contract/solc/SynthereumFactoryVersioning.abi --pkg=SynthereumFactoryVersioning --type=Contract --out=${PWD}/contract/SynthereumFactoryVersioning/contract.go"

// SynthereumDeployer contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/SynthereumDeployer:/src -v $(pwd)/contract:/dst ethereum/solc:0.8.4 -o /dst/solc/ --optimize --optimize-runs=200 --bin --abi --allow-paths /src --overwrite /src/Deployer.sol"
// NOTE: you have to use abigen after github.com/ethereum/go-ethereum/pull/23940, than fix contract/SynthereumDeployer/contract.go manually
//go:generate bash -c "cd ${GOPATH}/src/github.com/ethereum/go-ethereum && go run ./cmd/abigen --bin=${PWD}/contract/solc/SynthereumDeployer.bin --abi=${PWD}/contract/solc/SynthereumDeployer.abi --pkg=SynthereumDeployer --type=Contract --out=${PWD}/contract/SynthereumDeployer/contract.go"

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/SynthereumDeployer"
	"github.com/Fantom-foundation/go-opera/gossip/contract/SynthereumFactoryVersioning"
	"github.com/Fantom-foundation/go-opera/logger"
)

func TestTxTracing(t *testing.T) {
	require := require.New(t)

	logger.SetTestMode(t)
	logger.SetLevel("debug")

	env := newTestEnv()
	defer env.Close()
	backend := &EthAPIBackend{state: env.stateReader}

	var (
		addr       common.Address
		tx         *types.Transaction
		deployer   *SynthereumDeployer.Contract
		versioning *SynthereumFactoryVersioning.Contract
		err        error
	)
	var (
		admin      = env.Payer(1)
		maintainer = env.Payer(2)
		roles      = SynthereumFactoryVersioning.SynthereumFactoryVersioningRoles{
			Admin:      env.Address(1),
			Maintainer: env.Address(2),
		}
	)

	addr, tx, versioning, err = SynthereumFactoryVersioning.DeployContract(admin, env, roles)
	env.incNonce(roles.Admin)
	require.NoError(err)
	require.NotNil(versioning)
	env.ApplyBlock(time.Second, tx)

	_, tx, deployer, err = SynthereumDeployer.DeployContract(admin, env, addr, SynthereumDeployer.SynthereumDeployerRoles{
		Admin:      roles.Admin,
		Maintainer: roles.Maintainer,
	})
	env.incNonce(roles.Admin)
	require.NoError(err)
	require.NotNil(deployer)
	env.ApplyBlock(time.Second, tx)

	tx, err = deployer.DeployPool(maintainer,
		5,
		hexutils.HexToBytes("000000000000000000000000000000000000000000000000000000000000002000000000000000000000000004068da6c83afcfa0e13ba15a6696662335d5b7500000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000001e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000024435f5f12ea2977f4b4a3ad990600fd5387732f000000000000000000000000646877b5ea314627426429def0987b15fb8dbb9b000000000000000000000000c31249ba48763df46388ba5c4e7565d62ed4801c000000000000000000000000000000000000000000000000016345785d8a0000000000000000000000000000000000000000000000000000000000000000022045555255534400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000e398811bec6800000000000000000000000000000000000000000000000000009b6e64a8ec60000000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000154a61727669732053796e746865746963204575726f000000000000000000000000000000000000000000000000000000000000000000000000000000000000046a455552000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008e1bc9bf04000000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000001000000000000000000000000c31249ba48763df46388ba5c4e7565d62ed4801c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000064"),
	)
	env.incNonce(roles.Maintainer)
	require.NoError(err)
	require.NotNil(tx)
	receipts := env.ApplyBlock(time.Second, tx)
	require.NotEmpty(receipts)

	trace, err := backend.TxTraceByHash(context.Background(), tx.Hash())
	require.NoError(err)
	require.NotEmpty(trace)

	// visulization
	receiptStr, err := json.Marshal(receipts)
	require.NoError(err)
	t.Logf("receipts: %s", string(receiptStr))

	traceStr, err := json.Marshal(trace)
	require.NoError(err)
	t.Logf("trace: %s", string(traceStr))
}
