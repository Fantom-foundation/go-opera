package launcher

import (
	"fmt"
	"strings"
	"time"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
)

var (
	fixDirtyCommand = cli.Command{
		Action:      utils.MigrateFlags(fixDirty),
		Name:        "fixdirty",
		Usage:       "Experimental - try to fix dirty DB",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `Experimental - try to fix dirty DB.`,
	}
)

// maxEpochsToTry represents amount of last closed epochs to try (in case that the last one has the state unavailable)
const maxEpochsToTry = 10000

// fixDirty is the fixdirty command.
func fixDirty(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)

	log.Info("Opening databases")
	producer := makeRawDbsProducer(cfg)

	// reverts the gossip database state
	epochState, err := fixDirtyGossipDb(producer, cfg)
	if err != nil {
		return err
	}

	// drop epoch databases
	log.Info("Removing epoch DBs - will be recreated on next start")
	err = dropAllEpochDbs(producer)
	if err != nil {
		return err
	}

	// drop consensus database
	log.Info("Removing lachesis db")
	cMainDb := mustOpenDB(producer, "lachesis")
	_ = cMainDb.Close()
	cMainDb.Drop()

	// prepare consensus database from epochState
	log.Info("Recreating lachesis db")
	cMainDb = mustOpenDB(producer, "lachesis")
	cGetEpochDB := func(epoch idx.Epoch) kvdb.Store {
		return mustOpenDB(producer, fmt.Sprintf("lachesis-%d", epoch))
	}
	cdb := abft.NewStore(cMainDb, cGetEpochDB, panics("Lachesis store"), cfg.LachesisStore)
	err = cdb.ApplyGenesis(&abft.Genesis{
		Epoch:      epochState.Epoch,
		Validators: epochState.Validators,
	})
	if err != nil {
		return fmt.Errorf("failed to init consensus database: %v", err)
	}
	_ = cdb.Close()

	log.Info("Clearing dbs dirty flags")
	err = clearDirtyFlags(producer)
	if err != nil {
		return err
	}

	log.Info("Fixing done")
	return nil
}

// fixDirtyGossipDb reverts the gossip database into state, when was one of last epochs sealed
func fixDirtyGossipDb(producer kvdb.IterableDBProducer, cfg *config) (
	epochState *iblockproc.EpochState, err error) {
	gdb, err := makeRawGossipStore(producer, cfg) // requires FlushIDKey present (not clean) in all dbs
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	// find the last closed epoch with the state available
	epochIdx, blockState, epochState := getLastEpochWithState(gdb, maxEpochsToTry)
	if blockState == nil || epochState == nil {
		return nil, fmt.Errorf("state for last %d closed epochs is not available", maxEpochsToTry)
	}

	// set the historic state to be the current
	log.Info("Setting block epoch state", "epoch", epochIdx)
	gdb.SetBlockEpochState(*blockState, *epochState)
	gdb.FlushBlockEpochState()

	// Service.switchEpochTo
	gdb.SetHighestLamport(0)
	gdb.FlushHighestLamport()

	// removing excessive events (event epoch >= closed epoch)
	log.Info("Removing excessive events")
	gdb.ForEachEventRLP(epochIdx.Bytes(), func(id hash.Event, _ rlp.RawValue) bool {
		gdb.DelEvent(id)
		return true
	})

	return epochState, nil
}

// getLastEpochWithState finds the last closed epoch with the state available
func getLastEpochWithState(gdb *gossip.Store, epochsToTry idx.Epoch) (epochIdx idx.Epoch, blockState *iblockproc.BlockState, epochState *iblockproc.EpochState) {
	currentEpoch := gdb.GetEpoch()
	endEpoch := idx.Epoch(1)
	if currentEpoch > epochsToTry {
		endEpoch = currentEpoch - epochsToTry
	}

	for epochIdx = currentEpoch; epochIdx > endEpoch; epochIdx-- {
		blockState, epochState = gdb.GetHistoryBlockEpochState(epochIdx)
		if blockState == nil || epochState == nil {
			log.Info("Last closed epoch is not available", "epoch", epochIdx)
			continue
		}
		if !gdb.EvmStore().HasStateDB(blockState.FinalizedStateRoot) {
			log.Info("State for the last closed epoch is not available", "epoch", epochIdx)
			continue
		}
		log.Info("Last closed epoch with available state found", "epoch", epochIdx)
		return epochIdx, blockState, epochState
	}

	return 0, nil, nil
}

func dropAllEpochDbs(producer kvdb.IterableDBProducer) error {
	for _, name := range producer.Names() {
		if strings.HasPrefix(name, "gossip-") || strings.HasPrefix(name, "lachesis-") {
			log.Info("Removing db", "name", name)
			db, err := producer.OpenDB(name)
			if err != nil {
				return fmt.Errorf("unable to open db %s; %s", name, err)
			}
			_ = db.Close()
			db.Drop()
		}
	}
	return nil
}

// clearDirtyFlags - writes the CleanPrefix into all databases
func clearDirtyFlags(rawProducer kvdb.IterableDBProducer) error {
	id := bigendian.Uint64ToBytes(uint64(time.Now().UnixNano()))
	names := rawProducer.Names()
	for _, name := range names {
		db, err := rawProducer.OpenDB(name)
		if err != nil {
			return err
		}

		err = db.Put(integration.FlushIDKey, append([]byte{flushable.CleanPrefix}, id...))
		if err != nil {
			log.Crit("Failed to write CleanPrefix", "name", name)
			return err
		}
		log.Info("Database set clean", "name", name)
		_ = db.Close()
	}
	return nil
}

func mustOpenDB(producer kvdb.DBProducer, name string) kvdb.Store {
	db, err := producer.OpenDB(name)
	if err != nil {
		utils.Fatalf("Failed to open '%s' database: %v", name, err)
	}
	return db
}

func panics(name string) func(error) {
	return func(err error) {
		log.Crit(fmt.Sprintf("%s error", name), "err", err)
	}
}
