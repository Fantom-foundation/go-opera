package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/integration"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/utils/errlock"
	"github.com/Fantom-foundation/go-lachesis/version"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"gopkg.in/urfave/cli.v1"
	"io"
	"os"
	"strings"
	"time"
)

func exportChain(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	cfg := makeAllConfigs(ctx)

	if !cfg.Lachesis.NoCheckVersion {
		ver := params.VersionWithCommit(gitCommit, gitDate)
		status, msg, err := version.CheckRelease(nil, ver)
		applyVersionCheck(&cfg, status, msg, err)
	}

	// check errlock file
	errlock.SetDefaultDatadir(cfg.Node.DataDir)
	errlock.Check()

	stack := makeConfigNode(ctx, &cfg.Node)
	defer stack.Close()

	_, gdb := makeGDB(cfg.Node.DataDir, &cfg.Lachesis)
	defer gdb.Close()

	start := time.Now()

	var err error
	fp := ctx.Args().First()

	err = ExportChain(gdb, fp)
	if err != nil {
		utils.Fatalf("Export error: %v\n", err)
	}
	fmt.Printf("Export done in %v\n", time.Since(start))
	return nil
}

func makeGDB(dataDir string, gossipCfg *gossip.Config) (kvdb.KeyValueStore, *gossip.Store) {
	dbs := flushable.NewSyncedPool(integration.DBProducer(dataDir))

	gdb := gossip.NewStore(dbs, gossipCfg.StoreConfig)
	gdb.SetName("gossip-db")
	return dbs.GetDb("gossip-main"), gdb
}

// ExportChain exports a blockchain into the specified file, truncating any data
// already present in the file.
func ExportChain(gdb *gossip.Store, fn string) error {
	log.Info("Exporting events", "file", fn)

	// Open the file handle and potentially wrap with a gzip stream
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	var writer io.Writer = fh
	if strings.HasSuffix(fn, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}
	// Iterate over the blocks and export them
	if err := Export(gdb, writer); err != nil {
		return err
	}
	log.Info("Exported blockchain", "file", fn)

	return nil
}

// Export writes the active chain to the given writer.
func Export(gdb *gossip.Store, w io.Writer) error {
	log.Info("Exporting batch of events")
	var err error
	start, reported := time.Now(), time.Now()

	gdb.ForEachEventWithoutEpoch(func(event *inter.Event) bool {
		log.Debug("export--->", "event", event.String())
		if event == nil {
			err = errors.New("export failed, event not found")
			return false
		}
		err := event.EncodeRLP(w)
		if err != nil {
			err = fmt.Errorf("export failed, error: %v", err)
			return false
		}
		if time.Since(reported) >= statsReportLimit {
			log.Info("Exporting events", "exported", event.String(), "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
		return true
	})

	return err
}
