package gossip

// Called contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/called:/src -v $(pwd)/contract:/dst ethereum/solc:0.8.4 -o /dst/solc/ --optimize --optimize-runs=200 --bin --abi --allow-paths /src --overwrite /src/Called.sol"
//go:generate bash -c "go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/Called.bin --abi=./contract/solc/Called.abi --pkg=called --type=Contract --out=./contract/called/contract.go"

// Caller contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/caller:/src -v $(pwd)/contract:/dst ethereum/solc:0.8.4 -o /dst/solc/ --optimize --optimize-runs=200 --bin --abi --allow-paths /src --overwrite /src/Caller.sol"
//go:generate bash -c "go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./contract/solc/Caller.bin --abi=./contract/solc/Caller.abi --pkg=caller --type=Contract --out=./contract/caller/contract.go"

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/called"
	"github.com/Fantom-foundation/go-opera/gossip/contract/caller"
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
		addr     common.Address
		tx       *types.Transaction
		contract *caller.Contract
		library  *called.Contract
		err      error
	)

	const (
		admin = 1
		key   = 9
	)

	addr, tx, library, err = called.DeployContract(env.Payer(admin), env)
	require.NoError(err)
	require.NotNil(library)
	env.ApplyBlock(time.Second, tx)
	env.incNonce(env.Address(admin))

	_, tx, contract, err = caller.DeployContract(env.Payer(admin), env, addr)
	require.NoError(err)
	require.NotNil(contract)
	env.ApplyBlock(time.Second, tx)
	env.incNonce(env.Address(admin))

	tx, err = contract.Inc(env.Payer(admin), key)
	require.NoError(err)
	receipts := env.ApplyBlock(time.Second, tx)
	env.incNonce(env.Address(admin))
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
