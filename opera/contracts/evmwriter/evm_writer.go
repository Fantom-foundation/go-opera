package evmwriter

import (
	"bytes"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/opera/contracts/driver"
)

var (
	// ContractAddress is the EvmWriter pre-compiled contract address
	ContractAddress = common.HexToAddress("0xd100ec0000000000000000000000000000000000")
	// ContractABI is the input ABI used to generate the binding from
	ContractABI string = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"AdvanceEpochs\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"diff\",\"type\":\"bytes\"}],\"name\":\"UpdateNetworkRules\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"UpdateNetworkVersion\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"UpdateValidatorPubkey\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"weight\",\"type\":\"uint256\"}],\"name\":\"UpdateValidatorWeight\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"backend\",\"type\":\"address\"}],\"name\":\"UpdatedBackend\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_backend\",\"type\":\"address\"}],\"name\":\"setBackend\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_backend\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_evmWriterAddress\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"setBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"copyCode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"with\",\"type\":\"address\"}],\"name\":\"swapCode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"setStorage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"diff\",\"type\":\"uint256\"}],\"name\":\"incNonce\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"diff\",\"type\":\"bytes\"}],\"name\":\"updateNetworkRules\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"updateNetworkVersion\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"advanceEpochs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"updateValidatorWeight\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"updateValidatorPubkey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_auth\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"setGenesisValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupFromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupEndTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewards\",\"type\":\"uint256\"}],\"name\":\"setGenesisDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"deactivateValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nextValidatorIDs\",\"type\":\"uint256[]\"}],\"name\":\"sealEpochValidators\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"}],\"name\":\"sealEpoch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"usedGas\",\"type\":\"uint256\"}],\"name\":\"sealEpochV1\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

var (
	setBalanceMethodID []byte
	copyCodeMethodID   []byte
	swapCodeMethodID   []byte
	setStorageMethodID []byte
	incNonceMethodID   []byte
)

func init() {
	abi, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		panic(err)
	}

	for name, constID := range map[string]*[]byte{
		"setBalance": &setBalanceMethodID,
		"copyCode":   &copyCodeMethodID,
		"swapCode":   &swapCodeMethodID,
		"setStorage": &setStorageMethodID,
		"incNonce":   &incNonceMethodID,
	} {
		method, exist := abi.Methods[name]
		if !exist {
			panic("unknown EvmWriter method")
		}

		*constID = make([]byte, len(method.ID))
		copy(*constID, method.ID)
	}
}

type PreCompiledContract struct{}

func (_ PreCompiledContract) Run(stateDB vm.StateDB, _ vm.BlockContext, txCtx vm.TxContext, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	if caller != driver.ContractAddress {
		return nil, 0, vm.ErrExecutionReverted
	}
	if len(input) < 4 {
		return nil, 0, vm.ErrExecutionReverted
	}
	if bytes.Equal(input[:4], setBalanceMethodID) {
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

		if acc == txCtx.Origin {
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
	} else if bytes.Equal(input[:4], copyCodeMethodID) {
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
	} else if bytes.Equal(input[:4], swapCodeMethodID) {
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
	} else if bytes.Equal(input[:4], setStorageMethodID) {
		input = input[4:]
		// setStorage
		if suppliedGas < params.SstoreSetGasEIP2200 {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= params.SstoreSetGasEIP2200
		if len(input) != 96 {
			return nil, 0, vm.ErrExecutionReverted
		}
		acc := common.BytesToAddress(input[12:32])
		input = input[32:]
		key := common.BytesToHash(input[:32])
		input = input[32:]
		value := common.BytesToHash(input[:32])

		stateDB.SetState(acc, key, value)
	} else if bytes.Equal(input[:4], incNonceMethodID) {
		input = input[4:]
		// incNonce
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

		if acc == txCtx.Origin {
			// Origin nonce shouldn't change during his transaction
			return nil, 0, vm.ErrExecutionReverted
		}

		if value.Cmp(common.Big256) >= 0 {
			// Don't allow large nonce increasing to prevent a nonce overflow
			return nil, 0, vm.ErrExecutionReverted
		}
		if value.Sign() <= 0 {
			return nil, 0, vm.ErrExecutionReverted
		}

		stateDB.SetNonce(acc, stateDB.GetNonce(acc)+value.Uint64())
	} else {
		return nil, 0, vm.ErrExecutionReverted
	}
	return nil, suppliedGas, nil
}
