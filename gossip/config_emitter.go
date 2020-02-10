package gossip

import (
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_params "github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/params"
)

// EmitIntervals is the configuration of emit intervals.
type EmitIntervals struct {
	Min                time.Duration `json:"min"`
	Max                time.Duration `json:"max"`
	Confirming         time.Duration `json:"confirming"` // emit time when there's no txs to originate, but at least 1 tx to confirm
	SelfForkProtection time.Duration `json:"selfForkProtection"`
}

// EmitterConfig is the configuration of events emitter.
type EmitterConfig struct {
	VersionToPublish string

	Validator common.Address `json:"validator"`

	EmitIntervals EmitIntervals `json:"emitIntervals"` // event emission intervals

	MaxGasRateGrowthFactor float64 `json:"maxGasRateGrowthFactor"` // fine to use float, because no need in determinism

	MaxTxsFromSender int `json:"maxTxsFromSender"`

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

		EmitIntervals: EmitIntervals{
			Min:                200 * time.Millisecond,
			Max:                12 * time.Minute,
			Confirming:         200 * time.Millisecond,
			SelfForkProtection: 30 * time.Minute, // should be at least 2x of MaxEmitInterval
		},

		MaxGasRateGrowthFactor: 3.0,
		MaxTxsFromSender:       TxTurnNonces,
		EpochTailLength:        1,

		MaxParents: 7,

		SmoothTpsThreshold: (params.EventGas + params.TxGas) * 500,
		NoTxsThreshold:     params.EventGas * 30,
		EmergencyThreshold: params.EventGas * 5,
	}
}

// RandomizeEmitTime and return new config
func (cfg *EmitIntervals) RandomizeEmitTime(r *rand.Rand) *EmitIntervals {
	config := *cfg
	// value = value - 0.1 * value + 0.1 * random value
	if config.Max > 10 {
		config.Max = config.Max - config.Max/10 + time.Duration(r.Int63n(int64(config.Max/10)))
	}
	// value = value + 0.1 * random value
	if config.SelfForkProtection > 10 {
		config.SelfForkProtection = config.SelfForkProtection + time.Duration(r.Int63n(int64(config.SelfForkProtection/10)))
	}
	return &config
}

// FakeEmitterConfig returns the testing configurations for the events emitter.
func FakeEmitterConfig() EmitterConfig {
	cfg := DefaultEmitterConfig()
	cfg.EmitIntervals.Max = 10 * time.Second // don't wait long in fakenet
	cfg.EmitIntervals.SelfForkProtection = cfg.EmitIntervals.Max * 3 / 2
	return cfg
}
