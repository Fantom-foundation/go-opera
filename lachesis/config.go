package lachesis

import (
	"math/big"
	"time"

	ethparams "github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/params"
	"github.com/Fantom-foundation/go-lachesis/utils"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

const (
	MainNetworkID uint64 = 0xfa
	TestNetworkID uint64 = 0xfa2
	FakeNetworkID uint64 = 0xfa3
)

var (
	// PercentUnit is used to define ratios with integers, it's 1.0
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
	BlocksNum idx.Block
	Period    time.Duration
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

// BlocksConfig contains blocks constants
type BlocksConfig struct {
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
	Blocks BlocksConfig

	// Economy options
	Economy EconomyConfig
}

// EvmChainConfig returns ChainConfig for transaction signing and execution
func (c *Config) EvmChainConfig() *ethparams.ChainConfig {
	cfg := *ethparams.AllEthashProtocolChanges
	cfg.ChainID = new(big.Int).SetUint64(c.NetworkID)
	return &cfg
}

func MainNetConfig() Config {
	return Config{
		Name:      "main",
		NetworkID: MainNetworkID,
		Genesis:   genesis.MainGenesis(),
		Dag:       DefaultDagConfig(),
		Economy:   DefaultEconomyConfig(),
		Blocks: BlocksConfig{
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
		Blocks: BlocksConfig{
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
		Blocks: BlocksConfig{
			BlockGasHardLimit: 20000000,
		},
	}
}

// DefaultEconomyConfig returns mainnet economy
func DefaultEconomyConfig() EconomyConfig {
	// 45%
	txRewardPoiImpact := new(big.Int).Mul(big.NewInt(45), PercentUnit)
	txRewardPoiImpact.Div(txRewardPoiImpact, big.NewInt(100))

	initialRewardPerSecond := new(big.Int).Add(utils.ToFtm(16), big.NewInt(483988584467592592)) // 16.483988584467592592 FTM per sec, or 1424216,6136 FTM per day
	maxRewardPerSecond := new(big.Int).Mul(initialRewardPerSecond, big.NewInt(2))

	return EconomyConfig{
		PoiPeriodDuration:      30 * 24 * time.Hour,
		BlockMissedLatency:     3,
		TxRewardPoiImpact:      txRewardPoiImpact,
		InitialRewardPerSecond: initialRewardPerSecond,
		MaxRewardPerSecond:     maxRewardPerSecond,
		OfflinePenaltyThreshold: BlocksMissed{
			BlocksNum: 1000,
			Period:    24 * time.Hour,
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
	cfg.OfflinePenaltyThreshold.BlocksNum = 10
	cfg.ShortGasPower = FakeShortGasPowerConfig()
	cfg.LongGasPower = FakeLongGasPowerConfig()
	return cfg
}

func DefaultDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                10,
		MaxFreeParents:            3,
		MaxEpochBlocks:            1000,
		MaxEpochDuration:          4 * time.Hour,
		MaxValidatorEventsInBlock: 50,
		VectorClockConfig:         vector.DefaultIndexConfig(),
	}
}

func FakeNetDagConfig() DagConfig {
	cfg := DefaultDagConfig()
	cfg.MaxEpochBlocks = 200
	cfg.MaxEpochDuration = 10 * time.Minute
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
