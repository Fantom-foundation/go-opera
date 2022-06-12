package launcher

import (
	"fmt"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
)

var (
	experimentalFlag = cli.BoolFlag{
		Name:  "experimental",
		Usage: "Allow experimental DB fixing",
	}
	dbCommand = cli.Command{
		Name:        "db",
		Usage:       "A set of commands related to leveldb database",
		Category:    "DB COMMANDS",
		Description: "",
		Subcommands: []cli.Command{
			{
				Name:      "compact",
				Usage:     "Compact all databases",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(compact),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
				},
				Description: `
opera db compact
will compact all databases under datadir's chaindata.
`,
			},
			{
				Name:      "migrate",
				Usage:     "Migrate tables layout",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(dbMigrate),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
				},
				Description: `
opera db migrate
will migrate tables layout according to the configuration.
`,
			},
			{
				Name:      "fix",
				Usage:     "Experimental - try to fix dirty DB",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(fixDirty),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					experimentalFlag,
				},
				Description: `
opera db fix --experimental
Experimental - try to fix dirty DB.
`,
			},
		},
	}
)

func compact(ctx *cli.Context) error {

	cfg := makeAllConfigs(ctx)

	rawProducer := makeRawDbsProducer(cfg)
	for _, name := range rawProducer.Names() {
		db, err := rawProducer.OpenDB(name)
		defer db.Close()
		if err != nil {
			log.Error("Cannot open db or db does not exists", "db", name)
			return err
		}

		log.Info("Stats before compaction", "db", name)
		showLeveldbStats(db)

		log.Info("Triggering compaction", "db", name)
		for b := byte(0); b < 255; b++ {
			log.Trace("Compacting chain database", "db", name, "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
			if err := db.Compact([]byte{b}, []byte{b + 1}); err != nil {
				log.Error("Database compaction failed", "err", err)
				return err
			}
		}

		log.Info("Stats after compaction", "db", name)
		showLeveldbStats(db)
	}

	return nil
}

func showLeveldbStats(db ethdb.Stater) {
	if stats, err := db.Stat("leveldb.stats"); err != nil {
		log.Warn("Failed to read database stats", "error", err)
	} else {
		fmt.Println(stats)
	}
	if ioStats, err := db.Stat("leveldb.iostats"); err != nil {
		log.Warn("Failed to read database iostats", "error", err)
	} else {
		fmt.Println(ioStats)
	}
}
