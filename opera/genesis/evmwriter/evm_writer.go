package evmwriter

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/opera/genesis/driver"
)

// ContractAddress is the EvmWriter pre-compiled contract address
var ContractAddress = common.HexToAddress("0xd100ec0000000000000000000000000000000000")

type PreCompiledContract struct{}

func (_ PreCompiledContract) Run(stateDB vm.StateDB, ctx vm.Context, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	if caller != driver.ContractAddress {
		return nil, 0, vm.ErrExecutionReverted
	}
	if len(input) < 4 {
		return nil, 0, vm.ErrExecutionReverted
	}
	if bytes.Equal(input[:4], []byte{0xe3, 0x04, 0x43, 0xbc}) {
		input = input[4:]
		// setBalance
		if suppliedGas < params.CallValueTransferGas {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= params.CallValueTransferGas
		if len(input) != 64 {
			return nil, 0, vm.ErrExecutionReverted
		}

		acc := common.BytesToAddress(input[12:32])
		input = input[32:]
		value := new(big.Int).SetBytes(input[:32])

		if acc == ctx.Origin {
			// Origin balance shouldn't decrease during his transaction
			return nil, 0, vm.ErrExecutionReverted
		}

		balance := stateDB.GetBalance(acc)
		if balance.Cmp(value) >= 0 {
			diff := new(big.Int).Sub(balance, value)
			stateDB.SubBalance(acc, diff)
		} else {
			diff := new(big.Int).Sub(value, balance)
			stateDB.AddBalance(acc, diff)
		}
	} else if bytes.Equal(input[:4], []byte{0xd6, 0xa0, 0xc7, 0xaf}) {
		input = input[4:]
		// copyCode
		if suppliedGas < params.CreateGas {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= params.CreateGas
		if len(input) != 64 {
			return nil, 0, vm.ErrExecutionReverted
		}

		accTo := common.BytesToAddress(input[12:32])
		input = input[32:]
		accFrom := common.BytesToAddress(input[12:32])

		code := stateDB.GetCode(accFrom)
		if code == nil {
			code = []byte{}
		}
		cost := uint64(len(code)) * (params.CreateDataGas + params.MemoryGas)
		if suppliedGas < cost {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= cost
		if accTo != accFrom {
			stateDB.SetCode(accTo, code)
		}
	} else if bytes.Equal(input[:4], []byte{0x07, 0x69, 0x0b, 0x2a}) {
		input = input[4:]
		// swapCode
		cost := 2 * params.CreateGas
		if suppliedGas < cost {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= cost
		if len(input) != 64 {
			return nil, 0, vm.ErrExecutionReverted
		}

		acc0 := common.BytesToAddress(input[12:32])
		input = input[32:]
		acc1 := common.BytesToAddress(input[12:32])
		code0 := stateDB.GetCode(acc0)
		if code0 == nil {
			code0 = []byte{}
		}
		code1 := stateDB.GetCode(acc1)
		if code1 == nil {
			code1 = []byte{}
		}
		cost0 := uint64(len(code0)) * (params.CreateDataGas + params.MemoryGas)
		cost1 := uint64(len(code1)) * (params.CreateDataGas + params.MemoryGas)
		cost = (cost0 + cost1) / 2 // 50% discount because trie size won't increase after pruning
		if suppliedGas < cost {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= cost
		if acc0 != acc1 {
			stateDB.SetCode(acc0, code1)
			stateDB.SetCode(acc1, code0)
		}
	} else if bytes.Equal(input[:4], []byte{0x39, 0xe5, 0x03, 0xab}) {
		input = input[4:]
		// setStorage
		if suppliedGas < params.SstoreInitGasEIP2200 {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= params.SstoreInitGasEIP2200
		if len(input) != 96 {
			return nil, 0, vm.ErrExecutionReverted
		}
		acc := common.BytesToAddress(input[12:32])
		input = input[32:]
		key := common.BytesToHash(input[:32])
		input = input[32:]
		value := common.BytesToHash(input[:32])

		stateDB.SetState(acc, key, value)
	} else {
		return nil, 0, vm.ErrExecutionReverted
	}
	return nil, suppliedGas, nil
}
