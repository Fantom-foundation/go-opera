// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ballot

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
const ContractABI = "[{\"inputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"proposalNames\",\"type\":\"bytes32[]\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"constant\":true,\"inputs\":[],\"name\":\"chairperson\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"delegate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"}],\"name\":\"giveRightToVote\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"proposals\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"name\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"voteCount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"proposal\",\"type\":\"uint256\"}],\"name\":\"vote\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"voters\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"weight\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"voted\",\"type\":\"bool\"},{\"internalType\":\"address\",\"name\":\"delegate\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"vote\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"winnerName\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"winnerName_\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"winningProposal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"winningProposal_\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
var ContractBin = "0x608060405234801561001057600080fd5b5060405161089a38038061089a8339818101604052602081101561003357600080fd5b810190808051604051939291908464010000000082111561005357600080fd5b90830190602082018581111561006857600080fd5b825186602082028301116401000000008211171561008557600080fd5b82525081516020918201928201910280838360005b838110156100b257818101518382015260200161009a565b50505050919091016040908152600080546001600160a01b03191633178082556001600160a01b03168152600160208190529181209190915593505050505b8151811015610153576002604051806040016040528084848151811061011357fe5b60209081029190910181015182526000918101829052835460018181018655948352918190208351600290930201918255919091015190820155016100f1565b5050610736806101646000396000f3fe608060405234801561001057600080fd5b50600436106100885760003560e01c8063609ff1bd1161005b578063609ff1bd1461012c5780639e7b8d6114610146578063a3ec138d1461016c578063e2ba53f0146101c057610088565b80630121b93f1461008d578063013cf08b146100ac5780632e4176cf146100e25780635c19a95c14610106575b600080fd5b6100aa600480360360208110156100a357600080fd5b50356101c8565b005b6100c9600480360360208110156100c257600080fd5b50356102cb565b6040805192835260208301919091528051918290030190f35b6100ea6102f6565b604080516001600160a01b039092168252519081900360200190f35b6100aa6004803603602081101561011c57600080fd5b50356001600160a01b0316610305565b610134610516565b60408051918252519081900360200190f35b6100aa6004803603602081101561015c57600080fd5b50356001600160a01b031661057d565b6101926004803603602081101561018257600080fd5b50356001600160a01b0316610678565b6040805194855292151560208501526001600160a01b03909116838301526060830152519081900360800190f35b6101346106ac565b336000908152600160205260409020805461022a576040805162461bcd60e51b815260206004820152601460248201527f486173206e6f20726967687420746f20766f7465000000000000000000000000604482015290519081900360640190fd5b600181015460ff1615610284576040805162461bcd60e51b815260206004820152600e60248201527f416c726561647920766f7465642e000000000000000000000000000000000000604482015290519081900360640190fd5b6001818101805460ff19169091179055600280820183905581548154909190849081106102ad57fe5b60009182526020909120600160029092020101805490910190555050565b600281815481106102d857fe5b60009182526020909120600290910201805460019091015490915082565b6000546001600160a01b031681565b3360009081526001602081905260409091209081015460ff1615610370576040805162461bcd60e51b815260206004820152601260248201527f596f7520616c726561647920766f7465642e0000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0382163314156103ce576040805162461bcd60e51b815260206004820152601e60248201527f53656c662d64656c65676174696f6e20697320646973616c6c6f7765642e0000604482015290519081900360640190fd5b6001600160a01b038281166000908152600160208190526040909120015461010090041615610478576001600160a01b039182166000908152600160208190526040909120015461010090049091169033821415610473576040805162461bcd60e51b815260206004820152601960248201527f466f756e64206c6f6f7020696e2064656c65676174696f6e2e00000000000000604482015290519081900360640190fd5b6103ce565b6001818101805460ff191682177fffffffffffffffffffffff0000000000000000000000000000000000000000ff166101006001600160a01b0386169081029190911790915560009081526020829052604090209081015460ff1615610509578154600282810154815481106104ea57fe5b6000918252602090912060016002909202010180549091019055610511565b815481540181555b505050565b600080805b60025481101561057857816002828154811061053357fe5b9060005260206000209060020201600101541115610570576002818154811061055857fe5b90600052602060002090600202016001015491508092505b60010161051b565b505090565b6000546001600160a01b031633146105c65760405162461bcd60e51b81526004018080602001828103825260288152602001806106da6028913960400191505060405180910390fd5b6001600160a01b0381166000908152600160208190526040909120015460ff1615610638576040805162461bcd60e51b815260206004820152601860248201527f54686520766f74657220616c726561647920766f7465642e0000000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152600160205260409020541561065b57600080fd5b6001600160a01b0316600090815260016020819052604090912055565b600160208190526000918252604090912080549181015460029091015460ff82169161010090046001600160a01b03169084565b600060026106b8610516565b815481106106c257fe5b90600052602060002090600202016000015490509056fe4f6e6c79206368616972706572736f6e2063616e206769766520726967687420746f20766f74652ea265627a7a72315820a22e04a1172617645f3654f281b81f779805de1b7d7930b30443c2f276926a8564736f6c634300050c0032"

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend, proposalNames [][32]byte) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractBin), backend, proposalNames)
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

// Chairperson is a free data retrieval call binding the contract method 0x2e4176cf.
//
// Solidity: function chairperson() view returns(address)
func (_Contract *ContractCaller) Chairperson(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "chairperson")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Chairperson is a free data retrieval call binding the contract method 0x2e4176cf.
//
// Solidity: function chairperson() view returns(address)
func (_Contract *ContractSession) Chairperson() (common.Address, error) {
	return _Contract.Contract.Chairperson(&_Contract.CallOpts)
}

// Chairperson is a free data retrieval call binding the contract method 0x2e4176cf.
//
// Solidity: function chairperson() view returns(address)
func (_Contract *ContractCallerSession) Chairperson() (common.Address, error) {
	return _Contract.Contract.Chairperson(&_Contract.CallOpts)
}

// Proposals is a free data retrieval call binding the contract method 0x013cf08b.
//
// Solidity: function proposals(uint256 ) view returns(bytes32 name, uint256 voteCount)
func (_Contract *ContractCaller) Proposals(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Name      [32]byte
	VoteCount *big.Int
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "proposals", arg0)

	outstruct := new(struct {
		Name      [32]byte
		VoteCount *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Name = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.VoteCount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Proposals is a free data retrieval call binding the contract method 0x013cf08b.
//
// Solidity: function proposals(uint256 ) view returns(bytes32 name, uint256 voteCount)
func (_Contract *ContractSession) Proposals(arg0 *big.Int) (struct {
	Name      [32]byte
	VoteCount *big.Int
}, error) {
	return _Contract.Contract.Proposals(&_Contract.CallOpts, arg0)
}

// Proposals is a free data retrieval call binding the contract method 0x013cf08b.
//
// Solidity: function proposals(uint256 ) view returns(bytes32 name, uint256 voteCount)
func (_Contract *ContractCallerSession) Proposals(arg0 *big.Int) (struct {
	Name      [32]byte
	VoteCount *big.Int
}, error) {
	return _Contract.Contract.Proposals(&_Contract.CallOpts, arg0)
}

// Voters is a free data retrieval call binding the contract method 0xa3ec138d.
//
// Solidity: function voters(address ) view returns(uint256 weight, bool voted, address delegate, uint256 vote)
func (_Contract *ContractCaller) Voters(opts *bind.CallOpts, arg0 common.Address) (struct {
	Weight   *big.Int
	Voted    bool
	Delegate common.Address
	Vote     *big.Int
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "voters", arg0)

	outstruct := new(struct {
		Weight   *big.Int
		Voted    bool
		Delegate common.Address
		Vote     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Weight = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Voted = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.Delegate = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.Vote = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Voters is a free data retrieval call binding the contract method 0xa3ec138d.
//
// Solidity: function voters(address ) view returns(uint256 weight, bool voted, address delegate, uint256 vote)
func (_Contract *ContractSession) Voters(arg0 common.Address) (struct {
	Weight   *big.Int
	Voted    bool
	Delegate common.Address
	Vote     *big.Int
}, error) {
	return _Contract.Contract.Voters(&_Contract.CallOpts, arg0)
}

// Voters is a free data retrieval call binding the contract method 0xa3ec138d.
//
// Solidity: function voters(address ) view returns(uint256 weight, bool voted, address delegate, uint256 vote)
func (_Contract *ContractCallerSession) Voters(arg0 common.Address) (struct {
	Weight   *big.Int
	Voted    bool
	Delegate common.Address
	Vote     *big.Int
}, error) {
	return _Contract.Contract.Voters(&_Contract.CallOpts, arg0)
}

// WinnerName is a free data retrieval call binding the contract method 0xe2ba53f0.
//
// Solidity: function winnerName() view returns(bytes32 winnerName_)
func (_Contract *ContractCaller) WinnerName(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "winnerName")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// WinnerName is a free data retrieval call binding the contract method 0xe2ba53f0.
//
// Solidity: function winnerName() view returns(bytes32 winnerName_)
func (_Contract *ContractSession) WinnerName() ([32]byte, error) {
	return _Contract.Contract.WinnerName(&_Contract.CallOpts)
}

// WinnerName is a free data retrieval call binding the contract method 0xe2ba53f0.
//
// Solidity: function winnerName() view returns(bytes32 winnerName_)
func (_Contract *ContractCallerSession) WinnerName() ([32]byte, error) {
	return _Contract.Contract.WinnerName(&_Contract.CallOpts)
}

// WinningProposal is a free data retrieval call binding the contract method 0x609ff1bd.
//
// Solidity: function winningProposal() view returns(uint256 winningProposal_)
func (_Contract *ContractCaller) WinningProposal(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "winningProposal")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WinningProposal is a free data retrieval call binding the contract method 0x609ff1bd.
//
// Solidity: function winningProposal() view returns(uint256 winningProposal_)
func (_Contract *ContractSession) WinningProposal() (*big.Int, error) {
	return _Contract.Contract.WinningProposal(&_Contract.CallOpts)
}

// WinningProposal is a free data retrieval call binding the contract method 0x609ff1bd.
//
// Solidity: function winningProposal() view returns(uint256 winningProposal_)
func (_Contract *ContractCallerSession) WinningProposal() (*big.Int, error) {
	return _Contract.Contract.WinningProposal(&_Contract.CallOpts)
}

// Delegate is a paid mutator transaction binding the contract method 0x5c19a95c.
//
// Solidity: function delegate(address to) returns()
func (_Contract *ContractTransactor) Delegate(opts *bind.TransactOpts, to common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "delegate", to)
}

// Delegate is a paid mutator transaction binding the contract method 0x5c19a95c.
//
// Solidity: function delegate(address to) returns()
func (_Contract *ContractSession) Delegate(to common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Delegate(&_Contract.TransactOpts, to)
}

// Delegate is a paid mutator transaction binding the contract method 0x5c19a95c.
//
// Solidity: function delegate(address to) returns()
func (_Contract *ContractTransactorSession) Delegate(to common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Delegate(&_Contract.TransactOpts, to)
}

// GiveRightToVote is a paid mutator transaction binding the contract method 0x9e7b8d61.
//
// Solidity: function giveRightToVote(address voter) returns()
func (_Contract *ContractTransactor) GiveRightToVote(opts *bind.TransactOpts, voter common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "giveRightToVote", voter)
}

// GiveRightToVote is a paid mutator transaction binding the contract method 0x9e7b8d61.
//
// Solidity: function giveRightToVote(address voter) returns()
func (_Contract *ContractSession) GiveRightToVote(voter common.Address) (*types.Transaction, error) {
	return _Contract.Contract.GiveRightToVote(&_Contract.TransactOpts, voter)
}

// GiveRightToVote is a paid mutator transaction binding the contract method 0x9e7b8d61.
//
// Solidity: function giveRightToVote(address voter) returns()
func (_Contract *ContractTransactorSession) GiveRightToVote(voter common.Address) (*types.Transaction, error) {
	return _Contract.Contract.GiveRightToVote(&_Contract.TransactOpts, voter)
}

// Vote is a paid mutator transaction binding the contract method 0x0121b93f.
//
// Solidity: function vote(uint256 proposal) returns()
func (_Contract *ContractTransactor) Vote(opts *bind.TransactOpts, proposal *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "vote", proposal)
}

// Vote is a paid mutator transaction binding the contract method 0x0121b93f.
//
// Solidity: function vote(uint256 proposal) returns()
func (_Contract *ContractSession) Vote(proposal *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Vote(&_Contract.TransactOpts, proposal)
}

// Vote is a paid mutator transaction binding the contract method 0x0121b93f.
//
// Solidity: function vote(uint256 proposal) returns()
func (_Contract *ContractTransactorSession) Vote(proposal *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Vote(&_Contract.TransactOpts, proposal)
}
