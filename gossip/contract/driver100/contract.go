// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package driver100

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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"AdvanceEpochs\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"diff\",\"type\":\"bytes\"}],\"name\":\"UpdateNetworkRules\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"UpdateNetworkVersion\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"UpdateValidatorPubkey\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"weight\",\"type\":\"uint256\"}],\"name\":\"UpdateValidatorWeight\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"backend\",\"type\":\"address\"}],\"name\":\"UpdatedBackend\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_backend\",\"type\":\"address\"}],\"name\":\"setBackend\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_backend\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_evmWriterAddress\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"setBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"copyCode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"with\",\"type\":\"address\"}],\"name\":\"swapCode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"key\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"value\",\"type\":\"bytes32\"}],\"name\":\"setStorage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"diff\",\"type\":\"uint256\"}],\"name\":\"incNonce\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"diff\",\"type\":\"bytes\"}],\"name\":\"updateNetworkRules\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"updateNetworkVersion\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"advanceEpochs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"updateValidatorWeight\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"updateValidatorPubkey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_auth\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"setGenesisValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupFromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupEndTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewards\",\"type\":\"uint256\"}],\"name\":\"setGenesisDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"deactivateValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nextValidatorIDs\",\"type\":\"uint256[]\"}],\"name\":\"sealEpochValidators\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"}],\"name\":\"sealEpoch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"usedGas\",\"type\":\"uint256\"}],\"name\":\"sealEpochV1\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50611b5e806100206000396000f3fe608060405234801561001057600080fd5b50600436106101365760003560e01c806379bead38116100b2578063d6a0c7af11610081578063e08d7e6611610066578063e08d7e6614610676578063e30443bc146106e6578063ebdf104c1461071f57610136565b8063d6a0c7af14610608578063da7fc24f1461064357610136565b806379bead38146103d65780637f52e13e1461040f578063a4066fbe14610575578063b9cc6b1c1461059857610136565b8063242a6e3f1161010957806339e503ab116100ee57806339e503ab146102b1578063485cc955146102f05780634feb92f31461032b57610136565b8063242a6e3f1461021d578063267ab4461461029457610136565b806307690b2a1461013b5780630aeeca001461017857806318f628d4146101955780631e702f83146101fa575b600080fd5b6101766004803603604081101561015157600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516610885565b005b6101766004803603602081101561018e57600080fd5b5035610989565b61017660048036036101208110156101ac57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e0810135906101000135610a2b565b6101766004803603604081101561021057600080fd5b5080359060200135610b4f565b6101766004803603604081101561023357600080fd5b8135919081019060408101602082013564010000000081111561025557600080fd5b82018360208201111561026757600080fd5b8035906020019184600183028401116401000000008311171561028957600080fd5b509092509050610c1c565b610176600480360360208110156102aa57600080fd5b5035610cee565b610176600480360360608110156102c757600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060208101359060400135610d90565b6101766004803603604081101561030657600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516610e9b565b610176600480360361010081101561034257600080fd5b73ffffffffffffffffffffffffffffffffffffffff8235169160208101359181019060608101604082013564010000000081111561037f57600080fd5b82018360208201111561039157600080fd5b803590602001918460018302840111640100000000831117156103b357600080fd5b919350915080359060208101359060408101359060608101359060800135611043565b610176600480360360408110156103ec57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff813516906020013561119b565b610176600480360360a081101561042557600080fd5b81019060208101813564010000000081111561044057600080fd5b82018360208201111561045257600080fd5b8035906020019184602083028401116401000000008311171561047457600080fd5b91939092909160208101903564010000000081111561049257600080fd5b8201836020820111156104a457600080fd5b803590602001918460208302840111640100000000831117156104c657600080fd5b9193909290916020810190356401000000008111156104e457600080fd5b8201836020820111156104f657600080fd5b8035906020019184602083028401116401000000008311171561051857600080fd5b91939092909160208101903564010000000081111561053657600080fd5b82018360208201111561054857600080fd5b8035906020019184602083028401116401000000008311171561056a57600080fd5b919350915035611282565b6101766004803603604081101561058b57600080fd5b5080359060200135611411565b610176600480360360208110156105ae57600080fd5b8101906020810181356401000000008111156105c957600080fd5b8201836020820111156105db57600080fd5b803590602001918460018302840111640100000000831117156105fd57600080fd5b5090925090506114b7565b6101766004803603604081101561061e57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516611587565b6101766004803603602081101561065957600080fd5b503573ffffffffffffffffffffffffffffffffffffffff1661166f565b6101766004803603602081101561068c57600080fd5b8101906020810181356401000000008111156106a757600080fd5b8201836020820111156106b957600080fd5b803590602001918460208302840111640100000000831117156106db57600080fd5b509092509050611763565b610176600480360360408110156106fc57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060200135611859565b6101766004803603608081101561073557600080fd5b81019060208101813564010000000081111561075057600080fd5b82018360208201111561076257600080fd5b8035906020019184602083028401116401000000008311171561078457600080fd5b9193909290916020810190356401000000008111156107a257600080fd5b8201836020820111156107b457600080fd5b803590602001918460208302840111640100000000831117156107d657600080fd5b9193909290916020810190356401000000008111156107f457600080fd5b82018360208201111561080657600080fd5b8035906020019184602083028401116401000000008311171561082857600080fd5b91939092909160208101903564010000000081111561084657600080fd5b82018360208201111561085857600080fd5b8035906020019184602083028401116401000000008311171561087a57600080fd5b509092509050611940565b60345473ffffffffffffffffffffffffffffffffffffffff1633146108f1576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517f07690b2a00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff85811660048301528481166024830152915191909216916307690b2a91604480830192600092919082900301818387803b15801561096d57600080fd5b505af1158015610981573d6000803e3d6000fd5b505050505050565b60345473ffffffffffffffffffffffffffffffffffffffff1633146109f5576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b6040805182815290517f0151256d62457b809bbc891b1f81c6dd0b9987552c70ce915b519750cd434dd19181900360200190a150565b3315610a7e576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603454604080517f18f628d400000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8c81166004830152602482018c9052604482018b9052606482018a90526084820189905260a4820188905260c4820187905260e482018690526101048201859052915191909216916318f628d49161012480830192600092919082900301818387803b158015610b2c57600080fd5b505af1158015610b40573d6000803e3d6000fd5b50505050505050505050505050565b3315610ba2576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603454604080517f1e702f830000000000000000000000000000000000000000000000000000000081526004810185905260248101849052905173ffffffffffffffffffffffffffffffffffffffff90921691631e702f839160448082019260009290919082900301818387803b15801561096d57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff163314610c88576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b827f0f0ef1ab97439def0a9d2c6d9dc166207f1b13b99e62b442b2993d6153c63a6e838360405180806020018281038252848482818152602001925080828437600083820152604051601f909101601f19169092018290039550909350505050a2505050565b60345473ffffffffffffffffffffffffffffffffffffffff163314610d5a576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b6040805182815290517f2ccdfd47cf0c1f1069d949f1789bb79b2f12821f021634fc835af1de66ea2feb9181900360200190a150565b60345473ffffffffffffffffffffffffffffffffffffffff163314610dfc576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517f39e503ab00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff86811660048301526024820186905260448201859052915191909216916339e503ab91606480830192600092919082900301818387803b158015610e7e57600080fd5b505af1158015610e92573d6000803e3d6000fd5b50505050505050565b600054610100900460ff1680610eb45750610eb4611af5565b80610ec2575060005460ff16155b610efd5760405162461bcd60e51b815260040180806020018281038252602e815260200180611afc602e913960400191505060405180910390fd5b600054610100900460ff16158015610f6357600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b603480547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff85169081179091556040517f64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef24503583069390600090a2603580547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff8416179055801561103e57600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555b505050565b3315611096576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16634feb92f38a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001898152602001806020018781526020018681526020018581526020018481526020018381526020018281038252898982818152602001925080828437600081840152601f19601f8201169050808301925050509a5050505050505050505050600060405180830381600087803b158015610b2c57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff163314611207576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517f79bead3800000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff858116600483015260248201859052915191909216916379bead3891604480830192600092919082900301818387803b15801561096d57600080fd5b33156112d5576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663592fe0c08a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808060200180602001806020018060200186815260200185810385528e8e82818152602001925060200280828437600083820152601f01601f191690910186810385528c8152602090810191508d908d0280828437600083820152601f01601f191690910186810384528a8152602090810191508b908b0280828437600083820152601f01601f19169091018681038352888152602090810191508990890280828437600081840152601f19601f8201169050808301925050509d5050505050505050505050505050600060405180830381600087803b158015610b2c57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff16331461147d576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b60408051828152905183917fb975807576e3b1461be7de07ebf7d20e4790ed802d7a0c4fdd0a1a13df72a935919081900360200190a25050565b60345473ffffffffffffffffffffffffffffffffffffffff163314611523576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b7f47d10eed096a44e3d0abc586c7e3a5d6cb5358cc90e7d437cd0627f7e765fb99828260405180806020018281038252848482818152602001925080828437600083820152604051601f909101601f19169092018290039550909350505050a15050565b60345473ffffffffffffffffffffffffffffffffffffffff1633146115f3576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517fd6a0c7af00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff858116600483015284811660248301529151919092169163d6a0c7af91604480830192600092919082900301818387803b15801561096d57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff1633146116db576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b60405173ffffffffffffffffffffffffffffffffffffffff8216907f64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef24503583069390600090a2603480547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff92909216919091179055565b33156117b6576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b6034546040517fe08d7e660000000000000000000000000000000000000000000000000000000081526020600482018181526024830185905273ffffffffffffffffffffffffffffffffffffffff9093169263e08d7e6692869286929182916044909101908590850280828437600081840152601f19601f8201169050808301925050509350505050600060405180830381600087803b15801561096d57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff1633146118c5576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517fe30443bc00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8581166004830152602482018590529151919092169163e30443bc91604480830192600092919082900301818387803b15801561096d57600080fd5b3315611993576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663592fe0c0898989898989898963322adc3a6040518a63ffffffff1660e01b8152600401808060200180602001806020018060200186815260200185810385528e8e82818152602001925060200280828437600083820152601f01601f191690910186810385528c8152602090810191508d908d0280828437600083820152601f01601f191690910186810384528a8152602090810191508b908b0280828437600083820152601f01601f19169091018681038352888152602090810191508990890280828437600081840152601f19601f8201169050808301925050509d5050505050505050505050505050600060405180830381600087803b158015611ad357600080fd5b505af1158015611ae7573d6000803e3d6000fd5b505050505050505050505050565b303b159056fe436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a6564a265627a7a72315820962250f3ef56c54b3df8ce7213cfcf494e44135e9c70003b4f86fe5ef1bcbb9f64736f6c63430005110032",
}

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

// AdvanceEpochs is a paid mutator transaction binding the contract method 0x0aeeca00.
//
// Solidity: function advanceEpochs(uint256 num) returns()
func (_Contract *ContractTransactor) AdvanceEpochs(opts *bind.TransactOpts, num *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "advanceEpochs", num)
}

// AdvanceEpochs is a paid mutator transaction binding the contract method 0x0aeeca00.
//
// Solidity: function advanceEpochs(uint256 num) returns()
func (_Contract *ContractSession) AdvanceEpochs(num *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.AdvanceEpochs(&_Contract.TransactOpts, num)
}

// AdvanceEpochs is a paid mutator transaction binding the contract method 0x0aeeca00.
//
// Solidity: function advanceEpochs(uint256 num) returns()
func (_Contract *ContractTransactorSession) AdvanceEpochs(num *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.AdvanceEpochs(&_Contract.TransactOpts, num)
}

// CopyCode is a paid mutator transaction binding the contract method 0xd6a0c7af.
//
// Solidity: function copyCode(address acc, address from) returns()
func (_Contract *ContractTransactor) CopyCode(opts *bind.TransactOpts, acc common.Address, from common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "copyCode", acc, from)
}

// CopyCode is a paid mutator transaction binding the contract method 0xd6a0c7af.
//
// Solidity: function copyCode(address acc, address from) returns()
func (_Contract *ContractSession) CopyCode(acc common.Address, from common.Address) (*types.Transaction, error) {
	return _Contract.Contract.CopyCode(&_Contract.TransactOpts, acc, from)
}

// CopyCode is a paid mutator transaction binding the contract method 0xd6a0c7af.
//
// Solidity: function copyCode(address acc, address from) returns()
func (_Contract *ContractTransactorSession) CopyCode(acc common.Address, from common.Address) (*types.Transaction, error) {
	return _Contract.Contract.CopyCode(&_Contract.TransactOpts, acc, from)
}

// DeactivateValidator is a paid mutator transaction binding the contract method 0x1e702f83.
//
// Solidity: function deactivateValidator(uint256 validatorID, uint256 status) returns()
func (_Contract *ContractTransactor) DeactivateValidator(opts *bind.TransactOpts, validatorID *big.Int, status *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deactivateValidator", validatorID, status)
}

// DeactivateValidator is a paid mutator transaction binding the contract method 0x1e702f83.
//
// Solidity: function deactivateValidator(uint256 validatorID, uint256 status) returns()
func (_Contract *ContractSession) DeactivateValidator(validatorID *big.Int, status *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.DeactivateValidator(&_Contract.TransactOpts, validatorID, status)
}

// DeactivateValidator is a paid mutator transaction binding the contract method 0x1e702f83.
//
// Solidity: function deactivateValidator(uint256 validatorID, uint256 status) returns()
func (_Contract *ContractTransactorSession) DeactivateValidator(validatorID *big.Int, status *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.DeactivateValidator(&_Contract.TransactOpts, validatorID, status)
}

// IncNonce is a paid mutator transaction binding the contract method 0x79bead38.
//
// Solidity: function incNonce(address acc, uint256 diff) returns()
func (_Contract *ContractTransactor) IncNonce(opts *bind.TransactOpts, acc common.Address, diff *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "incNonce", acc, diff)
}

// IncNonce is a paid mutator transaction binding the contract method 0x79bead38.
//
// Solidity: function incNonce(address acc, uint256 diff) returns()
func (_Contract *ContractSession) IncNonce(acc common.Address, diff *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.IncNonce(&_Contract.TransactOpts, acc, diff)
}

// IncNonce is a paid mutator transaction binding the contract method 0x79bead38.
//
// Solidity: function incNonce(address acc, uint256 diff) returns()
func (_Contract *ContractTransactorSession) IncNonce(acc common.Address, diff *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.IncNonce(&_Contract.TransactOpts, acc, diff)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _backend, address _evmWriterAddress) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, _backend common.Address, _evmWriterAddress common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", _backend, _evmWriterAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _backend, address _evmWriterAddress) returns()
func (_Contract *ContractSession) Initialize(_backend common.Address, _evmWriterAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _backend, _evmWriterAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _backend, address _evmWriterAddress) returns()
func (_Contract *ContractTransactorSession) Initialize(_backend common.Address, _evmWriterAddress common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _backend, _evmWriterAddress)
}

// SealEpoch is a paid mutator transaction binding the contract method 0xebdf104c.
//
// Solidity: function sealEpoch(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee) returns()
func (_Contract *ContractTransactor) SealEpoch(opts *bind.TransactOpts, offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "sealEpoch", offlineTimes, offlineBlocks, uptimes, originatedTxsFee)
}

// SealEpoch is a paid mutator transaction binding the contract method 0xebdf104c.
//
// Solidity: function sealEpoch(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee) returns()
func (_Contract *ContractSession) SealEpoch(offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpoch(&_Contract.TransactOpts, offlineTimes, offlineBlocks, uptimes, originatedTxsFee)
}

// SealEpoch is a paid mutator transaction binding the contract method 0xebdf104c.
//
// Solidity: function sealEpoch(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee) returns()
func (_Contract *ContractTransactorSession) SealEpoch(offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpoch(&_Contract.TransactOpts, offlineTimes, offlineBlocks, uptimes, originatedTxsFee)
}

// SealEpochV1 is a paid mutator transaction binding the contract method 0x7f52e13e.
//
// Solidity: function sealEpochV1(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee, uint256 usedGas) returns()
func (_Contract *ContractTransactor) SealEpochV1(opts *bind.TransactOpts, offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int, usedGas *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "sealEpochV1", offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
}

// SealEpochV1 is a paid mutator transaction binding the contract method 0x7f52e13e.
//
// Solidity: function sealEpochV1(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee, uint256 usedGas) returns()
func (_Contract *ContractSession) SealEpochV1(offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int, usedGas *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpochV1(&_Contract.TransactOpts, offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
}

// SealEpochV1 is a paid mutator transaction binding the contract method 0x7f52e13e.
//
// Solidity: function sealEpochV1(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee, uint256 usedGas) returns()
func (_Contract *ContractTransactorSession) SealEpochV1(offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int, usedGas *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpochV1(&_Contract.TransactOpts, offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
}

// SealEpochValidators is a paid mutator transaction binding the contract method 0xe08d7e66.
//
// Solidity: function sealEpochValidators(uint256[] nextValidatorIDs) returns()
func (_Contract *ContractTransactor) SealEpochValidators(opts *bind.TransactOpts, nextValidatorIDs []*big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "sealEpochValidators", nextValidatorIDs)
}

// SealEpochValidators is a paid mutator transaction binding the contract method 0xe08d7e66.
//
// Solidity: function sealEpochValidators(uint256[] nextValidatorIDs) returns()
func (_Contract *ContractSession) SealEpochValidators(nextValidatorIDs []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpochValidators(&_Contract.TransactOpts, nextValidatorIDs)
}

// SealEpochValidators is a paid mutator transaction binding the contract method 0xe08d7e66.
//
// Solidity: function sealEpochValidators(uint256[] nextValidatorIDs) returns()
func (_Contract *ContractTransactorSession) SealEpochValidators(nextValidatorIDs []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpochValidators(&_Contract.TransactOpts, nextValidatorIDs)
}

// SetBackend is a paid mutator transaction binding the contract method 0xda7fc24f.
//
// Solidity: function setBackend(address _backend) returns()
func (_Contract *ContractTransactor) SetBackend(opts *bind.TransactOpts, _backend common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBackend", _backend)
}

// SetBackend is a paid mutator transaction binding the contract method 0xda7fc24f.
//
// Solidity: function setBackend(address _backend) returns()
func (_Contract *ContractSession) SetBackend(_backend common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetBackend(&_Contract.TransactOpts, _backend)
}

// SetBackend is a paid mutator transaction binding the contract method 0xda7fc24f.
//
// Solidity: function setBackend(address _backend) returns()
func (_Contract *ContractTransactorSession) SetBackend(_backend common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SetBackend(&_Contract.TransactOpts, _backend)
}

// SetBalance is a paid mutator transaction binding the contract method 0xe30443bc.
//
// Solidity: function setBalance(address acc, uint256 value) returns()
func (_Contract *ContractTransactor) SetBalance(opts *bind.TransactOpts, acc common.Address, value *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setBalance", acc, value)
}

// SetBalance is a paid mutator transaction binding the contract method 0xe30443bc.
//
// Solidity: function setBalance(address acc, uint256 value) returns()
func (_Contract *ContractSession) SetBalance(acc common.Address, value *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetBalance(&_Contract.TransactOpts, acc, value)
}

// SetBalance is a paid mutator transaction binding the contract method 0xe30443bc.
//
// Solidity: function setBalance(address acc, uint256 value) returns()
func (_Contract *ContractTransactorSession) SetBalance(acc common.Address, value *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetBalance(&_Contract.TransactOpts, acc, value)
}

// SetGenesisDelegation is a paid mutator transaction binding the contract method 0x18f628d4.
//
// Solidity: function setGenesisDelegation(address delegator, uint256 toValidatorID, uint256 stake, uint256 lockedStake, uint256 lockupFromEpoch, uint256 lockupEndTime, uint256 lockupDuration, uint256 earlyUnlockPenalty, uint256 rewards) returns()
func (_Contract *ContractTransactor) SetGenesisDelegation(opts *bind.TransactOpts, delegator common.Address, toValidatorID *big.Int, stake *big.Int, lockedStake *big.Int, lockupFromEpoch *big.Int, lockupEndTime *big.Int, lockupDuration *big.Int, earlyUnlockPenalty *big.Int, rewards *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setGenesisDelegation", delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
}

// SetGenesisDelegation is a paid mutator transaction binding the contract method 0x18f628d4.
//
// Solidity: function setGenesisDelegation(address delegator, uint256 toValidatorID, uint256 stake, uint256 lockedStake, uint256 lockupFromEpoch, uint256 lockupEndTime, uint256 lockupDuration, uint256 earlyUnlockPenalty, uint256 rewards) returns()
func (_Contract *ContractSession) SetGenesisDelegation(delegator common.Address, toValidatorID *big.Int, stake *big.Int, lockedStake *big.Int, lockupFromEpoch *big.Int, lockupEndTime *big.Int, lockupDuration *big.Int, earlyUnlockPenalty *big.Int, rewards *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisDelegation(&_Contract.TransactOpts, delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
}

// SetGenesisDelegation is a paid mutator transaction binding the contract method 0x18f628d4.
//
// Solidity: function setGenesisDelegation(address delegator, uint256 toValidatorID, uint256 stake, uint256 lockedStake, uint256 lockupFromEpoch, uint256 lockupEndTime, uint256 lockupDuration, uint256 earlyUnlockPenalty, uint256 rewards) returns()
func (_Contract *ContractTransactorSession) SetGenesisDelegation(delegator common.Address, toValidatorID *big.Int, stake *big.Int, lockedStake *big.Int, lockupFromEpoch *big.Int, lockupEndTime *big.Int, lockupDuration *big.Int, earlyUnlockPenalty *big.Int, rewards *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisDelegation(&_Contract.TransactOpts, delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
}

// SetGenesisValidator is a paid mutator transaction binding the contract method 0x4feb92f3.
//
// Solidity: function setGenesisValidator(address _auth, uint256 validatorID, bytes pubkey, uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime) returns()
func (_Contract *ContractTransactor) SetGenesisValidator(opts *bind.TransactOpts, _auth common.Address, validatorID *big.Int, pubkey []byte, status *big.Int, createdEpoch *big.Int, createdTime *big.Int, deactivatedEpoch *big.Int, deactivatedTime *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setGenesisValidator", _auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
}

// SetGenesisValidator is a paid mutator transaction binding the contract method 0x4feb92f3.
//
// Solidity: function setGenesisValidator(address _auth, uint256 validatorID, bytes pubkey, uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime) returns()
func (_Contract *ContractSession) SetGenesisValidator(_auth common.Address, validatorID *big.Int, pubkey []byte, status *big.Int, createdEpoch *big.Int, createdTime *big.Int, deactivatedEpoch *big.Int, deactivatedTime *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisValidator(&_Contract.TransactOpts, _auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
}

// SetGenesisValidator is a paid mutator transaction binding the contract method 0x4feb92f3.
//
// Solidity: function setGenesisValidator(address _auth, uint256 validatorID, bytes pubkey, uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime) returns()
func (_Contract *ContractTransactorSession) SetGenesisValidator(_auth common.Address, validatorID *big.Int, pubkey []byte, status *big.Int, createdEpoch *big.Int, createdTime *big.Int, deactivatedEpoch *big.Int, deactivatedTime *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisValidator(&_Contract.TransactOpts, _auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
}

// SetStorage is a paid mutator transaction binding the contract method 0x39e503ab.
//
// Solidity: function setStorage(address acc, bytes32 key, bytes32 value) returns()
func (_Contract *ContractTransactor) SetStorage(opts *bind.TransactOpts, acc common.Address, key [32]byte, value [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setStorage", acc, key, value)
}

// SetStorage is a paid mutator transaction binding the contract method 0x39e503ab.
//
// Solidity: function setStorage(address acc, bytes32 key, bytes32 value) returns()
func (_Contract *ContractSession) SetStorage(acc common.Address, key [32]byte, value [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.SetStorage(&_Contract.TransactOpts, acc, key, value)
}

// SetStorage is a paid mutator transaction binding the contract method 0x39e503ab.
//
// Solidity: function setStorage(address acc, bytes32 key, bytes32 value) returns()
func (_Contract *ContractTransactorSession) SetStorage(acc common.Address, key [32]byte, value [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.SetStorage(&_Contract.TransactOpts, acc, key, value)
}

// SwapCode is a paid mutator transaction binding the contract method 0x07690b2a.
//
// Solidity: function swapCode(address acc, address with) returns()
func (_Contract *ContractTransactor) SwapCode(opts *bind.TransactOpts, acc common.Address, with common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "swapCode", acc, with)
}

// SwapCode is a paid mutator transaction binding the contract method 0x07690b2a.
//
// Solidity: function swapCode(address acc, address with) returns()
func (_Contract *ContractSession) SwapCode(acc common.Address, with common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SwapCode(&_Contract.TransactOpts, acc, with)
}

// SwapCode is a paid mutator transaction binding the contract method 0x07690b2a.
//
// Solidity: function swapCode(address acc, address with) returns()
func (_Contract *ContractTransactorSession) SwapCode(acc common.Address, with common.Address) (*types.Transaction, error) {
	return _Contract.Contract.SwapCode(&_Contract.TransactOpts, acc, with)
}

// UpdateNetworkRules is a paid mutator transaction binding the contract method 0xb9cc6b1c.
//
// Solidity: function updateNetworkRules(bytes diff) returns()
func (_Contract *ContractTransactor) UpdateNetworkRules(opts *bind.TransactOpts, diff []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateNetworkRules", diff)
}

// UpdateNetworkRules is a paid mutator transaction binding the contract method 0xb9cc6b1c.
//
// Solidity: function updateNetworkRules(bytes diff) returns()
func (_Contract *ContractSession) UpdateNetworkRules(diff []byte) (*types.Transaction, error) {
	return _Contract.Contract.UpdateNetworkRules(&_Contract.TransactOpts, diff)
}

// UpdateNetworkRules is a paid mutator transaction binding the contract method 0xb9cc6b1c.
//
// Solidity: function updateNetworkRules(bytes diff) returns()
func (_Contract *ContractTransactorSession) UpdateNetworkRules(diff []byte) (*types.Transaction, error) {
	return _Contract.Contract.UpdateNetworkRules(&_Contract.TransactOpts, diff)
}

// UpdateNetworkVersion is a paid mutator transaction binding the contract method 0x267ab446.
//
// Solidity: function updateNetworkVersion(uint256 version) returns()
func (_Contract *ContractTransactor) UpdateNetworkVersion(opts *bind.TransactOpts, version *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateNetworkVersion", version)
}

// UpdateNetworkVersion is a paid mutator transaction binding the contract method 0x267ab446.
//
// Solidity: function updateNetworkVersion(uint256 version) returns()
func (_Contract *ContractSession) UpdateNetworkVersion(version *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateNetworkVersion(&_Contract.TransactOpts, version)
}

// UpdateNetworkVersion is a paid mutator transaction binding the contract method 0x267ab446.
//
// Solidity: function updateNetworkVersion(uint256 version) returns()
func (_Contract *ContractTransactorSession) UpdateNetworkVersion(version *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateNetworkVersion(&_Contract.TransactOpts, version)
}

// UpdateValidatorPubkey is a paid mutator transaction binding the contract method 0x242a6e3f.
//
// Solidity: function updateValidatorPubkey(uint256 validatorID, bytes pubkey) returns()
func (_Contract *ContractTransactor) UpdateValidatorPubkey(opts *bind.TransactOpts, validatorID *big.Int, pubkey []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateValidatorPubkey", validatorID, pubkey)
}

// UpdateValidatorPubkey is a paid mutator transaction binding the contract method 0x242a6e3f.
//
// Solidity: function updateValidatorPubkey(uint256 validatorID, bytes pubkey) returns()
func (_Contract *ContractSession) UpdateValidatorPubkey(validatorID *big.Int, pubkey []byte) (*types.Transaction, error) {
	return _Contract.Contract.UpdateValidatorPubkey(&_Contract.TransactOpts, validatorID, pubkey)
}

// UpdateValidatorPubkey is a paid mutator transaction binding the contract method 0x242a6e3f.
//
// Solidity: function updateValidatorPubkey(uint256 validatorID, bytes pubkey) returns()
func (_Contract *ContractTransactorSession) UpdateValidatorPubkey(validatorID *big.Int, pubkey []byte) (*types.Transaction, error) {
	return _Contract.Contract.UpdateValidatorPubkey(&_Contract.TransactOpts, validatorID, pubkey)
}

// UpdateValidatorWeight is a paid mutator transaction binding the contract method 0xa4066fbe.
//
// Solidity: function updateValidatorWeight(uint256 validatorID, uint256 value) returns()
func (_Contract *ContractTransactor) UpdateValidatorWeight(opts *bind.TransactOpts, validatorID *big.Int, value *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateValidatorWeight", validatorID, value)
}

// UpdateValidatorWeight is a paid mutator transaction binding the contract method 0xa4066fbe.
//
// Solidity: function updateValidatorWeight(uint256 validatorID, uint256 value) returns()
func (_Contract *ContractSession) UpdateValidatorWeight(validatorID *big.Int, value *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateValidatorWeight(&_Contract.TransactOpts, validatorID, value)
}

// UpdateValidatorWeight is a paid mutator transaction binding the contract method 0xa4066fbe.
//
// Solidity: function updateValidatorWeight(uint256 validatorID, uint256 value) returns()
func (_Contract *ContractTransactorSession) UpdateValidatorWeight(validatorID *big.Int, value *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateValidatorWeight(&_Contract.TransactOpts, validatorID, value)
}

// ContractAdvanceEpochsIterator is returned from FilterAdvanceEpochs and is used to iterate over the raw logs and unpacked data for AdvanceEpochs events raised by the Contract contract.
type ContractAdvanceEpochsIterator struct {
	Event *ContractAdvanceEpochs // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractAdvanceEpochsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractAdvanceEpochs)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractAdvanceEpochs)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractAdvanceEpochsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractAdvanceEpochsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractAdvanceEpochs represents a AdvanceEpochs event raised by the Contract contract.
type ContractAdvanceEpochs struct {
	Num *big.Int
	Raw types.Log // Blockchain specific contextual infos
}

// FilterAdvanceEpochs is a free log retrieval operation binding the contract event 0x0151256d62457b809bbc891b1f81c6dd0b9987552c70ce915b519750cd434dd1.
//
// Solidity: event AdvanceEpochs(uint256 num)
func (_Contract *ContractFilterer) FilterAdvanceEpochs(opts *bind.FilterOpts) (*ContractAdvanceEpochsIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "AdvanceEpochs")
	if err != nil {
		return nil, err
	}
	return &ContractAdvanceEpochsIterator{contract: _Contract.contract, event: "AdvanceEpochs", logs: logs, sub: sub}, nil
}

// WatchAdvanceEpochs is a free log subscription operation binding the contract event 0x0151256d62457b809bbc891b1f81c6dd0b9987552c70ce915b519750cd434dd1.
//
// Solidity: event AdvanceEpochs(uint256 num)
func (_Contract *ContractFilterer) WatchAdvanceEpochs(opts *bind.WatchOpts, sink chan<- *ContractAdvanceEpochs) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "AdvanceEpochs")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractAdvanceEpochs)
				if err := _Contract.contract.UnpackLog(event, "AdvanceEpochs", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAdvanceEpochs is a log parse operation binding the contract event 0x0151256d62457b809bbc891b1f81c6dd0b9987552c70ce915b519750cd434dd1.
//
// Solidity: event AdvanceEpochs(uint256 num)
func (_Contract *ContractFilterer) ParseAdvanceEpochs(log types.Log) (*ContractAdvanceEpochs, error) {
	event := new(ContractAdvanceEpochs)
	if err := _Contract.contract.UnpackLog(event, "AdvanceEpochs", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateNetworkRulesIterator is returned from FilterUpdateNetworkRules and is used to iterate over the raw logs and unpacked data for UpdateNetworkRules events raised by the Contract contract.
type ContractUpdateNetworkRulesIterator struct {
	Event *ContractUpdateNetworkRules // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdateNetworkRulesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateNetworkRules)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdateNetworkRules)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdateNetworkRulesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateNetworkRulesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateNetworkRules represents a UpdateNetworkRules event raised by the Contract contract.
type ContractUpdateNetworkRules struct {
	Diff []byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterUpdateNetworkRules is a free log retrieval operation binding the contract event 0x47d10eed096a44e3d0abc586c7e3a5d6cb5358cc90e7d437cd0627f7e765fb99.
//
// Solidity: event UpdateNetworkRules(bytes diff)
func (_Contract *ContractFilterer) FilterUpdateNetworkRules(opts *bind.FilterOpts) (*ContractUpdateNetworkRulesIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateNetworkRules")
	if err != nil {
		return nil, err
	}
	return &ContractUpdateNetworkRulesIterator{contract: _Contract.contract, event: "UpdateNetworkRules", logs: logs, sub: sub}, nil
}

// WatchUpdateNetworkRules is a free log subscription operation binding the contract event 0x47d10eed096a44e3d0abc586c7e3a5d6cb5358cc90e7d437cd0627f7e765fb99.
//
// Solidity: event UpdateNetworkRules(bytes diff)
func (_Contract *ContractFilterer) WatchUpdateNetworkRules(opts *bind.WatchOpts, sink chan<- *ContractUpdateNetworkRules) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateNetworkRules")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateNetworkRules)
				if err := _Contract.contract.UnpackLog(event, "UpdateNetworkRules", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdateNetworkRules is a log parse operation binding the contract event 0x47d10eed096a44e3d0abc586c7e3a5d6cb5358cc90e7d437cd0627f7e765fb99.
//
// Solidity: event UpdateNetworkRules(bytes diff)
func (_Contract *ContractFilterer) ParseUpdateNetworkRules(log types.Log) (*ContractUpdateNetworkRules, error) {
	event := new(ContractUpdateNetworkRules)
	if err := _Contract.contract.UnpackLog(event, "UpdateNetworkRules", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateNetworkVersionIterator is returned from FilterUpdateNetworkVersion and is used to iterate over the raw logs and unpacked data for UpdateNetworkVersion events raised by the Contract contract.
type ContractUpdateNetworkVersionIterator struct {
	Event *ContractUpdateNetworkVersion // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdateNetworkVersionIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateNetworkVersion)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdateNetworkVersion)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdateNetworkVersionIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateNetworkVersionIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateNetworkVersion represents a UpdateNetworkVersion event raised by the Contract contract.
type ContractUpdateNetworkVersion struct {
	Version *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUpdateNetworkVersion is a free log retrieval operation binding the contract event 0x2ccdfd47cf0c1f1069d949f1789bb79b2f12821f021634fc835af1de66ea2feb.
//
// Solidity: event UpdateNetworkVersion(uint256 version)
func (_Contract *ContractFilterer) FilterUpdateNetworkVersion(opts *bind.FilterOpts) (*ContractUpdateNetworkVersionIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateNetworkVersion")
	if err != nil {
		return nil, err
	}
	return &ContractUpdateNetworkVersionIterator{contract: _Contract.contract, event: "UpdateNetworkVersion", logs: logs, sub: sub}, nil
}

// WatchUpdateNetworkVersion is a free log subscription operation binding the contract event 0x2ccdfd47cf0c1f1069d949f1789bb79b2f12821f021634fc835af1de66ea2feb.
//
// Solidity: event UpdateNetworkVersion(uint256 version)
func (_Contract *ContractFilterer) WatchUpdateNetworkVersion(opts *bind.WatchOpts, sink chan<- *ContractUpdateNetworkVersion) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateNetworkVersion")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateNetworkVersion)
				if err := _Contract.contract.UnpackLog(event, "UpdateNetworkVersion", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdateNetworkVersion is a log parse operation binding the contract event 0x2ccdfd47cf0c1f1069d949f1789bb79b2f12821f021634fc835af1de66ea2feb.
//
// Solidity: event UpdateNetworkVersion(uint256 version)
func (_Contract *ContractFilterer) ParseUpdateNetworkVersion(log types.Log) (*ContractUpdateNetworkVersion, error) {
	event := new(ContractUpdateNetworkVersion)
	if err := _Contract.contract.UnpackLog(event, "UpdateNetworkVersion", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateValidatorPubkeyIterator is returned from FilterUpdateValidatorPubkey and is used to iterate over the raw logs and unpacked data for UpdateValidatorPubkey events raised by the Contract contract.
type ContractUpdateValidatorPubkeyIterator struct {
	Event *ContractUpdateValidatorPubkey // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdateValidatorPubkeyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateValidatorPubkey)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdateValidatorPubkey)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdateValidatorPubkeyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateValidatorPubkeyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateValidatorPubkey represents a UpdateValidatorPubkey event raised by the Contract contract.
type ContractUpdateValidatorPubkey struct {
	ValidatorID *big.Int
	Pubkey      []byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUpdateValidatorPubkey is a free log retrieval operation binding the contract event 0x0f0ef1ab97439def0a9d2c6d9dc166207f1b13b99e62b442b2993d6153c63a6e.
//
// Solidity: event UpdateValidatorPubkey(uint256 indexed validatorID, bytes pubkey)
func (_Contract *ContractFilterer) FilterUpdateValidatorPubkey(opts *bind.FilterOpts, validatorID []*big.Int) (*ContractUpdateValidatorPubkeyIterator, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateValidatorPubkey", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractUpdateValidatorPubkeyIterator{contract: _Contract.contract, event: "UpdateValidatorPubkey", logs: logs, sub: sub}, nil
}

// WatchUpdateValidatorPubkey is a free log subscription operation binding the contract event 0x0f0ef1ab97439def0a9d2c6d9dc166207f1b13b99e62b442b2993d6153c63a6e.
//
// Solidity: event UpdateValidatorPubkey(uint256 indexed validatorID, bytes pubkey)
func (_Contract *ContractFilterer) WatchUpdateValidatorPubkey(opts *bind.WatchOpts, sink chan<- *ContractUpdateValidatorPubkey, validatorID []*big.Int) (event.Subscription, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateValidatorPubkey", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateValidatorPubkey)
				if err := _Contract.contract.UnpackLog(event, "UpdateValidatorPubkey", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdateValidatorPubkey is a log parse operation binding the contract event 0x0f0ef1ab97439def0a9d2c6d9dc166207f1b13b99e62b442b2993d6153c63a6e.
//
// Solidity: event UpdateValidatorPubkey(uint256 indexed validatorID, bytes pubkey)
func (_Contract *ContractFilterer) ParseUpdateValidatorPubkey(log types.Log) (*ContractUpdateValidatorPubkey, error) {
	event := new(ContractUpdateValidatorPubkey)
	if err := _Contract.contract.UnpackLog(event, "UpdateValidatorPubkey", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdateValidatorWeightIterator is returned from FilterUpdateValidatorWeight and is used to iterate over the raw logs and unpacked data for UpdateValidatorWeight events raised by the Contract contract.
type ContractUpdateValidatorWeightIterator struct {
	Event *ContractUpdateValidatorWeight // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdateValidatorWeightIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdateValidatorWeight)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdateValidatorWeight)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdateValidatorWeightIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdateValidatorWeightIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdateValidatorWeight represents a UpdateValidatorWeight event raised by the Contract contract.
type ContractUpdateValidatorWeight struct {
	ValidatorID *big.Int
	Weight      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUpdateValidatorWeight is a free log retrieval operation binding the contract event 0xb975807576e3b1461be7de07ebf7d20e4790ed802d7a0c4fdd0a1a13df72a935.
//
// Solidity: event UpdateValidatorWeight(uint256 indexed validatorID, uint256 weight)
func (_Contract *ContractFilterer) FilterUpdateValidatorWeight(opts *bind.FilterOpts, validatorID []*big.Int) (*ContractUpdateValidatorWeightIterator, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdateValidatorWeight", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractUpdateValidatorWeightIterator{contract: _Contract.contract, event: "UpdateValidatorWeight", logs: logs, sub: sub}, nil
}

// WatchUpdateValidatorWeight is a free log subscription operation binding the contract event 0xb975807576e3b1461be7de07ebf7d20e4790ed802d7a0c4fdd0a1a13df72a935.
//
// Solidity: event UpdateValidatorWeight(uint256 indexed validatorID, uint256 weight)
func (_Contract *ContractFilterer) WatchUpdateValidatorWeight(opts *bind.WatchOpts, sink chan<- *ContractUpdateValidatorWeight, validatorID []*big.Int) (event.Subscription, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdateValidatorWeight", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdateValidatorWeight)
				if err := _Contract.contract.UnpackLog(event, "UpdateValidatorWeight", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdateValidatorWeight is a log parse operation binding the contract event 0xb975807576e3b1461be7de07ebf7d20e4790ed802d7a0c4fdd0a1a13df72a935.
//
// Solidity: event UpdateValidatorWeight(uint256 indexed validatorID, uint256 weight)
func (_Contract *ContractFilterer) ParseUpdateValidatorWeight(log types.Log) (*ContractUpdateValidatorWeight, error) {
	event := new(ContractUpdateValidatorWeight)
	if err := _Contract.contract.UnpackLog(event, "UpdateValidatorWeight", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractUpdatedBackendIterator is returned from FilterUpdatedBackend and is used to iterate over the raw logs and unpacked data for UpdatedBackend events raised by the Contract contract.
type ContractUpdatedBackendIterator struct {
	Event *ContractUpdatedBackend // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdatedBackendIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdatedBackend)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdatedBackend)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdatedBackendIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdatedBackendIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdatedBackend represents a UpdatedBackend event raised by the Contract contract.
type ContractUpdatedBackend struct {
	Backend common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUpdatedBackend is a free log retrieval operation binding the contract event 0x64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef245035830693.
//
// Solidity: event UpdatedBackend(address indexed backend)
func (_Contract *ContractFilterer) FilterUpdatedBackend(opts *bind.FilterOpts, backend []common.Address) (*ContractUpdatedBackendIterator, error) {

	var backendRule []interface{}
	for _, backendItem := range backend {
		backendRule = append(backendRule, backendItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdatedBackend", backendRule)
	if err != nil {
		return nil, err
	}
	return &ContractUpdatedBackendIterator{contract: _Contract.contract, event: "UpdatedBackend", logs: logs, sub: sub}, nil
}

// WatchUpdatedBackend is a free log subscription operation binding the contract event 0x64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef245035830693.
//
// Solidity: event UpdatedBackend(address indexed backend)
func (_Contract *ContractFilterer) WatchUpdatedBackend(opts *bind.WatchOpts, sink chan<- *ContractUpdatedBackend, backend []common.Address) (event.Subscription, error) {

	var backendRule []interface{}
	for _, backendItem := range backend {
		backendRule = append(backendRule, backendItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdatedBackend", backendRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdatedBackend)
				if err := _Contract.contract.UnpackLog(event, "UpdatedBackend", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdatedBackend is a log parse operation binding the contract event 0x64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef245035830693.
//
// Solidity: event UpdatedBackend(address indexed backend)
func (_Contract *ContractFilterer) ParseUpdatedBackend(log types.Log) (*ContractUpdatedBackend, error) {
	event := new(ContractUpdatedBackend)
	if err := _Contract.contract.UnpackLog(event, "UpdatedBackend", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

var ContractBinRuntime = "0x608060405234801561001057600080fd5b50600436106101365760003560e01c806379bead38116100b2578063d6a0c7af11610081578063e08d7e6611610066578063e08d7e6614610676578063e30443bc146106e6578063ebdf104c1461071f57610136565b8063d6a0c7af14610608578063da7fc24f1461064357610136565b806379bead38146103d65780637f52e13e1461040f578063a4066fbe14610575578063b9cc6b1c1461059857610136565b8063242a6e3f1161010957806339e503ab116100ee57806339e503ab146102b1578063485cc955146102f05780634feb92f31461032b57610136565b8063242a6e3f1461021d578063267ab4461461029457610136565b806307690b2a1461013b5780630aeeca001461017857806318f628d4146101955780631e702f83146101fa575b600080fd5b6101766004803603604081101561015157600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516610885565b005b6101766004803603602081101561018e57600080fd5b5035610989565b61017660048036036101208110156101ac57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e0810135906101000135610a2b565b6101766004803603604081101561021057600080fd5b5080359060200135610b4f565b6101766004803603604081101561023357600080fd5b8135919081019060408101602082013564010000000081111561025557600080fd5b82018360208201111561026757600080fd5b8035906020019184600183028401116401000000008311171561028957600080fd5b509092509050610c1c565b610176600480360360208110156102aa57600080fd5b5035610cee565b610176600480360360608110156102c757600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060208101359060400135610d90565b6101766004803603604081101561030657600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516610e9b565b610176600480360361010081101561034257600080fd5b73ffffffffffffffffffffffffffffffffffffffff8235169160208101359181019060608101604082013564010000000081111561037f57600080fd5b82018360208201111561039157600080fd5b803590602001918460018302840111640100000000831117156103b357600080fd5b919350915080359060208101359060408101359060608101359060800135611043565b610176600480360360408110156103ec57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff813516906020013561119b565b610176600480360360a081101561042557600080fd5b81019060208101813564010000000081111561044057600080fd5b82018360208201111561045257600080fd5b8035906020019184602083028401116401000000008311171561047457600080fd5b91939092909160208101903564010000000081111561049257600080fd5b8201836020820111156104a457600080fd5b803590602001918460208302840111640100000000831117156104c657600080fd5b9193909290916020810190356401000000008111156104e457600080fd5b8201836020820111156104f657600080fd5b8035906020019184602083028401116401000000008311171561051857600080fd5b91939092909160208101903564010000000081111561053657600080fd5b82018360208201111561054857600080fd5b8035906020019184602083028401116401000000008311171561056a57600080fd5b919350915035611282565b6101766004803603604081101561058b57600080fd5b5080359060200135611411565b610176600480360360208110156105ae57600080fd5b8101906020810181356401000000008111156105c957600080fd5b8201836020820111156105db57600080fd5b803590602001918460018302840111640100000000831117156105fd57600080fd5b5090925090506114b7565b6101766004803603604081101561061e57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff81358116916020013516611587565b6101766004803603602081101561065957600080fd5b503573ffffffffffffffffffffffffffffffffffffffff1661166f565b6101766004803603602081101561068c57600080fd5b8101906020810181356401000000008111156106a757600080fd5b8201836020820111156106b957600080fd5b803590602001918460208302840111640100000000831117156106db57600080fd5b509092509050611763565b610176600480360360408110156106fc57600080fd5b5073ffffffffffffffffffffffffffffffffffffffff8135169060200135611859565b6101766004803603608081101561073557600080fd5b81019060208101813564010000000081111561075057600080fd5b82018360208201111561076257600080fd5b8035906020019184602083028401116401000000008311171561078457600080fd5b9193909290916020810190356401000000008111156107a257600080fd5b8201836020820111156107b457600080fd5b803590602001918460208302840111640100000000831117156107d657600080fd5b9193909290916020810190356401000000008111156107f457600080fd5b82018360208201111561080657600080fd5b8035906020019184602083028401116401000000008311171561082857600080fd5b91939092909160208101903564010000000081111561084657600080fd5b82018360208201111561085857600080fd5b8035906020019184602083028401116401000000008311171561087a57600080fd5b509092509050611940565b60345473ffffffffffffffffffffffffffffffffffffffff1633146108f1576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517f07690b2a00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff85811660048301528481166024830152915191909216916307690b2a91604480830192600092919082900301818387803b15801561096d57600080fd5b505af1158015610981573d6000803e3d6000fd5b505050505050565b60345473ffffffffffffffffffffffffffffffffffffffff1633146109f5576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b6040805182815290517f0151256d62457b809bbc891b1f81c6dd0b9987552c70ce915b519750cd434dd19181900360200190a150565b3315610a7e576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603454604080517f18f628d400000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8c81166004830152602482018c9052604482018b9052606482018a90526084820189905260a4820188905260c4820187905260e482018690526101048201859052915191909216916318f628d49161012480830192600092919082900301818387803b158015610b2c57600080fd5b505af1158015610b40573d6000803e3d6000fd5b50505050505050505050505050565b3315610ba2576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603454604080517f1e702f830000000000000000000000000000000000000000000000000000000081526004810185905260248101849052905173ffffffffffffffffffffffffffffffffffffffff90921691631e702f839160448082019260009290919082900301818387803b15801561096d57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff163314610c88576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b827f0f0ef1ab97439def0a9d2c6d9dc166207f1b13b99e62b442b2993d6153c63a6e838360405180806020018281038252848482818152602001925080828437600083820152604051601f909101601f19169092018290039550909350505050a2505050565b60345473ffffffffffffffffffffffffffffffffffffffff163314610d5a576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b6040805182815290517f2ccdfd47cf0c1f1069d949f1789bb79b2f12821f021634fc835af1de66ea2feb9181900360200190a150565b60345473ffffffffffffffffffffffffffffffffffffffff163314610dfc576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517f39e503ab00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff86811660048301526024820186905260448201859052915191909216916339e503ab91606480830192600092919082900301818387803b158015610e7e57600080fd5b505af1158015610e92573d6000803e3d6000fd5b50505050505050565b600054610100900460ff1680610eb45750610eb4611af5565b80610ec2575060005460ff16155b610efd5760405162461bcd60e51b815260040180806020018281038252602e815260200180611afc602e913960400191505060405180910390fd5b600054610100900460ff16158015610f6357600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b603480547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff85169081179091556040517f64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef24503583069390600090a2603580547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff8416179055801561103e57600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555b505050565b3315611096576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16634feb92f38a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001898152602001806020018781526020018681526020018581526020018481526020018381526020018281038252898982818152602001925080828437600081840152601f19601f8201169050808301925050509a5050505050505050505050600060405180830381600087803b158015610b2c57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff163314611207576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517f79bead3800000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff858116600483015260248201859052915191909216916379bead3891604480830192600092919082900301818387803b15801561096d57600080fd5b33156112d5576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663592fe0c08a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808060200180602001806020018060200186815260200185810385528e8e82818152602001925060200280828437600083820152601f01601f191690910186810385528c8152602090810191508d908d0280828437600083820152601f01601f191690910186810384528a8152602090810191508b908b0280828437600083820152601f01601f19169091018681038352888152602090810191508990890280828437600081840152601f19601f8201169050808301925050509d5050505050505050505050505050600060405180830381600087803b158015610b2c57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff16331461147d576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b60408051828152905183917fb975807576e3b1461be7de07ebf7d20e4790ed802d7a0c4fdd0a1a13df72a935919081900360200190a25050565b60345473ffffffffffffffffffffffffffffffffffffffff163314611523576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b7f47d10eed096a44e3d0abc586c7e3a5d6cb5358cc90e7d437cd0627f7e765fb99828260405180806020018281038252848482818152602001925080828437600083820152604051601f909101601f19169092018290039550909350505050a15050565b60345473ffffffffffffffffffffffffffffffffffffffff1633146115f3576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517fd6a0c7af00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff858116600483015284811660248301529151919092169163d6a0c7af91604480830192600092919082900301818387803b15801561096d57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff1633146116db576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b60405173ffffffffffffffffffffffffffffffffffffffff8216907f64ee8f7bfc37fc205d7194ee3d64947ab7b57e663cd0d1abd3ef24503583069390600090a2603480547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff92909216919091179055565b33156117b6576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b6034546040517fe08d7e660000000000000000000000000000000000000000000000000000000081526020600482018181526024830185905273ffffffffffffffffffffffffffffffffffffffff9093169263e08d7e6692869286929182916044909101908590850280828437600081840152601f19601f8201169050808301925050509350505050600060405180830381600087803b15801561096d57600080fd5b60345473ffffffffffffffffffffffffffffffffffffffff1633146118c5576040805162461bcd60e51b815260206004820152601960248201527f63616c6c6572206973206e6f7420746865206261636b656e6400000000000000604482015290519081900360640190fd5b603554604080517fe30443bc00000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8581166004830152602482018590529151919092169163e30443bc91604480830192600092919082900301818387803b15801561096d57600080fd5b3315611993576040805162461bcd60e51b815260206004820152600c60248201527f6e6f742063616c6c61626c650000000000000000000000000000000000000000604482015290519081900360640190fd5b603460009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663592fe0c0898989898989898963322adc3a6040518a63ffffffff1660e01b8152600401808060200180602001806020018060200186815260200185810385528e8e82818152602001925060200280828437600083820152601f01601f191690910186810385528c8152602090810191508d908d0280828437600083820152601f01601f191690910186810384528a8152602090810191508b908b0280828437600083820152601f01601f19169091018681038352888152602090810191508990890280828437600081840152601f19601f8201169050808301925050509d5050505050505050505050505050600060405180830381600087803b158015611ad357600080fd5b505af1158015611ae7573d6000803e3d6000fd5b505050505050505050505050565b303b159056fe436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a6564a265627a7a72315820962250f3ef56c54b3df8ce7213cfcf494e44135e9c70003b4f86fe5ef1bcbb9f64736f6c63430005110032"
