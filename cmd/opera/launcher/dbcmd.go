package launcher

import (
	"path"

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	dbCommand = cli.Command{
		Name:        "db",
		Usage:       "A set of commands related to leveldb database",
		Category:    "DB COMMANDS",
		Description: "",
		Subcommands: []cli.Command{
			{
				Name:      "compact",
				Usage:     "Compact whole leveldb databases under chaindata",
				ArgsUsage: "<root>",
				Action:    utils.MigrateFlags(compact),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
				},
				Description: `
opera db compact
will compact entire data store under datadir's chaindata.
`,
			},
		},
	}
)

func compact(ctx *cli.Context) error {

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cfg.cachescale)
	for _, name := range rawProducer.Names() {
		db, err := rawProducer.OpenDB(name)
		if err != nil {
			return err
		}
		if err := db.Compact(nil, nil); err != nil {
			return err
		}
	}

	return nil
}
