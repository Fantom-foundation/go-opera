package app

import (
	"math/big"
)

// SfcConstants are constants which may be changed by SFC contract
type SfcConstants struct {
	ShortGasPowerAllocPerSec uint64
	LongGasPowerAllocPerSec  uint64
	BaseRewardPerSec         *big.Int
}
