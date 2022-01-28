package gossip

// SynthereumDeployer contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/SynthereumDeployer:/src -v $(pwd)/contract:/dst ethereum/solc:0.8.4 -o /dst/solc/ --optimize --optimize-runs=200 --bin --abi --allow-paths /src --overwrite /src/Deployer.sol"
// NOTE: you have to use abigen after github.com/ethereum/go-ethereum/pull/23940, than fix contract/SynthereumDeployer/contract.go manually
//go:generate bash -c "cd ${GOPATH}/src/github.com/ethereum/go-ethereum && go run ./cmd/abigen --bin=${PWD}/contract/solc/SynthereumDeployer.bin --abi=${PWD}/contract/solc/SynthereumDeployer.abi --pkg=SynthereumDeployer --type=Contract --out=${PWD}/contract/SynthereumDeployer/contract.go"

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/SynthereumDeployer"
	"github.com/Fantom-foundation/go-opera/logger"
)

func TestTxTracing(t *testing.T) {
	require := require.New(t)

	logger.SetTestMode(t)
	logger.SetLevel("debug")

	env := newTestEnv()
	defer env.Close()

	var (
		deployer *SynthereumDeployer.Contract
		err      error
	)

	roles := SynthereumDeployer.SynthereumDeployerRoles{
		Admin:      env.Address(1),
		Maintainer: env.Address(1),
	}
	_, _, deployer, err = SynthereumDeployer.DeployContract(env.Payer(1), env, common.Address{}, roles)
	require.NoError(err)
	require.NotEmpty(deployer)

}
