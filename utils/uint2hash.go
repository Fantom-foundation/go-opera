package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func U64to256(u64 uint64) common.Hash {
	return common.BytesToHash(new(big.Int).SetUint64(u64).Bytes())
}

func I64to256(i64 int64) common.Hash {
	return common.BytesToHash(new(big.Int).SetInt64(i64).Bytes())
}
