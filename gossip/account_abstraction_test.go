package gossip

import (
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/go-opera/gossip/contract/wallet"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestAccountAbstraction(t *testing.T) {
	logger.SetLevel("warn")
	logger.SetTestMode(t)
	require := require.New(t)

	env := newTestEnv(1, 1)
	defer env.Close()

	password := []byte("password")
	passwordHash := crypto.Keccak256Hash(password)

	// contract deploy
	payer := env.Pay(1)
	initialBalance := uint64(1e18)
	payer.Value = new(big.Int).SetUint64(initialBalance)
	addr, tx, cWallet, err := wallet.DeployContract(payer, env, passwordHash)
	require.NoError(err)
	require.NotNil(cWallet)

	receipts, err := env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)
	require.Equal(addr, receipts[0].ContractAddress)

	recipient := common.HexToAddress("0xcccccccccccccccccccccccccccccccccccccccc")
	amount := new(big.Int).SetUint64(100)
	gasPrice := new(big.Int).SetUint64(1e12)
	gasLimit := uint64(80000)

	count := 10
	gasUsed := uint64(0)
	for i := 0; i < count; i++ {
		newPassword := []byte("new_password")
		binary.BigEndian.AppendUint32(newPassword, uint32(count))
		newPasswordHash := crypto.Keccak256Hash(newPassword)

		data := cWallet.ContractTransactor.TransferData(password, newPasswordHash, recipient, amount, nil)
		state := env.State()
		nonce := state.GetNonce(addr)
		txdata := &types.AccountAbstractionTx{
			Nonce:    nonce,
			To:       &addr,
			Value:    common.Big0,
			Gas:      gasLimit,
			GasPrice: gasPrice,
			Data:     data,
		}

		password = newPassword
		tx = types.NewTx(txdata)
		signed := tx.WithAASignature()

		receipts, err = env.ApplyTxs(nextEpoch, signed)
		require.NoError(err)

		gasUsed += receipts[0].GasUsed
	}

	state := env.State()
	sentAmount := new(big.Int).Mul(amount, new(big.Int).SetUint64(uint64(count)))
	recipientBalance := state.GetBalance(recipient)
	require.Equal(recipientBalance, sentAmount)

	expectedBalance := new(big.Int).SetUint64(initialBalance)
	expectedBalance = expectedBalance.Sub(expectedBalance, new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasUsed)))
	expectedBalance = expectedBalance.Sub(expectedBalance, sentAmount)

	contractBalance := state.GetBalance(addr)
	require.Equal(contractBalance, expectedBalance)
}
