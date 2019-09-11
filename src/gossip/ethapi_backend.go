package gossip

import (
	"context"
	"errors"
	"math/big"

	//	"github.com/ethereum/go-ethereum/core/rawdb"
	//	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/gasprice"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
)

var ErrNotImplemented = func(name string) error { return errors.New(name + " method is not implemented yet") }

// EthAPIBackend implements ethapi.Backend.
type EthAPIBackend struct {
	extRPCEnabled bool
	svc           *Service
	state         *EvmStateReader
	gpo           *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *EthAPIBackend) ChainConfig() *params.ChainConfig {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.blockchain.Config()
	*/
	return nil
}

func (b *EthAPIBackend) CurrentBlock() *types.Block {
	blk := b.state.CurrentBlock()
	return types.NewBlock(
		convertHeader(&blk.EvmHeader),
		blk.Transactions,
		nil,
		nil,
	)
}

func (b *EthAPIBackend) SetHead(number uint64) {
	// TODO: implement or disable it. Origin:
	/*
		b.svc.protocolManager.downloader.Cancel()
		b.svc.blockchain.SetHead(number)
	*/
}

func (b *EthAPIBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		return nil, errors.New("Error: request pending block")
	}

	// Otherwise resolve and return the block
	var blk *evm_core.EvmBlock
	if number == rpc.LatestBlockNumber {
		blk = b.state.CurrentBlock()
	} else {
		n := uint64(number.Int64())
		blk = b.state.GetBlock(common.Hash{}, n)
	}

	return convertHeader(&blk.EvmHeader), nil
}

func (b *EthAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.blockchain.GetHeaderByHash(hash), nil
	*/
	return nil, ErrNotImplemented("HeaderByHash")
}

func (b *EthAPIBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if number == rpc.PendingBlockNumber {
		return nil, errors.New("Error: request pending block")
	}
	// Otherwise resolve and return the block
	var blk *evm_core.EvmBlock
	if number == rpc.LatestBlockNumber {
		blk = b.state.CurrentBlock()
	} else {
		n := uint64(number.Int64())
		blk = b.state.GetBlock(common.Hash{}, n)
	}

	return types.NewBlock(
		convertHeader(&blk.EvmHeader),
		blk.Transactions,
		nil,
		nil,
	), nil
}

func (b *EthAPIBackend) StateAndHeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// TODO: implement or disable it. Origin:
	/*
		// Pending state is only known by the miner
		if number == rpc.PendingBlockNumber {
			block, state := b.svc.miner.Pending()
			return state, block.Header(), nil
		}
		// Otherwise resolve the block number and return its state
		header, err := b.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, nil, err
		}
		if header == nil {
			return nil, nil, errors.New("header not found")
		}
		stateDb, err := b.svc.BlockChain().StateAt(header.Root)
		return stateDb, header, err
	*/
	return nil, nil, ErrNotImplemented("StateAndHeaderByNumber")
}

func (b *EthAPIBackend) GetHeader(ctx context.Context, hash common.Hash) *types.Header {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.blockchain.GetHeaderByHash(hash)
	*/
	return nil
}

func (b *EthAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.blockchain.GetBlockByHash(hash), nil
	*/
	return nil, ErrNotImplemented("GetBlock")
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
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.blockchain.GetTdByHash(blockHash)
	*/
	return big.NewInt(0)
}

func (b *EthAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error) {
	// TODO: implement or disable it. Origin:
	/*
		state.SetBalance(msg.From(), math.MaxBig256)
		vmError := func() error { return nil }

		context := core.NewEVMContext(msg, header, b.svc.BlockChain(), nil)
		return vm.NewEVM(context, state, b.svc.blockchain.Config(), *b.svc.blockchain.GetVMConfig()), vmError, nil
	*/
	return nil, nil, ErrNotImplemented("GetEVM")
}

func (b *EthAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeRemovedLogsEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeChainEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeChainHeadEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeChainSideEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.BlockChain().SubscribeLogsEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.txPool.AddLocal(signedTx)
	*/
	return ErrNotImplemented("SendTx")
}

func (b *EthAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	// TODO: implement or disable it. Origin:
	/*
		pending, err := b.svc.txPool.Pending()
		if err != nil {
			return nil, err
		}
		var txs types.Transactions
		for _, batch := range pending {
			txs = append(txs, batch...)
		}
		return txs, nil
	*/
	return nil, ErrNotImplemented("GetPoolTransactions")
}

func (b *EthAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.txPool.Get(hash)
	*/
	return nil
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
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.txPool.Stats()
	*/
	return 0, 0
}

func (b *EthAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.TxPool().Content()
	*/
	return nil, nil
}

func (b *EthAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.TxPool().SubscribeNewTxsEvent(ch)
	*/
	return nil
}

func (b *EthAPIBackend) Downloader() *downloader.Downloader {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.Downloader()
	*/
	return nil
}

func (b *EthAPIBackend) ProtocolVersion() int {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.EthVersion()
	*/
	return 0
}

func (b *EthAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	// TODO: implement or disable it. Origin:
	/*
		return b.gpo.SuggestPrice(ctx)
	*/
	return nil, ErrNotImplemented("SuggestPrice")
}

func (b *EthAPIBackend) ChainDb() ethdb.Database {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.ChainDb()
	*/
	return nil
}

func (b *EthAPIBackend) EventMux() *event.TypeMux {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.EventMux()
	*/
	return nil
}

func (b *EthAPIBackend) AccountManager() *accounts.Manager {
	return b.svc.AccountManager()
}

func (b *EthAPIBackend) ExtRPCEnabled() bool {
	return b.extRPCEnabled
}

func (b *EthAPIBackend) RPCGasCap() *big.Int {
	// TODO: implement or disable it. Origin:
	/*
		return b.svc.config.RPCGasCap
	*/
	return big.NewInt(0)
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

/*
 * utils:
 */

func convertHeader(h *evm_core.EvmHeader) *types.Header {
	// NOTE: incomplete conversion
	return &types.Header{
		Number:     h.Number,
		Coinbase:   h.Coinbase,
		GasLimit:   h.GasLimit,
		Root:       h.Root,
		ParentHash: h.ParentHash,
		Time:       uint64(h.Time.Unix()),
	}
}
