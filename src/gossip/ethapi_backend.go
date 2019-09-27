package gossip

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	notify "github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	errors2 "github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
	"github.com/Fantom-foundation/go-lachesis/src/gossip/gasprice"
)

var ErrNotImplemented = func(name string) error { return errors.New(name + " method is not implemented yet") }

// EthAPIBackend implements ethapi.Backend.
type EthAPIBackend struct {
	extRPCEnabled bool
	svc           *Service
	state         *EvmStateReader
	gpo           *gasprice.Oracle
	mux           *notify.TypeMux
}

// ChainConfig returns the active chain configuration.
func (b *EthAPIBackend) ChainConfig() *params.ChainConfig {
	return params.AllEthashProtocolChanges
}

func (b *EthAPIBackend) CurrentBlock() *evm_core.EvmBlock {
	return b.state.CurrentBlock()
}

func (b *EthAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*evm_core.EvmHeader, error) {
	blk, err := b.BlockByNumber(ctx, number)

	return blk.Header(), err
}

func (b *EthAPIBackend) HeaderByHash(ctx context.Context, h common.Hash) (*evm_core.EvmHeader, error) {
	index := b.svc.store.GetBlockIndex(hash.Event(h))
	if index == nil {
		return nil, errors.New("header wasn't found")
	}
	return b.HeaderByNumber(ctx, rpc.BlockNumber(*index))
}

func (b *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evm_core.EvmBlock, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		return nil, errors.New("pending block request isn't allowed")
	}
	// Otherwise resolve and return the block
	var blk *evm_core.EvmBlock
	if number == rpc.LatestBlockNumber {
		blk = b.state.CurrentBlock()
	} else {
		n := uint64(number.Int64())
		blk = b.state.GetBlock(common.Hash{}, n)
	}
	if blk == nil {
		return nil, errors.New("block wasn't found")
	}

	return blk, nil
}

func (b *EthAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *evm_core.EvmHeader, error) {
	if number == rpc.PendingBlockNumber {
		return nil, nil, errors.New("pending block request isn't allowed")
	}
	var header *evm_core.EvmHeader
	if number == rpc.LatestBlockNumber {
		header = &b.state.CurrentBlock().EvmHeader
	} else {
		header = b.svc.GetEvmStateReader().GetHeader(common.Hash{}, uint64(number))
	}
	if header == nil {
		return nil, nil, errors.New("header not found")
	}
	stateDb := b.svc.store.StateDB(header.Root)
	return stateDb, header, nil
}

// s is a string splitted by ":" separator
// example of a short ID: "5:26:a2395846", where 5 is epoch, 26 is lamport, a2395846 are first bytes of the hash
func decodeShortEventId(s []string) (idx.Epoch, idx.Lamport, []byte, error) {
	if len(s) != 3 {
		return 0, 0, nil, errors.New("incorrect format of short event ID (need Epoch:Lamport:Hash")
	}
	epoch, err := strconv.ParseUint(s[0], 10, 32)
	if err != nil {
		return 0, 0, nil, errors2.Wrap(err, "short hash parsing error (lamport)")
	}
	lamport, err := strconv.ParseUint(s[1], 10, 32)
	if err != nil {
		return 0, 0, nil, errors2.Wrap(err, "short hash parsing error (lamport)")
	}
	return idx.Epoch(epoch), idx.Lamport(lamport), common.FromHex(s[2]), nil
}

func (b *EthAPIBackend) GetFullEventId(shortEventId string) (hash.Event, error) {
	s := strings.Split(shortEventId, ":")
	if len(s) == 1 {
		// it's a full hash
		return hash.HexToEventHash(shortEventId), nil
	}
	// short hash
	epoch, lamport, prefix, err := decodeShortEventId(s)
	if err != nil {
		return hash.Event{}, err
	}

	options := b.svc.store.FindEventHashes(epoch, lamport, prefix)
	if len(options) == 0 {
		return hash.Event{}, errors.New("event wasn't found by short ID")
	}
	if len(options) > 1 {
		return hash.Event{}, errors.New("there're multiple events with the same short ID, please use full ID")
	}
	return options[0], nil
}

func (b *EthAPIBackend) GetEvent(ctx context.Context, shortEventId string) (*inter.Event, error) {
	id, err := b.GetFullEventId(shortEventId)
	if err != nil {
		return nil, err
	}
	e := b.svc.store.GetEvent(id)
	if e == nil {
		return nil, errors.New("event wasn't found")
	}
	return e, nil
}

func (b *EthAPIBackend) GetEventHeader(ctx context.Context, shortEventId string) (*inter.EventHeaderData, error) {
	id, err := b.GetFullEventId(shortEventId)
	if err != nil {
		return nil, err
	}
	epoch := id.Epoch()
	if epoch != b.svc.engine.GetEpoch() {
		return nil, errors.New("event headers are stored only for current epoch")
	}
	e := b.svc.store.GetEventHeader(epoch, id)
	if e == nil {
		return nil, errors.New("event header wasn't found")
	}
	return e, nil
}

func (b *EthAPIBackend) GetConsensusTime(ctx context.Context, shortEventId string) (inter.Timestamp, error) {
	id, err := b.GetFullEventId(shortEventId)
	if err != nil {
		return 0, err
	}
	return b.svc.engine.GetConsensusTime(id)
}

func (b *EthAPIBackend) GetHeads(ctx context.Context) hash.Events {
	heads := b.svc.store.GetHeads(b.svc.engine.GetEpoch())
	if heads == nil {
		heads = hash.Events{}
	}
	return heads
}

func (b *EthAPIBackend) GetHeader(ctx context.Context, h common.Hash) *evm_core.EvmHeader {
	header, err := b.HeaderByHash(ctx, h)
	if err != nil {
		return nil
	}
	return header
}

func (b *EthAPIBackend) GetBlock(ctx context.Context, h common.Hash) (*evm_core.EvmBlock, error) {
	index := b.svc.store.GetBlockIndex(hash.Event(h))
	if index == nil {
		return nil, errors.New("block wasn't found")
	}
	return b.BlockByNumber(ctx, rpc.BlockNumber(*index))
}

func (b *EthAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.blockchain.GetReceiptsByHash(hash), nil
	*/
	return nil, ErrNotImplemented("GetReceipts")
}

func (b *EthAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	// TODO: implement or disable it. Origin:
	/*
		receipts := b.svc.blockchain.GetReceiptsByHash(hash)
		if receipts == nil {
			return nil, nil
		}
		logs := make([][]*types.Log, len(receipts))
		for i, receipt := range receipts {
			logs[i] = receipt.Logs
		}
		return logs, nil
	*/
	return nil, ErrNotImplemented("GetLogs")
}

func (b *EthAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return big.NewInt(0)
}

func (b *EthAPIBackend) GetEVM(ctx context.Context, msg evm_core.Message, state *state.StateDB, header *evm_core.EvmHeader) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := evm_core.NewEVMContext(msg, header, b.svc.GetEvmStateReader(), nil)
	config := params.AllEthashProtocolChanges
	return vm.NewEVM(context, state, config, vm.Config{}), vmError, nil
}

func (b *EthAPIBackend) SubscribeRemovedLogsNotify(ch chan<- evm_core.RemovedLogsNotify) notify.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeRemovedLogsEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeChainNotify(ch chan<- evm_core.ChainNotify) notify.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeChainEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeChainHeadNotify(ch chan<- evm_core.ChainHeadNotify) notify.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeChainHeadEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeChainSideNotify(ch chan<- evm_core.ChainSideNotify) notify.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeChainSideEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeLogsNotify(ch chan<- []*types.Log) notify.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeLogsEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.svc.txpool.AddLocal(signedTx)
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

func (b *EthAPIBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	// TODO: implement or disable it. Origin:
	/*
		tx, blockHash, blockNumber, index := rawdb.ReadTransaction(b.svc.ChainDb(), txHash)
		return tx, blockHash, blockNumber, index, nil
	*/
	return nil, common.Hash{}, 0, 0, ErrNotImplemented("GetTransaction")
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

func (b *EthAPIBackend) SubscribeNewTxsNotify(ch chan<- evm_core.NewTxsNotify) notify.Subscription {
	return b.svc.txpool.SubscribeNewTxsNotify(ch)
}

func (b *EthAPIBackend) Progress() PeerProgress {
	return b.svc.pm.myProgress()
}

func (b *EthAPIBackend) ProtocolVersion() int {
	return int(ProtocolVersions[len(ProtocolVersions)-1])
}

func (b *EthAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *EthAPIBackend) ChainDb() ethdb.Database {
	return b.svc.store.table.Evm
}

func (b *EthAPIBackend) NotifyMux() *notify.TypeMux {
	return b.mux
}

func (b *EthAPIBackend) AccountManager() *accounts.Manager {
	return b.svc.AccountManager()
}

func (b *EthAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *EthAPIBackend) RPCGasCap() *big.Int {
	return b.svc.config.RPCGasCap
}

func (b *EthAPIBackend) BloomStatus() (uint64, uint64) {
	// TODO: implement or disable it. Origin:
	/*
		sections, _, _ := b.svc.bloomIndexer.Sections()
		return params.BloomBitsBlocks, sections
	*/
	return 0, 0
}

func (b *EthAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	// TODO: implement or disable it. Origin:
	/*
		for i := 0; i < bloomFilterThreads; i++ {
			go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.svc.bloomRequests)
		}
	*/
}
