package main

import (
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/node"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/cmd/lachesis/metrics"
	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/integration"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/utils/errlock"
)

const (
	// statsReportLimit is the time limit during import and export after which we
	// always print out progress. This avoids the user wondering what's going on.
	statsReportLimit = 8 * time.Second

	importBatchSize = 2500

	importFlushBatch = 300
)

var (
	importCommand = cli.Command{
		Action:    utils.MigrateFlags(importChain),
		Name:      "import",
		Usage:     "Import a blockchain file",
		ArgsUsage: "<filename> (<filename 2> ... <filename N>) ",
		Flags: []cli.Flag{
			DataDirFlag,
			utils.CacheFlag,
			utils.SyncModeFlag,
			utils.GCModeFlag,
			utils.CacheDatabaseFlag,
			utils.CacheGCFlag,
		},
		Category: "MISCELLANEOUS COMMANDS",
		Description: `
The import command imports blocks(events) from an RLP-encoded form. The form can be one file
with several RLP-encoded blocks(events), or several files can be used.
If only one file is used, import error will result in failure. If several files are used,
processing will proceed even if an individual RLP-file import failure occurs.`,
	}
	exportCommand = cli.Command{
		Action:    utils.MigrateFlags(exportChain),
		Name:      "export",
		Usage:     "Export blockchain into file",
		ArgsUsage: "<filename> [<blockNumFirst> <blockNumLast>]",
		Flags: []cli.Flag{
			DataDirFlag,
			utils.CacheFlag,
			utils.SyncModeFlag,
			utils.GCModeFlag,
		},
		Category: "MISCELLANEOUS COMMANDS",
		Description: `
Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last block to write. In this mode, the file will be appended
if already existing. If the file ends with .gz, the output will
be gzipped.`,
	}
)

func makeFNode(ctx *cli.Context, gossipCreate bool) (config, *node.Node, gossip.Consensus, kvdb.KeyValueStore, *gossip.Store) {
	cfg := makeAllConfigs(ctx)

	// check errlock file
	errlock.SetDefaultDatadir(cfg.Node.DataDir)
	errlock.Check()

	stack := makeConfigNode(ctx, &cfg.Node)
	defer stack.Close()

	engine, dbs, gdb := integration.MakeEngine(cfg.Node.DataDir, &cfg.Lachesis)
	metrics.SetDataDir(cfg.Node.DataDir)

	if gossipCreate {
		// Create and register a gossip network service. This is done through the definition
		// of a node.ServiceConstructor that will instantiate a node.Service. The reason for
		// the factory method approach is to support service restarts without relying on the
		// individual implementations' support for such operations.
		gossipService := func(ctx *node.ServiceContext) (node.Service, error) {
			return gossip.NewService(ctx, &cfg.Lachesis, gdb, engine)
		}

		if err := stack.Register(gossipService); err != nil {
			utils.Fatalf("Failed to register the service: %v", err)
		}
	}

	return cfg, stack, engine, dbs.GetDb("gossip-main"), gdb
}
