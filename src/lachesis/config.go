package lachesis

import (
	"math/big"

	"github.com/ethereum/go-ethereum/eth/gasprice"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

const (
	MainNetworkId uint64 = 1
	TestNetworkId uint64 = 2
	FakeNetworkId uint64 = 3
)

// DagConfig of DAG.
type DagConfig struct {
	MaxParents int       `json:"maxParents"`
	EpochLen   idx.Frame `json:"maxParents"`
}

// Config describes lachesis net.
type Config struct {
	Name      string
	NetworkId uint64

	Genesis genesis.Genesis

	// Graph options
	Dag DagConfig

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// Type of the EWASM interpreter ("" for default)
	EWASMInterpreter string

	// Type of the EVM interpreter ("" for default)
	EVMInterpreter string

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap *big.Int `toml:",omitempty"`
}

func MainNetConfig() Config {
	return Config{
		Name:      "main",
		NetworkId: MainNetworkId,
		Genesis:   genesis.MainGenesis(),
		Dag:       DefaultDagConfig(),
	}
}

func TestNetConfig() Config {
	return Config{
		Name:      "test",
		NetworkId: TestNetworkId,
		Genesis:   genesis.TestGenesis(),
		Dag:       DefaultDagConfig(),
	}
}

func FakeNetConfig(n int) Config {
	g := genesis.FakeGenesis(n)

	return Config{
		Name:      "fake",
		NetworkId: FakeNetworkId,
		Genesis:   g,
		Dag:       DefaultDagConfig(),
	}
}

func DefaultDagConfig() DagConfig {
	return DagConfig{
		MaxParents: 3,
		EpochLen:   100,
	}
}
