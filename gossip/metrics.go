package gossip

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	confirmBlocksMeter = metrics.NewRegisteredCounter("confirm/blocks", nil)
	confirmTxnsMeter   = metrics.NewRegisteredCounter("confirm/transactions", nil)
	//confirmTimeMeter = metrics.NewRegisteredHistogram("confirm/seconds", nil)
)
