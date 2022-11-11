package launcher

import (
	"context"
	//"fmt"
	"fmt"
	"path"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/erigon"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/logger"

	"github.com/ledgerwatch/erigon-lib/kv"
)

func readErigon(_ *cli.Context) error {

	db := erigon.MakeChainDatabase(logger.New("mdbx"))
	defer db.Close()

	tx, err := db.BeginRo(context.Background())
	if err != nil {
		return fmt.Errorf("unable to begin transaction, err: %q", err)
	}
	defer tx.Rollback()

	if err := erigon.ReadErigonTableNoDups(kv.PlainState, tx); err != nil {
		return fmt.Errorf("unable to read from Erigon table, err: %q", err)
	}

	// TODO handle flags
	return nil
}

func writeErigon(ctx *cli.Context) error {

	start := time.Now()
	log.Info("Writing of EVM accounts into Erigon database started")

	// initiate erigon database
	db := erigon.MakeChainDatabase(logger.New("mdbx"))
	defer db.Close()

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(erigon.DefaultDataDir(), "chaindata"), cacheScaler(ctx))

	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	if gdb.GetHighestLamport() != 0 {
		log.Warn("Attempting genesis export not in a beginning of an epoch. Genesis file output may contain excessive data.")
	}
	defer gdb.Close()

	log.Info("Getting EvmDb")
	chaindb := gdb.EvmStore().EvmDb

	log.Info("Getting FinalizedStateRoot")
	root := common.Hash(gdb.GetBlockState().FinalizedStateRoot)

	log.Info("Getting LastBlock")
	lastBlockIdx := gdb.GetBlockState().LastBlock.Idx

	log.Info("Generate Erigon Plain State...")
	if err := erigon.GeneratePlainState(root, chaindb, db, lastBlockIdx); err != nil {
		return err
	}
	log.Info("Generation of Erigon Plain State is complete", "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}
