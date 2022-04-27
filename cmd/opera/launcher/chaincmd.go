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
		Category:  "MISCELLANEOUS COMMANDS",
		Description: `
    opera import events

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
				},
				Description: `
The import command imports events from RLP-encoded files.
Events are fully verified by default, unless overridden by --check=false flag.`,
			},
			{
				Action:    utils.MigrateFlags(importEvm),
				Name:      "evm",
				Usage:     "Import EVM storage",
				ArgsUsage: "<filename> (<filename 2> ... <filename N>)",
				Flags: []cli.Flag{
					DataDirFlag,
					EventsCheckFlag,
				},
				Description: `
    opera import evm

The import command imports EVM storage (trie nodes, code, preimages) from files.`,
			},
			{
				Name:      "txtraces",
				Usage:     "Import transaction traces",
				ArgsUsage: "<filename>",
				Action:    utils.MigrateFlags(importTxTraces),
				Flags: []cli.Flag{
					DataDirFlag,
				},
				Description: `
    opera import txtraces

The import command imports transaction traces and replaces the old ones 
with traces from a file.
`,
			},
		},
	}
	exportCommand = cli.Command{
		Name:     "export",
		Usage:    "Export blockchain",
		Category: "MISCELLANEOUS COMMANDS",

		Subcommands: []cli.Command{
			{
				Name:      "events",
				Usage:     "Export blockchain events",
				ArgsUsage: "<filename> [<epochFrom> <epochTo>]",
				Action:    utils.MigrateFlags(exportEvents),
				Flags: []cli.Flag{
					DataDirFlag,
				},
				Description: `
    opera export events

Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last epoch to write. If the file ends with .gz, the output will
be gzipped
`,
			},
			{
				Name:      "txtraces",
				Usage:     "Export stored transaction traces",
				ArgsUsage: "<filename> [<blockFrom> <blockTo>]",
				Action:    utils.MigrateFlags(exportTxTraces),
				Flags: []cli.Flag{
					DataDirFlag,
				},
				Description: `
    opera export txtraces

Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last block to write transaction traces. If the file ends with .gz, the output will
be gzipped
`,
			},
		},
	}
	deleteCommand = cli.Command{
		Name:     "delete",
		Usage:    "Delete blockchain data",
		Category: "MISCELLANEOUS COMMANDS",

		Subcommands: []cli.Command{
			{
				Name:      "txtraces",
				Usage:     "Delete transaction traces",
				ArgsUsage: "[<blockFrom> <blockTo>]",
				Action:    utils.MigrateFlags(deleteTxTraces),
				Flags: []cli.Flag{
					DataDirFlag,
				},
				Description: `
    opera delete txtraces

Optional first and second arguments control the first and
last block to delete transaction traces from. If the file ends with .gz, the output will
be gzipped
`,
			},
		},
	}
	checkCommand = cli.Command{
		Name:     "check",
		Usage:    "Check blockchain",
		Category: "MISCELLANEOUS COMMANDS",

		Subcommands: []cli.Command{
			{
				Name:   "evm",
				Usage:  "Check EVM storage",
				Action: utils.MigrateFlags(checkEvm),
				Flags: []cli.Flag{
					DataDirFlag,
				},
				Description: `
    opera check evm

Checks EVM storage roots and code hashes
`,
			},
		},
	}
)
