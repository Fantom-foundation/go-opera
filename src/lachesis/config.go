package lachesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
	"math/big"

	"github.com/ethereum/go-ethereum/eth/gasprice"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/evm_core"
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

	Genesis genesis.Genesis

	// Graph options
	Dag DagConfig

	// Transaction pool options
	TxPool evm_core.TxPoolConfig

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
		Dag:       DagConfig{3},
		TxPool:    evm_core.DefaultTxPoolConfig,
	}
}

func TestNetConfig() Config {
	return Config{
		Name:      "test",
		NetworkId: TestNetworkId,
		Genesis:   genesis.TestGenesis(),
		Dag:       DagConfig{3},
		TxPool:    evm_core.DefaultTxPoolConfig,
	}
}

func FakeNetConfig(n int) (Config, []hash.Peer, []*crypto.PrivateKey) {
	g, nodes, keys := genesis.FakeGenesis(n)

	return Config{
		Name:      "fake",
		NetworkId: FakeNetworkId,
		Genesis:   g,
		Dag:       DagConfig{3},
		TxPool:    evm_core.DefaultTxPoolConfig,
	}, nodes, keys
}
