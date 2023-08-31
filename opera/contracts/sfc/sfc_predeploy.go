package sfc

import (
	"github.com/Fantom-foundation/go-opera/gossip/contract/sfc100"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// GetContractBin is SFC contract genesis implementation bin code
// Has to be compiled with flag bin-runtime
// Built from opera-sfc be4e79d5a5a425f08efd6d65b588a72ae90f706f, solc 0.5.17+commit.d19bba13.Emscripten.clang, optimize-runs 200
func GetContractBin() []byte {
	return hexutil.MustDecode(sfc100.ContractBinRuntime)
}

// ContractAddress is the SFC contract address
var ContractAddress = common.HexToAddress("0xfc00face00000000000000000000000000000000")
