package launcher

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	EventsCheckFlag = cli.BoolTFlag{
		Name:  "check",
		Usage: "true if events should be fully checked before importing",
	}
	importCommand = cli.Command{
		Name:      "import",
		Usage:     "Import a blockchain file",
		ArgsUsage: "<filename> (<filename 2> ... <filename N>) [check=false]",
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
The import command imports events from an RLP-encoded files.
Events are fully verified by default, unless overridden by check=false flag.`,

		Subcommands: []cli.Command{
			{
				Action:    utils.MigrateFlags(importEvents),
				Name:      "events",
				Usage:     "Import blockchain events",
				ArgsUsage: "<filename> (<filename 2> ... <filename N>) [--check=false]",
				Flags: []cli.Flag{
					DataDirFlag,
					EventsCheckFlag,
					utils.CacheFlag,
					utils.SyncModeFlag,
					utils.GCModeFlag,
					utils.CacheDatabaseFlag,
					utils.CacheGCFlag,
				},
				Description: `
The import command imports events from an RLP-encoded files.
Events are fully verified by default, unless overridden by --check=false flag.`,
			},
		},
	}
	exportCommand = cli.Command{
		Name:  "export",
		Usage: "Export blockchain",
		Flags: []cli.Flag{
			DataDirFlag,
			utils.CacheFlag,
			utils.SyncModeFlag,
			utils.GCModeFlag,
		},
		Category: "MISCELLANEOUS COMMANDS",

		Subcommands: []cli.Command{
			{
				Name:      "events",
				Usage:     "Export blockchain events",
				ArgsUsage: "<filename> [<epochFrom> <epochTo>]",
				Action:    utils.MigrateFlags(exportEvents),
				Flags: []cli.Flag{
					DataDirFlag,
					utils.CacheFlag,
					utils.SyncModeFlag,
					utils.GCModeFlag,
				},
				Description: `
    lachesis export events

Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last epoch to write. If the file ends with .gz, the output will
be gzipped
`,
			},
		},
	}
)
