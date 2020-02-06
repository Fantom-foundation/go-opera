package gossip

import (
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_params "github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/params"
)

// EmitterConfig is the configuration of events emitter.
type EmitterConfig struct {
	VersionToPublish string

	Validator common.Address `json:"validator"`

	MinEmitInterval time.Duration `json:"minEmitInterval"` // minimum event emission interval
	MaxEmitInterval time.Duration `json:"maxEmitInterval"` // maximum event emission interval

	MaxGasRateGrowthFactor float64 `json:"maxGasRateGrowthFactor"` // fine to use float, because no need in determinism

	MaxTxsFromSender int `json:"maxTxsFromSender"`

	SelfForkProtectionInterval time.Duration `json:"selfForkProtectionInterval"`

	EpochTailLength idx.Frame `json:"epochTailLength"` // number of frames before event is considered epoch

	MaxParents int `json:"maxParents"`

	// thresholds on GasLeft
	SmoothTpsThreshold uint64 `json:"smoothTpsThreshold"`
	NoTxsThreshold     uint64 `json:"noTxsThreshold"`
	EmergencyThreshold uint64 `json:"emergencyThreshold"`
}

// DefaultEmitterConfig returns the default configurations for the events emitter.
func DefaultEmitterConfig() EmitterConfig {
	return EmitterConfig{
		VersionToPublish: _params.VersionWithMeta(),

		MinEmitInterval: 200 * time.Millisecond,
		MaxEmitInterval: 12 * time.Minute,

		MaxGasRateGrowthFactor:     3.0,
		MaxTxsFromSender:           TxTurnNonces,
		SelfForkProtectionInterval: 30 * time.Minute, // should be at least 2x of MaxEmitInterval
		EpochTailLength:            1,

		MaxParents: 7,

		SmoothTpsThreshold: (params.EventGas + params.TxGas) * 500,
		NoTxsThreshold:     params.EventGas * 30,
		EmergencyThreshold: params.EventGas * 5,
	}
}

// RandomizeEmitTime and return new config
func (cfg *EmitterConfig) RandomizeEmitTime(r *rand.Rand) *EmitterConfig {
	config := *cfg
	// value = value - 0.1 * value + 0.1 * random value
	if config.MaxEmitInterval > 10 {
		config.MaxEmitInterval = config.MaxEmitInterval - config.MaxEmitInterval/10 + time.Duration(r.Int63n(int64(config.MaxEmitInterval/10)))
	}
	// value = value + 0.1 * random value
	if config.SelfForkProtectionInterval > 10 {
		config.SelfForkProtectionInterval = config.SelfForkProtectionInterval + time.Duration(r.Int63n(int64(config.SelfForkProtectionInterval/10)))
	}
	return &config
}

// FakeEmitterConfig returns the testing configurations for the events emitter.
func FakeEmitterConfig() EmitterConfig {
	cfg := DefaultEmitterConfig()
	cfg.MaxEmitInterval = 10 * time.Second // don't wait long in fakenet
	cfg.SelfForkProtectionInterval = cfg.MaxEmitInterval * 3 / 2
	return cfg
}
