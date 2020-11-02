package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BigTo256 converts big number to 32 bytes array
func BigTo256(b *big.Int) common.Hash {
	return common.BytesToHash(b.Bytes())
}

// U64to256 converts uint64 to 32 bytes array
func U64to256(u64 uint64) common.Hash {
	return BigTo256(new(big.Int).SetUint64(u64))
}

// U64toBig converts uint64 to big number
func U64toBig(u64 uint64) *big.Int {
	return new(big.Int).SetUint64(u64)
}

// I64to256 converts int64 to 32 bytes array
func I64to256(i64 int64) common.Hash {
	return BigTo256(new(big.Int).SetInt64(i64))
}
