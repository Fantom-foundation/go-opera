package gossip

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/gasprice"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/params"
)

//go:generate gencodec -type Config -formats toml -out gen_config.go

type Config struct {
	// The genesis object, which is inserted if the database is empty.
	// If nil, the Fantom main net genesis is used.
	Genesis hash.Hash `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode

	NoPruning  bool // Whether to disable pruning and flush everything to disk
	NoPrefetch bool // Whether to disable prefetching and only load state on demand

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

var DefaultConfig = Config{
	Emitter: EmitterConfig{
		MinEmitInterval: 1 * time.Second,
		MaxEmitInterval: 10 * time.Second,
	},
	Dag: params.DefaultDagConfig,
}
