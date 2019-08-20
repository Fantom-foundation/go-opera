package lachesis

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/gasprice"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

const (
	MainNetworkId uint64 = 1
	TestNetworkId uint64 = 2
	FakeNetworkId uint64 = 3
)

// DagConfig of DAG.
type DagConfig struct {
	MaxParents int `json:"maxParents"`
}

// Config describes lachesis net.
type Config struct {
	Name      string
	NetworkId uint64

	Genesis Genesis

	// Graph options
	Dag DagConfig

	// Transaction pool options
	TxPool core.TxPoolConfig

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
		Genesis:   MainGenesis(),
		Dag:       DagConfig{3},
	}
}

func TestNetConfig() Config {
	return Config{
		Name:      "test",
		NetworkId: TestNetworkId,
		Genesis:   TestGenesis(),
		Dag:       DagConfig{3},
	}
}

func FakeNetConfig(n int) (Config, []hash.Peer, []*crypto.PrivateKey) {
	g, nodes, keys := FakeGenesis(n)

	return Config{
		Name:      "fake",
		NetworkId: FakeNetworkId,
		Genesis:   g,
		Dag:       DagConfig{3},
	}, nodes, keys
}
