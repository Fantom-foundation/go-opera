package gossip

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-opera/ethapi"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/topicsdb"
	"github.com/Fantom-foundation/go-opera/tracing"
)

// EthAPIBackend implements ethapi.Backend.
type EthAPIBackend struct {
	extRPCEnabled       bool
	svc                 *Service
	state               *EvmStateReader
	signer              types.Signer
	allowUnprotectedTxs bool
}

// ChainConfig returns the active chain configuration.
func (b *EthAPIBackend) ChainConfig() *params.ChainConfig {
	return b.svc.store.GetRules().EvmChainConfig()
}

func (b *EthAPIBackend) CurrentBlock() *evmcore.EvmBlock {
	return b.state.CurrentBlock()
}

// HeaderByNumber returns evm block header by its number, or nil if not exists.
func (b *EthAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmHeader, error) {
	blk, err := b.BlockByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	if blk == nil {
		return nil, nil
	}
	return blk.Header(), err
}

// HeaderByHash returns evm block header by its (atropos) hash, or nil if not exists.
func (b *EthAPIBackend) HeaderByHash(ctx context.Context, h common.Hash) (*evmcore.EvmHeader, error) {
	index := b.svc.store.GetBlockIndex(hash.Event(h))
	if index == nil {
		return nil, nil
	}
	return b.HeaderByNumber(ctx, rpc.BlockNumber(*index))
}

// BlockByNumber returns evm block by its number, or nil if not exists.
func (b *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error) {
	if number == rpc.PendingBlockNumber {
		number = rpc.LatestBlockNumber
	}
	// Otherwise resolve and return the block
	var blk *evmcore.EvmBlock
	if number == rpc.LatestBlockNumber {
		blk = b.state.CurrentBlock()
	} else {
		n := uint64(number.Int64())
		blk = b.state.GetBlock(common.Hash{}, n)
	}

	return blk, nil
}

// StateAndHeaderByNumberOrHash returns evm state and block header by block number or block hash, err if not exists.
func (b *EthAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *evmcore.EvmHeader, error) {
	var header *evmcore.EvmHeader
	if number, ok := blockNrOrHash.Number(); ok && (number == rpc.LatestBlockNumber || number == rpc.PendingBlockNumber) {
		header = &b.state.CurrentBlock().EvmHeader
	} else if number, ok := blockNrOrHash.Number(); ok {
		header = b.state.GetHeader(common.Hash{}, uint64(number))
	} else if h, ok := blockNrOrHash.Hash(); ok {
		index := b.svc.store.GetBlockIndex(hash.Event(h))
		if index == nil {
			return nil, nil, errors.New("header not found")
		}
		header = b.state.GetHeader(common.Hash{}, uint64(*index))
	} else {
		return nil, nil, errors.New("unknown header selector")
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb, err := b.svc.store.evm.StateDB(hash.Hash(header.Root))
	if err != nil {
		return nil, nil, err
	}
	return stateDb, header, nil
}

// decodeShortEventID decodes ShortID
// example of a ShortID: "5:26:a2395846", where 5 is epoch, 26 is lamport, a2395846 are first bytes of the hash
// s is a string splitted by ":" separator
func decodeShortEventID(s []string) (idx.Epoch, idx.Lamport, []byte, error) {
	if len(s) != 3 {
		return 0, 0, nil, errors.New("incorrect format of short event ID (need Epoch:Lamport:Hash")
	}
	epoch, err := strconv.ParseUint(s[0], 10, 32)
	if err != nil {
		return 0, 0, nil, errors.Wrap(err, "short hash parsing error (lamport)")
	}
	lamport, err := strconv.ParseUint(s[1], 10, 32)
	if err != nil {
		return 0, 0, nil, errors.Wrap(err, "short hash parsing error (lamport)")
	}
	return idx.Epoch(epoch), idx.Lamport(lamport), common.FromHex(s[2]), nil
}

// GetFullEventID "converts" ShortID to full event's hash, by searching in events DB.
func (b *EthAPIBackend) GetFullEventID(shortEventID string) (hash.Event, error) {
	s := strings.Split(shortEventID, ":")
	if len(s) == 1 {
		// it's a full hash
		return hash.HexToEventHash(shortEventID), nil
	}
	// short hash
	epoch, lamport, prefix, err := decodeShortEventID(s)
	if err != nil {
		return hash.Event{}, err
	}

	options := b.svc.store.FindEventHashes(epoch, lamport, prefix)
	if len(options) == 0 {
		return hash.Event{}, errors.New("event not found by short ID")
	}
	if len(options) > 1 {
		return hash.Event{}, errors.New("there're multiple events with the same short ID, please use full ID")
	}
	return options[0], nil
}

// GetEventPayload returns Lachesis event by hash or short ID.
func (b *EthAPIBackend) GetEventPayload(ctx context.Context, shortEventID string) (*inter.EventPayload, error) {
	id, err := b.GetFullEventID(shortEventID)
	if err != nil {
		return nil, err
	}
	return b.svc.store.GetEventPayload(id), nil
}

// GetEvent returns the Lachesis event header by hash or short ID.
func (b *EthAPIBackend) GetEvent(ctx context.Context, shortEventID string) (*inter.Event, error) {
	id, err := b.GetFullEventID(shortEventID)
	if err != nil {
		return nil, err
	}
	return b.svc.store.GetEvent(id), nil
}

// GetHeads returns IDs of all the epoch events with no descendants.
// * When epoch is -2 the heads for latest epoch are returned.
// * When epoch is -1 the heads for latest sealed epoch are returned.
func (b *EthAPIBackend) GetHeads(ctx context.Context, epoch rpc.BlockNumber) (heads hash.Events, err error) {
	current := b.svc.store.GetEpoch()

	requested, err := b.epochWithDefault(ctx, epoch)
	if err != nil {
		return nil, err
	}

	if requested == current {
		heads = b.svc.store.GetHeadsSlice(requested)
	} else {
		err = errors.New("heads for previous epochs are not available")
		return
	}

	if heads == nil {
		heads = hash.Events{}
	}

	return
}

func (b *EthAPIBackend) epochWithDefault(ctx context.Context, epoch rpc.BlockNumber) (requested idx.Epoch, err error) {
	current := b.svc.store.GetEpoch()

	switch {
	case epoch == rpc.PendingBlockNumber:
		requested = current
	case epoch == rpc.LatestBlockNumber:
		requested = current - 1
	case epoch >= 0 && idx.Epoch(epoch) <= current:
		requested = idx.Epoch(epoch)
	default:
		err = errors.New("epoch is not in range")
		return
	}
	return requested, nil
}

// ForEachEpochEvent iterates all the events which are observed by head, and accepted by a filter.
// filter CANNOT called twice for the same event.
func (b *EthAPIBackend) ForEachEpochEvent(ctx context.Context, epoch rpc.BlockNumber, onEvent func(event *inter.EventPayload) bool) error {
	requested, err := b.epochWithDefault(ctx, epoch)
	if err != nil {
		return err
	}

	b.svc.store.ForEachEpochEvent(requested, onEvent)
	return nil
}

func (b *EthAPIBackend) BlockByHash(ctx context.Context, h common.Hash) (*evmcore.EvmBlock, error) {
	index := b.svc.store.GetBlockIndex(hash.Event(h))
	if index == nil {
		return nil, nil
	}

	if rpc.BlockNumber(*index) == rpc.PendingBlockNumber {
		return nil, errors.New("pending block request isn't allowed")
	}
	// Otherwise resolve and return the block
	var blk *evmcore.EvmBlock
	if rpc.BlockNumber(*index) == rpc.LatestBlockNumber {
		blk = b.state.CurrentBlock()
	} else {
		n := uint64(*index)
		blk = b.state.GetBlock(common.Hash{}, n)
	}

	return blk, nil
}

// GetReceiptsByNumber returns receipts by block number.
func (b *EthAPIBackend) GetReceiptsByNumber(ctx context.Context, number rpc.BlockNumber) (types.Receipts, error) {
	if !b.svc.config.TxIndex {
		return nil, errors.New("transactions index is disabled (enable TxIndex and re-process the DAGs)")
	}

	if number == rpc.PendingBlockNumber {
		number = rpc.LatestBlockNumber
	}
	if number == rpc.LatestBlockNumber {
		header := b.state.CurrentHeader()
		number = rpc.BlockNumber(header.Number.Uint64())
	}

	block := b.state.GetBlock(common.Hash{}, uint64(number))
	receipts := b.svc.store.evm.GetReceipts(idx.Block(number), b.signer, block.Hash, block.Transactions)
	return receipts, nil
}

// GetReceipts retrieves the receipts for all transactions in a given block.
func (b *EthAPIBackend) GetReceipts(ctx context.Context, block common.Hash) (types.Receipts, error) {
	number := b.svc.store.GetBlockIndex(hash.Event(block))
	if number == nil {
		return nil, nil
	}

	return b.GetReceiptsByNumber(ctx, rpc.BlockNumber(*number))
}

func (b *EthAPIBackend) GetLogs(ctx context.Context, block common.Hash) ([][]*types.Log, error) {
	receipts, err := b.GetReceipts(ctx, block)
	if receipts == nil || err != nil {
		return nil, err
	}
	logs := make([][]*types.Log, receipts.Len())
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *EthAPIBackend) GetTd(_ common.Hash) *big.Int {
	return big.NewInt(0)
}

func (b *EthAPIBackend) GetEVM(ctx context.Context, msg evmcore.Message, state *state.StateDB, header *evmcore.EvmHeader, vmConfig *vm.Config) (*vm.EVM, func() error, error) {
	vmError := func() error { return nil }

	if vmConfig == nil {
		vmConfig = &opera.DefaultVMConfig
	}
	txContext := evmcore.NewEVMTxContext(msg)
	context := evmcore.NewEVMBlockContext(header, b.state, nil)
	config := b.ChainConfig()
	return vm.NewEVM(context, txContext, state, config, *vmConfig), vmError, nil
}

func (b *EthAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	err := b.svc.txpool.AddLocal(signedTx)
	if err == nil {
		// NOTE: only sent txs tracing, see TxPool.addTxs() for all
		tracing.StartTx(signedTx.Hash(), "EthAPIBackend.SendTx()")
	}
	return err
}

func (b *EthAPIBackend) SubscribeLogsNotify(ch chan<- []*types.Log) notify.Subscription {
	return b.svc.feed.SubscribeNewLogs(ch)
}

func (b *EthAPIBackend) SubscribeNewBlockNotify(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return b.svc.feed.SubscribeNewBlock(ch)
}

func (b *EthAPIBackend) SubscribeNewTxsNotify(ch chan<- evmcore.NewTxsNotify) notify.Subscription {
	return b.svc.txpool.SubscribeNewTxsNotify(ch)
}

func (b *EthAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.svc.txpool.Pending(false)
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *EthAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.svc.txpool.Get(hash)
}

func (b *EthAPIBackend) GetTxPosition(txHash common.Hash) *evmstore.TxPosition {
	return b.svc.store.evm.GetTxPosition(txHash)
}

func (b *EthAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, uint64, uint64, error) {
	if !b.svc.config.TxIndex {
		return nil, 0, 0, errors.New("transactions index is disabled (enable TxIndex and re-process the DAG)")
	}

	position := b.svc.store.evm.GetTxPosition(txHash)
	if position == nil {
		return nil, 0, 0, nil
	}

	var tx *types.Transaction
	if position.Event.IsZero() {
		tx = b.svc.store.evm.GetTx(txHash)
	} else {
		event := b.svc.store.GetEventPayload(position.Event)
		if position.EventOffset > uint32(event.Txs().Len()) {
			return nil, 0, 0, fmt.Errorf("transactions index is corrupted (offset is larger than number of txs in event), event=%s, txid=%s, block=%d, offset=%d, txs_num=%d",
				position.Event.String(),
				txHash.String(),
				position.Block,
				position.EventOffset,
				event.Txs().Len())
		}
		tx = event.Txs()[position.EventOffset]
	}

	return tx, uint64(position.Block), uint64(position.BlockOffset), nil
}

func (b *EthAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.svc.txpool.Nonce(addr), nil
}

func (b *EthAPIBackend) Stats() (pending int, queued int) {
	return b.svc.txpool.Stats()
}

func (b *EthAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.svc.txpool.Content()
}

// Progress returns current synchronization status of this node
func (b *EthAPIBackend) Progress() ethapi.PeerProgress {
	p2pProgress := b.svc.handler.myProgress()
	highestP2pProgress := b.svc.handler.highestPeerProgress()
	lastBlock := b.svc.store.GetBlock(p2pProgress.LastBlockIdx)

	return ethapi.PeerProgress{
		CurrentEpoch:     p2pProgress.Epoch,
		CurrentBlock:     p2pProgress.LastBlockIdx,
		CurrentBlockHash: p2pProgress.LastBlockAtropos,
		CurrentBlockTime: lastBlock.Time,
		HighestBlock:     highestP2pProgress.LastBlockIdx,
		HighestEpoch:     highestP2pProgress.Epoch,
	}
}

func (b *EthAPIBackend) TxPoolContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	return b.svc.txpool.ContentFrom(addr)
}

func (b *EthAPIBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return b.svc.gpo.SuggestTipCap(), nil
}

func (b *EthAPIBackend) ChainDb() ethdb.Database {
	return b.svc.store.evm.EvmDb
}

func (b *EthAPIBackend) AccountManager() *accounts.Manager {
	return b.svc.AccountManager()
}

func (b *EthAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *EthAPIBackend) UnprotectedAllowed() bool {
	return b.allowUnprotectedTxs
}

func (b *EthAPIBackend) RPCGasCap() uint64 {
	return b.svc.config.RPCGasCap
}

func (b *EthAPIBackend) RPCTxFeeCap() float64 {
	return b.svc.config.RPCTxFeeCap
}

func (b *EthAPIBackend) EvmLogIndex() *topicsdb.Index {
	return b.svc.store.evm.EvmLogs
}

// CurrentEpoch returns current epoch number.
func (b *EthAPIBackend) CurrentEpoch(ctx context.Context) idx.Epoch {
	return b.svc.store.GetEpoch()
}

func (b *EthAPIBackend) MinGasPrice() *big.Int {
	return b.state.MinGasPrice()
}
func (b *EthAPIBackend) MaxGasLimit() uint64 {
	return b.state.MaxGasLimit()
}

func (b *EthAPIBackend) GetUptime(ctx context.Context, vid idx.ValidatorID) (*big.Int, error) {
	// Note: loads bs and es atomically to avoid a race condition
	bs, es := b.svc.store.GetBlockEpochState()
	if !es.Validators.Exists(vid) {
		return nil, nil
	}
	return new(big.Int).SetUint64(uint64(bs.GetValidatorState(vid, es.Validators).Uptime)), nil
}

func (b *EthAPIBackend) GetOriginatedFee(ctx context.Context, vid idx.ValidatorID) (*big.Int, error) {
	// Note: loads bs and es atomically to avoid a race condition
	bs, es := b.svc.store.GetBlockEpochState()
	if !es.Validators.Exists(vid) {
		return nil, nil
	}
	return bs.GetValidatorState(vid, es.Validators).Originated, nil
}

func (b *EthAPIBackend) GetDowntime(ctx context.Context, vid idx.ValidatorID) (idx.Block, inter.Timestamp, error) {
	// Note: loads bs and es atomically to avoid a race condition
	bs, es := b.svc.store.GetBlockEpochState()
	if !es.Validators.Exists(vid) {
		return 0, 0, nil
	}
	vs := bs.GetValidatorState(vid, es.Validators)
	missedBlocks := idx.Block(0)
	if bs.LastBlock.Idx > vs.LastBlock {
		missedBlocks = bs.LastBlock.Idx - vs.LastBlock
	}
	missedTime := inter.Timestamp(0)
	if bs.LastBlock.Time > vs.LastOnlineTime {
		missedTime = bs.LastBlock.Time - vs.LastOnlineTime
	}
	if missedBlocks < es.Rules.Economy.BlockMissedSlack {
		return 0, 0, nil
	}
	return missedBlocks, missedTime, nil
}

func (b *EthAPIBackend) CalcBlockExtApi() bool {
	return b.svc.config.RPCBlockExt
}

func (b *EthAPIBackend) SealedEpochTiming(ctx context.Context) (start inter.Timestamp, end inter.Timestamp) {
	es := b.svc.store.GetEpochState()
	return es.PrevEpochStart, es.EpochStart
}
