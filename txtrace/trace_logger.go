package txtrace

import (
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/Fantom-foundation/go-opera/gossip/txtrace"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

// TraceStructLogger is a transaction trace creator
type TraceStructLogger struct {
	store       *txtrace.Store
	from        *common.Address
	to          *common.Address
	newAddress  *common.Address
	blockHash   common.Hash
	tx          common.Hash
	txIndex     uint
	blockNumber big.Int
	value       big.Int

	gasUsed      uint64
	rootTrace    *CallTrace
	inputData    []byte
	state        []depthState
	traceAddress []uint32
	stack        []*big.Int
	reverted     bool
	output       []byte
	err          error
}

// NewTraceStructLogger creates new instance of trace creator
func NewTraceStructLogger(store *txtrace.Store) *TraceStructLogger {
	traceStructLogger := TraceStructLogger{
		store: store,
		stack: make([]*big.Int, 30),
	}
	return &traceStructLogger
}

// CaptureStart implements the tracer interface to initialize the tracing operation.
func (tr *TraceStructLogger) CaptureStart(from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) error {
	// Create main trace holder
	txTrace := CallTrace{
		Actions: make([]ActionTrace, 0),
	}

	// Check if To is defined. If not, it is create address call
	callType := CREATE
	var newAddress *common.Address
	if tr.to != nil {
		callType = CALL
	} else {
		newAddress = &to
	}

	// Store input data
	tr.inputData = input
	if gas == 0 && tr.gasUsed != 0 {
		gas = tr.gasUsed
	}

	// Make transaction trace root object
	blockTrace := NewActionTrace(tr.blockHash, tr.blockNumber, tr.tx, uint64(tr.txIndex), callType)
	var txAction *AddressAction
	if CREATE == callType {
		txAction = NewAddressAction(tr.from, gas, tr.inputData, nil, hexutil.Big(tr.value), nil)
		if newAddress != nil {
			blockTrace.Result.Address = newAddress
			blockTrace.Result.Code = hexutil.Bytes(tr.output)
		}
	} else {
		txAction = NewAddressAction(tr.from, gas, tr.inputData, tr.to, hexutil.Big(tr.value), &callType)
		out := hexutil.Bytes(tr.output)
		blockTrace.Result.Output = &out
	}
	blockTrace.Action = *txAction

	// Add root object into Tracer
	txTrace.AddTrace(blockTrace)
	tr.rootTrace = &txTrace

	// Init all needed variables
	tr.state = []depthState{{0, create}}
	tr.traceAddress = make([]uint32, 0)
	tr.rootTrace.Stack = append(tr.rootTrace.Stack, &tr.rootTrace.Actions[len(tr.rootTrace.Actions)-1])

	return nil
}

// stackPosFromEnd returns object from stack at givven position from end of stack
func stackPosFromEnd(stackData []uint256.Int, pos int) *big.Int {
	return new(big.Int).Set(stackData[len(stackData)-1-pos].ToBig())
}

// CaptureState implements creating of traces based on getting opCodes from evm during contract processing
func (tr *TraceStructLogger) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, rStack *vm.ReturnStack, rData []byte, contract *vm.Contract, depth int, err error) error {

	// When going back from inner call
	if lastState(tr.state).level == depth {
		result := tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1].Result
		if lastState(tr.state).create && result != nil {
			if len(stack.Data()) > 0 {
				addr := common.BytesToAddress(stackPosFromEnd(stack.Data(), 0).Bytes())
				result.Address = &addr
				result.GasUsed = hexutil.Uint64(gas)
			}
		}
		tr.traceAddress = removeTraceAddressLevel(tr.traceAddress, depth)
		tr.state = tr.state[:len(tr.state)-1]
		tr.rootTrace.Stack = tr.rootTrace.Stack[:len(tr.rootTrace.Stack)-1]
	}

	// Match processed instruction and create trace based on it
	switch op {
	case vm.CREATE, vm.CREATE2:
		tr.traceAddress = addTraceAddress(tr.traceAddress, depth)
		fromTrace := tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1]

		// Get input data from memory
		offset := stackPosFromEnd(stack.Data(), 1).Int64()
		inputSize := stackPosFromEnd(stack.Data(), 2).Int64()
		var input []byte
		if inputSize > 0 {
			input = make([]byte, inputSize)
			copy(input, memory.Data()[offset:offset+inputSize])
		}

		// Create new trace
		trace := NewActionTraceFromTrace(fromTrace, CREATE, tr.traceAddress)
		traceAction := NewAddressAction(nil, gas, input, nil, fromTrace.Action.Value, nil)
		trace.Action = *traceAction
		trace.Result.GasUsed = hexutil.Uint64(gas)
		fromTrace.childTraces = append(fromTrace.childTraces, trace)
		tr.rootTrace.Stack = append(tr.rootTrace.Stack, trace)
		tr.state = append(tr.state, depthState{depth, true})

	case vm.CALL, vm.CALLCODE, vm.DELEGATECALL, vm.STATICCALL:
		var (
			inOffset, inSize   int64
			retOffset, retSize uint64
			input              []byte
			value              = big.NewInt(0)
		)

		if vm.DELEGATECALL == op || vm.STATICCALL == op {
			inOffset = stackPosFromEnd(stack.Data(), 2).Int64()
			inSize = stackPosFromEnd(stack.Data(), 3).Int64()
			retOffset = stackPosFromEnd(stack.Data(), 4).Uint64()
			retSize = stackPosFromEnd(stack.Data(), 5).Uint64()
		} else {
			inOffset = stackPosFromEnd(stack.Data(), 3).Int64()
			inSize = stackPosFromEnd(stack.Data(), 4).Int64()
			retOffset = stackPosFromEnd(stack.Data(), 5).Uint64()
			retSize = stackPosFromEnd(stack.Data(), 6).Uint64()
			// only CALL and CALLCODE need `value` field
			value = stackPosFromEnd(stack.Data(), 2)
		}
		if inSize > 0 {
			input = make([]byte, inSize)
			copy(input, memory.Data()[inOffset:inOffset+inSize])
		}
		tr.traceAddress = addTraceAddress(tr.traceAddress, depth)
		fromTrace := tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1]
		// create new trace
		trace := NewActionTraceFromTrace(fromTrace, CALL, tr.traceAddress)
		action := fromTrace.Action
		addr := common.BytesToAddress(stackPosFromEnd(stack.Data(), 1).Bytes())
		callType := strings.ToLower(op.String())
		traceAction := NewAddressAction(action.To, gas, input, &addr, hexutil.Big(*value), &callType)
		trace.Action = *traceAction
		fromTrace.childTraces = append(fromTrace.childTraces, trace)
		trace.Result.RetOffset = retOffset
		trace.Result.RetSize = retSize
		tr.rootTrace.Stack = append(tr.rootTrace.Stack, trace)
		tr.state = append(tr.state, depthState{depth, false})

	case vm.RETURN, vm.STOP:
		if tr.reverted {
			tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1].Result = nil
			tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1].Error = "Reverted"
		} else {
			result := tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1].Result
			var data []byte

			if vm.STOP != op {
				offset := stackPosFromEnd(stack.Data(), 0).Int64()
				size := stackPosFromEnd(stack.Data(), 1).Int64()
				if size > 0 {
					data = make([]byte, size)
					copy(data, memory.Data()[offset:offset+size])
				}
			}

			if lastState(tr.state).create {
				result.Code = data
			} else {
				result.GasUsed = hexutil.Uint64(gas)
				out := hexutil.Bytes(data)
				result.Output = &out
			}
		}

	case vm.REVERT:
		tr.reverted = true
		tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1].Result = nil
		tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1].Error = "Reverted"

	case vm.SELFDESTRUCT:
		tr.traceAddress = addTraceAddress(tr.traceAddress, depth)
		fromTrace := tr.rootTrace.Stack[len(tr.rootTrace.Stack)-1]
		trace := NewActionTraceFromTrace(fromTrace, SELFDESTRUCT, tr.traceAddress)
		action := fromTrace.Action

		traceAction := NewAddressAction(nil, 0, nil, nil, action.Value, nil)
		traceAction.Address = action.To
		// set refund values
		refundAddress := common.BytesToAddress(stackPosFromEnd(stack.Data(), 0).Bytes())
		traceAction.RefundAddress = &refundAddress
		// Add the `to` field so that subsequent CALL traces can refer to the `fromTrace.To` field as the value of the `from` field
		traceAction.To = &refundAddress
		// Add `balance` field for convenient usage
		traceAction.Balance = &traceAction.Value
		trace.Action = *traceAction
		fromTrace.childTraces = append(fromTrace.childTraces, trace)

		tr.rootTrace.Stack = append(tr.rootTrace.Stack, trace)
		tr.state = append(tr.state, depthState{depth, false})
	}

	return nil
}

// CaptureEnd is called after the call finishes to finalize the tracing.
func (tr *TraceStructLogger) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	log.Debug("TraceStructLogger capture END", "tx hash", tr.tx.String(), "duration", t, "gasUsed", gasUsed)
	if gasUsed > 0 {
		if tr.rootTrace.Actions[0].Result != nil {
			tr.rootTrace.Actions[0].Result.GasUsed = hexutil.Uint64(gasUsed)
		}
		tr.rootTrace.lastTrace().Action.Gas = hexutil.Uint64(gasUsed)

		tr.gasUsed = gasUsed
	}
	tr.output = output
	return nil
}

// CaptureFault implements the Tracer interface to trace an execution fault
// while running an opcode.
func (tr *TraceStructLogger) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, rStack *vm.ReturnStack, contract *vm.Contract, depth int, err error) error {
	return nil
}

// Reset function to be able to reuse logger
func (tr *TraceStructLogger) reset() {
	tr.to = nil
	tr.from = nil
	tr.inputData = nil
	tr.rootTrace = nil
	tr.reverted = false
}

// SetTx basic setter
func (tr *TraceStructLogger) SetTx(tx common.Hash) {
	tr.tx = tx
}

// SetFrom basic setter
func (tr *TraceStructLogger) SetFrom(from common.Address) {
	tr.from = &from
}

// SetTo basic setter
func (tr *TraceStructLogger) SetTo(to *common.Address) {
	tr.to = to
}

// SetValue basic setter
func (tr *TraceStructLogger) SetValue(value big.Int) {
	tr.value = value
}

// SetBlockHash basic setter
func (tr *TraceStructLogger) SetBlockHash(blockHash common.Hash) {
	tr.blockHash = blockHash
}

// SetBlockNumber basic setter
func (tr *TraceStructLogger) SetBlockNumber(blockNumber *big.Int) {
	tr.blockNumber = *blockNumber
}

// SetTxIndex basic setter
func (tr *TraceStructLogger) SetTxIndex(txIndex uint) {
	tr.txIndex = txIndex
}

// SetNewAddress basic setter
func (tr *TraceStructLogger) SetNewAddress(newAddress common.Address) {
	tr.newAddress = &newAddress
}

// SetGasUsed basic setter
func (tr *TraceStructLogger) SetGasUsed(gasUsed uint64) {
	tr.gasUsed = gasUsed
}

// ProcessTx finalizes trace proces and stores result into key-value persistant store
func (tr *TraceStructLogger) ProcessTx() {
	if tr.rootTrace != nil {
		tr.rootTrace.lastTrace().Action.Gas = hexutil.Uint64(tr.gasUsed)
		if tr.rootTrace.lastTrace().Result != nil {
			tr.rootTrace.lastTrace().Result.GasUsed = hexutil.Uint64(tr.gasUsed)
		}
		tr.rootTrace.processLastTrace()
	}
}

// GetTraceActions returns action traces after recording evm process
func (tr *TraceStructLogger) SaveTrace() {
	if tr.rootTrace == nil {
		tr.rootTrace = &CallTrace{}
		tr.rootTrace.AddTrace(GetErrorTrace(tr.blockHash, tr.blockNumber, tr.to, tr.tx, tr.gasUsed, tr.err))

	}

	//if tr.store != nil && tr.rootTrace != nil {
	if tr.store != nil {
		// Convert trace objects to json byte array and save it
		tracesBytes, _ := json.Marshal(tr.rootTrace.Actions)
		tr.store.SetTxTrace(tr.tx, tracesBytes)
		log.Debug("Added tx trace", "txHash", tr.tx.String())
	}
	tr.reset()
}

// GetTraceActions returns action traces after recording evm process
func (tr *TraceStructLogger) GetTraceActions() *[]ActionTrace {
	if tr.rootTrace != nil {
		return &tr.rootTrace.Actions
	}
	empty := make([]ActionTrace, 0)
	return &empty
}

// CallTrace is struct for holding tracing results
type CallTrace struct {
	Actions []ActionTrace  `json:"result"`
	Stack   []*ActionTrace `json:"-"`
}

// AddTrace Append trace to call trace list
func (callTrace *CallTrace) AddTrace(blockTrace *ActionTrace) {
	if callTrace.Actions == nil {
		callTrace.Actions = make([]ActionTrace, 0)
	}
	callTrace.Actions = append(callTrace.Actions, *blockTrace)
}

// AddTraces Append traces to call trace list
func (callTrace *CallTrace) AddTraces(traces *[]ActionTrace, traceIndex *[]hexutil.Uint) {
	for _, trace := range *traces {
		if traceIndex == nil || equalContent(traceIndex, trace.TraceAddress) {
			callTrace.AddTrace(&trace)
		}
	}
}

// equalContent tells whether index and traceIndex are the same
func equalContent(index *[]hexutil.Uint, traceIndex []uint32) bool {
	if len(*index) != len(traceIndex) {
		return false
	}
	for i, v := range *index {
		if uint32(v) != traceIndex[i] {
			return false
		}
	}
	return true
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
		TraceAddress:        make([]uint32, 0),
		Result:              &TraceActionResult{},
	}
}

// NewActionTraceFromTrace creates new instance of type ActionTrace
// based on another trace
func NewActionTraceFromTrace(actionTrace *ActionTrace, tType string, traceAddress []uint32) *ActionTrace {
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
	CALL         = "call"
	CREATE       = "create"
	SELFDESTRUCT = "suicide"
)

// ActionTrace represents single interaction with blockchain
type ActionTrace struct {
	childTraces         []*ActionTrace     `json:"-"`
	Action              AddressAction      `json:"action"`
	BlockHash           common.Hash        `json:"blockHash"`
	BlockNumber         big.Int            `json:"blockNumber"`
	Result              *TraceActionResult `json:"result,omitempty"`
	Error               string             `json:"error,omitempty"`
	Subtraces           uint64             `json:"subtraces"`
	TraceAddress        []uint32           `json:"traceAddress"`
	TransactionHash     common.Hash        `json:"transactionHash"`
	TransactionPosition uint64             `json:"transactionPosition"`
	TraceType           string             `json:"type"`
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
	RefundAddress *common.Address `json:"refundAddress,omitempty"`
	Balance       *hexutil.Big    `json:"balance,omitempty"`
}

// TraceActionResult holds information related to result of the
// processed transaction
type TraceActionResult struct {
	GasUsed   hexutil.Uint64  `json:"gasUsed"`
	Output    *hexutil.Bytes  `json:"output,omitempty"`
	Code      hexutil.Bytes   `json:"code,omitempty"`
	Address   *common.Address `json:"address,omitempty"`
	RetOffset uint64          `json:"-"`
	RetSize   uint64          `json:"-"`
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
func addTraceAddress(traceAddress []uint32, depth int) []uint32 {
	index := depth - 1
	result := make([]uint32, len(traceAddress))
	copy(result, traceAddress)
	if len(result) <= index {
		result = append(result, 0)
	} else {
		result[index]++
	}
	return result
}

// removes trace address based on depth of process
func removeTraceAddressLevel(traceAddress []uint32, depth int) []uint32 {
	if len(traceAddress) > depth {
		result := make([]uint32, len(traceAddress))
		copy(result, traceAddress)

		result = result[:len(result)-1]
		return result
	}
	return traceAddress
}

// processLastTrace initiates final information distribution
// accros result traces
func (callTrace *CallTrace) processLastTrace() {
	trace := &callTrace.Actions[len(callTrace.Actions)-1]
	callTrace.processTrace(trace)
}

// processTrace goes thru all trace results and sets info
func (callTrace *CallTrace) processTrace(trace *ActionTrace) {
	trace.Subtraces = uint64(len(trace.childTraces))
	for _, childTrace := range trace.childTraces {
		if CALL == trace.TraceType {
			childTrace.Action.From = trace.Action.To
		} else {
			if trace.Result != nil {
				childTrace.Action.From = trace.Result.Address
			}
		}

		if childTrace.Result != nil {
			if trace.Action.Gas > childTrace.Result.GasUsed {
				childTrace.Action.Gas = trace.Action.Gas - childTrace.Result.GasUsed
			} else {
				childTrace.Action.Gas = childTrace.Result.GasUsed
			}
		}
		callTrace.AddTrace(childTrace)
		callTrace.processTrace(callTrace.lastTrace())
	}
}

// GetErrorTrace constructs filled error trace
func GetErrorTrace(blockHash common.Hash, blockNumber big.Int, to *common.Address, txHash common.Hash, index uint64, err error) *ActionTrace {

	var blockTrace *ActionTrace
	var txAction *AddressAction

	if to != nil {
		blockTrace = NewActionTrace(blockHash, blockNumber, txHash, index, "empty")
		txAction = NewAddressAction(&common.Address{}, 0, []byte{}, to, hexutil.Big{}, nil)
	} else {
		blockTrace = NewActionTrace(blockHash, blockNumber, txHash, index, "empty")
		txAction = NewAddressAction(&common.Address{}, 0, []byte{}, nil, hexutil.Big{}, nil)
	}
	blockTrace.Action = *txAction
	blockTrace.Result = nil
	if err != nil {
		blockTrace.Error = err.Error()
	} else {
		blockTrace.Error = "Reverted"
	}
	return blockTrace
}
