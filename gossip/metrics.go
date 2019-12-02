package gossip

import (
	"github.com/ethereum/go-ethereum/metrics"

	"github.com/Fantom-foundation/go-lachesis/cmd/tx-storm/meta"
)

var (
	confirmBlocksMeter    = metrics.NewRegisteredCounter("confirm/blocks", nil)
	confirmTxnsMeter      = metrics.NewRegisteredCounter("confirm/transactions", nil)
	confirmTxLatencyMeter = metrics.NewRegisteredHistogram("confirm/txlatency", nil, metrics.NewUniformSample(10000))
)

var txLatency = meta.NewTxs()
