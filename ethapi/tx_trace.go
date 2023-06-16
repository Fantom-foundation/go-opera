// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/contracts/sfc"
	"github.com/Fantom-foundation/go-opera/txtrace"
	"github.com/Fantom-foundation/go-opera/utils/signers/gsignercache"
)

// PublicTxTraceAPI provides an API to access transaction tracing.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicTxTraceAPI struct {
	b Backend
}

// NewPublicTxTraceAPI creates a new transaction trace API.
func NewPublicTxTraceAPI(b Backend) *PublicTxTraceAPI {
	return &PublicTxTraceAPI{b}
}

// Trace transaction and return processed result
func (s *PublicTxTraceAPI) traceTx(
	ctx context.Context, blockCtx vm.BlockContext, msg types.Message,
	state *state.StateDB, block *evmcore.EvmBlock, tx *types.Transaction, index uint64,
	status uint64, chainConfig *params.ChainConfig) (*[]txtrace.ActionTrace, error) {

	// Providing default config
	// In case of trace transaction node, this config is changed
	cfg := opera.DefaultVMConfig
	cfg.Debug = true
	txTracer := txtrace.NewTraceStructLogger(nil)
	cfg.Tracer = txTracer
	cfg.NoBaseFee = true

	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var timeout time.Duration = 5 * time.Second
	if s.b.RPCEVMTimeout() > 0 {
		timeout = s.b.RPCEVMTimeout()
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)

	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()
	txTracer.SetTx(tx.Hash())
	txTracer.SetFrom(msg.From())
	txTracer.SetTo(msg.To())
	txTracer.SetValue(*msg.Value())
	txTracer.SetBlockHash(block.Hash)
	txTracer.SetBlockNumber(block.Number)
	txTracer.SetTxIndex(uint(index))
	txTracer.SetGasUsed(tx.Gas())

	var txContext = evmcore.NewEVMTxContext(msg)
	vmenv := vm.NewEVM(blockCtx, txContext, state, chainConfig, cfg)

	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		vmenv.Cancel()
	}()

	// Setup the gas pool and stateDB
	gp := new(evmcore.GasPool).AddGas(msg.Gas())
	state.Prepare(tx.Hash(), int(index))
	result, err := evmcore.ApplyMessage(vmenv, msg, gp)

	if result != nil {
		txTracer.SetGasUsed(result.UsedGas)
	}
	// Process traces if any
	txTracer.ProcessTx()
	traceActions := txTracer.GetTraceActions()
	state.Finalise(true)
	if err != nil {
		errTrace := txtrace.GetErrorTraceFromMsg(&msg, block.Hash, *block.Number, tx.Hash(), index, err)
		at := make([]txtrace.ActionTrace, 0)
		at = append(at, *errTrace)
		if status == 1 {
			return nil, fmt.Errorf("invalid transaction replay state at %s", tx.Hash().String())
		}
		return &at, nil
	}
	// If the timer caused an abort, return an appropriate error message
	if vmenv.Cancelled() {
		log.Info("EVM was canceled due to timeout when replaying transaction ", "txHash", tx.Hash().String())
		return nil, fmt.Errorf("timeout when replaying tx")
	}

	if result != nil && result.Err != nil {
		if len(*traceActions) == 0 {
			log.Error("error in result when replaying transaction:", "txHash", tx.Hash().String(), " err", result.Err.Error())
			errTrace := txtrace.GetErrorTraceFromMsg(&msg, block.Hash, *block.Number, tx.Hash(), index, result.Err)
			at := make([]txtrace.ActionTrace, 0)
			at = append(at, *errTrace)
			return &at, nil
		}
		if status == 1 {
			return nil, fmt.Errorf("invalid transaction replay state at %s", tx.Hash().String())
		}
		return traceActions, nil
	}

	if status == 0 {
		return nil, fmt.Errorf("invalid transaction replay state at %s", tx.Hash().String())
	}
	return traceActions, nil
}

// Gets all transaction from specified block and process them
func (s *PublicTxTraceAPI) traceBlock(ctx context.Context, block *evmcore.EvmBlock, txHash *common.Hash, traceIndex *[]hexutil.Uint) (*[]txtrace.ActionTrace, error) {
	var (
		blockNumber   int64
		parentBlockNr rpc.BlockNumber
	)

	if block != nil && block.NumberU64() > 0 {
		blockNumber = block.Number.Int64()
		parentBlockNr = rpc.BlockNumber(blockNumber - 1)
	} else {
		return nil, fmt.Errorf("invalid block for tracing")
	}

	// Check if node is synced
	if s.b.CurrentBlock().Number.Int64() < blockNumber {
		return nil, fmt.Errorf("invalid block %v for tracing, current block is %v", blockNumber, s.b.CurrentBlock())
	}

	callTrace := txtrace.CallTrace{
		Actions: make([]txtrace.ActionTrace, 0),
	}

	signer := gsignercache.Wrap(types.MakeSigner(s.b.ChainConfig(), block.Number))
	blockCtx := s.b.GetBlockContext(block.Header())

	allTxOK := true
	// loop thru all transactions in the block get them from DB
	for _, tx := range block.Transactions {
		// Get transaction trace from backend persistent store
		// Otherwise replay transaction and save trace to db
		traces, err := s.b.TxTraceByHash(ctx, tx.Hash())
		if err == nil {
			if txHash == nil || *txHash == tx.Hash() {
				callTrace.AddTraces(traces, traceIndex)
				if txHash != nil {
					break
				}
			}
		} else {
			allTxOK = false
			break
		}
	}

	if !allTxOK {

		// get state from block parent, to be able to recreate correct nonces
		state, _, err := s.b.StateAndHeaderByNumberOrHash(ctx, rpc.BlockNumberOrHash{BlockNumber: &parentBlockNr})
		if err != nil {
			return nil, fmt.Errorf("cannot get state for block %v, error: %v", block.NumberU64(), err.Error())
		}
		receipts, err := s.b.GetReceiptsByNumber(ctx, rpc.BlockNumber(blockNumber))
		if err != nil {
			log.Debug("Cannot get receipts for block", "block", blockNumber, "err", err.Error())
			return nil, fmt.Errorf("cannot get receipts for block %v, error: %v", block.NumberU64(), err.Error())
		}

		callTrace = txtrace.CallTrace{
			Actions: make([]txtrace.ActionTrace, 0),
		}

		// loop thru all transactions in the block and process them
		for i, tx := range block.Transactions {
			if txHash == nil || *txHash == tx.Hash() {

				log.Info("Replaying transaction", "txHash", tx.Hash().String())
				// get full transaction info
				tx, _, index, err := s.b.GetTransaction(ctx, tx.Hash())
				if err != nil {
					log.Debug("Cannot get transaction", "txHash", tx.Hash().String(), "err", err.Error())
					callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, nil, tx.To(), tx.Hash(), index, err))
					continue
				}
				msg, err := tx.AsMessage(signer, block.BaseFee)
				if err != nil {
					callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, nil, tx.To(), tx.Hash(), index, errors.New("not able to decode tx")))
					continue
				}
				from := msg.From()
				if tx.To() != nil && *tx.To() == sfc.ContractAddress {
					errTrace := txtrace.GetErrorTrace(block.Hash, *block.Number, &from, tx.To(), tx.Hash(), index, errors.New("sfc tx"))
					at := make([]txtrace.ActionTrace, 0)
					at = append(at, *errTrace)
					callTrace.AddTrace(errTrace)
					jsonTraceBytes, _ := json.Marshal(&at)
					s.b.TxTraceSave(ctx, tx.Hash(), jsonTraceBytes)
				} else {
					txTraces, err := s.traceTx(ctx, blockCtx, msg, state, block, tx, index, receipts[i].Status, s.b.ChainConfig())
					if err != nil {
						log.Debug("Cannot get transaction trace for transaction", "txHash", tx.Hash().String(), "err", err.Error())
						callTrace.AddTrace(txtrace.GetErrorTraceFromMsg(&msg, block.Hash, *block.Number, tx.Hash(), index, err))
					} else {
						callTrace.AddTraces(txTraces, traceIndex)

						// Save trace result into persistent key-value store
						jsonTraceBytes, _ := json.Marshal(txTraces)
						s.b.TxTraceSave(ctx, tx.Hash(), jsonTraceBytes)
					}
				}
				if txHash != nil {
					break
				}
			} else if txHash != nil {
				log.Info("Replaying transaction without trace", "txHash", tx.Hash().String())
				// Generate the next state snapshot fast without tracing
				msg, _ := tx.AsMessage(signer, block.BaseFee)

				state.Prepare(tx.Hash(), i)
				vmConfig := opera.DefaultVMConfig
				vmConfig.NoBaseFee = true
				vmConfig.Debug = false
				vmConfig.Tracer = nil
				vmenv := vm.NewEVM(blockCtx, evmcore.NewEVMTxContext(msg), state, s.b.ChainConfig(), vmConfig)
				res, err := evmcore.ApplyMessage(vmenv, msg, new(evmcore.GasPool).AddGas(msg.Gas()))
				failed := false
				if err != nil {
					failed = true
					log.Error("Cannot replay transaction", "txHash", tx.Hash().String(), "err", err.Error())
				}
				if res != nil && res.Err != nil {
					failed = true
					log.Debug("Error replaying transaction", "txHash", tx.Hash().String(), "err", res.Err.Error())
				}
				// Finalize the state so any modifications are written to the trie
				state.Finalise(true)
				if (failed && receipts[i].Status == 1) || (!failed && receipts[i].Status == 0) {
					return nil, fmt.Errorf("invalid transaction replay state at %s", tx.Hash().String())
				}
			}
		}
	}

	// In case of empty result create empty trace for empty block
	if len(callTrace.Actions) == 0 {
		if traceIndex != nil || txHash != nil {
			return nil, nil
		} else {
			emptyTrace := txtrace.CallTrace{
				Actions: make([]txtrace.ActionTrace, 0),
			}
			blockTrace := txtrace.NewActionTrace(block.Hash, *block.Number, common.Hash{}, 0, "empty")
			txAction := txtrace.NewAddressAction(&common.Address{}, 0, []byte{}, nil, hexutil.Big{}, nil)
			blockTrace.Action = txAction
			blockTrace.Error = "Empty block"
			emptyTrace.AddTrace(blockTrace)
			return &emptyTrace.Actions, nil
		}
	}

	return &callTrace.Actions, nil
}

/* trace_block function returns transaction traces in givven block
* When blockNr is -1 the chain head is returned.
* When blockNr is -2 the pending chain head is returned.
* When fullTx is true all transactions in the block are returned, otherwise
* only the transaction hash is returned.
 */
func (s *PublicTxTraceAPI) Block(ctx context.Context, numberOrHash rpc.BlockNumberOrHash) (*[]txtrace.ActionTrace, error) {

	blockNr, _ := numberOrHash.Number()

	defer func(start time.Time) {
		log.Info("Executing trace_block call finished", "blockNr", blockNr.Int64(), "runtime", time.Since(start))
	}(time.Now())

	block, err := s.b.BlockByNumber(ctx, blockNr)
	if err != nil {
		log.Debug("Cannot get block from db", "blockNr", blockNr)
		return nil, err
	}

	return s.traceBlock(ctx, block, nil, nil)
}

// Transaction trace_transaction function returns transaction traces
func (s *PublicTxTraceAPI) Transaction(ctx context.Context, hash common.Hash) (*[]txtrace.ActionTrace, error) {
	defer func(start time.Time) {
		log.Info("Executing trace_transaction call finished", "txHash", hash.String(), "runtime", time.Since(start))
	}(time.Now())
	return s.traceTxHash(ctx, hash, nil)
}

// Get trace_get function returns transaction traces on specified index position of the traces
// If index is nil, then just root trace is returned
func (s *PublicTxTraceAPI) Get(ctx context.Context, hash common.Hash, traceIndex []hexutil.Uint) (*[]txtrace.ActionTrace, error) {
	defer func(start time.Time) {
		log.Info("Executing trace_get call finished", "txHash", hash.String(), "index", traceIndex, "runtime", time.Since(start))
	}(time.Now())
	return s.traceTxHash(ctx, hash, &traceIndex)
}

// traceTxHash looks for a block of this transaction hash and trace it
func (s *PublicTxTraceAPI) traceTxHash(ctx context.Context, hash common.Hash, traceIndex *[]hexutil.Uint) (*[]txtrace.ActionTrace, error) {
	_, blockNumber, _, _ := s.b.GetTransaction(ctx, hash)
	blkNr := rpc.BlockNumber(blockNumber)
	block, err := s.b.BlockByNumber(ctx, blkNr)
	if err != nil {
		log.Debug("Cannot get block from db", "blockNr", blkNr)
		return nil, err
	}
	callTrace := txtrace.CallTrace{
		Actions: make([]txtrace.ActionTrace, 0),
	}

	// Get transaction trace from backend persistent store
	// Otherwise replay transaction and save trace to db
	traces, err := s.b.TxTraceByHash(ctx, hash)
	if err == nil && len(*traces) > 0 {
		callTrace.AddTraces(traces, traceIndex)
	}
	if len(callTrace.Actions) != 0 {
		return &callTrace.Actions, nil
	}

	return s.traceBlock(ctx, block, &hash, traceIndex)
}

// FilterArgs represents the arguments for specifiing trace targets
type FilterArgs struct {
	FromAddress *[]common.Address      `json:"fromAddress"`
	ToAddress   *[]common.Address      `json:"toAddress"`
	FromBlock   *rpc.BlockNumberOrHash `json:"fromBlock"`
	ToBlock     *rpc.BlockNumberOrHash `json:"toBlock"`
	After       uint                   `json:"after"`
	Count       uint                   `json:"count"`
}

// Filter is function for trace_filter rpc call
func (s *PublicTxTraceAPI) Filter(ctx context.Context, args FilterArgs) (*[]txtrace.ActionTrace, error) {
	// add log after execution
	defer func(start time.Time) {

		var data []interface{}
		if args.FromBlock != nil {
			data = append(data, "fromBlock", args.FromBlock.BlockNumber.Int64())
		}
		if args.ToBlock != nil {
			data = append(data, "toBlock", args.ToBlock.BlockNumber.Int64())
		}
		if args.FromAddress != nil {
			adresses := make([]string, 0)
			for _, addr := range *args.FromAddress {
				adresses = append(adresses, addr.String())
			}
			data = append(data, "fromAddr", adresses)
		}
		if args.ToAddress != nil {
			adresses := make([]string, 0)
			for _, addr := range *args.ToAddress {
				adresses = append(adresses, addr.String())
			}
			data = append(data, "toAddr", adresses)
		}
		data = append(data, "time", time.Since(start))
		log.Info("Executing trace_filter call finished", data...)
	}(time.Now())

	// process arguments
	var (
		fromBlock, toBlock rpc.BlockNumber
		mainErr            error
	)
	if args.FromBlock != nil {
		fromBlock = *args.FromBlock.BlockNumber
	}
	if args.ToBlock != nil {
		toBlock = *args.ToBlock.BlockNumber
		if toBlock == rpc.LatestBlockNumber || toBlock == rpc.PendingBlockNumber {
			toBlock = rpc.BlockNumber(s.b.CurrentBlock().NumberU64())
		}
	} else {
		toBlock = rpc.BlockNumber(s.b.CurrentBlock().NumberU64())
	}

	// counter of processed traces
	var traceAdded, traceCount uint
	var fromAddresses, toAddresses map[common.Address]struct{}
	if args.FromAddress != nil {
		fromAddresses = make(map[common.Address]struct{})
		for _, addr := range *args.FromAddress {
			fromAddresses[addr] = struct{}{}
		}
	}
	if args.ToAddress != nil {
		toAddresses = make(map[common.Address]struct{})
		for _, addr := range *args.ToAddress {
			toAddresses[addr] = struct{}{}
		}
	}

	// check for context timeout
	contextDone := false
	go func() {
		<-ctx.Done()
		contextDone = true
	}()

	// struct for collecting result traces
	callTrace := txtrace.CallTrace{
		Actions: make([]txtrace.ActionTrace, 0),
	}

	// count of traces doesn't matter so use parallel workers
	if args.Count == 0 {
		workerCount := runtime.NumCPU() / 2
		blocks := make(chan rpc.BlockNumber, 10000)
		results := make(chan txtrace.ActionTrace, 100000)

		// create workers and their sync group
		var wg sync.WaitGroup
		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			wId := w
			go func() {
				defer wg.Done()
				worker(wId, s, ctx, blocks, results, fromAddresses, toAddresses)
			}()
		}

		// add all blocks in specified range for processing
		for i := fromBlock; i <= toBlock; i++ {
			blocks <- i
		}
		close(blocks)

		var wgResult sync.WaitGroup
		wgResult.Add(1)
		go func() {
			defer wgResult.Done()
			// collect results
			for trace := range results {
				callTrace.AddTrace(&trace)
			}
		}()

		// wait for proccessing all blocks
		wg.Wait()
		close(results)

		wgResult.Wait()
	} else {
	blocks:
		// go thru all blocks in specified range
		for i := fromBlock; i <= toBlock; i++ {
			block, err := s.b.BlockByNumber(ctx, i)
			if err != nil {
				mainErr = err
				break
			}

			// when block has any transaction, then process it
			if block != nil && block.Transactions.Len() > 0 {
				traces, err := s.traceBlock(ctx, block, nil, nil)
				if err != nil {
					mainErr = err
					break
				}

				// loop thru all traces from the block
				// and check
				for _, trace := range *traces {

					if args.Count == 0 || traceAdded < args.Count {
						addTrace := true

						if args.FromAddress != nil || args.ToAddress != nil {
							if args.FromAddress != nil {
								if trace.Action.From == nil {
									addTrace = false
								} else {
									if _, ok := fromAddresses[*trace.Action.From]; !ok {
										addTrace = false
									}
								}
							}
							if args.ToAddress != nil {
								if trace.Action.To == nil {
									addTrace = false
								} else if _, ok := toAddresses[*trace.Action.To]; !ok {
									addTrace = false
								}
							}
						}
						if addTrace {
							if traceCount >= args.After {
								callTrace.AddTrace(&trace)
								traceAdded++
							}
							traceCount++
						}
					} else {
						// already reached desired count of traces in batch
						break blocks
					}
				}
			}
			if contextDone {
				break
			}
		}
	}

	//when timeout occured or another error
	if contextDone || mainErr != nil {
		if mainErr != nil {
			return nil, mainErr
		}
		return nil, fmt.Errorf("timeout when scanning blocks")
	}

	return &callTrace.Actions, nil
}

func worker(id int,
	s *PublicTxTraceAPI,
	ctx context.Context,
	blocks <-chan rpc.BlockNumber,
	results chan<- txtrace.ActionTrace,
	fromAddresses map[common.Address]struct{},
	toAddresses map[common.Address]struct{}) {

	for i := range blocks {
		block, err := s.b.BlockByNumber(ctx, i)
		if err != nil {
			break
		}

		// when block has any transaction, then process it
		if block != nil && block.Transactions.Len() > 0 {
			traces, err := s.traceBlock(ctx, block, nil, nil)
			if err != nil {
				break
			}
			for _, trace := range *traces {
				addTrace := true

				if len(fromAddresses) > 0 {

					if trace.Action.From == nil {
						addTrace = false
					} else {
						if _, ok := fromAddresses[*trace.Action.From]; !ok {
							addTrace = false
						}
					}
				}
				if len(toAddresses) > 0 {
					if trace.Action.To == nil {
						addTrace = false
					} else if _, ok := toAddresses[*trace.Action.To]; !ok {
						addTrace = false
					}
				}
				if addTrace {
					results <- trace
				}
			}
		}
	}
}
