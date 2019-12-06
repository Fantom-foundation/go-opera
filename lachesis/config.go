package lachesis

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

const (
	MainNetworkID uint64 = 1
	TestNetworkID uint64 = 2
	FakeNetworkID uint64 = 3
)

var (
	PercentUnit = big.NewInt(1e6)
)

// GasPowerConfig defines gas power rules in the consensus.
type GasPowerConfig struct {
	TotalPerH          uint64          `json:"totalPerH"`
	MaxGasPowerPeriod  inter.Timestamp `json:"maxGasPowerPeriod"`
	StartupPeriod      inter.Timestamp `json:"startupPeriod"`
	MinStartupGasPower uint64          `json:"minStartupGasPower"`
}

// DagConfig of Lachesis DAG (directed acyclic graph).
type DagConfig struct {
	MaxParents                int       `json:"maxParents"`
	MaxFreeParents            int       `json:"maxFreeParents"` // maximum number of parents with no gas cost
	EpochLen                  idx.Frame `json:"epochLen"`
	MaxValidatorEventsInBlock idx.Event `json:"maxValidatorEventsInBlock"`

	GasPower GasPowerConfig `json:"gasPower"`

	IndexConfig vector.IndexConfig `json:"indexConfig"`
}

type EconomyConfig struct {
	ScoreCheckpointsInterval time.Duration
	PoiPeriodDuration        time.Duration
	BlockMissedLatency       idx.Block
	ValidatorPoiImpact       *big.Int
	RewardPerSecond          *big.Int
}

// Config describes lachesis net.
type Config struct {
	Name      string
	NetworkID uint64

	Genesis genesis.Genesis

	// Graph options
	Dag DagConfig

	// Economy options
	Economy EconomyConfig
}

func MainNetConfig() Config {
	return Config{
		Name:      "main",
		NetworkID: MainNetworkID,
		Genesis:   genesis.MainGenesis(),
		Dag:       DefaultDagConfig(),
		Economy:   DefaultEconomyConfig(),
	}
}

func TestNetConfig() Config {
	return Config{
		Name:      "test",
		NetworkID: TestNetworkID,
		Genesis:   genesis.TestGenesis(),
		Dag:       DefaultDagConfig(),
		Economy:   DefaultEconomyConfig(),
	}
}

func FakeNetConfig(accs genesis.VAccounts) Config {
	return Config{
		Name:      "fake",
		NetworkID: FakeNetworkID,
		Genesis:   genesis.FakeGenesis(accs),
		Dag:       FakeNetDagConfig(),
		Economy:   FakeEconomyConfig(),
	}
}

func DefaultEconomyConfig() EconomyConfig {
	// 30%
	validatorPoiImpact := big.NewInt(30)
	validatorPoiImpact.Mul(validatorPoiImpact, PercentUnit)
	validatorPoiImpact.Div(validatorPoiImpact, big.NewInt(100))

	return EconomyConfig{
		ScoreCheckpointsInterval: 30 * 24 * time.Hour,
		PoiPeriodDuration:        30 * 24 * time.Hour,
		BlockMissedLatency:       6,
		ValidatorPoiImpact:       validatorPoiImpact,
		RewardPerSecond:          big.NewInt(8.24199429223 * 1e18), // 712108.306849 FTM per day
	}
}

func FakeEconomyConfig() EconomyConfig {
	cfg := DefaultEconomyConfig()
	cfg.ScoreCheckpointsInterval = 5 * time.Minute
	cfg.PoiPeriodDuration = 1 * time.Minute
	return cfg
}

func DefaultDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                5,
		MaxFreeParents:            3,
		EpochLen:                  500,
		MaxValidatorEventsInBlock: 50,
		GasPower:                  DefaultGasPowerConfig(),
		IndexConfig:               vector.DefaultIndexConfig(),
	}
}

func FakeNetDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                5,
		MaxFreeParents:            3,
		EpochLen:                  500,
		MaxValidatorEventsInBlock: 50,
		GasPower:                  FakeNetGasPowerConfig(),
		IndexConfig:               vector.DefaultIndexConfig(),
	}
}

func DefaultGasPowerConfig() GasPowerConfig {
	return GasPowerConfig{
		TotalPerH:          50 * params.TxGas * 60 * 60,
		MaxGasPowerPeriod:  inter.Timestamp(10 * time.Minute),
		StartupPeriod:      inter.Timestamp(5 * time.Second),
		MinStartupGasPower: params.TxGas * 20,
	}
}

func FakeNetGasPowerConfig() GasPowerConfig {
	config := DefaultGasPowerConfig()
	config.TotalPerH *= 1000
	config.MinStartupGasPower *= 1000
	return config
}
