// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package SynthereumDeployer

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

// SynthereumDeployerRoles is an auto generated low-level Go binding around an user-defined struct.
type SynthereumDeployerRoles struct {
	Admin      common.Address
	Maintainer common.Address
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = "[{\"inputs\":[{\"internalType\":\"contractISynthereumFinder\",\"name\":\"_synthereumFinder\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"maintainer\",\"type\":\"address\"}],\"internalType\":\"structSynthereumDeployer.Roles\",\"name\":\"roles\",\"type\":\"tuple\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"fixedRateVersion\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"fixedRate\",\"type\":\"address\"}],\"name\":\"FixedRateDeployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"poolVersion\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newPool\",\"type\":\"address\"}],\"name\":\"PoolDeployed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"selfMintingDerivativeVersion\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"selfMintingDerivative\",\"type\":\"address\"}],\"name\":\"SelfMintingDerivativeDeployed\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAINTAINER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"fixedRateVersion\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"fixedRateParamsData\",\"type\":\"bytes\"}],\"name\":\"deployFixedRate\",\"outputs\":[{\"internalType\":\"contractISynthereumDeployment\",\"name\":\"fixedRate\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"poolVersion\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"poolParamsData\",\"type\":\"bytes\"}],\"name\":\"deployPool\",\"outputs\":[{\"internalType\":\"contractISynthereumDeployment\",\"name\":\"pool\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"selfMintingDerVersion\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"selfMintingDerParamsData\",\"type\":\"bytes\"}],\"name\":\"deploySelfMintingDerivative\",\"outputs\":[{\"internalType\":\"contractISynthereumDeployment\",\"name\":\"selfMintingDerivative\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRoleMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMemberCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"synthereumFinder\",\"outputs\":[{\"internalType\":\"contractISynthereumFinder\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = "0x60a06040523480156200001157600080fd5b506040516200276138038062002761833981016040819052620000349162000261565b60016002556001600160601b0319606083901b1660805262000058600080620000b1565b62000074600080516020620027418339815191526000620000b1565b80516200008490600090620000fc565b620000a9600080516020620027418339815191528260200151620000fc60201b60201c565b505062000314565b600082815260208190526040808220600101805490849055905190918391839186917fbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff9190a4505050565b6200011382826200013f60201b62000a781760201c565b60008281526001602090815260409091206200013a91839062000a866200014f821b17901c565b505050565b6200014b82826200016f565b5050565b600062000166836001600160a01b0384166200020f565b90505b92915050565b6000828152602081815260408083206001600160a01b038516845290915290205460ff166200014b576000828152602081815260408083206001600160a01b03851684529091529020805460ff19166001179055620001cb3390565b6001600160a01b0316816001600160a01b0316837f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d60405160405180910390a45050565b6000818152600183016020526040812054620002585750815460018181018455600084815260208082209093018490558454848252828601909352604090209190915562000169565b50600062000169565b600080828403606081121562000275578283fd5b83516200028281620002fb565b92506040601f198201121562000296578182fd5b50604080519081016001600160401b0381118282101715620002c657634e487b7160e01b83526041600452602483fd5b6040526020840151620002d981620002fb565b81526040840151620002eb81620002fb565b6020820152919491935090915050565b6001600160a01b03811681146200031157600080fd5b50565b60805160601c6123e46200035d6000396000818161021a01528181610bb101528181610d62015281816111140152818161126c015281816116960152611a6a01526123e46000f3fe608060405234801561001057600080fd5b50600436106100ea5760003560e01c806391d148541161008c578063d547741f11610066578063d547741f146101ef578063f07160f414610202578063f6bf3ef614610215578063f87422541461023c57600080fd5b806391d14854146101c1578063a217fddf146101d4578063ca15c873146101dc57600080fd5b806336568abe116100c857806336568abe1461015d578063552fc91e1461017057806368eb0dcd1461019b5780639010d07c146101ae57600080fd5b806301ffc9a7146100ef578063248a9ca3146101175780632f2ff15d14610148575b600080fd5b6101026100fd366004611e78565b610251565b60405190151581526020015b60405180910390f35b61013a610125366004611e10565b60009081526020819052604090206001015490565b60405190815260200161010e565b61015b610156366004611e28565b61027c565b005b61015b61016b366004611e28565b6102a3565b61018361017e366004611f6a565b6102c5565b6040516001600160a01b03909116815260200161010e565b6101836101a9366004611f6a565b610521565b6101836101bc366004611e57565b610774565b6101026101cf366004611e28565b610793565b61013a600081565b61013a6101ea366004611e10565b6107bc565b61015b6101fd366004611e28565b6107d3565b610183610210366004611f6a565b6107dd565b6101837f000000000000000000000000000000000000000000000000000000000000000081565b61013a60008051602061238f83398151915281565b60006001600160e01b03198216635a05180f60e01b1480610276575061027682610a9b565b92915050565b6102868282610ad0565b600082815260016020526040902061029e9082610a86565b505050565b6102ad8282610af6565b600082815260016020526040902061029e9082610b70565b60006102df60008051602061238f83398151915233610793565b6103045760405162461bcd60e51b81526004016102fb906121fa565b60405180910390fd5b6002805414156103265760405162461bcd60e51b81526004016102fb90612231565b60028055610372610335610b85565b8585858080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250610c3992505050565b905061037e8185610d60565b61038781610f16565b60006103916110e8565b9050806001600160a01b031663aab748b6836001600160a01b03166336815bb76040518163ffffffff1660e01b815260040160006040518083038186803b1580156103db57600080fd5b505afa1580156103ef573d6000803e3d6000fd5b505050506040513d6000823e601f3d908101601f191682016040526104179190810190611eb0565b846001600160a01b031663b2016bd46040518163ffffffff1660e01b815260040160206040518083038186803b15801561045057600080fd5b505afa158015610464573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104889190611dd4565b88866040518563ffffffff1660e01b81526004016104a994939291906121bb565b600060405180830381600087803b1580156104c357600080fd5b505af11580156104d7573d6000803e3d6000fd5b50506040516001600160a01b038516925060ff881691507f992db37443731f3f5a56acbcbb4bc0661c45df62ca7276d5e9da41dfb805e24390600090a35060016002559392505050565b600061053b60008051602061238f83398151915233610793565b6105575760405162461bcd60e51b81526004016102fb906121fa565b6002805414156105795760405162461bcd60e51b81526004016102fb90612231565b600280556105c5610588610b85565b8585858080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061114b92505050565b90506105d18185610d60565b6105da81610f16565b60006105e4611245565b9050806001600160a01b031663aab748b6836001600160a01b03166336815bb76040518163ffffffff1660e01b815260040160006040518083038186803b15801561062e57600080fd5b505afa158015610642573d6000803e3d6000fd5b505050506040513d6000823e601f3d908101601f1916820160405261066a9190810190611eb0565b846001600160a01b031663b2016bd46040518163ffffffff1660e01b815260040160206040518083038186803b1580156106a357600080fd5b505afa1580156106b7573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106db9190611dd4565b88866040518563ffffffff1660e01b81526004016106fc94939291906121bb565b600060405180830381600087803b15801561071657600080fd5b505af115801561072a573d6000803e3d6000fd5b50506040516001600160a01b038516925060ff881691507f077b1cf2873467dc041762beee77c09690af8842b7b6c958941d92cef7e22a3690600090a35060016002559392505050565b600082815260016020526040812061078c90836112a3565b9392505050565b6000918252602082815260408084206001600160a01b0393909316845291905290205460ff1690565b6000818152600160205260408120610276906112af565b6102ad82826112b9565b60006107f760008051602061238f83398151915233610793565b6108135760405162461bcd60e51b81526004016102fb906121fa565b6002805414156108355760405162461bcd60e51b81526004016102fb90612231565b600280556000610843610b85565b9050610851818686866112df565b915061085d8286610d60565b6000826001600160a01b0316638230ecd66040518163ffffffff1660e01b815260040160206040518083038186803b15801561089857600080fd5b505afa1580156108ac573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906108d09190611dd4565b90506108dc81846113f3565b60006108e6611668565b9050806001600160a01b031663aab748b6856001600160a01b03166336815bb76040518163ffffffff1660e01b815260040160006040518083038186803b15801561093057600080fd5b505afa158015610944573d6000803e3d6000fd5b505050506040513d6000823e601f3d908101601f1916820160405261096c9190810190611eb0565b866001600160a01b031663b2016bd46040518163ffffffff1660e01b815260040160206040518083038186803b1580156109a557600080fd5b505afa1580156109b9573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109dd9190611dd4565b8a886040518563ffffffff1660e01b81526004016109fe94939291906121bb565b600060405180830381600087803b158015610a1857600080fd5b505af1158015610a2c573d6000803e3d6000fd5b50506040516001600160a01b038716925060ff8a1691507f246f71ad275621aa67be570521951d4047b41afe97c13a14733dae8ec481331d90600090a350506001600255509392505050565b610a8282826116cd565b5050565b600061078c836001600160a01b038416611751565b60006001600160e01b03198216637965db0b60e01b148061027657506301ffc9a760e01b6001600160e01b0319831614610276565b600082815260208190526040902060010154610aec81336117a0565b61029e83836116cd565b6001600160a01b0381163314610b665760405162461bcd60e51b815260206004820152602f60248201527f416363657373436f6e74726f6c3a2063616e206f6e6c792072656e6f756e636560448201526e103937b632b9903337b91039b2b63360891b60648201526084016102fb565b610a828282611804565b600061078c836001600160a01b038416611869565b6040516302abf57960e61b815270466163746f727956657273696f6e696e6760781b60048201526000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063aafd5e40906024015b60206040518083038186803b158015610bfc57600080fd5b505afa158015610c10573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610c349190611dd4565b905090565b604051636839980160e11b81526f466978656452617465466163746f727960801b600482015260ff8316602482015260009081906001600160a01b0386169063d07330029060440160206040518083038186803b158015610c9957600080fd5b505afa158015610cad573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610cd19190611dd4565b90506000610d40610ce183611986565b85604051602001610cf392919061207d565b60408051601f19818403018152828201909152601b82527f57726f6e672066697865642072617465206465706c6f796d656e74000000000060208301526001600160a01b038516916119f9565b905080806020019051810190610d569190611dd4565b9695505050505050565b7f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316826001600160a01b031663f6bf3ef66040518163ffffffff1660e01b815260040160206040518083038186803b158015610dc357600080fd5b505afa158015610dd7573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610dfb9190611dd4565b6001600160a01b031614610e515760405162461bcd60e51b815260206004820152601a60248201527f57726f6e672066696e64657220696e206465706c6f796d656e7400000000000060448201526064016102fb565b8060ff16826001600160a01b03166354fd4d506040518163ffffffff1660e01b815260040160206040518083038186803b158015610e8e57600080fd5b505afa158015610ea2573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610ec69190611f4e565b60ff1614610a825760405162461bcd60e51b815260206004820152601b60248201527f57726f6e672076657273696f6e20696e206465706c6f796d656e74000000000060448201526064016102fb565b60008190506000826001600160a01b0316638230ecd66040518163ffffffff1660e01b815260040160206040518083038186803b158015610f5657600080fd5b505afa158015610f6a573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610f8e9190611dd4565b604051632474521560e21b81527f6e58ad548d72b425ea94c15f453bf26caddb061d82b2551db7fdd3cefe0e994060048201526001600160a01b038481166024830152919250908216906391d148549060440160206040518083038186803b158015610ff957600080fd5b505afa15801561100d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906110319190611df0565b15806110d95750604051632474521560e21b81527fe4b2a1ba12b0ae46fe120e095faea153cf269e4b012b647a52a09f4e0e45f17960048201526001600160a01b0383811660248301528216906391d148549060440160206040518083038186803b15801561109f57600080fd5b505afa1580156110b3573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906110d79190611df0565b155b1561029e5761029e81836113f3565b6040516302abf57960e61b815270466978656452617465526567697374727960781b60048201526000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063aafd5e4090602401610be4565b604051636839980160e11b81526a506f6f6c466163746f727960a81b600482015260ff8316602482015260009081906001600160a01b0386169063d07330029060440160206040518083038186803b1580156111a657600080fd5b505afa1580156111ba573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906111de9190611dd4565b90506000610d406111ee83611986565b8560405160200161120092919061207d565b60408051601f19818403018152828201909152601582527415dc9bdb99c81c1bdbdb0819195c1b1bde5b595b9d605a1b60208301526001600160a01b038516916119f9565b6040516302abf57960e61b81526b506f6f6c526567697374727960a01b60048201526000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063aafd5e4090602401610be4565b600061078c8383611a10565b6000610276825490565b6000828152602081905260409020600101546112d581336117a0565b61029e8383611804565b604051636839980160e11b81527153656c664d696e74696e67466163746f727960701b600482015260ff8416602482015260009081906001600160a01b0387169063d07330029060440160206040518083038186803b15801561134157600080fd5b505afa158015611355573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906113799190611dd4565b905060006113d261138983611986565b868660405160200161139d93929190612059565b60408051601f1981840301815260608301909152602880835290919061236760208301396001600160a01b03851691906119f9565b9050808060200190518101906113e89190611dd4565b979650505050505050565b60006113fd611a48565b6040805160028082526060820183529293506000929091602083019080368337505060408051600280825260608201835293945060009390925090602083019080368337505060408051600280825260608201835293945060009390925090602083019080368337019050509050858360008151811061148d57634e487b7160e01b600052603260045260246000fd5b60200260200101906001600160a01b031690816001600160a01b03168152505085836001815181106114cf57634e487b7160e01b600052603260045260246000fd5b60200260200101906001600160a01b031690816001600160a01b0316815250507f6e58ad548d72b425ea94c15f453bf26caddb061d82b2551db7fdd3cefe0e99408260008151811061153157634e487b7160e01b600052603260045260246000fd5b6020026020010181815250507fe4b2a1ba12b0ae46fe120e095faea153cf269e4b012b647a52a09f4e0e45f1798260018151811061157f57634e487b7160e01b600052603260045260246000fd5b60200260200101818152505084816000815181106115ad57634e487b7160e01b600052603260045260246000fd5b60200260200101906001600160a01b031690816001600160a01b03168152505084816001815181106115ef57634e487b7160e01b600052603260045260246000fd5b6001600160a01b03928316602091820292909201015260405163ee8106cf60e01b81529085169063ee8106cf9061162e9086908690869060040161213f565b600060405180830381600087803b15801561164857600080fd5b505af115801561165c573d6000803e3d6000fd5b50505050505050505050565b6040516302abf57960e61b81527253656c664d696e74696e67526567697374727960681b60048201526000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063aafd5e4090602401610be4565b6116d78282610793565b610a82576000828152602081815260408083206001600160a01b03851684529091529020805460ff1916600117905561170d3390565b6001600160a01b0316816001600160a01b0316837f2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d60405160405180910390a45050565b600081815260018301602052604081205461179857508154600181810184556000848152602080822090930184905584548482528286019093526040902091909155610276565b506000610276565b6117aa8282610793565b610a82576117c2816001600160a01b03166014611aa1565b6117cd836020611aa1565b6040516020016117de9291906120ca565b60408051601f198184030181529082905262461bcd60e51b82526102fb916004016121a8565b61180e8282610793565b15610a82576000828152602081815260408083206001600160a01b0385168085529252808320805460ff1916905551339285917ff6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b9190a45050565b6000818152600183016020526040812054801561197c57600061188d60018361229f565b85549091506000906118a19060019061229f565b90508181146119225760008660000182815481106118cf57634e487b7160e01b600052603260045260246000fd5b906000526020600020015490508087600001848154811061190057634e487b7160e01b600052603260045260246000fd5b6000918252602080832090910192909255918252600188019052604090208390555b855486908061194157634e487b7160e01b600052603160045260246000fd5b600190038181906000526020600020016000905590558560010160008681526020019081526020016000206000905560019350505050610276565b6000915050610276565b6000816001600160a01b031663b084033c6040518163ffffffff1660e01b815260040160206040518083038186803b1580156119c157600080fd5b505afa1580156119d5573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102769190611e94565b6060611a088484600085611c83565b949350505050565b6000826000018281548110611a3557634e487b7160e01b600052603260045260246000fd5b9060005260206000200154905092915050565b6040516302abf57960e61b81526626b0b730b3b2b960c91b60048201526000907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063aafd5e4090602401610be4565b60606000611ab0836002612280565b611abb906002612268565b67ffffffffffffffff811115611ae157634e487b7160e01b600052604160045260246000fd5b6040519080825280601f01601f191660200182016040528015611b0b576020820181803683370190505b509050600360fc1b81600081518110611b3457634e487b7160e01b600052603260045260246000fd5b60200101906001600160f81b031916908160001a905350600f60fb1b81600181518110611b7157634e487b7160e01b600052603260045260246000fd5b60200101906001600160f81b031916908160001a9053506000611b95846002612280565b611ba0906001612268565b90505b6001811115611c34576f181899199a1a9b1b9c1cb0b131b232b360811b85600f1660108110611be257634e487b7160e01b600052603260045260246000fd5b1a60f81b828281518110611c0657634e487b7160e01b600052603260045260246000fd5b60200101906001600160f81b031916908160001a90535060049490941c93611c2d816122e6565b9050611ba3565b50831561078c5760405162461bcd60e51b815260206004820181905260248201527f537472696e67733a20686578206c656e67746820696e73756666696369656e7460448201526064016102fb565b606082471015611ce45760405162461bcd60e51b815260206004820152602660248201527f416464726573733a20696e73756666696369656e742062616c616e636520666f6044820152651c8818d85b1b60d21b60648201526084016102fb565b843b611d325760405162461bcd60e51b815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e747261637400000060448201526064016102fb565b600080866001600160a01b03168587604051611d4e91906120ae565b60006040518083038185875af1925050503d8060008114611d8b576040519150601f19603f3d011682016040523d82523d6000602084013e611d90565b606091505b50915091506113e882828660608315611daa57508161078c565b825115611dba5782518084602001fd5b8160405162461bcd60e51b81526004016102fb91906121a8565b600060208284031215611de5578081fd5b815161078c81612329565b600060208284031215611e01578081fd5b8151801515811461078c578182fd5b600060208284031215611e21578081fd5b5035919050565b60008060408385031215611e3a578081fd5b823591506020830135611e4c81612329565b809150509250929050565b60008060408385031215611e69578182fd5b50508035926020909101359150565b600060208284031215611e89578081fd5b813561078c81612341565b600060208284031215611ea5578081fd5b815161078c81612341565b600060208284031215611ec1578081fd5b815167ffffffffffffffff80821115611ed8578283fd5b818401915084601f830112611eeb578283fd5b815181811115611efd57611efd612313565b604051601f8201601f19908116603f01168101908382118183101715611f2557611f25612313565b81604052828152876020848701011115611f3d578586fd5b6113e88360208301602088016122b6565b600060208284031215611f5f578081fd5b815161078c81612357565b600080600060408486031215611f7e578081fd5b8335611f8981612357565b9250602084013567ffffffffffffffff80821115611fa5578283fd5b818601915086601f830112611fb8578283fd5b813581811115611fc6578384fd5b876020828501011115611fd7578384fd5b6020830194508093505050509250925092565b6000815180845260208085019450808401835b838110156120225781516001600160a01b031687529582019590820190600101611ffd565b509495945050505050565b600081518084526120458160208601602086016122b6565b601f01601f19169290920160200192915050565b6001600160e01b031984168152818360048301376000910160040190815292915050565b6001600160e01b03198316815281516000906120a08160048501602087016122b6565b919091016004019392505050565b600082516120c08184602087016122b6565b9190910192915050565b7f416363657373436f6e74726f6c3a206163636f756e74200000000000000000008152600083516121028160178501602088016122b6565b7001034b99036b4b9b9b4b733903937b6329607d1b60179184019182015283516121338160288401602088016122b6565b01602801949350505050565b6060815260006121526060830186611fea565b828103602084810191909152855180835286820192820190845b818110156121885784518352938301939183019160010161216c565b5050848103604086015261219c8187611fea565b98975050505050505050565b60208152600061078c602083018461202d565b6080815260006121ce608083018761202d565b6001600160a01b03958616602084015260ff949094166040830152509216606090920191909152919050565b6020808252601d908201527f53656e646572206d75737420626520746865206d61696e7461696e6572000000604082015260600190565b6020808252601f908201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604082015260600190565b6000821982111561227b5761227b6122fd565b500190565b600081600019048311821515161561229a5761229a6122fd565b500290565b6000828210156122b1576122b16122fd565b500390565b60005b838110156122d15781810151838201526020016122b9565b838111156122e0576000848401525b50505050565b6000816122f5576122f56122fd565b506000190190565b634e487b7160e01b600052601160045260246000fd5b634e487b7160e01b600052604160045260246000fd5b6001600160a01b038116811461233e57600080fd5b50565b6001600160e01b03198116811461233e57600080fd5b60ff8116811461233e57600080fdfe57726f6e672073656c662d6d696e74696e672064657269766174697665206465706c6f796d656e74126303c860ea810f85e857ad8768056e2eebc24b7796655ff3107e4af18e3f1ea2646970667358221220c234f5ffce42f2895b6973c2208ef83f9f66d0aef4bc0e3100d376a33898d2ec64736f6c63430008040033126303c860ea810f85e857ad8768056e2eebc24b7796655ff3107e4af18e3f1e"

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend, _synthereumFinder common.Address, roles SynthereumDeployerRoles) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractBin), backend, _synthereumFinder, roles)
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

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Contract *ContractCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Contract *ContractSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Contract.Contract.DEFAULTADMINROLE(&_Contract.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Contract *ContractCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Contract.Contract.DEFAULTADMINROLE(&_Contract.CallOpts)
}

// MAINTAINERROLE is a free data retrieval call binding the contract method 0xf8742254.
//
// Solidity: function MAINTAINER_ROLE() view returns(bytes32)
func (_Contract *ContractCaller) MAINTAINERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "MAINTAINER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MAINTAINERROLE is a free data retrieval call binding the contract method 0xf8742254.
//
// Solidity: function MAINTAINER_ROLE() view returns(bytes32)
func (_Contract *ContractSession) MAINTAINERROLE() ([32]byte, error) {
	return _Contract.Contract.MAINTAINERROLE(&_Contract.CallOpts)
}

// MAINTAINERROLE is a free data retrieval call binding the contract method 0xf8742254.
//
// Solidity: function MAINTAINER_ROLE() view returns(bytes32)
func (_Contract *ContractCallerSession) MAINTAINERROLE() ([32]byte, error) {
	return _Contract.Contract.MAINTAINERROLE(&_Contract.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Contract *ContractCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Contract *ContractSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Contract.Contract.GetRoleAdmin(&_Contract.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Contract *ContractCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Contract.Contract.GetRoleAdmin(&_Contract.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Contract *ContractCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Contract *ContractSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Contract.Contract.GetRoleMember(&_Contract.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Contract *ContractCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Contract.Contract.GetRoleMember(&_Contract.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Contract *ContractCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Contract *ContractSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Contract.Contract.GetRoleMemberCount(&_Contract.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Contract *ContractCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Contract.Contract.GetRoleMemberCount(&_Contract.CallOpts, role)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Contract *ContractCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Contract *ContractSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Contract.Contract.HasRole(&_Contract.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Contract *ContractCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Contract.Contract.HasRole(&_Contract.CallOpts, role, account)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Contract *ContractCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Contract *ContractSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Contract.Contract.SupportsInterface(&_Contract.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Contract *ContractCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Contract.Contract.SupportsInterface(&_Contract.CallOpts, interfaceId)
}

// SynthereumFinder is a free data retrieval call binding the contract method 0xf6bf3ef6.
//
// Solidity: function synthereumFinder() view returns(address)
func (_Contract *ContractCaller) SynthereumFinder(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "synthereumFinder")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SynthereumFinder is a free data retrieval call binding the contract method 0xf6bf3ef6.
//
// Solidity: function synthereumFinder() view returns(address)
func (_Contract *ContractSession) SynthereumFinder() (common.Address, error) {
	return _Contract.Contract.SynthereumFinder(&_Contract.CallOpts)
}

// SynthereumFinder is a free data retrieval call binding the contract method 0xf6bf3ef6.
//
// Solidity: function synthereumFinder() view returns(address)
func (_Contract *ContractCallerSession) SynthereumFinder() (common.Address, error) {
	return _Contract.Contract.SynthereumFinder(&_Contract.CallOpts)
}

// DeployFixedRate is a paid mutator transaction binding the contract method 0x552fc91e.
//
// Solidity: function deployFixedRate(uint8 fixedRateVersion, bytes fixedRateParamsData) returns(address fixedRate)
func (_Contract *ContractTransactor) DeployFixedRate(opts *bind.TransactOpts, fixedRateVersion uint8, fixedRateParamsData []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deployFixedRate", fixedRateVersion, fixedRateParamsData)
}

// DeployFixedRate is a paid mutator transaction binding the contract method 0x552fc91e.
//
// Solidity: function deployFixedRate(uint8 fixedRateVersion, bytes fixedRateParamsData) returns(address fixedRate)
func (_Contract *ContractSession) DeployFixedRate(fixedRateVersion uint8, fixedRateParamsData []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeployFixedRate(&_Contract.TransactOpts, fixedRateVersion, fixedRateParamsData)
}

// DeployFixedRate is a paid mutator transaction binding the contract method 0x552fc91e.
//
// Solidity: function deployFixedRate(uint8 fixedRateVersion, bytes fixedRateParamsData) returns(address fixedRate)
func (_Contract *ContractTransactorSession) DeployFixedRate(fixedRateVersion uint8, fixedRateParamsData []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeployFixedRate(&_Contract.TransactOpts, fixedRateVersion, fixedRateParamsData)
}

// DeployPool is a paid mutator transaction binding the contract method 0x68eb0dcd.
//
// Solidity: function deployPool(uint8 poolVersion, bytes poolParamsData) returns(address pool)
func (_Contract *ContractTransactor) DeployPool(opts *bind.TransactOpts, poolVersion uint8, poolParamsData []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deployPool", poolVersion, poolParamsData)
}

// DeployPool is a paid mutator transaction binding the contract method 0x68eb0dcd.
//
// Solidity: function deployPool(uint8 poolVersion, bytes poolParamsData) returns(address pool)
func (_Contract *ContractSession) DeployPool(poolVersion uint8, poolParamsData []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeployPool(&_Contract.TransactOpts, poolVersion, poolParamsData)
}

// DeployPool is a paid mutator transaction binding the contract method 0x68eb0dcd.
//
// Solidity: function deployPool(uint8 poolVersion, bytes poolParamsData) returns(address pool)
func (_Contract *ContractTransactorSession) DeployPool(poolVersion uint8, poolParamsData []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeployPool(&_Contract.TransactOpts, poolVersion, poolParamsData)
}

// DeploySelfMintingDerivative is a paid mutator transaction binding the contract method 0xf07160f4.
//
// Solidity: function deploySelfMintingDerivative(uint8 selfMintingDerVersion, bytes selfMintingDerParamsData) returns(address selfMintingDerivative)
func (_Contract *ContractTransactor) DeploySelfMintingDerivative(opts *bind.TransactOpts, selfMintingDerVersion uint8, selfMintingDerParamsData []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deploySelfMintingDerivative", selfMintingDerVersion, selfMintingDerParamsData)
}

// DeploySelfMintingDerivative is a paid mutator transaction binding the contract method 0xf07160f4.
//
// Solidity: function deploySelfMintingDerivative(uint8 selfMintingDerVersion, bytes selfMintingDerParamsData) returns(address selfMintingDerivative)
func (_Contract *ContractSession) DeploySelfMintingDerivative(selfMintingDerVersion uint8, selfMintingDerParamsData []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeploySelfMintingDerivative(&_Contract.TransactOpts, selfMintingDerVersion, selfMintingDerParamsData)
}

// DeploySelfMintingDerivative is a paid mutator transaction binding the contract method 0xf07160f4.
//
// Solidity: function deploySelfMintingDerivative(uint8 selfMintingDerVersion, bytes selfMintingDerParamsData) returns(address selfMintingDerivative)
func (_Contract *ContractTransactorSession) DeploySelfMintingDerivative(selfMintingDerVersion uint8, selfMintingDerParamsData []byte) (*types.Transaction, error) {
	return _Contract.Contract.DeploySelfMintingDerivative(&_Contract.TransactOpts, selfMintingDerVersion, selfMintingDerParamsData)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Contract *ContractTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Contract *ContractSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.Contract.GrantRole(&_Contract.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Contract *ContractTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.Contract.GrantRole(&_Contract.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Contract *ContractTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Contract *ContractSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RenounceRole(&_Contract.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Contract *ContractTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RenounceRole(&_Contract.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Contract *ContractTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Contract *ContractSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RevokeRole(&_Contract.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Contract *ContractTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Contract.Contract.RevokeRole(&_Contract.TransactOpts, role, account)
}

// ContractFixedRateDeployedIterator is returned from FilterFixedRateDeployed and is used to iterate over the raw logs and unpacked data for FixedRateDeployed events raised by the Contract contract.
type ContractFixedRateDeployedIterator struct {
	Event *ContractFixedRateDeployed // Event containing the contract specifics and raw log

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
func (it *ContractFixedRateDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractFixedRateDeployed)
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
		it.Event = new(ContractFixedRateDeployed)
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
func (it *ContractFixedRateDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractFixedRateDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractFixedRateDeployed represents a FixedRateDeployed event raised by the Contract contract.
type ContractFixedRateDeployed struct {
	FixedRateVersion uint8
	FixedRate        common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterFixedRateDeployed is a free log retrieval operation binding the contract event 0x992db37443731f3f5a56acbcbb4bc0661c45df62ca7276d5e9da41dfb805e243.
//
// Solidity: event FixedRateDeployed(uint8 indexed fixedRateVersion, address indexed fixedRate)
func (_Contract *ContractFilterer) FilterFixedRateDeployed(opts *bind.FilterOpts, fixedRateVersion []uint8, fixedRate []common.Address) (*ContractFixedRateDeployedIterator, error) {

	var fixedRateVersionRule []interface{}
	for _, fixedRateVersionItem := range fixedRateVersion {
		fixedRateVersionRule = append(fixedRateVersionRule, fixedRateVersionItem)
	}
	var fixedRateRule []interface{}
	for _, fixedRateItem := range fixedRate {
		fixedRateRule = append(fixedRateRule, fixedRateItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "FixedRateDeployed", fixedRateVersionRule, fixedRateRule)
	if err != nil {
		return nil, err
	}
	return &ContractFixedRateDeployedIterator{contract: _Contract.contract, event: "FixedRateDeployed", logs: logs, sub: sub}, nil
}

// WatchFixedRateDeployed is a free log subscription operation binding the contract event 0x992db37443731f3f5a56acbcbb4bc0661c45df62ca7276d5e9da41dfb805e243.
//
// Solidity: event FixedRateDeployed(uint8 indexed fixedRateVersion, address indexed fixedRate)
func (_Contract *ContractFilterer) WatchFixedRateDeployed(opts *bind.WatchOpts, sink chan<- *ContractFixedRateDeployed, fixedRateVersion []uint8, fixedRate []common.Address) (event.Subscription, error) {

	var fixedRateVersionRule []interface{}
	for _, fixedRateVersionItem := range fixedRateVersion {
		fixedRateVersionRule = append(fixedRateVersionRule, fixedRateVersionItem)
	}
	var fixedRateRule []interface{}
	for _, fixedRateItem := range fixedRate {
		fixedRateRule = append(fixedRateRule, fixedRateItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "FixedRateDeployed", fixedRateVersionRule, fixedRateRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractFixedRateDeployed)
				if err := _Contract.contract.UnpackLog(event, "FixedRateDeployed", log); err != nil {
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

// ParseFixedRateDeployed is a log parse operation binding the contract event 0x992db37443731f3f5a56acbcbb4bc0661c45df62ca7276d5e9da41dfb805e243.
//
// Solidity: event FixedRateDeployed(uint8 indexed fixedRateVersion, address indexed fixedRate)
func (_Contract *ContractFilterer) ParseFixedRateDeployed(log types.Log) (*ContractFixedRateDeployed, error) {
	event := new(ContractFixedRateDeployed)
	if err := _Contract.contract.UnpackLog(event, "FixedRateDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractPoolDeployedIterator is returned from FilterPoolDeployed and is used to iterate over the raw logs and unpacked data for PoolDeployed events raised by the Contract contract.
type ContractPoolDeployedIterator struct {
	Event *ContractPoolDeployed // Event containing the contract specifics and raw log

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
func (it *ContractPoolDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractPoolDeployed)
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
		it.Event = new(ContractPoolDeployed)
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
func (it *ContractPoolDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractPoolDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractPoolDeployed represents a PoolDeployed event raised by the Contract contract.
type ContractPoolDeployed struct {
	PoolVersion uint8
	NewPool     common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterPoolDeployed is a free log retrieval operation binding the contract event 0x077b1cf2873467dc041762beee77c09690af8842b7b6c958941d92cef7e22a36.
//
// Solidity: event PoolDeployed(uint8 indexed poolVersion, address indexed newPool)
func (_Contract *ContractFilterer) FilterPoolDeployed(opts *bind.FilterOpts, poolVersion []uint8, newPool []common.Address) (*ContractPoolDeployedIterator, error) {

	var poolVersionRule []interface{}
	for _, poolVersionItem := range poolVersion {
		poolVersionRule = append(poolVersionRule, poolVersionItem)
	}
	var newPoolRule []interface{}
	for _, newPoolItem := range newPool {
		newPoolRule = append(newPoolRule, newPoolItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "PoolDeployed", poolVersionRule, newPoolRule)
	if err != nil {
		return nil, err
	}
	return &ContractPoolDeployedIterator{contract: _Contract.contract, event: "PoolDeployed", logs: logs, sub: sub}, nil
}

// WatchPoolDeployed is a free log subscription operation binding the contract event 0x077b1cf2873467dc041762beee77c09690af8842b7b6c958941d92cef7e22a36.
//
// Solidity: event PoolDeployed(uint8 indexed poolVersion, address indexed newPool)
func (_Contract *ContractFilterer) WatchPoolDeployed(opts *bind.WatchOpts, sink chan<- *ContractPoolDeployed, poolVersion []uint8, newPool []common.Address) (event.Subscription, error) {

	var poolVersionRule []interface{}
	for _, poolVersionItem := range poolVersion {
		poolVersionRule = append(poolVersionRule, poolVersionItem)
	}
	var newPoolRule []interface{}
	for _, newPoolItem := range newPool {
		newPoolRule = append(newPoolRule, newPoolItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "PoolDeployed", poolVersionRule, newPoolRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractPoolDeployed)
				if err := _Contract.contract.UnpackLog(event, "PoolDeployed", log); err != nil {
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

// ParsePoolDeployed is a log parse operation binding the contract event 0x077b1cf2873467dc041762beee77c09690af8842b7b6c958941d92cef7e22a36.
//
// Solidity: event PoolDeployed(uint8 indexed poolVersion, address indexed newPool)
func (_Contract *ContractFilterer) ParsePoolDeployed(log types.Log) (*ContractPoolDeployed, error) {
	event := new(ContractPoolDeployed)
	if err := _Contract.contract.UnpackLog(event, "PoolDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Contract contract.
type ContractRoleAdminChangedIterator struct {
	Event *ContractRoleAdminChanged // Event containing the contract specifics and raw log

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
func (it *ContractRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRoleAdminChanged)
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
		it.Event = new(ContractRoleAdminChanged)
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
func (it *ContractRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRoleAdminChanged represents a RoleAdminChanged event raised by the Contract contract.
type ContractRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Contract *ContractFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*ContractRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &ContractRoleAdminChangedIterator{contract: _Contract.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Contract *ContractFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *ContractRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRoleAdminChanged)
				if err := _Contract.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
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

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Contract *ContractFilterer) ParseRoleAdminChanged(log types.Log) (*ContractRoleAdminChanged, error) {
	event := new(ContractRoleAdminChanged)
	if err := _Contract.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Contract contract.
type ContractRoleGrantedIterator struct {
	Event *ContractRoleGranted // Event containing the contract specifics and raw log

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
func (it *ContractRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRoleGranted)
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
		it.Event = new(ContractRoleGranted)
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
func (it *ContractRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRoleGranted represents a RoleGranted event raised by the Contract contract.
type ContractRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Contract *ContractFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*ContractRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &ContractRoleGrantedIterator{contract: _Contract.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Contract *ContractFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *ContractRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRoleGranted)
				if err := _Contract.contract.UnpackLog(event, "RoleGranted", log); err != nil {
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

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Contract *ContractFilterer) ParseRoleGranted(log types.Log) (*ContractRoleGranted, error) {
	event := new(ContractRoleGranted)
	if err := _Contract.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Contract contract.
type ContractRoleRevokedIterator struct {
	Event *ContractRoleRevoked // Event containing the contract specifics and raw log

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
func (it *ContractRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRoleRevoked)
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
		it.Event = new(ContractRoleRevoked)
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
func (it *ContractRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRoleRevoked represents a RoleRevoked event raised by the Contract contract.
type ContractRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Contract *ContractFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*ContractRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &ContractRoleRevokedIterator{contract: _Contract.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Contract *ContractFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *ContractRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRoleRevoked)
				if err := _Contract.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
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

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Contract *ContractFilterer) ParseRoleRevoked(log types.Log) (*ContractRoleRevoked, error) {
	event := new(ContractRoleRevoked)
	if err := _Contract.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractSelfMintingDerivativeDeployedIterator is returned from FilterSelfMintingDerivativeDeployed and is used to iterate over the raw logs and unpacked data for SelfMintingDerivativeDeployed events raised by the Contract contract.
type ContractSelfMintingDerivativeDeployedIterator struct {
	Event *ContractSelfMintingDerivativeDeployed // Event containing the contract specifics and raw log

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
func (it *ContractSelfMintingDerivativeDeployedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractSelfMintingDerivativeDeployed)
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
		it.Event = new(ContractSelfMintingDerivativeDeployed)
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
func (it *ContractSelfMintingDerivativeDeployedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractSelfMintingDerivativeDeployedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractSelfMintingDerivativeDeployed represents a SelfMintingDerivativeDeployed event raised by the Contract contract.
type ContractSelfMintingDerivativeDeployed struct {
	SelfMintingDerivativeVersion uint8
	SelfMintingDerivative        common.Address
	Raw                          types.Log // Blockchain specific contextual infos
}

// FilterSelfMintingDerivativeDeployed is a free log retrieval operation binding the contract event 0x246f71ad275621aa67be570521951d4047b41afe97c13a14733dae8ec481331d.
//
// Solidity: event SelfMintingDerivativeDeployed(uint8 indexed selfMintingDerivativeVersion, address indexed selfMintingDerivative)
func (_Contract *ContractFilterer) FilterSelfMintingDerivativeDeployed(opts *bind.FilterOpts, selfMintingDerivativeVersion []uint8, selfMintingDerivative []common.Address) (*ContractSelfMintingDerivativeDeployedIterator, error) {

	var selfMintingDerivativeVersionRule []interface{}
	for _, selfMintingDerivativeVersionItem := range selfMintingDerivativeVersion {
		selfMintingDerivativeVersionRule = append(selfMintingDerivativeVersionRule, selfMintingDerivativeVersionItem)
	}
	var selfMintingDerivativeRule []interface{}
	for _, selfMintingDerivativeItem := range selfMintingDerivative {
		selfMintingDerivativeRule = append(selfMintingDerivativeRule, selfMintingDerivativeItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "SelfMintingDerivativeDeployed", selfMintingDerivativeVersionRule, selfMintingDerivativeRule)
	if err != nil {
		return nil, err
	}
	return &ContractSelfMintingDerivativeDeployedIterator{contract: _Contract.contract, event: "SelfMintingDerivativeDeployed", logs: logs, sub: sub}, nil
}

// WatchSelfMintingDerivativeDeployed is a free log subscription operation binding the contract event 0x246f71ad275621aa67be570521951d4047b41afe97c13a14733dae8ec481331d.
//
// Solidity: event SelfMintingDerivativeDeployed(uint8 indexed selfMintingDerivativeVersion, address indexed selfMintingDerivative)
func (_Contract *ContractFilterer) WatchSelfMintingDerivativeDeployed(opts *bind.WatchOpts, sink chan<- *ContractSelfMintingDerivativeDeployed, selfMintingDerivativeVersion []uint8, selfMintingDerivative []common.Address) (event.Subscription, error) {

	var selfMintingDerivativeVersionRule []interface{}
	for _, selfMintingDerivativeVersionItem := range selfMintingDerivativeVersion {
		selfMintingDerivativeVersionRule = append(selfMintingDerivativeVersionRule, selfMintingDerivativeVersionItem)
	}
	var selfMintingDerivativeRule []interface{}
	for _, selfMintingDerivativeItem := range selfMintingDerivative {
		selfMintingDerivativeRule = append(selfMintingDerivativeRule, selfMintingDerivativeItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "SelfMintingDerivativeDeployed", selfMintingDerivativeVersionRule, selfMintingDerivativeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractSelfMintingDerivativeDeployed)
				if err := _Contract.contract.UnpackLog(event, "SelfMintingDerivativeDeployed", log); err != nil {
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

// ParseSelfMintingDerivativeDeployed is a log parse operation binding the contract event 0x246f71ad275621aa67be570521951d4047b41afe97c13a14733dae8ec481331d.
//
// Solidity: event SelfMintingDerivativeDeployed(uint8 indexed selfMintingDerivativeVersion, address indexed selfMintingDerivative)
func (_Contract *ContractFilterer) ParseSelfMintingDerivativeDeployed(log types.Log) (*ContractSelfMintingDerivativeDeployed, error) {
	event := new(ContractSelfMintingDerivativeDeployed)
	if err := _Contract.contract.UnpackLog(event, "SelfMintingDerivativeDeployed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
