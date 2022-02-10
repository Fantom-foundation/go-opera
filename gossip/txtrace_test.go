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

	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/called"
	"github.com/Fantom-foundation/go-opera/gossip/contract/caller"
	"github.com/Fantom-foundation/go-opera/logger"
)

func TestTxTracing(t *testing.T) {
	logger.SetTestMode(t)

	env := newTestEnv(2, 3, func(cfg *StoreConfig) {
		cfg.TraceTransactions = true
	})
	defer env.Close()

	var (
		contract *caller.Contract
		library  *called.Contract
	)
	const (
		admin = 1
		key   = 9
	)

	addr, tx, library, err := called.DeployContract(env.Payer(admin), env)
	require.NoError(t, err)
	require.NotNil(t, library)
	_, err = env.ApplyTxs(time.Second, tx)
	require.NoError(t, err)
	env.incNonce(env.Address(admin))

	_, tx, contract, err = caller.DeployContract(env.Payer(admin), env, addr)
	require.NoError(t, err)
	require.NotNil(t, contract)
	_, err = env.ApplyTxs(time.Second, tx)
	require.NoError(t, err)
	env.incNonce(env.Address(admin))

	t.Run("SubReverts", func(t *testing.T) {
		require := require.New(t)

		tx, err := contract.Inc(env.Payer(admin), key)
		require.NoError(err)
		receipts, err := env.ApplyTxs(time.Second, tx)
		require.NoError(err)
		require.NotEmpty(receipts)
		env.incNonce(env.Address(admin))

		actions, err := env.EthAPI.TxTraceByHash(context.Background(), tx.Hash())
		require.NoError(err)
		require.NotNil(actions)
		require.Len(*actions, 3)

		reverted := 0
		for _, action := range *actions {
			if action.Error == "Reverted" {
				reverted += 1
			}
		}
		require.Equal(1, reverted)

		// visulization
		traceStr, err := json.Marshal(actions)
		require.NoError(err)
		t.Logf("traces: %s", string(traceStr))
	})
}
