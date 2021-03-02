package emitter

import (
	"math/rand"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
	"github.com/Fantom-foundation/go-opera/opera"
)

// EmitIntervals is the configuration of emit intervals.
type EmitIntervals struct {
	Min                        time.Duration
	Max                        time.Duration
	Confirming                 time.Duration // emit time when there's no txs to originate, but at least 1 tx to confirm
	ParallelInstanceProtection time.Duration
	DoublesignProtection       time.Duration
}

type ValidatorConfig struct {
	ID     idx.ValidatorID
	PubKey validatorpk.PubKey
}

// Config is the configuration of events emitter.
type Config struct {
	VersionToPublish string

	Validator ValidatorConfig

	EmitIntervals EmitIntervals // event emission intervals

	MaxGasRateGrowthFactor float64 // fine to use float, because no need in determinism

	MaxTxsPerAddress int

	EpochTailLength idx.Frame // number of frames before event is considered epoch

	MaxParents idx.Event

	// thresholds on GasLeft
	LimitedTpsThreshold uint64
	NoTxsThreshold      uint64
	EmergencyThreshold  uint64
}

// DefaultConfig returns the default configurations for the events emitter.
func DefaultConfig() Config {
	return Config{
		VersionToPublish: params.VersionWithMeta(),

		EmitIntervals: EmitIntervals{
			Min:                        110 * time.Millisecond,
			Max:                        10 * time.Minute,
			Confirming:                 120 * time.Millisecond,
			DoublesignProtection:       27 * time.Minute, // should be greater than MaxEmitInterval
			ParallelInstanceProtection: 1 * time.Minute,
		},

		MaxGasRateGrowthFactor: 3.0,
		MaxTxsPerAddress:       TxTurnNonces / 3,
		EpochTailLength:        1,

		MaxParents: 0,

		LimitedTpsThreshold: (opera.DefaultEventMaxGas + params.TxGas) * 500,
		NoTxsThreshold:      opera.DefaultEventMaxGas * 30,
		EmergencyThreshold:  opera.DefaultEventMaxGas * 5,
	}
}

// RandomizeEmitTime and return new config
func (cfg EmitIntervals) RandomizeEmitTime(r *rand.Rand) EmitIntervals {
	config := cfg
	// value = value - 0.1 * value + 0.1 * random value
	if config.Max > 10 {
		config.Max = config.Max - config.Max/10 + time.Duration(r.Int63n(int64(config.Max/10)))
	}
	// value = value + 0.33 * random value
	if config.DoublesignProtection > 3 {
		config.DoublesignProtection = config.DoublesignProtection + time.Duration(r.Int63n(int64(config.DoublesignProtection/3)))
	}
	return config
}

// FakeConfig returns the testing configurations for the events emitter.
func FakeConfig(num int) Config {
	cfg := DefaultConfig()
	cfg.EmitIntervals.Max = 10 * time.Second // don't wait long in fakenet
	cfg.EmitIntervals.DoublesignProtection = cfg.EmitIntervals.Max / 2
	if num <= 1 {
		// disable self-fork protection if fakenet 1/1
		cfg.EmitIntervals.DoublesignProtection = 0
	}
	return cfg
}
