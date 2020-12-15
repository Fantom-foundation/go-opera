package gossip

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
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
	"github.com/Fantom-foundation/go-opera/gossip/gasprice"
	"github.com/Fantom-foundation/go-opera/inter"
	operaparams "github.com/Fantom-foundation/go-opera/opera/params"
	"github.com/Fantom-foundation/go-opera/topicsdb"
	"github.com/Fantom-foundation/go-opera/tracing"
)

// EthAPIBackend implements ethapi.Backend.
type EthAPIBackend struct {
	extRPCEnabled bool
	svc           *Service
	store         *Store
	state         *EvmStateReader
	gpo           *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *EthAPIBackend) ChainConfig() *params.ChainConfig {
	return b.svc.net.EvmChainConfig()
}

func (b *EthAPIBackend) CurrentBlock() *evmcore.EvmBlock {
	return b.state.CurrentBlock()
}

func (b *EthAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmHeader, error) {
	blk, err := b.BlockByNumber(ctx, number)
	return blk.Header(), err
}

// HeaderByHash returns evm header by its (atropos) hash.
func (b *EthAPIBackend) HeaderByHash(ctx context.Context, h common.Hash) (*evmcore.EvmHeader, error) {
	index := b.store.GetBlockIndex(hash.Event(h))
	if index == nil {
		return nil, nil
	}
	return b.HeaderByNumber(ctx, rpc.BlockNumber(*index))
}

// BlockByNumber returns block by its number.
func (b *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error) {
	if number == rpc.PendingBlockNumber {
		return nil, errors.New("pending block request isn't allowed")
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

func (b *EthAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *evmcore.EvmHeader, error) {
	if number == rpc.PendingBlockNumber {
		return nil, nil, errors.New("pending block request isn't allowed")
	}
	var header *evmcore.EvmHeader
	if number == rpc.LatestBlockNumber {
		header = &b.state.CurrentBlock().EvmHeader
	} else {
		header = b.state.GetHeader(common.Hash{}, uint64(number))
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb := b.store.evm.StateDB(hash.Hash(header.Root))
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

	options := b.store.FindEventHashes(epoch, lamport, prefix)
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
	return b.store.GetEventPayload(id), nil
}

// GetEvent returns the Lachesis event header by hash or short ID.
func (b *EthAPIBackend) GetEvent(ctx context.Context, shortEventID string) (*inter.Event, error) {
	id, err := b.GetFullEventID(shortEventID)
	if err != nil {
		return nil, err
	}
	return b.store.GetEvent(id), nil
}

// GetHeads returns IDs of all the epoch events with no descendants.
// * When epoch is -2 the heads for latest epoch are returned.
// * When epoch is -1 the heads for latest sealed epoch are returned.
func (b *EthAPIBackend) GetHeads(ctx context.Context, epoch rpc.BlockNumber) (heads hash.Events, err error) {
	current := b.store.GetEpoch()

	requested, err := b.epochWithDefault(ctx, epoch)
	if err != nil {
		return nil, err
	}

	if requested == current {
		heads = b.store.GetHeads(requested)
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
	current := b.store.GetEpoch()

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

	b.store.ForEachEpochEvent(requested, onEvent)
	return nil
}

func (b *EthAPIBackend) GetValidators(ctx context.Context) *pos.Validators {
	return b.store.GetValidators()
}

func (b *EthAPIBackend) GetBlock(ctx context.Context, h common.Hash) (*evmcore.EvmBlock, error) {
	index := b.store.GetBlockIndex(hash.Event(h))
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
		return nil, errors.New("pending block request isn't allowed")
	}
	if number == rpc.LatestBlockNumber {
		header := b.state.CurrentHeader()
		number = rpc.BlockNumber(header.Number.Uint64())
	}

	receipts := b.store.evm.GetReceipts(idx.Block(number))
	return receipts, nil
}

// GetReceipts retrieves the receipts for all transactions in a given block.
func (b *EthAPIBackend) GetReceipts(ctx context.Context, block common.Hash) (types.Receipts, error) {
	number := b.store.GetBlockIndex(hash.Event(block))
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
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *EthAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return big.NewInt(0)
}

func (b *EthAPIBackend) GetEVM(ctx context.Context, msg evmcore.Message, state *state.StateDB, header *evmcore.EvmHeader) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := evmcore.NewEVMContext(msg, header, b.state, nil)
	config := b.ChainConfig()
	return vm.NewEVM(context, state, config, vm.Config{}), vmError, nil
}

func (b *EthAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	err := b.svc.txpool.AddLocal(signedTx)
	if err == nil {
		// NOTE: only sent txs tracing, see TxPool.addTxs() for all
		tracing.StartTx(signedTx.Hash(), "EthAPIBackend.SendTx()")
	}
	return err
}

func (b *EthAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) notify.Subscription {
	return b.svc.feed.SubscribeFinalLogs(ch)
}

func (b *EthAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) notify.Subscription {
	return b.svc.feed.SubscribeFinalTxs(ch)
}

func (b *EthAPIBackend) SubscribeNewBlockEvent(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return b.svc.feed.SubscribeFinalBlock(ch)
}

func (b *EthAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.svc.txpool.Pending()
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

func (b *EthAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, uint64, uint64, error) {
	if !b.svc.config.TxIndex {
		return nil, 0, 0, errors.New("transactions index is disabled (enable TxIndex and re-process the DAG)")
	}

	position := b.store.evm.GetTxPosition(txHash)
	if position == nil {
		return nil, 0, 0, nil
	}

	var tx *types.Transaction
	if position.Event.IsZero() {
		tx = b.store.evm.GetTx(txHash)
	} else {
		event := b.store.GetEventPayload(position.Event)
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

func (b *EthAPIBackend) SubscribeNewTxsNotify(ch chan<- evmcore.NewTxsNotify) notify.Subscription {
	return b.svc.txpool.SubscribeNewTxsNotify(ch)
}

// Progress returns current synchronization status of this node
func (b *EthAPIBackend) Progress() ethapi.PeerProgress {
	p2pProgress := b.svc.pm.myProgress()
	highestP2pProgress := b.svc.pm.highestPeerProgress()
	b.svc.engineMu.RLock()
	lastBlock := b.svc.store.GetBlock(p2pProgress.NumOfBlocks)
	b.svc.engineMu.RUnlock()

	return ethapi.PeerProgress{
		CurrentEpoch:     p2pProgress.Epoch,
		CurrentBlock:     p2pProgress.NumOfBlocks,
		CurrentBlockHash: p2pProgress.LastBlock,
		CurrentBlockTime: lastBlock.Time,
		HighestBlock:     highestP2pProgress.NumOfBlocks,
		HighestEpoch:     highestP2pProgress.Epoch,
	}
}

func (b *EthAPIBackend) ProtocolVersion() int {
	return int(ProtocolVersions[len(ProtocolVersions)-1])
}

func (b *EthAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *EthAPIBackend) MinGasPrice() *big.Int {
	return operaparams.MinGasPrice
}

func (b *EthAPIBackend) ChainDb() ethdb.Database {
	return b.store.evm.EvmTable()
}

func (b *EthAPIBackend) AccountManager() *accounts.Manager {
	return b.svc.AccountManager()
}

func (b *EthAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *EthAPIBackend) RPCGasCap() uint64 {
	return b.svc.config.RPCGasCap
}

func (b *EthAPIBackend) RPCTxFeeCap() float64 {
	return b.svc.config.RPCTxFeeCap
}

func (b *EthAPIBackend) EvmLogIndex() *topicsdb.Index {
	return b.store.evm.EvmLogs()
}

// CurrentEpoch returns current epoch number.
func (b *EthAPIBackend) CurrentEpoch(ctx context.Context) idx.Epoch {
	return b.store.GetEpoch()
}
