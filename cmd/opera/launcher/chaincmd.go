package launcher

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	EvmExportMode = cli.StringFlag{
		Name:  "export.evm.mode",
		Usage: `EVM export mode ("full" or "ext-mpt" or "mpt")`,
		Value: "mpt",
	}
	EvmExportExclude = cli.StringFlag{
		Name:  "export.evm.exclude",
		Usage: `DB of EVM keys to exclude from genesis`,
	}
	GenesisExportSections = cli.StringFlag{
		Name:  "export.sections",
		Usage: `Genesis sections to export separated by comma (e.g. "brs-1" or "ers" or "evm-2")`,
		Value: "brs,ers,evm",
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
				ArgsUsage: "<filename> (<filename 2> ... <filename N>)",
				Flags: []cli.Flag{
					DataDirFlag,
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
				},
				Description: `
    opera import evm

The import command imports EVM storage (trie nodes, code, preimages) from files.`,
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
last epoch to write. If the file ends with .gz, the output will be gzipped.
End the rest of file name with .dot to export events graph as DOT
`,
			},
			{
				Name:      "genesis",
				Usage:     "Export current state into a genesis file",
				ArgsUsage: "<filename or dry-run> [<epochFrom> <epochTo>] [--export.evm.mode=MODE --export.evm.exclude=DB_PATH --export.sections=A,B,C]",
				Action:    utils.MigrateFlags(exportGenesis),
				Flags: []cli.Flag{
					DataDirFlag,
					EvmExportMode,
					EvmExportExclude,
					GenesisExportSections,
				},
				Description: `
    opera export genesis

Export current state into a genesis file.
Requires a first argument of the file to write to.
Optional second and third arguments control the first and
last epoch to write.
Pass dry-run instead of filename for calculation of hashes without exporting data.
EVM export mode is configured with --export.evm.mode.
`,
			},
			{
				Name:      "evm-keys",
				Usage:     "Export EVM node keys",
				ArgsUsage: "<directory>",
				Action:    utils.MigrateFlags(exportEvmKeys),
				Flags: []cli.Flag{
					DataDirFlag,
				},
				Description: `
    opera export evm-keys

Requires a first argument of the DB directory to write to.
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
