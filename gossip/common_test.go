package gossip

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis/gpos"
	"github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/gsignercache"
)

const (
	gasLimit       = uint64(21000)
	genesisStakers = 3
	genesisBalance = 1e18
	genesisStake   = 2 * 4e6

	maxEpochDuration = time.Hour
	sameEpoch        = maxEpochDuration / 1000
	nextEpoch        = maxEpochDuration
)

type testEnv struct {
	store *Store

	blockProcWg      sync.WaitGroup
	blockProcTasks   *workers.Workers
	blockProcModules BlockProc

	signer types.Signer

	lastBlock     idx.Block
	lastBlockTime time.Time
	lastState     hash.Hash
	validators    gpos.Validators
	stateReader   *EvmStateReader

	nonces map[common.Address]uint64

	epoch    idx.Epoch
	eventSeq idx.Event

	wg   sync.WaitGroup
	done chan struct{}
}

func newTestEnv() *testEnv {
	genStore := makegenesis.FakeGenesisStore(genesisStakers, utils.ToFtm(genesisBalance), utils.ToFtm(genesisStake))
	genesis := genStore.GetGenesis()

	genesis.Rules.Epochs.MaxEpochDuration = inter.Timestamp(maxEpochDuration)
	genesis.Rules.Blocks.MaxEmptyBlockSkipPeriod = 0

	store := NewMemStore()
	blockProc := DefaultBlockProc(genesis)
	_, err := store.ApplyGenesis(blockProc, genesis)
	if err != nil {
		panic(err)
	}

	env := &testEnv{
		blockProcModules: blockProc,
		store:            store,
		signer:           gsignercache.Wrap(types.NewEIP2930Signer(big.NewInt(int64(genesis.Rules.NetworkID)))),

		lastBlock:     1,
		lastState:     store.GetBlockState().FinalizedStateRoot,
		lastBlockTime: genesis.Time.Time(),
		validators:    genesis.Validators,
		stateReader: &EvmStateReader{
			store: store,
		},

		nonces: make(map[common.Address]uint64),

		done: make(chan struct{}),
	}

	env.blockProcTasks = workers.New(&env.wg, env.done, 1)
	env.blockProcTasks.Start(1)

	return env
}

func (env *testEnv) Close() {
	close(env.done)
	env.store.Close()
	env.wg.Wait()
}

func (env *testEnv) GetEvmStateReader() *EvmStateReader {
	return &EvmStateReader{
		store: env.store,
	}
}

// consensusCallbackBeginBlockFn returns single (for testEnv) callback instance.
// Note that onBlockEnd overwrites previous.
// Note that onBlockEnd would be run async.
func (env *testEnv) consensusCallbackBeginBlockFn(
	onBlockEnd func(block *inter.Block, preInternalReceipts, internalReceipts, externalReceipts types.Receipts),
) lachesis.BeginBlockFn {
	const txIndex = true
	callback := consensusCallbackBeginBlockFn(
		env.blockProcTasks,
		&env.blockProcWg,
		new(uint32),
		env.store,
		env.blockProcModules,
		txIndex,
		nil,
		nil,
		nil,
		onBlockEnd,
	)
	return callback
}

func (env *testEnv) ApplyBlock(spent time.Duration, txs ...*types.Transaction) (receipts types.Receipts) {
	env.lastBlock++
	env.lastBlockTime = env.lastBlockTime.Add(spent)

	eBuilder := inter.MutableEventPayload{}
	eBuilder.SetMedianTime(inter.Timestamp(env.lastBlockTime.UnixNano()))
	eBuilder.SetTxs(txs)
	event := eBuilder.Build()
	env.store.SetEvent(event)

	var waitForBlockEnd sync.WaitGroup
	waitForBlockEnd.Add(1)
	onBlockEnd := func(block *inter.Block, preInternalReceipts, internalReceipts, externalReceipts types.Receipts) {
		receipts = externalReceipts
		env.lastState = block.Root
		waitForBlockEnd.Done()
	}

	beginBlock := env.consensusCallbackBeginBlockFn(onBlockEnd)
	process := beginBlock(&lachesis.Block{
		Atropos: event.ID(),
	})

	process.ApplyEvent(event)
	_ = process.EndBlock()
	waitForBlockEnd.Wait()

	return
}

func (env *testEnv) Transfer(from int, to int, amount *big.Int) *types.Transaction {
	sender := env.Address(from)
	nonce, _ := env.PendingNonceAt(nil, sender)
	env.incNonce(sender)
	key := env.privateKey(from)
	receiver := env.Address(to)
	gp := env.store.GetRules().Economy.MinGasPrice
	tx := types.NewTransaction(nonce, receiver, amount, gasLimit, gp, nil)
	tx, err := types.SignTx(tx, env.signer, key)
	if err != nil {
		panic(err)
	}

	return tx
}

func (env *testEnv) Contract(from int, amount *big.Int, hex string) *types.Transaction {
	sender := env.Address(from)
	nonce, _ := env.PendingNonceAt(nil, sender)
	env.incNonce(sender)
	key := env.privateKey(from)
	gp := env.store.GetRules().Economy.MinGasPrice
	data := hexutil.MustDecode(hex)
	tx := types.NewContractCreation(nonce, amount, gasLimit*100000, gp, data)
	tx, err := types.SignTx(tx, env.signer, key)
	if err != nil {
		panic(err)
	}

	return tx
}

func (env *testEnv) privateKey(n int) *ecdsa.PrivateKey {
	key := makegenesis.FakeKey(n)
	return key
}

func (env *testEnv) Address(n int) common.Address {
	key := makegenesis.FakeKey(n)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return addr
}

func (env *testEnv) Payer(n int, amounts ...*big.Int) *bind.TransactOpts {
	key := env.privateKey(n)
	t := bind.NewKeyedTransactor(key)
	nonce, _ := env.PendingNonceAt(nil, env.Address(n))
	t.Nonce = big.NewInt(int64(nonce))
	t.Value = big.NewInt(0)
	for _, amount := range amounts {
		t.Value.Add(t.Value, amount)
	}
	t.GasLimit = env.stateReader.MaxGasLimit()
	t.GasPrice = env.stateReader.MinGasPrice()

	return t
}

func (env *testEnv) ReadOnly() *bind.CallOpts {
	return &bind.CallOpts{}
}

func (env *testEnv) State() *state.StateDB {
	statedb, _ := env.store.evm.StateDB(env.lastState)
	return statedb
}

func (env *testEnv) incNonce(account common.Address) {
	env.nonces[account] += 1
}

/*
 * bind.ContractCaller interface
 */

var (
	errBlockNumberUnsupported = errors.New("simulatedBackend cannot access blocks other than the latest block")
)

// CodeAt returns the code of the given account. This is needed to differentiate
// between contract internal errors and the local chain being out of sync.
func (env *testEnv) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	if blockNumber != nil && idx.Block(blockNumber.Uint64()) != env.lastBlock {
		return nil, errBlockNumberUnsupported
	}

	code := env.State().GetCode(contract)
	return code, nil
}

// ContractCall executes an Ethereum contract call with the specified data as the
// input.
func (env *testEnv) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if blockNumber != nil && idx.Block(blockNumber.Uint64()) != env.lastBlock {
		return nil, errBlockNumberUnsupported
	}

	h := env.GetEvmStateReader().GetHeader(common.Hash{}, uint64(env.lastBlock))
	block := &evmcore.EvmBlock{
		EvmHeader: *h,
	}

	rval, _, _, err := env.callContract(ctx, call, block, env.State())
	return rval, err
}

// callContract implements common code between normal and pending contract calls.
// state is modified during execution, make sure to copy it if necessary.
func (env *testEnv) callContract(
	ctx context.Context, call ethereum.CallMsg, block *evmcore.EvmBlock, state *state.StateDB,
) (
	ret []byte, usedGas uint64, failed bool, err error,
) {
	// Ensure message is initialized properly.
	if call.GasPrice == nil {
		call.GasPrice = big.NewInt(1)
	}
	if call.Gas == 0 {
		call.Gas = 50000000
	}
	if call.Value == nil {
		call.Value = new(big.Int)
	}
	// Set infinite balance to the fake caller account.
	from := state.GetOrNewStateObject(call.From)
	from.SetBalance(big.NewInt(math.MaxInt64))

	msg := callmsg{call}

	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	txContext := evmcore.NewEVMTxContext(msg)
	context := evmcore.NewEVMBlockContext(block.Header(), env.GetEvmStateReader(), nil)
	vmenv := vm.NewEVM(context, txContext, state, env.store.GetRules().EvmChainConfig(), opera.DefaultVMConfig)
	gaspool := new(evmcore.GasPool).AddGas(math.MaxUint64)
	res, err := evmcore.NewStateTransition(vmenv, msg, gaspool).TransitionDb()

	ret, usedGas, failed = res.Return(), res.UsedGas, res.Failed()
	return
}

/*
 * bind.ContractTransactor interface
 */

// PendingCodeAt returns the code of the given account in the pending state.
func (env *testEnv) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	code := env.State().GetCode(account)
	return code, nil
}

// PendingNonceAt retrieves the current pending nonce associated with an account.
func (env *testEnv) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	nonce := env.nonces[account]
	return nonce, nil
}

// SuggestGasPrice retrieves the currently suggested gas price to allow a timely
// execution of a transaction.
func (env *testEnv) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return env.store.GetRules().Economy.MinGasPrice, nil
}

// EstimateGas tries to estimate the gas needed to execute a specific
// transaction based on the current pending state of the backend blockchain.
// There is no guarantee that this is the true gas limit requirement as other
// transactions may be added or removed by miners, but it should provide a basis
// for setting a reasonable default.
func (env *testEnv) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	if call.To == nil {
		gas = gasLimit * 10000
	} else {
		gas = gasLimit * 10
	}
	return
}

// SendTransaction injects the transaction into the pending pool for execution.
func (env *testEnv) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	// do nothing to avoid executing by transactor, only generating needed
	return nil
}

/*
 *  bind.ContractFilterer interface
 */

// FilterLogs executes a log filter operation, blocking during execution and
// returning all the results in one batch.
func (env *testEnv) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	panic("not implemented yet")
	return nil, nil
}

// SubscribeFilterLogs creates a background log filtering operation, returning
// a subscription immediately, which can be used to stream the found events.
func (env *testEnv) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	panic("not implemented yet")
	return nil, nil
}

// callmsg implements evmcore.Message to allow passing it as a transaction simulator.
type callmsg struct {
	ethereum.CallMsg
}

func (m callmsg) From() common.Address         { return m.CallMsg.From }
func (m callmsg) To() *common.Address          { return m.CallMsg.To }
func (m callmsg) GasPrice() *big.Int           { return m.CallMsg.GasPrice }
func (m callmsg) Gas() uint64                  { return m.CallMsg.Gas }
func (m callmsg) Value() *big.Int              { return m.CallMsg.Value }
func (m callmsg) Nonce() uint64                { return 0 }
func (m callmsg) CheckNonce() bool             { return false }
func (m callmsg) Data() []byte                 { return m.CallMsg.Data }
func (m callmsg) AccessList() types.AccessList { return nil }
