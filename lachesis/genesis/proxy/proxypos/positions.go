package proxypos

import (
	"github.com/ethereum/go-ethereum/common"
)

// Admin is position of Admin variable
func Admin() common.Hash {
	return common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")
}

// Implementation is position of Implementation variable
func Implementation() common.Hash {
	return common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")
}
