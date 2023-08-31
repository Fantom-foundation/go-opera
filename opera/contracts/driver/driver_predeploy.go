package driver

import (
	"github.com/Fantom-foundation/go-opera/gossip/contract/driver100"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// GetContractBin is NodeDriver contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
// Built from opera-sfc be4e79d5a5a425f08efd6d65b588a72ae90f706f, solc 0.5.17+commit.d19bba13.Emscripten.clang, optimize-runs 10000
func GetContractBin() []byte {
	return hexutil.MustDecode(driver100.ContractBinRuntime)
}

// ContractAddress is the NodeDriver contract address
var ContractAddress = common.HexToAddress("0xd100a01e00000000000000000000000000000000")
