// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package aatester

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

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"payable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_contract\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_method\",\"type\":\"string\"}],\"name\":\"call\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gasLeft\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gasLeftAfterPaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gasLeftBeforePaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"gasPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"origin\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"reset\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"revertBeforePaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sender\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setBalanceAfterPaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setBalanceBeforePaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setGasPrice\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setNonce\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setNonceBeforePaygasAndRevert\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setOriginAfterPaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setOriginBeforePaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setSenderAfterPaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"setSenderBeforePaygas\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040526108b0806100136000396000f3fe3273ffffffffffffffffffffffffffffffffffffffff1460245736601f57005b600080fd5b608060405234801561003557600080fd5b50600436106101515760003560e01c8063b4adeaff116100d2578063d826f88f11610096578063d826f88f14610262578063e6409c6f1461026c578063ea1bdf4d14610276578063f3f403fb14610280578063fe173b971461028a57610151565b8063b4adeaff1461021c578063b69ef8a814610226578063c32bafd114610244578063c69fa5771461024e578063cae67e791461025857610151565b80638f67cd3c116101195780638f67cd3c146101c25780639099df1b146101cc578063938b5f32146101d6578063ae224cbf146101f4578063affed0e0146101fe57610151565b806309ee8dba146101565780632ddb301b1461016057806344e1ce0d1461017e57806367e404ce1461019a5780637078aa28146101b8575b600080fd5b61015e6102a8565b005b6101686102b3565b6040516101759190610628565b60405180910390f35b61019860048036038101906101939190610710565b6102b9565b005b6101a26103b2565b6040516101af919061077f565b60405180910390f35b6101c06103d8565b005b6101ca6103e2565b005b6101d46103ef565b005b6101de6103f9565b6040516101eb919061077f565b60405180910390f35b6101fc61041f565b005b610206610429565b6040516102139190610628565b60405180910390f35b61022461042f565b005b61022e610473565b60405161023b9190610628565b60405180910390f35b61024c610479565b005b6102566104bd565b005b6102606104c7565b005b61026a6104d1565b005b610274610577565b005b61027e6105bb565b005b6102886105c5565b005b610292610609565b60405161029f9190610628565b60405180910390f35b496000819055600080fd5b60025481565b5c60008373ffffffffffffffffffffffffffffffffffffffff16838360405160240160405160208183030381529060405291906040516102fa9291906107d9565b60405180910390207bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff838183161783525050505060405161035c9190610863565b6000604051808303816000865af19150503d8060008114610399576040519150601f19603f3d011682016040523d82523d6000602084013e61039e565b606091505b50509050806103ac57600080fd5b50505050565b600360009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5c3a600181905550565b496000819055505c600080fd5b496000819055505c565b600460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5c5a600281905550565b60005481565b33600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505c565b60055481565b5c33600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550565b476005819055505c565b5c47600581905550565b5c60008081905550600060018190555060006002819055506000600360006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000600581905550565b32600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505c565b5a6002819055505c565b5c32600460006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550565b60015481565b6000819050919050565b6106228161060f565b82525050565b600060208201905061063d6000830184610619565b92915050565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006106788261064d565b9050919050565b6106888161066d565b811461069357600080fd5b50565b6000813590506106a58161067f565b92915050565b600080fd5b600080fd5b600080fd5b60008083601f8401126106d0576106cf6106ab565b5b8235905067ffffffffffffffff8111156106ed576106ec6106b0565b5b602083019150836001820283011115610709576107086106b5565b5b9250929050565b60008060006040848603121561072957610728610643565b5b600061073786828701610696565b935050602084013567ffffffffffffffff81111561075857610757610648565b5b610764868287016106ba565b92509250509250925092565b6107798161066d565b82525050565b60006020820190506107946000830184610770565b92915050565b600081905092915050565b82818337600083830152505050565b60006107c0838561079a565b93506107cd8385846107a5565b82840190509392505050565b60006107e68284866107b4565b91508190509392505050565b600081519050919050565b600081905092915050565b60005b8381101561082657808201518184015260208101905061080b565b60008484015250505050565b600061083d826107f2565b61084781856107fd565b9350610857818560208601610808565b80840191505092915050565b600061086f8284610832565b91508190509291505056fea264697066735822122055449b6ae02445607e4d1d25278fb510e39e3892355fc7ebd3e72bbb96ecb96964736f6c63430008130033",
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
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend)
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

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Contract *ContractCaller) Balance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "balance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Contract *ContractSession) Balance() (*big.Int, error) {
	return _Contract.Contract.Balance(&_Contract.CallOpts)
}

// Balance is a free data retrieval call binding the contract method 0xb69ef8a8.
//
// Solidity: function balance() view returns(uint256)
func (_Contract *ContractCallerSession) Balance() (*big.Int, error) {
	return _Contract.Contract.Balance(&_Contract.CallOpts)
}

// GasLeft is a free data retrieval call binding the contract method 0x2ddb301b.
//
// Solidity: function gasLeft() view returns(uint256)
func (_Contract *ContractCaller) GasLeft(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "gasLeft")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GasLeft is a free data retrieval call binding the contract method 0x2ddb301b.
//
// Solidity: function gasLeft() view returns(uint256)
func (_Contract *ContractSession) GasLeft() (*big.Int, error) {
	return _Contract.Contract.GasLeft(&_Contract.CallOpts)
}

// GasLeft is a free data retrieval call binding the contract method 0x2ddb301b.
//
// Solidity: function gasLeft() view returns(uint256)
func (_Contract *ContractCallerSession) GasLeft() (*big.Int, error) {
	return _Contract.Contract.GasLeft(&_Contract.CallOpts)
}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_Contract *ContractCaller) GasPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "gasPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_Contract *ContractSession) GasPrice() (*big.Int, error) {
	return _Contract.Contract.GasPrice(&_Contract.CallOpts)
}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_Contract *ContractCallerSession) GasPrice() (*big.Int, error) {
	return _Contract.Contract.GasPrice(&_Contract.CallOpts)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_Contract *ContractCaller) Nonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "nonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_Contract *ContractSession) Nonce() (*big.Int, error) {
	return _Contract.Contract.Nonce(&_Contract.CallOpts)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_Contract *ContractCallerSession) Nonce() (*big.Int, error) {
	return _Contract.Contract.Nonce(&_Contract.CallOpts)
}

// Origin is a free data retrieval call binding the contract method 0x938b5f32.
//
// Solidity: function origin() view returns(address)
func (_Contract *ContractCaller) Origin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "origin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Origin is a free data retrieval call binding the contract method 0x938b5f32.
//
// Solidity: function origin() view returns(address)
func (_Contract *ContractSession) Origin() (common.Address, error) {
	return _Contract.Contract.Origin(&_Contract.CallOpts)
}

// Origin is a free data retrieval call binding the contract method 0x938b5f32.
//
// Solidity: function origin() view returns(address)
func (_Contract *ContractCallerSession) Origin() (common.Address, error) {
	return _Contract.Contract.Origin(&_Contract.CallOpts)
}

// Sender is a free data retrieval call binding the contract method 0x67e404ce.
//
// Solidity: function sender() view returns(address)
func (_Contract *ContractCaller) Sender(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "sender")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Sender is a free data retrieval call binding the contract method 0x67e404ce.
//
// Solidity: function sender() view returns(address)
func (_Contract *ContractSession) Sender() (common.Address, error) {
	return _Contract.Contract.Sender(&_Contract.CallOpts)
}

// Sender is a free data retrieval call binding the contract method 0x67e404ce.
//
// Solidity: function sender() view returns(address)
func (_Contract *ContractCallerSession) Sender() (common.Address, error) {
	return _Contract.Contract.Sender(&_Contract.CallOpts)
}

func (_Contract *Contract) CallSetOrigin(contract common.Address) []byte {
	data, _ := sAbi.Pack("call", contract, "setOriginAfterPaygas()")
	return data
}

func (_Contract *Contract) CallSetSender(contract common.Address) []byte {
	data, _ := sAbi.Pack("call", contract, "setSenderAfterPaygas()")
	return data
}

func (_Contract *Contract) SetGasLeftAfterPaygas() []byte {
	data, _ := sAbi.Pack("gasLeftAfterPaygas")
	return data
}

func (_Contract *Contract) SetGasLeftBeforePaygas() []byte {
	data, _ := sAbi.Pack("gasLeftBeforePaygas")
	return data
}

func (_Contract *Contract) SetGasPrice() []byte {
	data, _ := sAbi.Pack("setGasPrice")
	return data
}

func (_Contract *Contract) SetBalanceAfterPaygas() []byte {
	data, _ := sAbi.Pack("setBalanceAfterPaygas")
	return data
}

func (_Contract *Contract) SetBalanceBeforePaygas() []byte {
	data, _ := sAbi.Pack("setBalanceBeforePaygas")
	return data
}

func (_Contract *Contract) SetNonce() []byte {
	data, _ := sAbi.Pack("setNonce")
	return data
}

func (_Contract *Contract) SetNonceBeforePaygasAndRevert() []byte {
	data, _ := sAbi.Pack("setNonceBeforePaygasAndRevert")
	return data
}
func (_Contract *Contract) SetOriginAfterPaygas() []byte {
	data, _ := sAbi.Pack("setOriginAfterPaygas")
	return data
}

func (_Contract *Contract) SetOriginBeforePaygas() []byte {
	data, _ := sAbi.Pack("setOriginBeforePaygas")
	return data
}

func (_Contract *Contract) SetSenderAfterPaygas() []byte {
	data, _ := sAbi.Pack("setSenderAfterPaygas")
	return data
}

func (_Contract *Contract) SetSenderBeforePaygas() []byte {
	data, _ := sAbi.Pack("setSenderAfterPaygas")
	return data
}

func (_Contract *Contract) Reset() []byte {
	data, _ := sAbi.Pack("reset")
	return data
}
