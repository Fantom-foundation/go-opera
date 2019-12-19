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
	// PercentUnit is used to define ratios with integers
	PercentUnit = big.NewInt(1e6)
)

// GasPowerConfig defines gas power rules in the consensus.
type GasPowerConfig struct {
	InitialAllocPerSec uint64          `json:"initialAllocPerSec"`
	MaxAllocPeriod     inter.Timestamp `json:"maxAllocPeriod"`
	StartupAllocPeriod inter.Timestamp `json:"startupAllocPeriod"`
	MinStartup         uint64          `json:"minStartup"`
}

// DagConfig of Lachesis DAG (directed acyclic graph).
type DagConfig struct {
	MaxParents     int       `json:"maxParents"`
	MaxFreeParents int       `json:"maxFreeParents"` // maximum number of parents with no gas cost
	EpochLen       idx.Frame `json:"epochLen"`

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
	RewardPerSecond         *big.Int

	GasPower GasPowerConfig `json:"gasPower"`
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

	return EconomyConfig{
		PoiPeriodDuration:  30 * 24 * time.Hour,
		BlockMissedLatency: 3,
		TxRewardPoiImpact:  txRewardPoiImpact,
		RewardPerSecond:    big.NewInt(8241994292233796296), // 8.241994 FTM per sec, 712108.306849 FTM per day
		OfflinePenaltyThreshold: BlocksMissed{
			Num:    1000,
			Period: 24 * time.Hour,
		},
		GasPower: DefaultGasPowerConfig(),
	}
}

// FakeEconomyConfig returns fakenet economy
func FakeEconomyConfig() EconomyConfig {
	cfg := DefaultEconomyConfig()
	cfg.PoiPeriodDuration = 15 * time.Minute
	cfg.OfflinePenaltyThreshold.Period = 10 * time.Minute
	cfg.OfflinePenaltyThreshold.Num = 10
	cfg.GasPower = FakeNetGasPowerConfig()
	return cfg
}

func DefaultDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                5,
		MaxFreeParents:            3,
		EpochLen:                  500,
		MaxValidatorEventsInBlock: 50,
	}
}

func FakeNetDagConfig() DagConfig {
	cfg := DefaultDagConfig()
	cfg.VectorClockConfig = vector.DefaultIndexConfig()
	return cfg
}

func DefaultGasPowerConfig() GasPowerConfig {
	return GasPowerConfig{
		InitialAllocPerSec: 50 * params.TxGas,
		MaxAllocPeriod:     inter.Timestamp(10 * time.Minute),
		StartupAllocPeriod: inter.Timestamp(5 * time.Second),
		MinStartup:         params.TxGas * 20,
	}
}

func FakeNetGasPowerConfig() GasPowerConfig {
	config := DefaultGasPowerConfig()
	config.InitialAllocPerSec *= 1000
	config.MinStartup *= 1000
	return config
}
