package lachesis

import (
	"time"

	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
)

const (
	MainNetworkID uint64 = 1
	TestNetworkID uint64 = 2
	FakeNetworkID uint64 = 3
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
}

// Config describes lachesis net.
type Config struct {
	Name      string
	NetworkID uint64

	Genesis genesis.Genesis

	// Graph options
	Dag DagConfig
}

func MainNetConfig() Config {
	return Config{
		Name:      "main",
		NetworkID: MainNetworkID,
		Genesis:   genesis.MainGenesis(),
		Dag:       DefaultDagConfig(),
	}
}

func TestNetConfig() Config {
	return Config{
		Name:      "test",
		NetworkID: TestNetworkID,
		Genesis:   genesis.TestGenesis(),
		Dag:       DefaultDagConfig(),
	}
}

func FakeNetConfig(n int) Config {
	g := genesis.FakeGenesis(n)

	return Config{
		Name:      "fake",
		NetworkID: FakeNetworkID,
		Genesis:   g,
		Dag:       FakeNetDagConfig(),
	}
}

func DefaultDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                5,
		MaxFreeParents:            3,
		EpochLen:                  100,
		MaxValidatorEventsInBlock: 50,
		GasPower:                  DefaultGasPowerConfig(),
	}
}

func FakeNetDagConfig() DagConfig {
	return DagConfig{
		MaxParents:                5,
		MaxFreeParents:            3,
		EpochLen:                  50,
		MaxValidatorEventsInBlock: 50,
		GasPower:                  FakeNetGasPowerConfig(),
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
	config.TotalPerH *= 10
	config.MinStartupGasPower *= 10
	return config
}
