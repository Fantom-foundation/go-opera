package gossip

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type EmitterConfig struct {
	Emitbase common.Address

	MinEmitInterval time.Duration // minimum event emission interval
	MaxEmitInterval time.Duration // maximum event emission interval

	MaxGasRateGrowthFactor float64 // fine to use float, because no need in determinism

	// thresholds on GasLeft
	SmoothTpsThreshold uint64 `json:"smoothTpsThreshold"`
	NoTxsThreshold     uint64 `json:"noTxsThreshold"`
	EmergencyThreshold uint64 `json:"emergencyThreshold"`
}

func DefaultEmitterConfig() EmitterConfig {
	return EmitterConfig{
		MinEmitInterval:        1 * time.Second,
		MaxEmitInterval:        60 * time.Second,
		MaxGasRateGrowthFactor: 3.0,

		SmoothTpsThreshold: params.TxGas * 500,
		NoTxsThreshold:     params.TxGas * 100,
		EmergencyThreshold: params.TxGas * 5,
	}
}
