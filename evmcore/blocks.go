package evmcore

import (
	"github.com/ethereum/go-ethereum/common"
)

// BadHashes represent a set of manually tracked bad hashes (usually hard forks)
var BadHashes = map[common.Hash]bool{}
