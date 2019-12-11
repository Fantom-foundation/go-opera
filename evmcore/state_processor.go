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
	"math/big"

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
func (p *StateProcessor) Process(block *EvmBlock, statedb *state.StateDB, cfg vm.Config, strict bool) (types.Receipts, []*types.Log, uint64, *big.Int, []uint, error) {
	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit)
		skipped  = make([]uint, 0, len(block.Transactions))
		totalFee = new(big.Int)
	)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions {
		statedb.Prepare(tx.Hash(), block.Hash, i)
		receipt, _, fee, skip, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, block.Header(), tx, usedGas, cfg, strict)
		if !strict && (skip || err != nil) {
			skipped = append(skipped, uint(i))
			continue
		}
		totalFee.Add(totalFee, fee)
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}

	return receipts, allLogs, *usedGas, totalFee, skipped, nil
}

func TransactionPreCheck(statedb *state.StateDB, msg types.Message, tx *types.Transaction) error {
	nonce := statedb.GetNonce(msg.From())
	if nonce < msg.Nonce() {
		return ErrNonceTooHigh
	} else if nonce > msg.Nonce() {
		return ErrNonceTooLow
	}

	balance := statedb.GetBalance(msg.From())
	if balance.Cmp(tx.Cost()) < 0 {
		return ErrInsufficientFunds
	}

	return nil
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(
	config *params.ChainConfig,
	bc DummyChain,
	author *common.Address,
	gp *GasPool,
	statedb *state.StateDB,
	header *EvmHeader,
	tx *types.Transaction,
	usedGas *uint64,
	cfg vm.Config,
	strict bool,
) (
	*types.Receipt,
	uint64,
	*big.Int,
	bool,
	error,
) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, 0, common.Big0, false, err
	}

	if !strict {
		// the reason why we check here is to avoid spending sender's gas in a case if tx failed (due to insufficient balance or wrong nonce)
		// the transaction has already spent validator's gas power
		err = TransactionPreCheck(statedb, msg, tx)
		if err != nil {
			return nil, 0, common.Big0, true, err
		}
	}

	// Create a new context to be used in the EVM environment
	context := NewEVMContext(msg, header, bc, author)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewEVM(context, statedb, config, cfg)
	// Apply the transaction to the current state (included in the env)
	_, gas, fee, failed, err := ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, 0, fee, false, err
	}
	// Update the state with pending changes
	var root []byte
	if config.IsByzantium(header.Number) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes()
	}
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
	}
	// Set the receipt logs
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.BlockHash = statedb.BlockHash()
	receipt.BlockNumber = header.Number
	receipt.TransactionIndex = uint(statedb.TxIndex())

	return receipt, gas, fee, false, err
}
