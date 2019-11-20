package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func BigTo256(b *big.Int) common.Hash {
	return common.BytesToHash(b.Bytes())
}

func U64to256(u64 uint64) common.Hash {
	return BigTo256(new(big.Int).SetUint64(u64))
}

func I64to256(i64 int64) common.Hash {
	return BigTo256(new(big.Int).SetInt64(i64))
}
