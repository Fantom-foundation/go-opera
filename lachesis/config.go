package lachesis

import (
	"math/big"
	"time"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/params"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

const (
	MainNetworkID uint64 = 1
	TestNetworkID uint64 = 2
	FakeNetworkID uint64 = 3
)

var (
	// PercentUnit is used to define ratios with integers
	PercentUnit = big.NewInt(1e6)
)

// GasPowerConfig defines gas power rules in the consensus.
type GasPowerConfig struct {
	InitialAllocPerSec uint64          `json:"initialAllocPerSec"`
	MaxAllocPerSec     uint64          `json:"maxAllocPerSec"`
	MinAllocPerSec     uint64          `json:"minAllocPerSec"`
	MaxAllocPeriod     inter.Timestamp `json:"maxAllocPeriod"`
	StartupAllocPeriod inter.Timestamp `json:"startupAllocPeriod"`
	MinStartupGas      uint64          `json:"minStartupGas"`
}

// DagConfig of Lachesis DAG (directed acyclic graph).
type DagConfig struct {
	MaxParents     int `json:"maxParents"`
	MaxFreeParents int `json:"maxFreeParents"` // maximum number of parents with no gas cost

	MaxEpochBlocks   idx.Frame     `json:"maxEpochBlocks"`
	MaxEpochDuration time.Duration `json:"maxEpochDuration"`

	VectorClockConfig vector.IndexConfig `json:"vectorClockConfig"`

	MaxValidatorEventsInBlock idx.Event `json:"maxValidatorEventsInBlock"`
}

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	Num    idx.Block
	Period time.Duration
}

// EconomyConfig contains economy constants
type EconomyConfig struct {
	PoiPeriodDuration       time.Duration
	BlockMissedLatency      idx.Block
	OfflinePenaltyThreshold BlocksMissed
	TxRewardPoiImpact       *big.Int
	InitialRewardPerSecond  *big.Int
	MaxRewardPerSecond      *big.Int

	ShortGasPower GasPowerConfig `json:"shortGasPower"`
	LongGasPower  GasPowerConfig `json:"longGasPower"`
}

// BlockchainConfig contains transactions model constants
type BlockchainConfig struct {
	BlockGasHardLimit uint64 `json:"maxBlockGasLimit"` // technical hard limit, gas is mostly governed by gas power allocation
}

// Config describes lachesis net.
type Config struct {
	Name      string
	NetworkID uint64

	Genesis genesis.Genesis

	// Graph options
	Dag DagConfig

	// Blockchain options
	Blockchain BlockchainConfig

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
		Blockchain: BlockchainConfig{
			BlockGasHardLimit: 20000000,
		},
	}
}

func TestNetConfig() Config {
	return Config{
		Name:      "test",
		NetworkID: TestNetworkID,
		Genesis:   genesis.TestGenesis(),
		Dag:       DefaultDagConfig(),
		Economy:   DefaultEconomyConfig(),
		Blockchain: BlockchainConfig{
			BlockGasHardLimit: 20000000,
		},
	}
}

func FakeNetConfig(accs genesis.VAccounts) Config {
	return Config{
		Name:      "fake",
		NetworkID: FakeNetworkID,
		Genesis:   genesis.FakeGenesis(accs),
		Dag:       FakeNetDagConfig(),
		Economy:   FakeEconomyConfig(),
		Blockchain: BlockchainConfig{
			BlockGasHardLimit: 20000000,
		},
	}
}

// DefaultEconomyConfig returns mainnet economy
func DefaultEconomyConfig() EconomyConfig {
	// 45%
	txRewardPoiImpact := new(big.Int).Mul(big.NewInt(45), PercentUnit)
	txRewardPoiImpact.Div(txRewardPoiImpact, big.NewInt(100))

	initialRewardPerSecond := big.NewInt(8241994292233796296) // 8.241994 FTM per sec, 712108.306849 FTM per day
	maxRewardPerSecond := new(big.Int).Mul(initialRewardPerSecond, big.NewInt(100))

	return EconomyConfig{
		PoiPeriodDuration:      30 * 24 * time.Hour,
		BlockMissedLatency:     3,
		TxRewardPoiImpact:      txRewardPoiImpact,
		InitialRewardPerSecond: big.NewInt(8241994292233796296), // 8.241994 FTM per sec, 712108.306849 FTM per day
		MaxRewardPerSecond:     maxRewardPerSecond,
		OfflinePenaltyThreshold: BlocksMissed{
			Num:    1000,
			Period: 24 * time.Hour,
		},
		ShortGasPower: DefaultShortGasPowerConfig(),
		LongGasPower:  DefaulLongGasPowerConfig(),
	}
}

// FakeEconomyConfig returns fakenet economy
func FakeEconomyConfig() EconomyConfig {
	cfg := DefaultEconomyConfig()
	cfg.PoiPeriodDuration = 15 * time.Minute
	cfg.OfflinePenaltyThreshold.Period = 10 * time.Minute
	cfg.OfflinePenaltyThreshold.Num = 10
	cfg.ShortGasPower = FakeShortGasPowerConfig()
	cfg.LongGasPower = FakeLongGasPowerConfig()
	return cfg
}

func DefaultDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                5,
		MaxFreeParents:            3,
		MaxEpochBlocks:            1000,
		MaxValidatorEventsInBlock: 50,
		VectorClockConfig:         vector.DefaultIndexConfig(),
	}
}

func FakeNetDagConfig() DagConfig {
	cfg := DefaultDagConfig()
	cfg.MaxEpochBlocks = 200
	return cfg
}

// DefaulLongGasPowerConfig is long-window config
func DefaulLongGasPowerConfig() GasPowerConfig {
	return GasPowerConfig{
		InitialAllocPerSec: 100 * params.EventGas,
		MaxAllocPerSec:     1000 * params.EventGas,
		MinAllocPerSec:     10 * params.EventGas,
		MaxAllocPeriod:     inter.Timestamp(60 * time.Minute),
		StartupAllocPeriod: inter.Timestamp(5 * time.Second),
		MinStartupGas:      params.EventGas * 20,
	}
}

// DefaultShortGasPowerConfig is short-window config
func DefaultShortGasPowerConfig() GasPowerConfig {
	// 5x faster allocation rate, 12x lower max accumulated gas power
	cfg := DefaulLongGasPowerConfig()
	cfg.InitialAllocPerSec *= 5
	cfg.MaxAllocPerSec *= 5
	cfg.MinAllocPerSec *= 5
	cfg.StartupAllocPeriod /= 5
	cfg.MaxAllocPeriod /= 5 * 12
	return cfg
}

// FakeLongGasPowerConfig is fake long-window config
func FakeLongGasPowerConfig() GasPowerConfig {
	config := DefaulLongGasPowerConfig()
	config.InitialAllocPerSec *= 1000
	config.MaxAllocPerSec *= 1000
	return config
}

// FakeShortGasPowerConfig is fake short-window config
func FakeShortGasPowerConfig() GasPowerConfig {
	config := DefaultShortGasPowerConfig()
	config.InitialAllocPerSec *= 1000
	config.MaxAllocPerSec *= 1000
	return config
}
