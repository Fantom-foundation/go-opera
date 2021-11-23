package gossip

// Simple ballot contract
//go:generate bash -c "docker run --rm -v $(pwd)/contract/ballot:/src -v $(pwd)/contract:/dst ethereum/solc:0.5.12 -o /dst/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src --overwrite /src/Ballot.sol"
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=contract/solc/Ballot.bin --abi=contract/solc/Ballot.abi --pkg=ballot --type=Contract --out=contract/ballot/contract.go

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/contract/ballot"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
)

func BenchmarkBallotTxsProcessing(b *testing.B) {
	logger.SetLevel("warn")
	logger.SetTestMode(b)
	require := require.New(b)

	env := newTestEnv(3)
	defer env.Close()

	for i := 0; i < b.N; i++ {
		count := 10

		proposals := [][32]byte{
			ballotOption("Option 1"),
			ballotOption("Option 2"),
			ballotOption("Option 3"),
		}

		// contract deploy
		addr, tx, cBallot, err := ballot.DeployContract(env.Pay(1), env, proposals)
		require.NoError(err)
		require.NotNil(cBallot)
		r := env.ApplyTxs(nextEpoch, tx)

		require.Equal(addr, r[0].ContractAddress)

		admin, err := cBallot.Chairperson(env.ReadOnly())
		require.NoError(err)
		require.Equal(env.Address(1), admin)

		txs := make([]*types.Transaction, 0, count-1)
		flushTxs := func() {
			if len(txs) != 0 {
				env.ApplyTxs(nextEpoch, txs...)
			}
			txs = txs[:0]
		}

		// Init accounts
		for i := 2; i <= count; i++ {
			tx := env.Transfer(1, idx.ValidatorID(i), utils.ToFtm(10))
			txs = append(txs, tx)
			if len(txs) > 2 {
				flushTxs()
			}
		}
		flushTxs()

		// GiveRightToVote
		for i := 1; i <= count; i++ {
			tx, err := cBallot.GiveRightToVote(env.Pay(1), env.Address(idx.ValidatorID(i)))
			require.NoError(err)
			txs = append(txs, tx)
			if len(txs) > 2 {
				flushTxs()
			}
		}
		flushTxs()

		// Vote
		for i := 1; i <= count; i++ {
			proposal := big.NewInt(int64(i % len(proposals)))
			tx, err := cBallot.Vote(env.Pay(idx.ValidatorID(i)), proposal)
			require.NoError(err)
			txs = append(txs, tx)
			if len(txs) > 2 {
				flushTxs()
			}
		}
		flushTxs()

		// Winner
		_, err = cBallot.WinnerName(env.ReadOnly())
		require.NoError(err)
	}
}

func ballotOption(str string) (res [32]byte) {
	buf := []byte(str)
	if len(buf) > 32 {
		panic("string too long")
	}
	copy(res[:], buf)
	return
}

func uniqName() string {
	return hash.FakeHash(rand.Int63()).Hex()
}
