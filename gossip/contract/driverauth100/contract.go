// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package driverauth100

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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_sfc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_driver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"newDriverAuth\",\"type\":\"address\"}],\"name\":\"migrateTo\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"executable\",\"type\":\"address\"}],\"name\":\"execute\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"executable\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"selfCodeHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"driverCodeHash\",\"type\":\"bytes32\"}],\"name\":\"mutExecute\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"diff\",\"type\":\"uint256\"}],\"name\":\"incBalance\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"upgradeCode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"copyCode\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"acc\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"diff\",\"type\":\"uint256\"}],\"name\":\"incNonce\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"diff\",\"type\":\"bytes\"}],\"name\":\"updateNetworkRules\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minGasPrice\",\"type\":\"uint256\"}],\"name\":\"updateMinGasPrice\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"version\",\"type\":\"uint256\"}],\"name\":\"updateNetworkVersion\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"advanceEpochs\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"updateValidatorWeight\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"updateValidatorPubkey\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_auth\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"setGenesisValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupFromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupEndTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewards\",\"type\":\"uint256\"}],\"name\":\"setGenesisDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"deactivateValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nextValidatorIDs\",\"type\":\"uint256[]\"}],\"name\":\"sealEpochValidators\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256\",\"name\":\"usedGas\",\"type\":\"uint256\"}],\"name\":\"sealEpoch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50612269806100206000396000f3fe608060405234801561001057600080fd5b506004361061018d5760003560e01c806366e7ea0f116100e3578063b9cc6b1c1161008c578063e08d7e6611610066578063e08d7e6614610702578063f2fde38b14610772578063fd1b6ec1146107985761018d565b8063b9cc6b1c1461062c578063c0c53b8b1461069c578063d6a0c7af146106d45761018d565b80638da5cb5b116100bd5780638da5cb5b146105c95780638f32d59b146105ed578063a4066fbe146106095761018d565b806366e7ea0f14610569578063715018a61461059557806379bead381461059d5761018d565b8063242a6e3f116101455780634ddaf8f21161011f5780634ddaf8f21461033f5780634feb92f314610365578063592fe0c0146104035761018d565b8063242a6e3f14610285578063267ab446146102fc5780634b64e492146103195761018d565b806318f628d41161017657806318f628d4146101ce5780631cef4fab146102265780631e702f83146102625761018d565b806307aaf344146101925780630aeeca00146101b1575b600080fd5b6101af600480360360208110156101a857600080fd5b50356107c6565b005b6101af600480360360208110156101c757600080fd5b5035610969565b6101af60048036036101208110156101e557600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e0810135906101000135610a28565b6101af6004803603608081101561023c57600080fd5b506001600160a01b03813581169160208101359091169060408101359060600135610b35565b6101af6004803603604081101561027857600080fd5b5080359060200135610ba0565b6101af6004803603604081101561029b57600080fd5b813591908101906040810160208201356401000000008111156102bd57600080fd5b8201836020820111156102cf57600080fd5b803590602001918460018302840111640100000000831117156102f157600080fd5b509092509050610c72565b6101af6004803603602081101561031257600080fd5b5035610d87565b6101af6004803603602081101561032f57600080fd5b50356001600160a01b0316610e46565b6101af6004803603602081101561035557600080fd5b50356001600160a01b0316610ed1565b6101af600480360361010081101561037c57600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156103ac57600080fd5b8201836020820111156103be57600080fd5b803590602001918460018302840111640100000000831117156103e057600080fd5b919350915080359060208101359060408101359060608101359060800135610f91565b6101af600480360360a081101561041957600080fd5b81019060208101813564010000000081111561043457600080fd5b82018360208201111561044657600080fd5b8035906020019184602083028401116401000000008311171561046857600080fd5b91939092909160208101903564010000000081111561048657600080fd5b82018360208201111561049857600080fd5b803590602001918460208302840111640100000000831117156104ba57600080fd5b9193909290916020810190356401000000008111156104d857600080fd5b8201836020820111156104ea57600080fd5b8035906020019184602083028401116401000000008311171561050c57600080fd5b91939092909160208101903564010000000081111561052a57600080fd5b82018360208201111561053c57600080fd5b8035906020019184602083028401116401000000008311171561055e57600080fd5b9193509150356110ab565b6101af6004803603604081101561057f57600080fd5b506001600160a01b038135169060200135611270565b6101af611394565b6101af600480360360408110156105b357600080fd5b506001600160a01b03813516906020013561144f565b6105d1611516565b604080516001600160a01b039092168252519081900360200190f35b6105f5611525565b604080519115158252519081900360200190f35b6101af6004803603604081101561061f57600080fd5b5080359060200135611536565b6101af6004803603602081101561064257600080fd5b81019060208101813564010000000081111561065d57600080fd5b82018360208201111561066f57600080fd5b8035906020019184600183028401116401000000008311171561069157600080fd5b509092509050611602565b6101af600480360360608110156106b257600080fd5b506001600160a01b0381358116916020810135821691604090910135166116eb565b6101af600480360360408110156106ea57600080fd5b506001600160a01b0381358116916020013516611838565b6101af6004803603602081101561071857600080fd5b81019060208101813564010000000081111561073357600080fd5b82018360208201111561074557600080fd5b8035906020019184602083028401116401000000008311171561076757600080fd5b509092509050611900565b6101af6004803603602081101561078857600080fd5b50356001600160a01b03166119df565b6101af600480360360408110156107ae57600080fd5b506001600160a01b0381358116916020013516611a41565b6066546001600160a01b03163314610825576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b60675460408051808201909152601a81527f7b2245636f6e6f6d79223a7b224d696e4761735072696365223a00000000000060208201526001600160a01b039091169063b9cc6b1c906108b69061087b85611b04565b6040518060400160405280600281526020017f7d7d000000000000000000000000000000000000000000000000000000000000815250611c28565b6040518263ffffffff1660e01b81526004018080602001828103825283818151815260200191508051906020019080838360005b838110156109025781810151838201526020016108ea565b50505050905090810190601f16801561092f5780820380516001836020036101000a031916815260200191505b5092505050600060405180830381600087803b15801561094e57600080fd5b505af1158015610962573d6000803e3d6000fd5b5050505050565b610971611525565b6109c2576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517f0aeeca000000000000000000000000000000000000000000000000000000000081526004810184905290516001600160a01b0390921691630aeeca009160248082019260009290919082900301818387803b15801561094e57600080fd5b6067546001600160a01b03163314610a715760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606654604080517f18f628d40000000000000000000000000000000000000000000000000000000081526001600160a01b038c81166004830152602482018c9052604482018b9052606482018a90526084820189905260a4820188905260c4820187905260e482018690526101048201859052915191909216916318f628d49161012480830192600092919082900301818387803b158015610b1257600080fd5b505af1158015610b26573d6000803e3d6000fd5b50505050505050505050505050565b610b3d611525565b610b8e576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b610b9a84848484611dc6565b50505050565b6067546001600160a01b03163314610be95760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606654604080517f1e702f83000000000000000000000000000000000000000000000000000000008152600481018590526024810184905290516001600160a01b0390921691631e702f839160448082019260009290919082900301818387803b158015610c5657600080fd5b505af1158015610c6a573d6000803e3d6000fd5b505050505050565b6066546001600160a01b03163314610cd1576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b606754604080517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810186815260248201928352604482018590526001600160a01b039093169263242a6e3f928792879287929091606401848480828437600081840152601f19601f820116905080830192505050945050505050600060405180830381600087803b158015610d6a57600080fd5b505af1158015610d7e573d6000803e3d6000fd5b50505050505050565b610d8f611525565b610de0576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517f267ab4460000000000000000000000000000000000000000000000000000000081526004810184905290516001600160a01b039092169163267ab4469160248082019260009290919082900301818387803b15801561094e57600080fd5b610e4e611525565b610e9f576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b610ece81610eab611516565b610eb430611ef0565b606754610ec9906001600160a01b0316611ef0565b611dc6565b50565b610ed9611525565b610f2a576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517fda7fc24f0000000000000000000000000000000000000000000000000000000081526001600160a01b0384811660048301529151919092169163da7fc24f91602480830192600092919082900301818387803b15801561094e57600080fd5b6067546001600160a01b03163314610fda5760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606660009054906101000a90046001600160a01b03166001600160a01b0316634feb92f38a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808a6001600160a01b03166001600160a01b03168152602001898152602001806020018781526020018681526020018581526020018481526020018381526020018281038252898982818152602001925080828437600081840152601f19601f8201169050808301925050509a5050505050505050505050600060405180830381600087803b158015610b1257600080fd5b6067546001600160a01b031633146110f45760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606660009054906101000a90046001600160a01b03166001600160a01b031663592fe0c08a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808060200180602001806020018060200186815260200185810385528e8e82818152602001925060200280828437600083820152601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe01690910186810385528c8152602090810191508d908d0280828437600083820152601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe01690910186810384528a8152602090810191508b908b0280828437600083820152601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169091018681038352888152602090810191508990890280828437600081840152601f19601f8201169050808301925050509d5050505050505050505050505050600060405180830381600087803b158015610b1257600080fd5b6066546001600160a01b031633146112cf576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b6066546001600160a01b0383811691161461131b5760405162461bcd60e51b81526004018080602001828103825260218152602001806121ef6021913960400191505060405180910390fd5b6067546001600160a01b039081169063e30443bc908490611345908216318563ffffffff611ef416565b6040518363ffffffff1660e01b815260040180836001600160a01b03166001600160a01b0316815260200182815260200192505050600060405180830381600087803b158015610c5657600080fd5b61139c611525565b6113ed576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055565b611457611525565b6114a8576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517f79bead380000000000000000000000000000000000000000000000000000000081526001600160a01b03858116600483015260248201859052915191909216916379bead3891604480830192600092919082900301818387803b158015610c5657600080fd5b6033546001600160a01b031690565b6033546001600160a01b0316331490565b6066546001600160a01b03163314611595576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b606754604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018590526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b158015610c5657600080fd5b61160a611525565b61165b576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6067546040517fb9cc6b1c000000000000000000000000000000000000000000000000000000008152602060048201908152602482018490526001600160a01b039092169163b9cc6b1c91859185918190604401848480828437600081840152601f19601f8201169050808301925050509350505050600060405180830381600087803b158015610c5657600080fd5b600054610100900460ff16806117045750611704611f55565b80611712575060005460ff16155b61174d5760405162461bcd60e51b815260040180806020018281038252602e8152602001806121c1602e913960400191505060405180910390fd5b600054610100900460ff161580156117b357600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b6117bc82611f5b565b606780546001600160a01b038086167fffffffffffffffffffffffff00000000000000000000000000000000000000009283161790925560668054928716929091169190911790558015610b9a57600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff16905550505050565b611840611525565b611891576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517fd6a0c7af0000000000000000000000000000000000000000000000000000000081526001600160a01b03858116600483015284811660248301529151919092169163d6a0c7af91604480830192600092919082900301818387803b158015610c5657600080fd5b6067546001600160a01b031633146119495760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b6066546040517fe08d7e66000000000000000000000000000000000000000000000000000000008152602060048201818152602483018590526001600160a01b039093169263e08d7e6692869286929182916044909101908590850280828437600081840152601f19601f8201169050808301925050509350505050600060405180830381600087803b158015610c5657600080fd5b6119e7611525565b611a38576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b610ece816120bd565b611a49611525565b611a9a576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b611aa382612176565b8015611ab35750611ab381612176565b611891576040805162461bcd60e51b815260206004820152600e60248201527f6e6f74206120636f6e7472616374000000000000000000000000000000000000604482015290519081900360640190fd5b606081611b45575060408051808201909152600181527f30000000000000000000000000000000000000000000000000000000000000006020820152611c23565b6000611b508361217c565b90506060816040519080825280601f01601f191660200182016040528015611b7f576020820181803883390190505b5090507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82015b8415611c1e57600a850660300160f81b828281518110611bc257fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600a850494507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01611ba6565b509150505b919050565b60608084905060608490506060849050606081518351855101016040519080825280601f01601f191660200182016040528015611c6c576020820181803883390190505b509050806000805b8651811015611cdd57868181518110611c8957fe5b602001015160f81c60f81b838380600101945081518110611ca657fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600101611c74565b5060005b8551811015611d4a57858181518110611cf657fe5b602001015160f81c60f81b838380600101945081518110611d1357fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600101611ce1565b5060005b8451811015611db757848181518110611d6357fe5b602001015160f81c60f81b838380600101945081518110611d8057fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600101611d4e565b50909998505050505050505050565b611dcf846120bd565b836001600160a01b031663614619546040518163ffffffff1660e01b8152600401600060405180830381600087803b158015611e0a57600080fd5b505af1158015611e1e573d6000803e3d6000fd5b50505050611e2b836120bd565b81611e3530611ef0565b14611e87576040805162461bcd60e51b815260206004820152601c60248201527f73656c6620636f6465206861736820646f65736e2774206d6174636800000000604482015290519081900360640190fd5b6067548190611e9e906001600160a01b0316611ef0565b14610b9a576040805162461bcd60e51b815260206004820152601e60248201527f64726976657220636f6465206861736820646f65736e2774206d617463680000604482015290519081900360640190fd5b3f90565b600082820183811015611f4e576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b9392505050565b303b1590565b600054610100900460ff1680611f745750611f74611f55565b80611f82575060005460ff16155b611fbd5760405162461bcd60e51b815260040180806020018281038252602e8152602001806121c1602e913960400191505060405180910390fd5b600054610100900460ff1615801561202357600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0384811691909117918290556040519116906000907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a380156120b957600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555b5050565b6001600160a01b0381166121025760405162461bcd60e51b815260040180806020018281038252602681526020018061219b6026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b3b151590565b6000805b821561219457600101600a83049250612180565b9291505056fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f2061646472657373436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a6564726563697069656e74206973206e6f74207468652053464320636f6e747261637463616c6c6572206973206e6f7420746865204e6f646544726976657220636f6e7472616374a265627a7a72315820e32e2bd9d4306bc53a2f1341ed34b7c36a48a87713cbfbf664b6cde5edf9f98e64736f6c63430005110032",
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

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_Contract *ContractCaller) IsOwner(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isOwner")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_Contract *ContractSession) IsOwner() (bool, error) {
	return _Contract.Contract.IsOwner(&_Contract.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_Contract *ContractCallerSession) IsOwner() (bool, error) {
	return _Contract.Contract.IsOwner(&_Contract.CallOpts)
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

// Execute is a paid mutator transaction binding the contract method 0x4b64e492.
//
// Solidity: function execute(address executable) returns()
func (_Contract *ContractTransactor) Execute(opts *bind.TransactOpts, executable common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "execute", executable)
}

// Execute is a paid mutator transaction binding the contract method 0x4b64e492.
//
// Solidity: function execute(address executable) returns()
func (_Contract *ContractSession) Execute(executable common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Execute(&_Contract.TransactOpts, executable)
}

// Execute is a paid mutator transaction binding the contract method 0x4b64e492.
//
// Solidity: function execute(address executable) returns()
func (_Contract *ContractTransactorSession) Execute(executable common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Execute(&_Contract.TransactOpts, executable)
}

// IncBalance is a paid mutator transaction binding the contract method 0x66e7ea0f.
//
// Solidity: function incBalance(address acc, uint256 diff) returns()
func (_Contract *ContractTransactor) IncBalance(opts *bind.TransactOpts, acc common.Address, diff *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "incBalance", acc, diff)
}

// IncBalance is a paid mutator transaction binding the contract method 0x66e7ea0f.
//
// Solidity: function incBalance(address acc, uint256 diff) returns()
func (_Contract *ContractSession) IncBalance(acc common.Address, diff *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.IncBalance(&_Contract.TransactOpts, acc, diff)
}

// IncBalance is a paid mutator transaction binding the contract method 0x66e7ea0f.
//
// Solidity: function incBalance(address acc, uint256 diff) returns()
func (_Contract *ContractTransactorSession) IncBalance(acc common.Address, diff *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.IncBalance(&_Contract.TransactOpts, acc, diff)
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

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _sfc, address _driver, address _owner) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, _sfc common.Address, _driver common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", _sfc, _driver, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _sfc, address _driver, address _owner) returns()
func (_Contract *ContractSession) Initialize(_sfc common.Address, _driver common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _sfc, _driver, _owner)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _sfc, address _driver, address _owner) returns()
func (_Contract *ContractTransactorSession) Initialize(_sfc common.Address, _driver common.Address, _owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, _sfc, _driver, _owner)
}

// MigrateTo is a paid mutator transaction binding the contract method 0x4ddaf8f2.
//
// Solidity: function migrateTo(address newDriverAuth) returns()
func (_Contract *ContractTransactor) MigrateTo(opts *bind.TransactOpts, newDriverAuth common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "migrateTo", newDriverAuth)
}

// MigrateTo is a paid mutator transaction binding the contract method 0x4ddaf8f2.
//
// Solidity: function migrateTo(address newDriverAuth) returns()
func (_Contract *ContractSession) MigrateTo(newDriverAuth common.Address) (*types.Transaction, error) {
	return _Contract.Contract.MigrateTo(&_Contract.TransactOpts, newDriverAuth)
}

// MigrateTo is a paid mutator transaction binding the contract method 0x4ddaf8f2.
//
// Solidity: function migrateTo(address newDriverAuth) returns()
func (_Contract *ContractTransactorSession) MigrateTo(newDriverAuth common.Address) (*types.Transaction, error) {
	return _Contract.Contract.MigrateTo(&_Contract.TransactOpts, newDriverAuth)
}

// MutExecute is a paid mutator transaction binding the contract method 0x1cef4fab.
//
// Solidity: function mutExecute(address executable, address newOwner, bytes32 selfCodeHash, bytes32 driverCodeHash) returns()
func (_Contract *ContractTransactor) MutExecute(opts *bind.TransactOpts, executable common.Address, newOwner common.Address, selfCodeHash [32]byte, driverCodeHash [32]byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "mutExecute", executable, newOwner, selfCodeHash, driverCodeHash)
}

// MutExecute is a paid mutator transaction binding the contract method 0x1cef4fab.
//
// Solidity: function mutExecute(address executable, address newOwner, bytes32 selfCodeHash, bytes32 driverCodeHash) returns()
func (_Contract *ContractSession) MutExecute(executable common.Address, newOwner common.Address, selfCodeHash [32]byte, driverCodeHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.MutExecute(&_Contract.TransactOpts, executable, newOwner, selfCodeHash, driverCodeHash)
}

// MutExecute is a paid mutator transaction binding the contract method 0x1cef4fab.
//
// Solidity: function mutExecute(address executable, address newOwner, bytes32 selfCodeHash, bytes32 driverCodeHash) returns()
func (_Contract *ContractTransactorSession) MutExecute(executable common.Address, newOwner common.Address, selfCodeHash [32]byte, driverCodeHash [32]byte) (*types.Transaction, error) {
	return _Contract.Contract.MutExecute(&_Contract.TransactOpts, executable, newOwner, selfCodeHash, driverCodeHash)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _Contract.Contract.RenounceOwnership(&_Contract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Contract.Contract.RenounceOwnership(&_Contract.TransactOpts)
}

// SealEpoch is a paid mutator transaction binding the contract method 0x592fe0c0.
//
// Solidity: function sealEpoch(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee, uint256 usedGas) returns()
func (_Contract *ContractTransactor) SealEpoch(opts *bind.TransactOpts, offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int, usedGas *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "sealEpoch", offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
}

// SealEpoch is a paid mutator transaction binding the contract method 0x592fe0c0.
//
// Solidity: function sealEpoch(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee, uint256 usedGas) returns()
func (_Contract *ContractSession) SealEpoch(offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int, usedGas *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpoch(&_Contract.TransactOpts, offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
}

// SealEpoch is a paid mutator transaction binding the contract method 0x592fe0c0.
//
// Solidity: function sealEpoch(uint256[] offlineTimes, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee, uint256 usedGas) returns()
func (_Contract *ContractTransactorSession) SealEpoch(offlineTimes []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int, usedGas *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpoch(&_Contract.TransactOpts, offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
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

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferOwnership(&_Contract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferOwnership(&_Contract.TransactOpts, newOwner)
}

// UpdateMinGasPrice is a paid mutator transaction binding the contract method 0x07aaf344.
//
// Solidity: function updateMinGasPrice(uint256 minGasPrice) returns()
func (_Contract *ContractTransactor) UpdateMinGasPrice(opts *bind.TransactOpts, minGasPrice *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateMinGasPrice", minGasPrice)
}

// UpdateMinGasPrice is a paid mutator transaction binding the contract method 0x07aaf344.
//
// Solidity: function updateMinGasPrice(uint256 minGasPrice) returns()
func (_Contract *ContractSession) UpdateMinGasPrice(minGasPrice *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateMinGasPrice(&_Contract.TransactOpts, minGasPrice)
}

// UpdateMinGasPrice is a paid mutator transaction binding the contract method 0x07aaf344.
//
// Solidity: function updateMinGasPrice(uint256 minGasPrice) returns()
func (_Contract *ContractTransactorSession) UpdateMinGasPrice(minGasPrice *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateMinGasPrice(&_Contract.TransactOpts, minGasPrice)
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

// UpgradeCode is a paid mutator transaction binding the contract method 0xfd1b6ec1.
//
// Solidity: function upgradeCode(address acc, address from) returns()
func (_Contract *ContractTransactor) UpgradeCode(opts *bind.TransactOpts, acc common.Address, from common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "upgradeCode", acc, from)
}

// UpgradeCode is a paid mutator transaction binding the contract method 0xfd1b6ec1.
//
// Solidity: function upgradeCode(address acc, address from) returns()
func (_Contract *ContractSession) UpgradeCode(acc common.Address, from common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpgradeCode(&_Contract.TransactOpts, acc, from)
}

// UpgradeCode is a paid mutator transaction binding the contract method 0xfd1b6ec1.
//
// Solidity: function upgradeCode(address acc, address from) returns()
func (_Contract *ContractTransactorSession) UpgradeCode(acc common.Address, from common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpgradeCode(&_Contract.TransactOpts, acc, from)
}

// ContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Contract contract.
type ContractOwnershipTransferredIterator struct {
	Event *ContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractOwnershipTransferred)
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
		it.Event = new(ContractOwnershipTransferred)
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
func (it *ContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractOwnershipTransferred represents a OwnershipTransferred event raised by the Contract contract.
type ContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractOwnershipTransferredIterator{contract: _Contract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractOwnershipTransferred)
				if err := _Contract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) ParseOwnershipTransferred(log types.Log) (*ContractOwnershipTransferred, error) {
	event := new(ContractOwnershipTransferred)
	if err := _Contract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

var ContractBinRuntime = "0x608060405234801561001057600080fd5b506004361061018d5760003560e01c806366e7ea0f116100e3578063b9cc6b1c1161008c578063e08d7e6611610066578063e08d7e6614610702578063f2fde38b14610772578063fd1b6ec1146107985761018d565b8063b9cc6b1c1461062c578063c0c53b8b1461069c578063d6a0c7af146106d45761018d565b80638da5cb5b116100bd5780638da5cb5b146105c95780638f32d59b146105ed578063a4066fbe146106095761018d565b806366e7ea0f14610569578063715018a61461059557806379bead381461059d5761018d565b8063242a6e3f116101455780634ddaf8f21161011f5780634ddaf8f21461033f5780634feb92f314610365578063592fe0c0146104035761018d565b8063242a6e3f14610285578063267ab446146102fc5780634b64e492146103195761018d565b806318f628d41161017657806318f628d4146101ce5780631cef4fab146102265780631e702f83146102625761018d565b806307aaf344146101925780630aeeca00146101b1575b600080fd5b6101af600480360360208110156101a857600080fd5b50356107c6565b005b6101af600480360360208110156101c757600080fd5b5035610969565b6101af60048036036101208110156101e557600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e0810135906101000135610a28565b6101af6004803603608081101561023c57600080fd5b506001600160a01b03813581169160208101359091169060408101359060600135610b35565b6101af6004803603604081101561027857600080fd5b5080359060200135610ba0565b6101af6004803603604081101561029b57600080fd5b813591908101906040810160208201356401000000008111156102bd57600080fd5b8201836020820111156102cf57600080fd5b803590602001918460018302840111640100000000831117156102f157600080fd5b509092509050610c72565b6101af6004803603602081101561031257600080fd5b5035610d87565b6101af6004803603602081101561032f57600080fd5b50356001600160a01b0316610e46565b6101af6004803603602081101561035557600080fd5b50356001600160a01b0316610ed1565b6101af600480360361010081101561037c57600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156103ac57600080fd5b8201836020820111156103be57600080fd5b803590602001918460018302840111640100000000831117156103e057600080fd5b919350915080359060208101359060408101359060608101359060800135610f91565b6101af600480360360a081101561041957600080fd5b81019060208101813564010000000081111561043457600080fd5b82018360208201111561044657600080fd5b8035906020019184602083028401116401000000008311171561046857600080fd5b91939092909160208101903564010000000081111561048657600080fd5b82018360208201111561049857600080fd5b803590602001918460208302840111640100000000831117156104ba57600080fd5b9193909290916020810190356401000000008111156104d857600080fd5b8201836020820111156104ea57600080fd5b8035906020019184602083028401116401000000008311171561050c57600080fd5b91939092909160208101903564010000000081111561052a57600080fd5b82018360208201111561053c57600080fd5b8035906020019184602083028401116401000000008311171561055e57600080fd5b9193509150356110ab565b6101af6004803603604081101561057f57600080fd5b506001600160a01b038135169060200135611270565b6101af611394565b6101af600480360360408110156105b357600080fd5b506001600160a01b03813516906020013561144f565b6105d1611516565b604080516001600160a01b039092168252519081900360200190f35b6105f5611525565b604080519115158252519081900360200190f35b6101af6004803603604081101561061f57600080fd5b5080359060200135611536565b6101af6004803603602081101561064257600080fd5b81019060208101813564010000000081111561065d57600080fd5b82018360208201111561066f57600080fd5b8035906020019184600183028401116401000000008311171561069157600080fd5b509092509050611602565b6101af600480360360608110156106b257600080fd5b506001600160a01b0381358116916020810135821691604090910135166116eb565b6101af600480360360408110156106ea57600080fd5b506001600160a01b0381358116916020013516611838565b6101af6004803603602081101561071857600080fd5b81019060208101813564010000000081111561073357600080fd5b82018360208201111561074557600080fd5b8035906020019184602083028401116401000000008311171561076757600080fd5b509092509050611900565b6101af6004803603602081101561078857600080fd5b50356001600160a01b03166119df565b6101af600480360360408110156107ae57600080fd5b506001600160a01b0381358116916020013516611a41565b6066546001600160a01b03163314610825576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b60675460408051808201909152601a81527f7b2245636f6e6f6d79223a7b224d696e4761735072696365223a00000000000060208201526001600160a01b039091169063b9cc6b1c906108b69061087b85611b04565b6040518060400160405280600281526020017f7d7d000000000000000000000000000000000000000000000000000000000000815250611c28565b6040518263ffffffff1660e01b81526004018080602001828103825283818151815260200191508051906020019080838360005b838110156109025781810151838201526020016108ea565b50505050905090810190601f16801561092f5780820380516001836020036101000a031916815260200191505b5092505050600060405180830381600087803b15801561094e57600080fd5b505af1158015610962573d6000803e3d6000fd5b5050505050565b610971611525565b6109c2576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517f0aeeca000000000000000000000000000000000000000000000000000000000081526004810184905290516001600160a01b0390921691630aeeca009160248082019260009290919082900301818387803b15801561094e57600080fd5b6067546001600160a01b03163314610a715760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606654604080517f18f628d40000000000000000000000000000000000000000000000000000000081526001600160a01b038c81166004830152602482018c9052604482018b9052606482018a90526084820189905260a4820188905260c4820187905260e482018690526101048201859052915191909216916318f628d49161012480830192600092919082900301818387803b158015610b1257600080fd5b505af1158015610b26573d6000803e3d6000fd5b50505050505050505050505050565b610b3d611525565b610b8e576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b610b9a84848484611dc6565b50505050565b6067546001600160a01b03163314610be95760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606654604080517f1e702f83000000000000000000000000000000000000000000000000000000008152600481018590526024810184905290516001600160a01b0390921691631e702f839160448082019260009290919082900301818387803b158015610c5657600080fd5b505af1158015610c6a573d6000803e3d6000fd5b505050505050565b6066546001600160a01b03163314610cd1576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b606754604080517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810186815260248201928352604482018590526001600160a01b039093169263242a6e3f928792879287929091606401848480828437600081840152601f19601f820116905080830192505050945050505050600060405180830381600087803b158015610d6a57600080fd5b505af1158015610d7e573d6000803e3d6000fd5b50505050505050565b610d8f611525565b610de0576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517f267ab4460000000000000000000000000000000000000000000000000000000081526004810184905290516001600160a01b039092169163267ab4469160248082019260009290919082900301818387803b15801561094e57600080fd5b610e4e611525565b610e9f576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b610ece81610eab611516565b610eb430611ef0565b606754610ec9906001600160a01b0316611ef0565b611dc6565b50565b610ed9611525565b610f2a576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517fda7fc24f0000000000000000000000000000000000000000000000000000000081526001600160a01b0384811660048301529151919092169163da7fc24f91602480830192600092919082900301818387803b15801561094e57600080fd5b6067546001600160a01b03163314610fda5760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606660009054906101000a90046001600160a01b03166001600160a01b0316634feb92f38a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808a6001600160a01b03166001600160a01b03168152602001898152602001806020018781526020018681526020018581526020018481526020018381526020018281038252898982818152602001925080828437600081840152601f19601f8201169050808301925050509a5050505050505050505050600060405180830381600087803b158015610b1257600080fd5b6067546001600160a01b031633146110f45760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b606660009054906101000a90046001600160a01b03166001600160a01b031663592fe0c08a8a8a8a8a8a8a8a8a6040518a63ffffffff1660e01b8152600401808060200180602001806020018060200186815260200185810385528e8e82818152602001925060200280828437600083820152601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe01690910186810385528c8152602090810191508d908d0280828437600083820152601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe01690910186810384528a8152602090810191508b908b0280828437600083820152601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169091018681038352888152602090810191508990890280828437600081840152601f19601f8201169050808301925050509d5050505050505050505050505050600060405180830381600087803b158015610b1257600080fd5b6066546001600160a01b031633146112cf576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b6066546001600160a01b0383811691161461131b5760405162461bcd60e51b81526004018080602001828103825260218152602001806121ef6021913960400191505060405180910390fd5b6067546001600160a01b039081169063e30443bc908490611345908216318563ffffffff611ef416565b6040518363ffffffff1660e01b815260040180836001600160a01b03166001600160a01b0316815260200182815260200192505050600060405180830381600087803b158015610c5657600080fd5b61139c611525565b6113ed576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055565b611457611525565b6114a8576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517f79bead380000000000000000000000000000000000000000000000000000000081526001600160a01b03858116600483015260248201859052915191909216916379bead3891604480830192600092919082900301818387803b158015610c5657600080fd5b6033546001600160a01b031690565b6033546001600160a01b0316331490565b6066546001600160a01b03163314611595576040805162461bcd60e51b815260206004820152601e60248201527f63616c6c6572206973206e6f74207468652053464320636f6e74726163740000604482015290519081900360640190fd5b606754604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018590526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b158015610c5657600080fd5b61160a611525565b61165b576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6067546040517fb9cc6b1c000000000000000000000000000000000000000000000000000000008152602060048201908152602482018490526001600160a01b039092169163b9cc6b1c91859185918190604401848480828437600081840152601f19601f8201169050808301925050509350505050600060405180830381600087803b158015610c5657600080fd5b600054610100900460ff16806117045750611704611f55565b80611712575060005460ff16155b61174d5760405162461bcd60e51b815260040180806020018281038252602e8152602001806121c1602e913960400191505060405180910390fd5b600054610100900460ff161580156117b357600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b6117bc82611f5b565b606780546001600160a01b038086167fffffffffffffffffffffffff00000000000000000000000000000000000000009283161790925560668054928716929091169190911790558015610b9a57600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff16905550505050565b611840611525565b611891576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b606754604080517fd6a0c7af0000000000000000000000000000000000000000000000000000000081526001600160a01b03858116600483015284811660248301529151919092169163d6a0c7af91604480830192600092919082900301818387803b158015610c5657600080fd5b6067546001600160a01b031633146119495760405162461bcd60e51b81526004018080602001828103825260258152602001806122106025913960400191505060405180910390fd5b6066546040517fe08d7e66000000000000000000000000000000000000000000000000000000008152602060048201818152602483018590526001600160a01b039093169263e08d7e6692869286929182916044909101908590850280828437600081840152601f19601f8201169050808301925050509350505050600060405180830381600087803b158015610c5657600080fd5b6119e7611525565b611a38576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b610ece816120bd565b611a49611525565b611a9a576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b611aa382612176565b8015611ab35750611ab381612176565b611891576040805162461bcd60e51b815260206004820152600e60248201527f6e6f74206120636f6e7472616374000000000000000000000000000000000000604482015290519081900360640190fd5b606081611b45575060408051808201909152600181527f30000000000000000000000000000000000000000000000000000000000000006020820152611c23565b6000611b508361217c565b90506060816040519080825280601f01601f191660200182016040528015611b7f576020820181803883390190505b5090507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82015b8415611c1e57600a850660300160f81b828281518110611bc257fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600a850494507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01611ba6565b509150505b919050565b60608084905060608490506060849050606081518351855101016040519080825280601f01601f191660200182016040528015611c6c576020820181803883390190505b509050806000805b8651811015611cdd57868181518110611c8957fe5b602001015160f81c60f81b838380600101945081518110611ca657fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600101611c74565b5060005b8551811015611d4a57858181518110611cf657fe5b602001015160f81c60f81b838380600101945081518110611d1357fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600101611ce1565b5060005b8451811015611db757848181518110611d6357fe5b602001015160f81c60f81b838380600101945081518110611d8057fe5b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350600101611d4e565b50909998505050505050505050565b611dcf846120bd565b836001600160a01b031663614619546040518163ffffffff1660e01b8152600401600060405180830381600087803b158015611e0a57600080fd5b505af1158015611e1e573d6000803e3d6000fd5b50505050611e2b836120bd565b81611e3530611ef0565b14611e87576040805162461bcd60e51b815260206004820152601c60248201527f73656c6620636f6465206861736820646f65736e2774206d6174636800000000604482015290519081900360640190fd5b6067548190611e9e906001600160a01b0316611ef0565b14610b9a576040805162461bcd60e51b815260206004820152601e60248201527f64726976657220636f6465206861736820646f65736e2774206d617463680000604482015290519081900360640190fd5b3f90565b600082820183811015611f4e576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b9392505050565b303b1590565b600054610100900460ff1680611f745750611f74611f55565b80611f82575060005460ff16155b611fbd5760405162461bcd60e51b815260040180806020018281038252602e8152602001806121c1602e913960400191505060405180910390fd5b600054610100900460ff1615801561202357600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0384811691909117918290556040519116906000907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a380156120b957600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555b5050565b6001600160a01b0381166121025760405162461bcd60e51b815260040180806020018281038252602681526020018061219b6026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b3b151590565b6000805b821561219457600101600a83049250612180565b9291505056fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f2061646472657373436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a6564726563697069656e74206973206e6f74207468652053464320636f6e747261637463616c6c6572206973206e6f7420746865204e6f646544726976657220636f6e7472616374a265627a7a72315820e32e2bd9d4306bc53a2f1341ed34b7c36a48a87713cbfbf664b6cde5edf9f98e64736f6c63430005110032"
