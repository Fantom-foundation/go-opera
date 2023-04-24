package gossip

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/Fantom-foundation/go-opera/gossip/contract/aatester"
	"github.com/Fantom-foundation/go-opera/gossip/contract/wallet"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

var aaGasLimit uint64 = params.VerificationGasCap + 30_000
var aaGasPrice = new(big.Int).SetUint64(1e12)

func (env *testEnv) walletSignature(address common.Address, privateKey *ecdsa.PrivateKey) wallet.WalletSignature {
	nonce := new(uint256.Int).SetUint64(env.State().GetNonce(address))
	hashedData := crypto.Keccak256(address.Bytes(), common.LeftPadBytes(nonce.Bytes(), 32))
	signature, _ := crypto.Sign(hashedData, privateKey)
	var (
		r [32]byte
		s [32]byte
	)
	copy(r[:], signature[:32])
	copy(s[:], signature[32:64])
	return wallet.WalletSignature{
		V: uint8(signature[64]) + 27,
		R: r,
		S: s,
	}
}

func TestSimpleAAWallet(t *testing.T) {
	logger.SetLevel("warn")
	logger.SetTestMode(t)
	require := require.New(t)

	env := newTestEnv(1, 1)
	defer env.Close()

	payer := env.Pay(1)
	owner := env.Address(2)
	ownerKey := env.privateKey(2)

	initialBalance := uint64(1e18)
	payer.Value = new(big.Int).SetUint64(initialBalance)
	addr, tx, cWallet, err := wallet.DeployContract(payer, env, owner)
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
		signature := env.walletSignature(addr, ownerKey)
		data := cWallet.Transfer(recipient, amount, nil, signature)
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

	callOpts := &bind.CallOpts{From: types.AAEntryPoint}
	returnedOwner, err := cWallet.Owner(callOpts)
	require.NoError(err)
	require.Equal(returnedOwner, owner)

	newOwner := env.Address(3)
	newOwnerKey := env.privateKey(3)

	signature := env.walletSignature(addr, newOwnerKey)
	data := cWallet.Transfer(recipient, amount, nil, signature)
	nonce := env.State().GetNonce(addr)
	txdata := &types.AccountAbstractionTx{
		Nonce:    nonce,
		To:       &addr,
		Value:    common.Big0,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	}

	tx = types.NewTx(txdata)
	signed := tx.WithAASignature()

	// Execution reverted - wrong signature
	receipts, err = env.ApplyTxs(nextEpoch, signed)
	require.NoError(err)

	state = env.State()
	recipientBalance = state.GetBalance(recipient)
	require.Equal(recipientBalance, sentAmount)

	contractBalance = state.GetBalance(addr)
	require.Equal(contractBalance, expectedBalance)

	signature = env.walletSignature(addr, ownerKey)
	data = cWallet.ChangeOwner(newOwner, signature)
	nonce = env.State().GetNonce(addr)
	txdata = &types.AccountAbstractionTx{
		Nonce:    nonce,
		To:       &addr,
		Value:    common.Big0,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	}

	tx = types.NewTx(txdata)
	signed = tx.WithAASignature()

	receipts, err = env.ApplyTxs(nextEpoch, signed)
	require.NoError(err)

	returnedOwner, err = cWallet.Owner(callOpts)
	require.NoError(err)
	require.Equal(returnedOwner, newOwner)
}

func (env *testEnv) buildAATx(to common.Address, data []byte) *types.Transaction {
	state := env.State()
	nonce := state.GetNonce(to)

	txdata := &types.AccountAbstractionTx{
		Nonce:    nonce,
		To:       &to,
		Value:    common.Big0,
		Gas:      aaGasLimit,
		GasPrice: aaGasPrice,
		Data:     data,
	}

	tx := types.NewTx(txdata)
	return tx.WithAASignature()
}

func resetTester(env *testEnv, t *testing.T, contract common.Address) {
	require := require.New(t)
	cTester, err := aatester.NewContract(contract, env)
	require.NoError(err)

	data := cTester.Reset()
	tx := env.buildAATx(contract, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	callOpts := &bind.CallOpts{From: types.AAEntryPoint}

	nonce, err := cTester.Nonce(callOpts)
	require.NoError(err)
	require.Equal(nonce.Sign(), 0)

	balance, err := cTester.Balance(callOpts)
	require.NoError(err)
	require.Equal(balance.Sign(), 0)

	gasPrice, err := cTester.GasPrice(callOpts)
	require.NoError(err)
	require.Equal(gasPrice.Sign(), 0)

	gasLeft, err := cTester.GasLeft(callOpts)
	require.NoError(err)
	require.Equal(gasLeft.Sign(), 0)

	origin, err := cTester.Origin(callOpts)
	require.NoError(err)
	require.Equal(origin, common.Address{})

	sender, err := cTester.Sender(callOpts)
	require.NoError(err)
	require.Equal(sender, common.Address{})
}

func TestAATransactions(t *testing.T) {
	logger.SetLevel("warn")
	logger.SetTestMode(t)
	require := require.New(t)

	env := newTestEnv(1, 1)
	defer env.Close()

	payer := env.Pay(1)
	initialBalance := uint64(1e18)
	payer.Value = new(big.Int).SetUint64(initialBalance)
	addr, tx, cTester, err := aatester.DeployContract(payer, env)
	require.NoError(err)
	require.NotNil(cTester)

	receipts, err := env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)
	require.Equal(addr, receipts[0].ContractAddress)

	resetTester(env, t, addr)

	data := cTester.SetNonce()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	expectedNonce := tx.Nonce()

	data = cTester.SetOriginBeforePaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	data = cTester.SetSenderBeforePaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	callOpts := &bind.CallOpts{From: types.AAEntryPoint}

	nonce, err := cTester.Nonce(callOpts)
	require.NoError(err)
	require.Equal(expectedNonce, nonce.Uint64())

	origin, err := cTester.Origin(callOpts)
	require.NoError(err)
	require.Equal(types.AAEntryPoint, origin)

	sender, err := cTester.Sender(callOpts)
	require.NoError(err)
	require.Equal(types.AAEntryPoint, sender)

	resetTester(env, t, addr)

	data = cTester.SetNonceBeforePaygasAndRevert()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	expectedNonce = tx.Nonce()
	nonce, err = cTester.Nonce(callOpts)
	require.NoError(err)
	require.Equal(expectedNonce, nonce.Uint64())

	state := env.State()
	require.Equal(state.GetNonce(addr), expectedNonce+1)

	resetTester(env, t, addr)

	data = cTester.SetOriginAfterPaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	data = cTester.SetSenderAfterPaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	data = cTester.SetGasPrice()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	origin, err = cTester.Origin(callOpts)
	require.NoError(err)
	require.Equal(types.AAEntryPoint, origin)

	sender, err = cTester.Sender(callOpts)
	require.NoError(err)
	require.Equal(types.AAEntryPoint, sender)

	gasPrice, err := cTester.GasPrice(callOpts)
	require.NoError(err)
	require.Equal(aaGasPrice, gasPrice)

	resetTester(env, t, addr)

	data = cTester.SetBalanceBeforePaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	balance, err := cTester.Balance(callOpts)
	require.NoError(err)
	require.Equal(balance.Sign(), 0)

	data = cTester.SetBalanceAfterPaygas()
	tx = env.buildAATx(addr, data)

	state = env.State()
	receipts, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	balance, err = cTester.Balance(callOpts)
	require.NoError(err)
	require.Equal(balance.Cmp(common.Big0), 1)

	resetTester(env, t, addr)

	data = cTester.SetGasLeftBeforePaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	gasLeftBefore, err := cTester.GasLeft(callOpts)
	require.NoError(err)

	data = cTester.SetGasLeftAfterPaygas()
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	gasLeftAfter, err := cTester.GasLeft(callOpts)
	require.NoError(err)

	limitDifference := aaGasLimit - params.VerificationGasCap
	require.True(new(big.Int).Sub(gasLeftAfter, gasLeftBefore).Uint64() < limitDifference+uint64(100))

	resetTester(env, t, addr)

	data = cTester.CallSetOrigin(addr)
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	origin, err = cTester.Origin(callOpts)
	require.NoError(err)
	require.Equal(types.AAEntryPoint, origin)

	data = cTester.CallSetSender(addr)
	tx = env.buildAATx(addr, data)
	_, err = env.ApplyTxs(nextEpoch, tx)
	require.NoError(err)

	sender, err = cTester.Sender(callOpts)
	require.NoError(err)
	require.Equal(sender, addr)
}
