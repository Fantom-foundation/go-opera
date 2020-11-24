package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

// gas settings
const (
	// MaxGasPowerUsed - max value of Gas Power used in one event
	MaxGasPowerUsed = 10000000 + EventGas
	MaxExtraData    = 128

	EventGas  = 28000
	ParentGas = 2400
	// ExtraDataGas is cost per byte of extra event data. It's higher than regular data price, because it's a part of the header
	ExtraDataGas = 25

	TxGas = params.TxGas
)

var (
	// MinGasPrice is minimum possible gas price for a transaction
	MinGasPrice = big.NewInt(1e9)
)
