package gossip

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// EmitterConfig is the configuration of events emitter.
type EmitterConfig struct {
	Coinbase common.Address `json:"coinbase"`

	MinEmitInterval time.Duration `json:"minEmitInterval"` // minimum event emission interval
	MaxEmitInterval time.Duration `json:"maxEmitInterval"` // maximum event emission interval

	MaxGasRateGrowthFactor float64 `json:"maxGasRateGrowthFactor"` // fine to use float, because no need in determinism

	MaxTxsFromSender int `json:"maxTxsFromSender"`

	SelfForkProtectionInterval time.Duration `json:"selfForkProtectionInterval"`

	EpochTailLength idx.Frame `json:"epochTailLength"` // number of frames before event is considered epoch tail

	// thresholds on GasLeft
	SmoothTpsThreshold uint64 `json:"smoothTpsThreshold"`
	NoTxsThreshold     uint64 `json:"noTxsThreshold"`
	EmergencyThreshold uint64 `json:"emergencyThreshold"`
}

// DefaultEmitterConfig returns the default configurations for the events emitter.
func DefaultEmitterConfig() EmitterConfig {
	return EmitterConfig{
		MinEmitInterval:            500 * time.Millisecond,
		MaxEmitInterval:            10 * time.Minute,
		MaxGasRateGrowthFactor:     3.0,
		MaxTxsFromSender:           2,
		SelfForkProtectionInterval: 30 * time.Minute, // should be at least 2x of MaxEmitInterval
		EpochTailLength:            1,

		SmoothTpsThreshold: params.TxGas * 500,
		NoTxsThreshold:     params.TxGas * 100,
		EmergencyThreshold: params.TxGas * 5,
	}
}

// FakeEmitterConfig returns the testing configurations for the events emitter.
func FakeEmitterConfig() EmitterConfig {
	cfg := DefaultEmitterConfig()
	cfg.MaxEmitInterval = 10 * time.Second // don't wait long in fakenet
	cfg.SelfForkProtectionInterval = cfg.MaxEmitInterval * 3 / 2
	return cfg
}
