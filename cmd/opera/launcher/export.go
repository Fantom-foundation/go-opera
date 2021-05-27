package launcher

import (
	"context"
	"errors"
	"path"
	"strconv"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/ethapi"
	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
)

func exportEvents(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}
	fileName := ctx.Args().First()

	from := rpc.BlockNumber(1)
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 32)
		if err != nil {
			return err
		}
		from = rpc.BlockNumber(n)
	}

	to := rpc.BlockNumber(0)
	if len(ctx.Args()) > 2 {
		n, err := strconv.ParseUint(ctx.Args().Get(2), 10, 32)
		if err != nil {
			return err
		}
		to = rpc.BlockNumber(n)
	}

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	backend := gossip.NewEthAPIBackendOverStore(gdb)
	api := ethapi.NewPrivateDebugAPI(backend)

	return api.ExportEvents(context.Background(), from, to, fileName)
}

func checkStateInitialized(rawProducer kvdb.IterableDBProducer) error {
	names := rawProducer.Names()
	if len(names) == 0 {
		return errors.New("datadir is not initialized")
	}
	// if flushID is not written, then previous genesis processing attempt was interrupted
	for _, name := range names {
		db, err := rawProducer.OpenDB(name)
		if err != nil {
			return err
		}
		flushID, _ := db.Get(integration.FlushIDKey)
		_ = db.Close()
		if flushID != nil {
			return nil
		}
	}
	return errors.New("datadir is not initialized")
}

func makeRawGossipStore(rawProducer kvdb.IterableDBProducer, cfg *config) (*gossip.Store, error) {
	if err := checkStateInitialized(rawProducer); err != nil {
		return nil, err
	}
	dbs := &integration.DummyFlushableProducer{rawProducer}
	gdb := gossip.NewStore(dbs, cfg.OperaStore)
	gdb.SetName("gossip-db")
	return gdb, nil
}
