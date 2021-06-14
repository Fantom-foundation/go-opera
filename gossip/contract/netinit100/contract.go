// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package netinit100

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sealedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_sfc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_auth\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_driver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_evmWriter\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"initializeAll\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
var ContractBin = "0x608060405234801561001057600080fd5b50610296806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c8063c80e151314610030575b600080fd5b610091600480360360e081101561004657600080fd5b5080359060208101359073ffffffffffffffffffffffffffffffffffffffff60408201358116916060810135821691608082013581169160a081013582169160c09091013516610093565b005b604080517f485cc95500000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8681166004830152848116602483015291519185169163485cc9559160448082019260009290919082900301818387803b15801561010c57600080fd5b505af1158015610120573d6000803e3d6000fd5b5050604080517fc0c53b8b00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8981166004830152878116602483015285811660448301529151918816935063c0c53b8b925060648082019260009290919082900301818387803b1580156101a557600080fd5b505af11580156101b9573d6000803e3d6000fd5b5050604080517f019e2729000000000000000000000000000000000000000000000000000000008152600481018b9052602481018a905273ffffffffffffffffffffffffffffffffffffffff888116604483015285811660648301529151918916935063019e2729925060848082019260009290919082900301818387803b15801561024457600080fd5b505af1158015610258573d6000803e3d6000fd5b50600092505050fffea265627a7a72315820335fde84a547de34d8d678665aeca278de801945c81fdc1ce389633b38fa0cef64736f6c63430005110032"

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractBin), backend)
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

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// InitializeAll is a paid mutator transaction binding the contract method 0xc80e1513.
//
// Solidity: function initializeAll(uint256 sealedEpoch, uint256 totalSupply, address _sfc, address _auth, address _driver, address _evmWriter, address _owner) returns()
func (_Contract *ContractTransactor) InitializeAll(opts *bind.TransactOpts, sealedEpoch *big.Int, totalSupply *big.Int, _sfc common.Address, _auth common.Address, _driver common.Address, _evmWriter common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initializeAll", sealedEpoch, totalSupply, _sfc, _auth, _driver, _evmWriter, _owner)
}

// InitializeAll is a paid mutator transaction binding the contract method 0xc80e1513.
//
// Solidity: function initializeAll(uint256 sealedEpoch, uint256 totalSupply, address _sfc, address _auth, address _driver, address _evmWriter, address _owner) returns()
func (_Contract *ContractSession) InitializeAll(sealedEpoch *big.Int, totalSupply *big.Int, _sfc common.Address, _auth common.Address, _driver common.Address, _evmWriter common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.InitializeAll(&_Contract.TransactOpts, sealedEpoch, totalSupply, _sfc, _auth, _driver, _evmWriter, _owner)
}

// InitializeAll is a paid mutator transaction binding the contract method 0xc80e1513.
//
// Solidity: function initializeAll(uint256 sealedEpoch, uint256 totalSupply, address _sfc, address _auth, address _driver, address _evmWriter, address _owner) returns()
func (_Contract *ContractTransactorSession) InitializeAll(sealedEpoch *big.Int, totalSupply *big.Int, _sfc common.Address, _auth common.Address, _driver common.Address, _evmWriter common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.InitializeAll(&_Contract.TransactOpts, sealedEpoch, totalSupply, _sfc, _auth, _driver, _evmWriter, _owner)
}

var ContractBinRuntime = "0x608060405234801561001057600080fd5b506004361061002b5760003560e01c8063c80e151314610030575b600080fd5b610091600480360360e081101561004657600080fd5b5080359060208101359073ffffffffffffffffffffffffffffffffffffffff60408201358116916060810135821691608082013581169160a081013582169160c09091013516610093565b005b604080517f485cc95500000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8681166004830152848116602483015291519185169163485cc9559160448082019260009290919082900301818387803b15801561010c57600080fd5b505af1158015610120573d6000803e3d6000fd5b5050604080517fc0c53b8b00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8981166004830152878116602483015285811660448301529151918816935063c0c53b8b925060648082019260009290919082900301818387803b1580156101a557600080fd5b505af11580156101b9573d6000803e3d6000fd5b5050604080517f019e2729000000000000000000000000000000000000000000000000000000008152600481018b9052602481018a905273ffffffffffffffffffffffffffffffffffffffff888116604483015285811660648301529151918916935063019e2729925060848082019260009290919082900301818387803b15801561024457600080fd5b505af1158015610258573d6000803e3d6000fd5b50600092505050fffea265627a7a72315820335fde84a547de34d8d678665aeca278de801945c81fdc1ce389633b38fa0cef64736f6c63430005110032"
