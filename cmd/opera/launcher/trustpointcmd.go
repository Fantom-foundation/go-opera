package launcher

import (
	"fmt"
	"os"
	"path"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/opera/trustpoint"
)

var (
	trustpointCommand = cli.Command{
		Name:        "trustpoint",
		Usage:       "A set of commands based on the trustpoint",
		Category:    "MISCELLANEOUS COMMANDS",
		Description: "",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Prune stale EVM state data and save trustpoint into file",
				ArgsUsage: "<filename>",
				Action:    utils.MigrateFlags(trustpointCreate),
				Category:  "MISCELLANEOUS COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.CacheTrieJournalFlag,
					utils.BloomFilterSizeFlag,
				},
				Description: `
opera trustpoint create

Note: command also prunes EVM state data.
<TODO>
`,
			},
			{
				Name:      "apply",
				Usage:     "Initialize datadir from trustpoint file",
				ArgsUsage: "<filename>",
				Action:    utils.MigrateFlags(trustpointApply),
				Category:  "MISCELLANEOUS COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.CacheTrieJournalFlag,
				},
				Description: `
opera trustpoint apply

Note: datadir shoul be empty.
<TODO>
`,
			},
		},
	}
)

func trustpointCreate(ctx *cli.Context) error {
	if ctx.NArg() > 1 {
		log.Error("Too many arguments given")
		return errors.New("too many arguments")
	}
	if ctx.NArg() < 1 {
		log.Error("File name argument required")
		return errors.New("file name argument required")
	}
	file := ctx.Args()[0]

	cfg := makeAllConfigs(ctx)
	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	gdb, _, err := makeRawStores(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	if false { // TODO: enable later
		es := gdb.GetEpochState()
		_, err = gdb.EvmStore().StateDB(es.EpochStateRoot)
		if err != nil {
			log.Error("State not found, probably pruned before", "err", err, "root", es.EpochStateRoot)
			return err
		}
		// prune state db
		bloomFilterSize := ctx.GlobalUint64(utils.BloomFilterSizeFlag.Name)
		// TODO: how about last 256 roots?
		err = pruneStateTo(gdb, common.Hash(es.EpochStateRoot), common.Hash{}, bloomFilterSize)
		if err != nil {
			return err
		}
	}

	db, err := rawProducer.OpenDB("trustpoint")
	if err != nil {
		return err
	}
	defer db.Drop()

	store := trustpoint.NewStore(db)
	defer store.Close()

	err = gdb.SaveTrustpoint(store)
	if err != nil {
		return err
	}

	return trustpointSaveTo(store, file)
}

func trustpointApply(ctx *cli.Context) error {
	if ctx.NArg() > 1 {
		log.Error("Too many arguments given")
		return errors.New("too many arguments")
	}
	if ctx.NArg() < 1 {
		log.Error("File name argument required")
		return errors.New("file name argument required")
	}
	file := ctx.Args()[0]

	cfg := makeAllConfigs(ctx)

	gdbDatadir := path.Join(cfg.Node.DataDir, "chaindata")
	err := os.MkdirAll(gdbDatadir, 0750)
	if err != nil {
		return err
	}

	rawProducer := integration.DBProducer(gdbDatadir, cacheScaler(ctx))
	if len(rawProducer.Names()) > 0 {
		return fmt.Errorf("datadir is not empty")
	}

	// apply genesis
	genesis := getOperaGenesis(ctx)
	err = integration.ApplyGenesis(rawProducer, gossip.DefaultBlockProc(), genesis, cfg.AppConfigs())
	if err != nil {
		return err
	}
	_ = genesis.Close()

	// drop abft genesis
	// TODO: modify the cdb.ApplyGenesis() to ignore existing genesis instead of drop
	dbs1 := integration.NewDummyFlushableProducer(rawProducer)
	abftDb, _ := dbs1.OpenDB("lachesis")
	abftDb.Close()
	abftDb.Drop()

	// apply trustpoint
	dbs := integration.NewDummyFlushableProducer(rawProducer)
	gdb, cdb := integration.MakeStores(dbs, cfg.AppConfigs())
	defer gdb.Close()

	db, err := dbs.OpenDB("trustpoint")
	if err != nil {
		return err
	}
	defer db.Drop()

	store := trustpoint.NewStore(db)
	defer store.Close()

	err = trustpointReadFrom(file, store)
	if err != nil {
		log.Error("Trustpoint file read", "err", err)
		return err
	}

	err = gdb.ApplyTrustpoint(store)
	if err != nil {
		return err
	}

	err = cdb.ApplyGenesis(&abft.Genesis{
		Epoch:      gdb.GetEpoch(),
		Validators: gdb.GetValidators(),
	})
	if err != nil {
		return err
	}

	return gdb.Commit()
}

func trustpointSaveTo(store *trustpoint.Store, file string) error {
	log.Info("Encoding go-opera trustpoint", "path", file)
	fh, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fh.Close()

	err = trustpoint.WriteStore(store, fh)
	if err != nil {
		return err
	}

	return nil
}

func trustpointReadFrom(file string, store *trustpoint.Store) error {
	log.Info("Decoding go-opera trustpoint", "path", file)
	fh, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fh.Close()

	return trustpoint.ReadStore(fh, store)
}
