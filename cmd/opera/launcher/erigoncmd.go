package launcher

import (
	"context"
	"fmt"
	"path"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/erigon"
	"github.com/Fantom-foundation/go-opera/integration"
)

func writeEVMToErigon(ctx *cli.Context) error {

	start := time.Now()
	log.Info("Writing of EVM accounts into Erigon database started")
	// initiate erigon lmdb
	//db, tmpDir, err := erigon.SetupDB()
	db, tmpDir, err := erigon.SetupDB()
	if err != nil {
		return err
	}
	defer db.Close()

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))

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
	accountLimitFlag := ctx.Int(erigonAccountLimitFlag.Name)
	mptFlag := ctx.String(mptTraversalMode.Name)

	log.Info("Generate Erigon Plain State...")
	if err := erigon.GeneratePlainState(mptFlag, accountLimitFlag, root, chaindb, db, lastBlockIdx); err != nil {
		return err
	}
	log.Info("Generation of Erigon Plain State is complete")

	log.Info("Generate Erigon Hash State")
	if err := erigon.GenerateHashedState("HashedState", db, tmpDir, context.Background()); err != nil {
		log.Error("GenerateHashedState error: ", err)
		return err
	}
	log.Info("Generation Hash State is complete")

	log.Info("Generate Intermediate Hashes state and compute State Root")
	trieCfg := erigon.StageTrieCfg(db, true, true, "")
	hash, err := erigon.ComputeStateRoot("Intermediate Hashes", db, trieCfg)
	if err != nil {
		log.Error("GenerateIntermediateHashes error: ", err)
		return err
	}
	log.Info(fmt.Sprintf("[%s] Trie root", "GenerateStateRoot"), "hash", hash.Hex())
	log.Info("Generation of Intermediate Hashes state and computation of State Root Complete")

	log.Info("Writing of EVM accounts into Erigon database completed", "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}
