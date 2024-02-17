package gossip

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/ethapi"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/gossip/emitter"
	"github.com/Fantom-foundation/go-opera/integration/makefakegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/adapters/vecmt2dagidx"
	"github.com/Fantom-foundation/go-opera/valkeystore"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

const (
	gasPrice = uint64(1e12)

	genesisBalance = 1e18
	genesisStake   = 2 * 4e6

	maxEpochDuration = time.Hour
	sameEpoch        = maxEpochDuration / 1000
	nextEpoch        = maxEpochDuration
)

type callbacks struct {
	buildEvent       func(e *inter.MutableEventPayload)
	onEventConfirmed func(e inter.EventI)
}

type testEnv struct {
	t        time.Time
	nonces   map[common.Address]uint64
	callback callbacks
	*Service
	transactor *ethapi.PublicTransactionPoolAPI
	signer     valkeystore.SignerI
	pubkeys    []validatorpk.PubKey
}

func panics(name string) func(error) {
	return func(err error) {
		log.Crit(fmt.Sprintf("%s error", name), "err", err)
	}
}

type testGossipStoreAdapter struct {
	*Store
}

func (g *testGossipStoreAdapter) GetEvent(id hash.Event) dag.Event {
	e := g.Store.GetEvent(id)
	if e == nil {
		return nil
	}
	return e
}

func makeTestEngine(gdb *Store) (*abft.Lachesis, *vecmt.Index) {
	cdb := abft.NewMemStore()
	_ = cdb.ApplyGenesis(&abft.Genesis{
		Epoch:      gdb.GetEpoch(),
		Validators: gdb.GetValidators(),
	})
	vecClock := vecmt.NewIndex(panics("Vector clock"), vecmt.LiteConfig())
	engine := abft.NewLachesis(cdb, &testGossipStoreAdapter{gdb}, vecmt2dagidx.Wrap(vecClock), panics("Lachesis"), abft.LiteConfig())
	return engine, vecClock
}

type testEmitterWorldExternal struct {
	emitter.External
	env *testEnv
}

func (em testEmitterWorldExternal) Build(e *inter.MutableEventPayload, onIndexed func()) error {
	e.SetCreationTime(inter.Timestamp(em.env.t.UnixNano()))
	if em.env.callback.buildEvent != nil {
		em.env.callback.buildEvent(e)
	}
	err := em.External.Build(e, onIndexed)

	return err
}

func (em testEmitterWorldExternal) Broadcast(emitted *inter.EventPayload) {
	// PM listens and will broadcast it
	em.env.feed.newEmittedEvent.Send(emitted)
}

type testConfirmedEventsProcessor struct {
	blockproc.ConfirmedEventsProcessor
	env *testEnv
}

func (p testConfirmedEventsProcessor) ProcessConfirmedEvent(e inter.EventI) {
	if p.env.callback.onEventConfirmed != nil {
		p.env.callback.onEventConfirmed(e)
	}

	p.ConfirmedEventsProcessor.ProcessConfirmedEvent(e)
}

type testConfirmedEventsModule struct {
	blockproc.ConfirmedEventsModule
	env *testEnv
}

func (m testConfirmedEventsModule) Start(bs iblockproc.BlockState, es iblockproc.EpochState) blockproc.ConfirmedEventsProcessor {
	p := m.ConfirmedEventsModule.Start(bs, es)
	return testConfirmedEventsProcessor{p, m.env}
}

func newTestEnv(firstEpoch idx.Epoch, validatorsNum idx.Validator) *testEnv {
	rules := opera.FakeNetRules()
	rules.Epochs.MaxEpochDuration = inter.Timestamp(maxEpochDuration)
	rules.Blocks.MaxEmptyBlockSkipPeriod = 0

	genStore := makefakegenesis.FakeGenesisStoreWithRulesAndStart(validatorsNum, utils.ToFtm(genesisBalance), utils.ToFtm(genesisStake), rules, firstEpoch, 2)
	genesis := genStore.Genesis()

	store := NewMemStore()
	_, err := store.ApplyGenesis(genesis)
	if err != nil {
		panic(err)
	}

	// install blockProc callbacks
	env := &testEnv{
		t:      store.GetGenesisTime().Time(),
		nonces: make(map[common.Address]uint64),
	}
	blockProc := DefaultBlockProc()
	blockProc.EventsModule = testConfirmedEventsModule{blockProc.EventsModule, env}

	engine, vecClock := makeTestEngine(store)

	// create the service
	txPool := newDummyTxPool()
	env.Service, err = newService(DefaultConfig(cachescale.Identity), store, blockProc, engine, vecClock, func(_ evmcore.StateReader) TxPool {
		return txPool
	})
	if err != nil {
		panic(err)
	}
	txPool.Signer = env.Service.EthAPI.signer
	env.transactor = ethapi.NewPublicTransactionPoolAPI(env.Service.EthAPI, new(ethapi.AddrLocker))

	err = engine.Bootstrap(env.GetConsensusCallbacks())
	if err != nil {
		panic(err)
	}

	valKeystore := valkeystore.NewDefaultMemKeystore()
	env.signer = valkeystore.NewSigner(valKeystore)

	// register emitters
	for i := idx.Validator(0); i < validatorsNum; i++ {
		cfg := emitter.DefaultConfig()
		vid := store.GetValidators().GetID(i)
		pubkey := store.GetEpochState().ValidatorProfiles[vid].PubKey
		cfg.Validator = emitter.ValidatorConfig{
			ID:     vid,
			PubKey: pubkey,
		}
		cfg.EmitIntervals = emitter.EmitIntervals{}
		cfg.MaxParents = idx.Event(validatorsNum/2 + 1)
		cfg.MaxTxsPerAddress = 10000000
		_ = valKeystore.Add(pubkey, crypto.FromECDSA(makefakegenesis.FakeKey(vid)), validatorpk.FakePassword)
		_ = valKeystore.Unlock(pubkey, validatorpk.FakePassword)
		world := env.EmitterWorld(env.signer)
		world.External = testEmitterWorldExternal{world.External, env}
		em := emitter.NewEmitter(cfg, world)
		env.RegisterEmitter(em)
		env.pubkeys = append(env.pubkeys, pubkey)
		em.Start()
		em.Stop() // to control emitting manually
	}

	_ = env.store.GenerateSnapshotAt(common.Hash(store.GetBlockState().FinalizedStateRoot), false)
	env.blockProcTasks.Start(1)
	env.verWatcher.Start()

	return env
}

func (env *testEnv) Close() {
	env.verWatcher.Stop()
	env.store.Close()
	env.tflusher.Stop()
}

func (env *testEnv) GetEvmStateReader() *EvmStateReader {
	return &EvmStateReader{
		store: env.store,
	}
}

func (env *testEnv) BlockTxs(spent time.Duration, txs ...*types.Transaction) (types.Receipts, error) {
	// Just `env.applyTxs(spent, true, txs...)` does not work guaranteed because of gas rules.
	// So we make single-event block and skip emitting process.
	env.t = env.t.Add(spent)

	mutEvent := &inter.MutableEventPayload{}
	mutEvent.SetVersion(1) // LLR
	mutEvent.SetEpoch(env.store.GetEpoch())
	mutEvent.SetCreationTime(inter.Timestamp(env.t.UnixNano()))
	mutEvent.SetTxs(types.Transactions(txs))
	event := mutEvent.Build()
	env.store.SetEvent(event)

	consensus := env.Service.GetConsensusCallbacks()
	blockCallback := consensus.BeginBlock(&lachesis.Block{
		Atropos:  event.ID(),
		Cheaters: make([]idx.ValidatorID, 0),
	})
	blockCallback.ApplyEvent(event)
	blockCallback.EndBlock()
	env.blockProcWg.Wait()

	number := env.store.GetBlockIndex(event.ID())
	block := env.GetEvmStateReader().GetBlock(common.Hash{}, uint64(*number))
	receipts := env.store.evm.GetReceipts(*number, env.EthAPI.signer, block.Hash, block.Transactions)
	return receipts, nil
}

func (env *testEnv) ApplyTxs(spent time.Duration, txs ...*types.Transaction) (types.Receipts, error) {
	return env.applyTxs(spent, txs...)
}

func (env *testEnv) applyTxs(spent time.Duration, txs ...*types.Transaction) (types.Receipts, error) {
	env.t = env.t.Add(spent)

	waitForCount := int64(len(txs))
	waitForTxs := make(map[common.Hash]struct{}, len(txs))
	for _, tx := range txs {
		waitForTxs[tx.Hash()] = struct{}{}
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	defer wg.Wait()

	receipts := make(types.Receipts, 0, len(txs))

	newBlocks := make(chan evmcore.ChainHeadNotify)
	defer close(newBlocks)
	blocksSub := env.feed.SubscribeNewBlock(newBlocks)
	defer blocksSub.Unsubscribe()
	go func() {
		defer wg.Done()
		for b := range newBlocks {

			n := idx.Block(b.Block.Number.Uint64())
			// valid txs
			if len(b.Block.Transactions) > 0 {
				rr := env.store.evm.GetReceipts(n, env.EthAPI.signer, b.Block.Hash, b.Block.Transactions)
				for _, r := range rr {
					env.txpool.(*dummyTxPool).Delete(r.TxHash)
					if _, ok := waitForTxs[r.TxHash]; ok {
						receipts = append(receipts, r)
						delete(waitForTxs, r.TxHash)
						atomic.AddInt64(&waitForCount, -1)
					}
				}
			}
			// invalid txs
			block := env.store.GetBlock(n)
			if len(block.SkippedTxs) > 0 {
				var (
					internalsLen = uint32(len(block.InternalTxs))
					externalsLen = uint32(len(block.InternalTxs) + len(block.Txs))
					eventstxsLen = uint32(len(block.InternalTxs) + len(block.Txs) + 0)
					e            = 0
				)
				for _, txi := range block.SkippedTxs {
					var tx common.Hash

					switch {
					case txi < internalsLen:
						tx = block.InternalTxs[txi+0]
					case txi < externalsLen:
						tx = block.Txs[txi-internalsLen]
					default:
						for {
							etxs := env.store.GetEventPayload(block.Events[e]).Txs()
							if txi < (eventstxsLen + uint32(len(etxs))) {
								tx = etxs[txi-eventstxsLen].Hash()
								break
							} else {
								e += 1 // next event
								eventstxsLen += uint32(len(etxs))
							}
						}
					}
					env.txpool.(*dummyTxPool).Delete(tx)
					if _, ok := waitForTxs[tx]; ok {
						delete(waitForTxs, tx)
						atomic.AddInt64(&waitForCount, -1)
					}
				}
			}
		}
	}()

	byEmitters := func() []*emitter.Emitter {
		if atomic.LoadInt64(&waitForCount) > 0 {
			// allow block creation
			return env.emitters
		}
		// ready to stop
		return nil
	}

	env.txpool.AddRemotes(txs)
	defer env.txpool.(*dummyTxPool).Clear()
	err := env.EmitUntil(byEmitters)

	return receipts, err
}

func (env *testEnv) ApplyMPs(spent time.Duration, mps ...inter.MisbehaviourProof) error {
	env.t = env.t.Add(spent)

	// all callbacks are non-async
	lastEpoch := idx.Epoch(0)
	env.callback.buildEvent = func(e *inter.MutableEventPayload) {
		if e.Epoch() > lastEpoch {
			e.SetMisbehaviourProofs(mps)
			lastEpoch = e.Epoch()
		}
	}
	confirmed := false
	env.callback.onEventConfirmed = func(_e inter.EventI) {
		// ensure that not only MPs were confirmed, but also no new MPs will be confirmed in future
		if _e.AnyMisbehaviourProofs() && _e.Epoch() == lastEpoch {
			confirmed = true
			// sanity check for gas used
			e := env.store.GetEventPayload(_e.ID())
			rule := env.store.GetRules().Economy.Gas
			if e.GasPowerUsed() < rule.EventGas+uint64(len(e.MisbehaviourProofs()))*rule.MisbehaviourProofGas {
				panic("GasPowerUsed calculation doesn't include MisbehaviourProofGas")
			}
		}
	}
	defer func() {
		env.callback.buildEvent = nil
	}()

	byEmitters := func() []*emitter.Emitter {
		if !confirmed {
			return env.emitters
		}
		return nil
	}

	return env.EmitUntil(byEmitters)
}

func (env *testEnv) EmitUntil(by func() []*emitter.Emitter) error {
	start := time.Now()
	for {
		emitters := by()
		if len(emitters) < 1 {
			break
		}
		for _, em := range emitters {
			_, err := em.EmitEvent()
			if err != nil {
				return err
			}
		}
		env.WaitBlockEnd()
		env.t = env.t.Add(time.Second)
		if time.Since(start) > 30*time.Second {
			panic("block doesn't get processed")
		}
	}
	return nil
}

func (env *testEnv) Transfer(from, to idx.ValidatorID, amount *big.Int) *types.Transaction {
	sender := env.Address(from)
	nonce := env.nextNonce(sender)
	receiver := env.Address(to)
	gasLimit := env.GetEvmStateReader().MaxGasLimit()

	raw, err := env.transactor.FillTransaction(context.Background(), ethapi.TransactionArgs{
		From:     &sender,
		To:       &receiver,
		Value:    (*hexutil.Big)(amount),
		Nonce:    (*hexutil.Uint64)(&nonce),
		Gas:      (*hexutil.Uint64)(&gasLimit),
		GasPrice: (*hexutil.Big)(new(big.Int).SetUint64(gasPrice)),
	})
	if err != nil {
		panic(err)
	}

	key := env.privateKey(from)
	tx, err := types.SignTx(raw.Tx, env.EthAPI.signer, key)
	if err != nil {
		panic(err)
	}

	return tx
}

func (env *testEnv) Contract(from idx.ValidatorID, amount *big.Int, hex string) *types.Transaction {
	sender := env.Address(from)
	nonce := env.nextNonce(sender)
	gasLimit := env.GetEvmStateReader().MaxGasLimit()
	data := hexutil.MustDecode(hex)

	raw, err := env.transactor.FillTransaction(context.Background(), ethapi.TransactionArgs{
		From:     &sender,
		Value:    (*hexutil.Big)(amount),
		Nonce:    (*hexutil.Uint64)(&nonce),
		Gas:      (*hexutil.Uint64)(&gasLimit),
		GasPrice: (*hexutil.Big)(new(big.Int).SetUint64(gasPrice)),
		Data:     (*hexutil.Bytes)(&data),
	})
	if err != nil {
		panic(err)
	}

	key := env.privateKey(from)
	tx, err := types.SignTx(raw.Tx, env.EthAPI.signer, key)
	if err != nil {
		panic(err)
	}

	return tx
}

func (env *testEnv) privateKey(n idx.ValidatorID) *ecdsa.PrivateKey {
	key := makefakegenesis.FakeKey(n)
	return key
}

func (env *testEnv) Address(n idx.ValidatorID) common.Address {
	key := makefakegenesis.FakeKey(n)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return addr
}

func (env *testEnv) Pay(from idx.ValidatorID, amounts ...*big.Int) *bind.TransactOpts {
	sender := env.Address(from)
	nonce := env.nextNonce(sender)
	amount := big.NewInt(0)
	for _, a := range amounts {
		amount.Add(amount, a)
	}
	gasLimit := env.GetEvmStateReader().MaxGasLimit()
	data := []byte{0} // fake

	raw, err := env.transactor.FillTransaction(context.Background(), ethapi.TransactionArgs{
		From:     &sender,
		Value:    (*hexutil.Big)(amount),
		Nonce:    (*hexutil.Uint64)(&nonce),
		Gas:      (*hexutil.Uint64)(&gasLimit),
		GasPrice: (*hexutil.Big)(new(big.Int).SetUint64(gasPrice)),
		Data:     (*hexutil.Bytes)(&data),
	})
	if err != nil {
		panic(err)
	}

	key := env.privateKey(from)
	t, _ := bind.NewKeyedTransactorWithChainID(key, new(big.Int).SetUint64(env.store.GetRules().NetworkID))
	{
		t.Nonce = big.NewInt(int64(nonce))
		t.Value = amount
		t.NoSend = true
		t.GasLimit = raw.Tx.Gas()
		t.GasPrice = raw.Tx.GasPrice()
		//t.GasFeeCap = raw.Tx.GasFeeCap()
		//t.GasTipCap = raw.Tx.GasTipCap()
	}
	return t
}

func withLowGas(opts *bind.TransactOpts) *bind.TransactOpts {
	originSigner := opts.Signer

	opts.Signer = func(from common.Address, tx *types.Transaction) (*types.Transaction, error) {
		gas, err := evmcore.IntrinsicGas(tx.Data(), tx.AccessList(), tx.To() == nil)
		if err != nil {
			return nil, err
		}
		if tx.Gas() >= gas {
			repack := &types.LegacyTx{
				To:       tx.To(),
				Nonce:    tx.Nonce(),
				GasPrice: tx.GasPrice(),
				Gas:      gas + 1,
				Value:    tx.Value(),
				Data:     tx.Data(),
			}
			tx = types.NewTx(repack)
		}
		return originSigner(from, tx)
	}

	return opts

}

func (env *testEnv) ReadOnly() *bind.CallOpts {
	return &bind.CallOpts{}
}

func (env *testEnv) State() *state.StateDB {
	statedb, _ := env.store.evm.StateDB(env.store.GetBlockState().FinalizedStateRoot)
	return statedb
}

func (env *testEnv) nextNonce(account common.Address) uint64 {
	nonce := env.nonces[account]
	env.nonces[account] = nonce + 1
	return nonce
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
	if blockNumber != nil && idx.Block(blockNumber.Uint64()) != env.store.GetLatestBlockIndex() {
		return nil, errBlockNumberUnsupported
	}

	code := env.State().GetCode(contract)
	return code, nil
}

// ContractCall executes an Ethereum contract call with the specified data as the
// input.
func (env *testEnv) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if blockNumber != nil && idx.Block(blockNumber.Uint64()) != env.store.GetLatestBlockIndex() {
		return nil, errBlockNumberUnsupported
	}

	h := env.GetEvmStateReader().GetHeader(common.Hash{}, uint64(env.store.GetLatestBlockIndex()))
	block := &evmcore.EvmBlock{
		EvmHeader: *h,
	}

	rval, _, _, err := env.callContract(ctx, call, block, env.State())
	return rval, err
}

func (env *testEnv) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	var num64 uint64
	if number == nil {
		num64 = uint64(env.store.GetLatestBlockIndex())
	} else {
		num64 = number.Uint64()
	}
	return env.GetEvmStateReader().GetHeader(common.Hash{}, num64).EthHeader(), nil
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
	vmenv := vm.NewEVM(context, txContext, state, env.store.GetEvmChainConfig(), opera.DefaultVMConfig)
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

// SuggestGasTipCap retrieves the currently suggested gas price tip to allow a timely
// execution of a transaction.
func (env *testEnv) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return new(big.Int), nil
}

// SuggestGasTipCap retrieves the currently suggested gas price to allow a timely
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
	panic("is not implemented")
	return 0, nil
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
func (m callmsg) GasTipCap() *big.Int          { return m.CallMsg.GasTipCap }
func (m callmsg) GasFeeCap() *big.Int          { return m.CallMsg.GasFeeCap }
func (m callmsg) Gas() uint64                  { return m.CallMsg.Gas }
func (m callmsg) Value() *big.Int              { return m.CallMsg.Value }
func (m callmsg) Nonce() uint64                { return 0 }
func (m callmsg) IsFake() bool                 { return true }
func (m callmsg) Data() []byte                 { return m.CallMsg.Data }
func (m callmsg) AccessList() types.AccessList { return nil }
