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

package evmcore

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     DummyChain          // Canonical block chain
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc DummyChain) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(
	block *EvmBlock, statedb *state.StateDB, cfg vm.Config, usedGas *uint64, internal bool, onNewLog func(*types.Log, *state.StateDB),
) (
	receipts types.Receipts, allLogs []*types.Log, skipped []uint32, err error,
) {
	skipped = make([]uint32, 0, len(block.Transactions))
	var (
		gp           = new(GasPool).AddGas(block.GasLimit)
		receipt      *types.Receipt
		skip         bool
		header       = block.Header()
		blockContext = NewEVMBlockContext(header, p.bc, nil)
		vmenv        = vm.NewEVM(blockContext, vm.TxContext{}, statedb, p.config, cfg)
	)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions {
		var msg types.Message
		if !internal {
			msg, err = tx.AsMessage(types.MakeSigner(p.config, header.Number))
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			msg = types.NewMessage(common.Address{}, tx.To(), tx.Nonce(), tx.Value(), tx.Gas(), tx.GasPrice(), tx.Data(), tx.AccessList(), false)
		}

		statedb.Prepare(tx.Hash(), block.Hash, i)
		receipt, _, skip, err = applyTransaction(msg, p.config, gp, statedb, block.Header(), tx, usedGas, vmenv, onNewLog)
		if skip {
			skipped = append(skipped, uint32(i))
			err = nil
			continue
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("could not apply tx %d [%v]: %w", i, tx.Hash().Hex(), err)
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	return
}

func applyTransaction(
	msg types.Message,
	config *params.ChainConfig,
	gp *GasPool,
	statedb *state.StateDB,
	header *EvmHeader,
	tx *types.Transaction,
	usedGas *uint64,
	evm *vm.EVM,
	onNewLog func(*types.Log, *state.StateDB),
) (
	*types.Receipt,
	uint64,
	bool,
	error,
) {
	// Create a new context to be used in the EVM environment.
	txContext := NewEVMTxContext(msg)
	evm.Reset(txContext, statedb)

	// Apply the transaction to the current state (included in the env).
	result, err := ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, 0, result == nil, err
	}
	// Notify about logs with potential state changes
	logs := statedb.GetLogs(tx.Hash())
	for _, l := range logs {
		onNewLog(l, statedb)
	}

	// Update the state with pending changes.
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	}
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used
	// by the tx.
	receipt := &types.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: *usedGas}
	if result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = result.UsedGas

	// If the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	}

	// Set the receipt logs.
	receipt.Logs = logs
	receipt.BlockHash = statedb.BlockHash()
	receipt.BlockNumber = header.Number
	receipt.TransactionIndex = uint(statedb.TxIndex())
	return receipt, result.UsedGas, false, err
}
