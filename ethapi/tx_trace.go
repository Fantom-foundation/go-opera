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
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
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
func traceTx(ctx context.Context, state *state.StateDB, header *evmcore.EvmHeader, backend Backend, block *evmcore.EvmBlock, tx *types.Transaction, index uint64) (*[]txtrace.ActionTrace, error) {

	var mainErr error
	// Providing default config
	// In case of trace transaction node, this config is changed
	cfg := opera.DefaultVMConfig
	cfg.Debug = true
	txTracer := txtrace.NewTraceStructLogger(nil)
	cfg.Tracer = txTracer

	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	// TODO add time into the server configuration
	var timeout time.Duration = 3 * time.Second
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)

	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// create a signer of this transaction
	signer := gsignercache.Wrap(types.MakeSigner(backend.ChainConfig(), block.Number))
	var from common.Address

	// reconstruct message from transaction
	msg, err := tx.AsMessage(signer, header.BaseFee)
	if err != nil {
		log.Debug("Can't recreate message for transaction:", "txHash", tx.Hash().String(), " err", err.Error())
		if v, _, _ := tx.RawSignatureValues(); new(big.Int).Cmp(v) == 0 {
			from = common.Address{}
		}
	} else {
		from = msg.From()
	}

	txTracer.SetTx(tx.Hash())
	txTracer.SetFrom(msg.From())
	txTracer.SetTo(msg.To())
	txTracer.SetValue(*msg.Value())
	txTracer.SetBlockHash(block.Hash)
	txTracer.SetBlockNumber(block.Number)
	txTracer.SetTxIndex(uint(index))
	txTracer.SetGasUsed(tx.Gas())

	// Changing some variables for replay and get a new instance of the EVM.
	replayMsg := types.NewMessage(from, msg.To(), 0, msg.Value(), tx.Gas(), tx.GasPrice(), tx.GasFeeCap(), tx.GasTipCap(), msg.Data(), tx.AccessList(), true)
	evm, vmError, err := backend.GetEVM(nil, replayMsg, state, header, &cfg)
	if err != nil {
		log.Error("Can't get evm for processing transaction:", "txHash", tx.Hash().String(), " err", err.Error())
		mainErr = err
	}
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(evmcore.GasPool).AddGas(math.MaxUint64)
	state.Prepare(tx.Hash(), int(index))
	result, err := evmcore.ApplyMessage(evm, replayMsg, gp)
	if err = vmError(); err != nil {
		log.Error("Error when replaying transaction:", "txHash", tx.Hash().String(), " err", err.Error())
		mainErr = err
	}
	// If the timer caused an abort, return an appropriate error message
	if evm.Cancelled() {
		log.Info("EVM was canceled due to timeout when replaying transaction ", "txHash", tx.Hash().String())
		mainErr = err
	}

	if mainErr != nil {
		return nil, mainErr
	}

	if result.Err != nil {
		return nil, result.Err
	}

	txTracer.ProcessTx()

	return txTracer.GetTraceActions(), nil
}

// Gets all transaction from specified block and process them
func traceBlock(ctx context.Context, block *evmcore.EvmBlock, backend Backend, txHash *common.Hash, traceIndex *[]hexutil.Uint) (*[]txtrace.ActionTrace, error) {
	var (
		blockNumber   int64
		parentBlockNr rpc.BlockNumber
	)

	if block != nil && block.NumberU64() > 0 {
		blockNumber = block.Number.Int64()
		parentBlockNr = rpc.BlockNumber(blockNumber - 1)
	} else {
		return nil, fmt.Errorf("Invalid block for tracing")
	}

	// Check if node is synced
	if backend.CurrentBlock().Number.Int64() < blockNumber {
		return nil, fmt.Errorf("Invalid block %v for tracing, current block is %v", blockNumber, backend.CurrentBlock())
	}

	callTrace := txtrace.CallTrace{
		Actions: make([]txtrace.ActionTrace, 0),
	}

	// loop thru all transactions in the block and process them
	for _, tx := range block.Transactions {
		if txHash == nil || *txHash == tx.Hash() {

			// Get transaction trace from backend persistent store
			// Otherwise replay transaction and save trace to db
			traces, err := backend.TxTraceByHash(ctx, tx.Hash())
			if err == nil {
				callTrace.AddTraces(traces, traceIndex)
			} else {
				log.Info("Replaying transaction", "txHash", tx.Hash().String())
				// get full transaction info
				tx, _, index, err := backend.GetTransaction(ctx, tx.Hash())
				if err != nil {
					log.Debug("Cannot get transaction", "txHash", tx.Hash().String(), "err", err.Error())
					callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, tx.To(), tx.Hash(), index, err))
					continue
				}

				receipts, err := backend.GetReceiptsByNumber(ctx, rpc.BlockNumber(blockNumber))
				if err != nil {
					log.Debug("Cannot get receipts for block", "block", blockNumber, "err", err.Error())
					callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, tx.To(), tx.Hash(), index, err))
					continue
				}

				if len(receipts) < int(index) {
					receipt := receipts[index]
					if receipt.Status == types.ReceiptStatusFailed {
						log.Debug("Transaction has status failed", "block", blockNumber, "err", err.Error())
						callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, tx.To(), tx.Hash(), index, nil))
						continue
					}
				}

				if tx.To() != nil && *tx.To() == sfc.ContractAddress {
					callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, tx.To(), tx.Hash(), index, err))
				} else {

					// get state and header from block parent, to be able to recreate correct nonces
					state, header, err := backend.StateAndHeaderByNumberOrHash(ctx, rpc.BlockNumberOrHash{BlockNumber: &parentBlockNr})
					if err != nil {
						log.Debug("Cannot get state for blockblock ", "block", block.NumberU64(), "err", err.Error())
						callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, nil, common.Hash{}, 0, err))
						continue
					}

					txTraces, err := traceTx(ctx, state, header, backend, block, tx, index)
					if err != nil {
						log.Debug("Cannot get transaction trace for transaction", "txHash", tx.Hash().String(), "err", err.Error())
						callTrace.AddTrace(txtrace.GetErrorTrace(block.Hash, *block.Number, tx.To(), tx.Hash(), index, err))
					} else {
						callTrace.AddTraces(txTraces, traceIndex)

						// Save trace result into persistent key-value store
						jsonTraceBytes, _ := json.Marshal(txTraces)
						backend.TxTraceSave(ctx, tx.Hash(), jsonTraceBytes)
					}
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
			blockTrace.Action = *txAction
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

	return traceBlock(ctx, block, s.b, nil, nil)
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
	return traceBlock(ctx, block, s.b, &hash, traceIndex)
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

	// TODO put timeout to server configuration
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

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
			traces, err := traceBlock(ctx, block, s.b, nil, nil)
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

	//when timeout occured or another error
	if contextDone || mainErr != nil {
		if mainErr != nil {
			return nil, mainErr
		}
		return nil, fmt.Errorf("Timeout when scanning blocks")
	}

	return &callTrace.Actions, nil
}
