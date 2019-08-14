package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"log"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/node"
	"go.etcd.io/bbolt"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

const (
	// clientIdentifier to advertise over the network.
	clientIdentifier = "go-lachesis"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags).
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, gitDate, "the go-lachesis command line interface")
	// Flags that configure the node.
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.AncientFlag,
		utils.KeyStoreDirFlag,
		utils.ExternalSignerFlag,
		utils.NoUSBFlag,
		utils.SmartCardDaemonPathFlag,
		utils.DashboardEnabledFlag,
		utils.DashboardAddrFlag,
		utils.DashboardPortFlag,
		utils.DashboardRefreshFlag,
		utils.EthashCacheDirFlag,
		utils.EthashCachesInMemoryFlag,
		utils.EthashCachesOnDiskFlag,
		utils.EthashDatasetDirFlag,
		utils.EthashDatasetsInMemoryFlag,
		utils.EthashDatasetsOnDiskFlag,
		utils.TxPoolLocalsFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.SyncModeFlag,
		utils.ExitWhenSyncedFlag,
		utils.GCModeFlag,
		utils.LightServeFlag,
		utils.LightLegacyServFlag,
		utils.LightIngressFlag,
		utils.LightEgressFlag,
		utils.LightMaxPeersFlag,
		utils.LightLegacyPeersFlag,
		utils.LightKDFFlag,
		utils.UltraLightServersFlag,
		utils.UltraLightFractionFlag,
		utils.UltraLightOnlyAnnounceFlag,
		utils.WhitelistFlag,
		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheTrieFlag,
		utils.CacheGCFlag,
		utils.CacheNoPrefetchFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.RinkebyFlag,
		utils.GoerliFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.EthStatsURLFlag,
		utils.FakePoWFlag,
		utils.NoCompactionFlag,
		utils.GpoBlocksFlag,
		utils.GpoPercentileFlag,
		utils.EWASMInterpreterFlag,
		utils.EVMInterpreterFlag,
		configFileFlag,
	}
)

// init the CLI app.
func init() {
	overrideParams()

	app.Action = lachesis
	app.HideVersion = true // we have a command to print the version
	app.Commands = []cli.Command{
		dumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)

	app.Before = func(ctx *cli.Context) error {
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// lachesis is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func lachesis(ctx *cli.Context) error {
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	cfg := makeNodeConfig(ctx)

	var db *bbolt.DB
	if cfg.DataDir != "inmemory" {
		var stop func()
		var err error
		db, stop, err = openDB(cfg.DataDir)
		if err != nil {
			return err
		}
		defer stop()
	}

	node := makeFullNode(cfg, db)
	defer node.Close()

	utils.StartNode(node)
	node.Wait()
	return nil
}

func makeFullNode(cfg *node.Config, db *bbolt.DB) *node.Node {
	// Create BD.
	gdb, cdb := makeStorages(db)

	// Create consensus.
	concensus := posposet.New(cdb, gdb)

	// Create and register a gossip network service. This is done through the definition
	// of a node.ServiceConstructor that will instantiate a node.Service. The reason for
	// the factory method approach is to support service restarts without relying on the
	// individual implementations' support for such operations.
	constructor := func(ctx *node.ServiceContext) (node.Service, error) {
		return gossip.NewService(&gossip.DefaultConfig, new(event.TypeMux), gdb, concensus)
	}

	// Create node.
	stack, err := node.New(cfg)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	if err := stack.Register(constructor); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	return stack
}

func makeStorages(db *bbolt.DB) (*gossip.Store, *posposet.Store) {
	var (
		p      kvdb.Database
		n      kvdb.Database
		cached bool
	)
	if db == nil {
		p = kvdb.NewMemDatabase()
		n = kvdb.NewMemDatabase()
		cached = false
	} else {
		db := kvdb.NewBoltDatabase(db)
		p = db.NewTable([]byte("p_"))
		n = db.NewTable([]byte("n_"))
		cached = true
	}

	return gossip.NewStore(n),
		posposet.NewStore(p, cached)
}
