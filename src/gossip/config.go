package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/gasprice"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/inter/genesis"
	"github.com/Fantom-foundation/go-lachesis/src/params"
)

//go:generate gencodec -type Config -formats toml -out gen_config.go

type Config struct {
	// The genesis object, which is inserted if the database is empty.
	// If nil, the Fantom main net genesis is used.
	Genesis *genesis.Config `toml:",omitempty"`

	// Protocol options
	SyncMode downloader.SyncMode

	NoPruning       bool // Whether to disable pruning and flush everything to disk
	NoPrefetch      bool // Whether to disable prefetching and only load state on demand
	ForcedBroadcast bool

	// Database options
	SkipDagVersionCheck bool `toml:"-"`
	DatabaseHandles     int  `toml:"-"`
	DatabaseCache       int
	DatabaseFreezer     string

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    time.Duration

	// Emitter options
	Emitter EmitterConfig

	// Dag options
	Dag params.DagConfig

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

func MainNet() *Config {
	return &Config{
		Genesis: genesis.MainNet(),
		Emitter: EmitterConfig{
			MinEmitInterval: 1 * time.Second,
			MaxEmitInterval: 10 * time.Second,
		},
		ForcedBroadcast: true,
		Dag:             params.DefaultDagConfig,
	}
}

func TestNet() *Config {
	config := MainNet()
	config.Genesis = genesis.TestNet()
	return config
}

func FakeNet(n int) (*Config, []hash.Peer, []*crypto.PrivateKey) {
	config := MainNet()
	g, nodes, keys := genesis.FakeNet(n)
	config.Genesis = g
	return config, nodes, keys
}

func EmptyFakeNet() *Config {
	config := MainNet()
	config.Genesis = genesis.EmptyFakeNet()
	return config
}
