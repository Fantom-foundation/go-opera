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
	"fmt"
	"math/big"
	"strings"
	"time"
	"errors"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera/genesis/sfc"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
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

// CallTrace is struct for holding tracing results
type CallTrace struct {
	Actions []ActionTrace  `json:"result"`
	Stack   []*ActionTrace `json:"-"`
}

// addTrace Append trace to call trace list
func (callTrace *CallTrace) addTrace(blockTrace *ActionTrace) {
	callTrace.Actions = append(callTrace.Actions, *blockTrace)
}

// addTraces Append traces to call trace list
func (callTrace *CallTrace) addTraces(traces *[]ActionTrace) {
	for _, trace := range *traces {
		callTrace.addTrace(&trace)
	}
}

// lastTrace Get last trace in call trace list
func (callTrace *CallTrace) lastTrace() *ActionTrace {
	if len(callTrace.Actions) > 0 {
		return &callTrace.Actions[len(callTrace.Actions)-1]
	}
	return nil
}

// NewActionTrace creates new instance of type ActionTrace
func NewActionTrace(bHash common.Hash, bNumber big.Int, tHash common.Hash, tPos uint64, tType string) *ActionTrace {
	return &ActionTrace{
		BlockHash:           bHash,
		BlockNumber:         bNumber,
		TransactionHash:     tHash,
		TransactionPosition: tPos,
		TraceType:           tType,
		TraceAddress:        make([]int, 0),
		Result:              TraceActionResult{},
	}
}

// NewActionTraceFromTrace creates new instance of type ActionTrace
// based on another trace
func NewActionTraceFromTrace(actionTrace *ActionTrace, tType string, traceAddress []int) *ActionTrace {
	trace := NewActionTrace(
		actionTrace.BlockHash,
		actionTrace.BlockNumber,
		actionTrace.TransactionHash,
		actionTrace.TransactionPosition,
		tType)
	trace.TraceAddress = traceAddress
	return trace
}

const (
	CALL   = "call"
	CREATE = "create"
)

// ActionTrace represents single interaction with blockchain
type ActionTrace struct {
	childTraces         []*ActionTrace    `json:"-"`
	Action              AddressAction     `json:"action"`
	BlockHash           common.Hash       `json:"blockHash"`
	BlockNumber         big.Int           `json:"blockNumber"`
	Result              TraceActionResult `json:"result"`
	Error               string            `json:"error,omitempty"`
	Subtraces           uint64            `json:"subtraces"`
	TraceAddress        []int             `json:"traceAddress"`
	TransactionHash     common.Hash       `json:"transactionHash"`
	TransactionPosition uint64            `json:"transactionPosition"`
	TraceType           string            `json:"type"`
}

// NewAddressAction creates specific information about trace addresses
func NewAddressAction(from *common.Address, gas uint64, data []byte, to *common.Address, value hexutil.Big, callType *string) *AddressAction {
	action := AddressAction{
		From:     from,
		To:       to,
		Gas:      hexutil.Uint64(gas),
		Value:    value,
		CallType: callType,
	}
	if callType == nil {
		action.Init = hexutil.Bytes(data)
	} else {
		action.Input = hexutil.Bytes(data)
	}
	return &action
}

// AddressAction represents more specific information about
// account interaction
type AddressAction struct {
	CallType      *string         `json:"callType,omitempty"`
	From          *common.Address `json:"from"`
	To            *common.Address `json:"to,omitempty"`
	Value         hexutil.Big     `json:"value"`
	Gas           hexutil.Uint64  `json:"gas"`
	Init          hexutil.Bytes   `json:"init,omitempty"`
	Input         hexutil.Bytes   `json:"input,omitempty"`
	Address       *common.Address `json:"address,omitempty"`
	RefundAddress *common.Address `json:"refund_address,omitempty"`
	Balance       *hexutil.Big    `json:"balance,omitempty"`
}

// TraceActionResult holds information related to result of the
// processed transaction
type TraceActionResult struct {
	GasUsed   hexutil.Uint64  `json:"gasUsed"`
	Output    *hexutil.Bytes  `json:"output,omitempty"`
	Code      hexutil.Bytes   `json:"code,omitempty"`
	Address   *common.Address `json:"address,omitempty"`
	RetOffset int64           `json:"-"`
	RetSize   int64           `json:"-"`
}

// depthState is struct for having state of logs processing
type depthState struct {
	level  int
	create bool
}

// returns last state
func lastState(state []depthState) *depthState {
	return &state[len(state)-1]
}

// adds trace address and retuns it
func addTraceAddress(traceAddress []int, depth int) []int {
	index := depth - 1
	result := make([]int, len(traceAddress))
	copy(result, traceAddress)
	if len(result) <= index {
		result = append(result, 0)
	} else {
		result[index]++
	}
	return result
}

// removes trace address based on depth of process
func removeTraceAddressLevel(traceAddress []int, depth int) []int {
	if len(traceAddress) > depth {
		result := make([]int, len(traceAddress))
		copy(result, traceAddress)

		result = result[:len(result)-1]
		return result
	}
	return traceAddress
}

// processStructLog itterates thru instruction log and produces all
// transaction tracing informations
func processStructLog(ctx context.Context, backend Backend, structLogger *TraceStructLogger, callTrace *CallTrace, mainBlockTrace *ActionTrace, create bool) {

	state := []depthState{{0, create}}
	traceAddress := make([]int, 0)

	callTrace.Stack = append(callTrace.Stack, &callTrace.Actions[len(callTrace.Actions)-1])

	for i, logg := range structLogger.StructLogs() {

		// when going back from inner call
		if lastState(state).level == logg.Depth {

			result := &callTrace.Stack[len(callTrace.Stack)-1].Result
			if lastState(state).create {
				if len(logg.Stack) > 0 {
					addr := common.BytesToAddress(logg.Stack[len(logg.Stack)-1].Bytes())
					result.Address = &addr
				}
				result.GasUsed = result.GasUsed - hexutil.Uint64(logg.Gas)
			}
			traceAddress = removeTraceAddressLevel(traceAddress, logg.Depth)
			state = state[:len(state)-1]
			callTrace.Stack = callTrace.Stack[:len(callTrace.Stack)-1]
		}

		// match processed instruction and create trace based on it
		switch logg.Op {
		case vm.CREATE, vm.CREATE2:
			traceAddress = addTraceAddress(traceAddress, logg.Depth)
			fromTrace := callTrace.Stack[len(callTrace.Stack)-1]

			// get input data from memory
			stackLastIndex := len(logg.Stack) - 1
			offset := logg.Stack[stackLastIndex-1].Int64()
			inputSize := logg.Stack[stackLastIndex-2].Int64()
			var input []byte
			if inputSize > 0 {
				input = logg.Memory[offset : offset+inputSize]
			}

			// create new trace
			trace := NewActionTraceFromTrace(fromTrace, CREATE, traceAddress)
			traceAction := NewAddressAction(nil, logg.Gas, input, nil, fromTrace.Action.Value, nil)
			trace.Action = *traceAction
			trace.Result.GasUsed = hexutil.Uint64(logg.Gas)
			fromTrace.childTraces = append(fromTrace.childTraces, trace)
			callTrace.Stack = append(callTrace.Stack, trace)
			state = append(state, depthState{logg.Depth, true})

		case vm.CALL, vm.CALLCODE, vm.DELEGATECALL, vm.STATICCALL:
			// get input data from memory
			stackLastIndex := len(logg.Stack) - 1
			var (
				inOffset, inSize, retOffset, retSize int64
				input                                []byte
			)

			if vm.DELEGATECALL == logg.Op || vm.STATICCALL == logg.Op {
				inOffset = logg.Stack[stackLastIndex-2].Int64()
				inSize = logg.Stack[stackLastIndex-3].Int64()
				retOffset = logg.Stack[stackLastIndex-4].Int64()
				retSize = logg.Stack[stackLastIndex-5].Int64()
			} else {
				inOffset = logg.Stack[stackLastIndex-3].Int64()
				inSize = logg.Stack[stackLastIndex-4].Int64()
				retOffset = logg.Stack[stackLastIndex-5].Int64()
				retSize = logg.Stack[stackLastIndex-6].Int64()
			}
			if inSize > 0 {
				input = logg.Memory[inOffset : inOffset+inSize]
			}
			traceAddress = addTraceAddress(traceAddress, logg.Depth)
			fromTrace := callTrace.Stack[len(callTrace.Stack)-1]
			// create new trace
			trace := NewActionTraceFromTrace(fromTrace, CALL, traceAddress)
			action := fromTrace.Action
			addr := common.BytesToAddress(logg.Stack[len(logg.Stack)-2].Bytes())
			callType := strings.ToLower(logg.OpName())
			traceAction := NewAddressAction(action.To, logg.Gas, input, &addr, action.Value, &callType)
			trace.Action = *traceAction
			fromTrace.childTraces = append(fromTrace.childTraces, trace)
			trace.Result.RetOffset = retOffset
			trace.Result.RetSize = retSize
			if len(structLogger.StructLogs()) > i+1 {
				trace.Result.GasUsed = hexutil.Uint64((structLogger.StructLogs()[i+1]).Gas)
			}

			callTrace.Stack = append(callTrace.Stack, trace)
			state = append(state, depthState{logg.Depth, false})

		case vm.RETURN, vm.STOP, vm.REVERT:
			result := &callTrace.Stack[len(callTrace.Stack)-1].Result
			var data []byte

			if vm.STOP != logg.Op {
				offset := logg.Stack[len(logg.Stack)-1].Int64()
				size := logg.Stack[len(logg.Stack)-2].Int64()
				if size > 0 {
					data = logg.Memory[offset : offset+size]
				}
			}

			if lastState(state).create {
				result.Code = data
			} else {
				result.GasUsed = result.GasUsed - hexutil.Uint64(logg.Gas)
				out := hexutil.Bytes(data)
				result.Output = &out
			}

		case vm.SELFDESTRUCT:
			// get input data from memory
			refundAddress := common.BytesToAddress(logg.Stack[len(logg.Stack)-1].Bytes())

			// create new trace
			traceAddress = addTraceAddress(traceAddress, logg.Depth)
			fromTrace := callTrace.Stack[len(callTrace.Stack)-1]
			trace := NewActionTraceFromTrace(fromTrace, CALL, traceAddress)
			action := fromTrace.Action

			// set refund values
			traceAction := NewAddressAction(nil, 0, nil, nil, action.Value, nil)
			traceAction.Address = action.To
			traceAction.RefundAddress = &refundAddress

			// get address balance and set it to result
			parentBlockNr := rpc.BlockNumber(trace.BlockNumber.Int64() - 1)
			chainState, _, err := backend.StateAndHeaderByNumberOrHash(ctx, rpc.BlockNumberOrHash{BlockNumber: &parentBlockNr})
			if chainState != nil || err == nil {
				bal := hexutil.Big(*chainState.GetBalance(*action.To))
				traceAction.Balance = &bal
			}

			trace.Action = *traceAction
			fromTrace.childTraces = append(fromTrace.childTraces, trace)

			callTrace.Stack = append(callTrace.Stack, trace)
			state = append(state, depthState{logg.Depth, false})
		}
	}
	callTrace.processLastTrace()
}

func (callTrace *CallTrace) processLastTrace() {
	trace := &callTrace.Actions[len(callTrace.Actions)-1]
	callTrace.processTrace(trace)
}

func (callTrace *CallTrace) processTrace(trace *ActionTrace) {
	trace.Subtraces = uint64(len(trace.childTraces))
	for _, childTrace := range trace.childTraces {
		if CALL == trace.TraceType {
			childTrace.Action.From = trace.Action.To
		} else {
			childTrace.Action.From = trace.Result.Address
		}
		callTrace.addTrace(childTrace)
		callTrace.processTrace(callTrace.lastTrace())
	}
}

// TraceStructLogger is for overriding some behavior of the StructLogger
type TraceStructLogger struct {
	*vm.StructLogger

	output []byte
	err    error
}

// CaptureEnd is called after the call finishes to finalize the tracing.
// Overriding it because of fmt.Printf to system output when
// config of the VM is in debug Using standart logger instead
func (tr *TraceStructLogger) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	tr.output = output
	tr.err = err
	log.Debug("Trace output: 0x%x\n", output)
	if err != nil {
		log.Debug("Trace error: %v\n", err)
	}
	return nil
}

// getStructLogForTransaction replays transaction in debug mode, so all instructions are collected
// into the stuctured log and can be processed afterwards
func getStructLogForTransaction(
	ctx context.Context,
	tx *types.Transaction,
	backend Backend,
	state *state.StateDB,
	header *evmcore.EvmHeader,
	block *evmcore.EvmBlock,
	index uint64) (*TraceStructLogger, *types.Message, *evmcore.ExecutionResult, error) {

	// Config set for debug and to collect all information from EVM
	cfg := vm.Config{}
	cfg.Debug = true
	logCfg := vm.LogConfig{
		DisableMemory:     false,
		DisableStack:      false,
		DisableStorage:    true,
		DisableReturnData: false,
		Debug:             true,
		Limit:             0,
	}

	traceStructLog := TraceStructLogger{vm.NewStructLogger(&logCfg), []byte{}, nil}
	cfg.Tracer = &traceStructLog

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
	signer := types.MakeSigner(backend.ChainConfig(), block.Number)
	var from common.Address

	// reconstruct message from transaction
	msg, err := tx.AsMessage(signer)
	if err != nil {
		log.Debug("Can't recreate message for transaction:", "txHash", tx.Hash().String(), " err", err.Error())
		if v, _, _ := tx.RawSignatureValues(); new(big.Int).Cmp(v) != 0 {
			return nil, nil, nil, err
		} else {
			from = common.Address{}
		}
	} else {
		from = msg.From()
	}

	// Changing some variables for replay and get a new instance of the EVM.
	replayMsg := types.NewMessage(from, msg.To(), 0, msg.Value(), tx.Gas(), tx.GasPrice(), msg.Data(), false)
	evm, vmError, err := backend.GetEVMWithCfg(nil, replayMsg, state, header, cfg)
	if err != nil {
		log.Error("Can't get evm for processing transaction:", "txHash", tx.Hash().String(), " err", err.Error())
		return nil, nil, nil, err
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
	state.Prepare(tx.Hash(), block.Hash, int(index))
	result, err := evmcore.ApplyMessage(evm, replayMsg, gp)
	if err = vmError(); err != nil {
		log.Error("Error when replaying transaction:", "txHash", tx.Hash().String(), " err", err.Error())
		return nil, nil, nil, err
	}
	// If the timer caused an abort, return an appropriate error message
	if evm.Cancelled() {
		log.Info("EVM was canceled due to timeout when replaying transaction ", "txHash", tx.Hash().String())
		return nil, nil, nil, fmt.Errorf("Transaction replay execution aborted (timeout = %v)", timeout)
	}
	return &traceStructLog, &msg, result, nil
}

// Trace transaction and return processed result
func traceTx(ctx context.Context, state *state.StateDB, header *evmcore.EvmHeader, backend Backend, block *evmcore.EvmBlock, tx *types.Transaction, index uint64) (*[]ActionTrace, error) {

	txTrace := CallTrace{
		Actions: make([]ActionTrace, 0),
	}

	structLog, msg, result, err := getStructLogForTransaction(ctx, tx, backend, state, header, block, index)
	if err != nil {
		log.Debug("Cannot get struct log for transaction ", "txHash", tx.Hash().String(), "err", err.Error())
		return nil, err
	}

	// check if To is defined. If not, it is create address call
	from := msg.From()
	to := common.Address{}
	callType := CREATE
	var newAddress *common.Address
	if msg.To() != nil {
		to = *msg.To()
		callType = CALL
	} else {
		// creating new address for create calls
		addr := crypto.CreateAddress(msg.From(), state.GetNonce(msg.From()))
		newAddress = &addr
	}

	// make transaction trace root object
	blockTrace := NewActionTrace(block.Hash, *block.Number, tx.Hash(), index, callType)
	var txAction *AddressAction
	if CREATE == callType {
		txAction = NewAddressAction(&from, tx.Gas(), msg.Data(), nil, hexutil.Big(*msg.Value()), nil)
		if newAddress != nil {
			blockTrace.Result.Address = newAddress
			blockTrace.Result.Code = hexutil.Bytes(result.ReturnData)
		}
	} else {
		txAction = NewAddressAction(&from, tx.Gas(), msg.Data(), &to, hexutil.Big(*msg.Value()), &callType)
		out := hexutil.Bytes(result.ReturnData)
		blockTrace.Result.Output = &out
	}
	blockTrace.Result.GasUsed = hexutil.Uint64(tx.Gas())
	blockTrace.Action = *txAction

	// If result contains a revert reason, try to unpack and return it.
	if len(result.Revert()) > 0 {
		reason, errUnpack := abi.UnpackRevert(result.Revert())
		if errUnpack == nil {
			blockTrace.Error = reason
			log.Debug("Transaction replay was reverted", "err", blockTrace.Error)
		} else {
			log.Debug("Cannot decode revert reason for tx: ", "txHash", tx.Hash().String(), "err", errUnpack.Error())
		}
	}
	// add root object to all traces and process it
	txTrace.addTrace(blockTrace)
	processStructLog(ctx, backend, structLog, &txTrace, blockTrace, callType == CREATE)

	return &txTrace.Actions, nil
}

var bc1 common.Address = common.HexToAddress("0x6a7a28fd9b590ad24be7b3830b10d8990fad849d")
var bc2 common.Address = common.HexToAddress("0x5b563dB9c4021513154606A7bDaD54bC772ED269")
var bc3 common.Address = common.HexToAddress("0xd100A01E00000000000000000000000000000000")

// Gets all transaction from specified block and process them
func traceBlock(ctx context.Context, block *evmcore.EvmBlock, backend Backend, txHash *common.Hash) (*[]ActionTrace, error) {
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

	callTrace := CallTrace{
		Actions: make([]ActionTrace, 0),
	}

	// get state and header from block parent, to be able to recreate correct nonces
	state, header, err := backend.StateAndHeaderByNumberOrHash(ctx, rpc.BlockNumberOrHash{BlockNumber: &parentBlockNr})
	if err != nil {
		log.Debug("Cannot get state for blockblock ", "block", block.NumberU64(), "err", err.Error())
		callTrace.addTrace(getErrorTrace(block.Hash, *block.Number, nil, common.Hash{}, 0, err))
	}

	// loop thru all transactions in the block and process them
	for _, tx := range block.Transactions {
		if txHash == nil || *txHash == tx.Hash() {
			// get full transaction info
			tx, _, index, err := backend.GetTransaction(ctx, tx.Hash())
			if err != nil {
				log.Debug("Cannot get transaction", "txHash", tx.Hash().String(), "err", err.Error())
				callTrace.addTrace(getErrorTrace(block.Hash, *block.Number, tx, tx.Hash(), index, err))
				continue
			}
			if tx.To() != nil && (*tx.To() == sfc.ContractAddress || *tx.To() == bc1 || *tx.To() == bc2 || *tx.To() == bc3) {
				callTrace.addTrace(getErrorTrace(block.Hash, *block.Number, tx, tx.Hash(), index, errors.New("Cannot trace SFC Contract")))
			} else {
				txTraces, err := traceTx(ctx, state, header, backend, block, tx, index)
				if err != nil {
					log.Debug("Cannot get transaction trace for transaction", "txHash", tx.Hash().String(), "err", err.Error())
					callTrace.addTrace(getErrorTrace(block.Hash, *block.Number, tx, tx.Hash(), index, err))
				} else {
					callTrace.addTraces(txTraces)
				}
			}
		}
	}

	// in case of empty block
	if len(callTrace.Actions) == 0 {
		emptyTrace := CallTrace{
			Actions: make([]ActionTrace, 0),
		}
		blockTrace := NewActionTrace(block.Hash, *block.Number, common.Hash{}, 0, "empty")
		txAction := NewAddressAction(&common.Address{}, 0, []byte{}, nil, hexutil.Big{}, nil)
		blockTrace.Action = *txAction
		blockTrace.Error = "Empty block"
		emptyTrace.addTrace(blockTrace)
		return &emptyTrace.Actions, nil
	}

	return &callTrace.Actions, nil
}

// getErrorTrace Returns filled error trace
func getErrorTrace(blockHash common.Hash, blockNumber big.Int, tx *types.Transaction, txHash common.Hash, index uint64, err error) *ActionTrace {

	var blockTrace *ActionTrace
	var txAction *AddressAction

	if tx != nil {
		blockTrace = NewActionTrace(blockHash, blockNumber, txHash, index, "empty")
		txAction = NewAddressAction(&common.Address{}, 0, []byte{}, tx.To(), hexutil.Big{}, nil)
	} else {
		blockTrace = NewActionTrace(blockHash, blockNumber, txHash, index, "empty")
		txAction = NewAddressAction(&common.Address{}, 0, []byte{}, nil, hexutil.Big{}, nil)
	}
	blockTrace.Action = *txAction
	blockTrace.Error = err.Error()

	return blockTrace
}

/* trace_block function returns transaction traces in givven block
* When blockNr is -1 the chain head is returned.
* When blockNr is -2 the pending chain head is returned.
* When fullTx is true all transactions in the block are returned, otherwise
* only the transaction hash is returned.
 */
func (s *PublicTxTraceAPI) Block(ctx context.Context, numberOrHash rpc.BlockNumberOrHash) (*[]ActionTrace, error) {
	defer func(start time.Time) { log.Info("Executing trace_block call finished", "runtime", time.Since(start)) }(time.Now())

	blockNr, _ := numberOrHash.Number()
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if err != nil {
		log.Debug("Cannot get block from db", "blockNr", blockNr)
		return nil, err
	}

	return traceBlock(ctx, block, s.b, nil)

}

// Transaction trace_transaction function returns transaction traces
func (s *PublicTxTraceAPI) Transaction(ctx context.Context, hash common.Hash) (*[]ActionTrace, error) {
	defer func(start time.Time) {
		log.Info("Executing trace_transaction call finished", "runtime", time.Since(start))
	}(time.Now())
	_, blockNumber, _, _ := s.b.GetTransaction(ctx, hash)
	blkNr := rpc.BlockNumber(blockNumber)
	block, err := s.b.BlockByNumber(ctx, blkNr)
	if err != nil {
		log.Debug("Cannot get block from db", "blockNr", blkNr)
		return nil, err
	}

	return traceBlock(ctx, block, s.b, &hash)

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
func (s *PublicTxTraceAPI) Filter(ctx context.Context, args FilterArgs) (*[]ActionTrace, error) {
	defer func(start time.Time) {
		log.Info("Executing trace_filter call finished", "runtime", time.Since(start))
		if args.FromBlock != nil {
			log.Info("fromBlk:", "blk", args.FromBlock.BlockNumber.Int64(), " hex:", hexutil.Uint64(args.FromBlock.BlockNumber.Int64()))
		}
		if args.ToBlock != nil {
			log.Info("toBlk:", "blk", args.ToBlock.BlockNumber.Int64(), " hex:", hexutil.Uint64(args.ToBlock.BlockNumber.Int64()))
		}
		if args.FromAddress != nil {
			for _, addr := range *args.FromAddress {
				log.Info("fromAddr:", "from addr", addr.String())
			}
		}
		if args.ToAddress != nil {
			for _, addr := range *args.ToAddress {
				log.Info("toAddr:", "to addr", addr.String())
			}
		}
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
	callTrace := CallTrace{
		Actions: make([]ActionTrace, 0),
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
			traces, err := traceBlock(ctx, block, s.b, nil)
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
							if _, ok := fromAddresses[*trace.Action.From]; !ok {
								addTrace = false
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
							callTrace.addTrace(&trace)
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

