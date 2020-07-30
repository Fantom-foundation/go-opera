package main

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	importCommand = cli.Command{
		Action:    utils.MigrateFlags(importChain),
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
	}
	exportCommand = cli.Command{
		Action:    utils.MigrateFlags(exportChain),
		Name:      "export",
		Usage:     "Export blockchain into file",
		ArgsUsage: "<filename> [<epochFrom> <epochTo>]",
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
last epoch to write. If the file ends with .gz, the output will
be gzipped.`,
	}
)
