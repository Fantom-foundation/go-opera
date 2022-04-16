package launcher

import (
	"fmt"
	"path"

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
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
			log.Error("Cannot open db or db does not exists", "db", name)
			return err
		}
		for b := byte(0); b < 255; b++ {
			log.Info("Compacting chain database", "db", name, "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
			if err := db.Compact([]byte{b}, []byte{b + 1}); err != nil {
				log.Error("Database compaction failed", "err", err)
				return err
			}
		}
	}

	return nil
}
