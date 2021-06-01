package ethapi

import (
	"context"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
	"hash"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/crypto/sha3"
)

type helpHash struct {
	hashed hash.Hash
}

func newHash() *helpHash {

	return &helpHash{hashed: sha3.NewLegacyKeccak256()}
}

func (h *helpHash) Reset() {
	h.hashed.Reset()
}

func (h *helpHash) Update(key, val []byte) {
	h.hashed.Write(key)
	h.hashed.Write(val)
}

func (h *helpHash) Hash() common.Hash {
	return common.BytesToHash(h.hashed.Sum(nil))
}

type PreExecTx struct {
	ChainId                                     *big.Int
	From, To, Data, Value, Gas, GasPrice, Nonce string
}

type preData struct {
	block   *evmcore.EvmBlock
	tx      *types.Transaction
	msg     types.Message
	stateDb *state.StateDB
	header  *evmcore.EvmHeader
}

// PreExecAPI provides pre exec info for rpc
type PreExecAPI struct {
	b Backend
}

func NewPreExecAPI(b Backend) *PreExecAPI {
	return &PreExecAPI{b}
}

func (api *PreExecAPI) getBlockAndMsg(origin *PreExecTx, number *big.Int) (*evmcore.EvmBlock, types.Message) {
	fromAddr := common.HexToAddress(origin.From)
	toAddr := common.HexToAddress(origin.To)

	tx := types.NewTransaction(
		hexutil.MustDecodeUint64(origin.Nonce),
		toAddr,
		hexutil.MustDecodeBig(origin.Value),
		hexutil.MustDecodeUint64(origin.Gas),
		hexutil.MustDecodeBig(origin.GasPrice),
		hexutil.MustDecode(origin.Data),
	)

	number.Add(number, big.NewInt(1))
	block := &evmcore.EvmBlock{}
	block.EvmHeader = *evmcore.ConvertFromEthHeader(&types.Header{Number: number})
	block.Transactions = []*types.Transaction{tx}

	msg := types.NewMessage(
		fromAddr,
		&toAddr,
		hexutil.MustDecodeUint64(origin.Nonce),
		hexutil.MustDecodeBig(origin.Value),
		hexutil.MustDecodeUint64(origin.Gas),
		hexutil.MustDecodeBig(origin.GasPrice),
		hexutil.MustDecode(origin.Data),
		false,
	)

	return block, msg
}

func (api *PreExecAPI) prepareData(ctx context.Context, origin *PreExecTx) (*preData, error) {
	var (
		d             preData
		err           error
		parentBlockNr rpc.BlockNumber
	)
	parentBlockNr = rpc.LatestBlockNumber

	d.stateDb, d.header, err = api.b.StateAndHeaderByNumberOrHash(ctx, rpc.BlockNumberOrHash{BlockNumber: &parentBlockNr})
	if err != nil {
		return nil, err
	}
	d.block, d.msg = api.getBlockAndMsg(origin, big.NewInt(parentBlockNr.Int64()))
	d.tx = d.block.Transactions[0]
	return &d, nil
}

func (api *PreExecAPI) GetLogs(ctx context.Context, origin *PreExecTx) (*types.Receipt, error) {
	d, err := api.prepareData(ctx, origin)
	if err != nil {
		return nil, err
	}
	d.stateDb.Prepare(d.tx.Hash(), d.block.Hash, 0)
	gas := d.tx.Gas()
	d.stateDb.Prepare(d.tx.Hash(), d.block.Hash, 0)

	cfg := opera.DefaultVMConfig
	cfg.Debug = true
	var timeout = 3 * time.Second
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	evm, vmError, err := api.b.GetEVMWithCfg(nil, d.msg, d.stateDb, d.block.Header(), cfg)
	defer cancel()
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()
	if err != nil {
		return nil, err
	}

	d.stateDb.Prepare(d.tx.Hash(), d.block.Hash, 0)
	result, err := evmcore.ApplyMessage(evm, d.msg, new(evmcore.GasPool).AddGas(gas))
	if err = vmError(); err != nil {

	}

	config := api.b.ChainConfig()
	var root []byte
	if config.IsByzantium(d.header.Number) {
		d.stateDb.Finalise(true)
	} else {
		root = d.stateDb.IntermediateRoot(config.IsEIP158(d.header.Number)).Bytes()
	}

	receipt := types.NewReceipt(root, result.Failed(), result.UsedGas)
	receipt.TxHash = d.tx.Hash()
	receipt.GasUsed = result.UsedGas
	receipt.Logs = d.stateDb.GetLogs(d.tx.Hash())
	receipt.BlockHash = d.stateDb.BlockHash()
	receipt.BlockNumber = d.block.Header().Number
	receipt.TransactionIndex = 0
	return receipt, nil
}

// TraceTransaction tracing pre-exec transaction object.
func (api *PreExecAPI) TraceTransaction(ctx context.Context, origin *PreExecTx) (interface{}, error) {
	d, err := api.prepareData(ctx, origin)
	if err != nil {
		return nil, err
	}
	d.stateDb.Prepare(d.tx.Hash(), d.block.Hash, 0)
	return traceTx(ctx, d.stateDb, d.header, api.b, d.block, d.tx, 0)
}
