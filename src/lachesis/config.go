package lachesis

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/cfg_emitter"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/cfg_gossip"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/gasprice"
	"math/big"
	"time"
)

type DagConfig struct {
	MaxParents int `json:"maxParents"`
}

// Net describes lachesis net.
type Net struct {
	Name    string
	Genesis *genesis.Config

	// Gossip options
	Gossip *cfg_gossip.Config

	// Emitter options
	Emitter *cfg_emitter.Config

	// Graph options
	Dag *DagConfig

	// Database options
	SkipDagVersionCheck bool `toml:"-"`
	DatabaseHandles     int  `toml:"-"`
	DatabaseCache       int
	DatabaseFreezer     string

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration

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

func MainNet() *Net {
	return &Net{
		Genesis: genesis.MainNet(),
		Gossip: &cfg_gossip.Config{
			ForcedBroadcast: true,
		},
		Emitter: &cfg_emitter.Config{
			MinEmitInterval: 1 * time.Second,
			MaxEmitInterval: 60 * time.Second,
		},
		Dag: &DagConfig{3},
	}
}

func TestNet() *Net {
	config := MainNet()
	config.Genesis = genesis.TestNet()
	config.Emitter.MaxEmitInterval = 3 * time.Second
	return config
}

func EmptyFakeNet() *Net {
	config := TestNet()
	config.Genesis = genesis.EmptyFakeNet()
	return config
}

func FakeNet(n int) (*Net, []hash.Peer, []*crypto.PrivateKey) {
	config := EmptyFakeNet()
	g, nodes, keys := genesis.FakeNet(n)
	config.Genesis = g
	return config, nodes, keys
}
