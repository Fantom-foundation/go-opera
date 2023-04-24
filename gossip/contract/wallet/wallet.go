// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package wallet

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// WalletSignature is an auto generated low-level Go binding around an user-defined struct.
type WalletSignature struct {
	V uint8
	R [32]byte
	S [32]byte
}

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structWallet.Signature\",\"name\":\"signature\",\"type\":\"tuple\"}],\"name\":\"changeOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"internalType\":\"structWallet.Signature\",\"name\":\"signature\",\"type\":\"tuple\"}],\"name\":\"transfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526040516109ea3803806109ea833981810160405281019061002591906100ce565b806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550506100fb565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061009b82610070565b9050919050565b6100ab81610090565b81146100b657600080fd5b50565b6000815190506100c8816100a2565b92915050565b6000602082840312156100e4576100e361006b565b5b60006100f2848285016100b9565b91505092915050565b6108e08061010a6000396000f3fe3273ffffffffffffffffffffffffffffffffffffffff1460245736601f57005b600080fd5b608060405234801561003557600080fd5b50600436106100665760003560e01c806354a1a1671461006b5780638d5da5b6146100875780638da5cb5b146100a3575b600080fd5b610085600480360381019061008091906104ed565b6100c1565b005b6100a1600480360381019061009c91906105b3565b610217565b005b6100ab61032f565b6040516100b89190610602565b60405180910390f35b80600030496040516020016100d79291906106d6565b604051602081830303815290604052905060008180519060200120905060006101008285610353565b90508073ffffffffffffffffffffffffffffffffffffffff1660008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610190576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101879061075f565b60405180910390fd5b5c60008973ffffffffffffffffffffffffffffffffffffffff168989896040516101bb9291906107be565b60006040518083038185875af1925050503d80600081146101f8576040519150601f19603f3d011682016040523d82523d6000602084013e6101fd565b606091505b505090508061020b57600080fd5b50505050505050505050565b806000304960405160200161022d9291906106d6565b604051602081830303815290604052905060008180519060200120905060006102568285610353565b90508073ffffffffffffffffffffffffffffffffffffffff1660008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16146102e6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102dd9061075f565b60405180910390fd5b5c856000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550505050505050565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600060018383600001602081019061036b9190610810565b84602001358560400135604051600081526020016040526040516103929493929190610865565b6020604051602081039080840390855afa1580156103b4573d6000803e3d6000fd5b50505060206040510351905092915050565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006103fb826103d0565b9050919050565b61040b816103f0565b811461041657600080fd5b50565b60008135905061042881610402565b92915050565b6000819050919050565b6104418161042e565b811461044c57600080fd5b50565b60008135905061045e81610438565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f84011261048957610488610464565b5b8235905067ffffffffffffffff8111156104a6576104a5610469565b5b6020830191508360018202830111156104c2576104c161046e565b5b9250929050565b600080fd5b6000606082840312156104e4576104e36104c9565b5b81905092915050565b600080600080600060c08688031215610509576105086103c6565b5b600061051788828901610419565b95505060206105288882890161044f565b945050604086013567ffffffffffffffff811115610549576105486103cb565b5b61055588828901610473565b93509350506060610568888289016104ce565b9150509295509295909350565b6000610580826103d0565b9050919050565b61059081610575565b811461059b57600080fd5b50565b6000813590506105ad81610587565b92915050565b600080608083850312156105ca576105c96103c6565b5b60006105d88582860161059e565b92505060206105e9858286016104ce565b9150509250929050565b6105fc81610575565b82525050565b600060208201905061061760008301846105f3565b92915050565b6000819050919050565b600061064261063d610638846103d0565b61061d565b6103d0565b9050919050565b600061065482610627565b9050919050565b600061066682610649565b9050919050565b60008160601b9050919050565b60006106858261066d565b9050919050565b60006106978261067a565b9050919050565b6106af6106aa8261065b565b61068c565b82525050565b6000819050919050565b6106d06106cb8261042e565b6106b5565b82525050565b60006106e2828561069e565b6014820191506106f282846106bf565b6020820191508190509392505050565b600082825260208201905092915050565b7f4e6f74206f776e65720000000000000000000000000000000000000000000000600082015250565b6000610749600983610702565b915061075482610713565b602082019050919050565b600060208201905081810360008301526107788161073c565b9050919050565b600081905092915050565b82818337600083830152505050565b60006107a5838561077f565b93506107b283858461078a565b82840190509392505050565b60006107cb828486610799565b91508190509392505050565b600060ff82169050919050565b6107ed816107d7565b81146107f857600080fd5b50565b60008135905061080a816107e4565b92915050565b600060208284031215610826576108256103c6565b5b6000610834848285016107fb565b91505092915050565b6000819050919050565b6108508161083d565b82525050565b61085f816107d7565b82525050565b600060808201905061087a6000830187610847565b6108876020830186610856565b6108946040830185610847565b6108a16060830184610847565b9594505050505056fea2646970667358221220fc4d60b963fb07591e356b1c0ce07cc87c7f147f36a86b6b98c5fc7c0c80159e64736f6c63430008130033",
}

var (
	sAbi, _ = abi.JSON(strings.NewReader(ContractMetaData.ABI))
)

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = ContractMetaData.Bin

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend, _owner common.Address) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend, _owner)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCallerSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

func (_Contract *Contract) Transfer(to common.Address, amount *big.Int, payload []byte, signature WalletSignature) []byte {
	data, _ := sAbi.Pack("transfer", to, amount, payload, signature)
	return data
}

func (_Contract *Contract) ChangeOwner(newOwner common.Address, signature WalletSignature) []byte {
	data, _ := sAbi.Pack("changeOwner", newOwner, signature)
	return data
}
