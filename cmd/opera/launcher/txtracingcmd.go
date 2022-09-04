package launcher

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/inter"
)

type TracePayload struct {
	Key    common.Hash
	Traces []byte
}

// importTxTraces imports transaction traces from a specified file
func importTxTraces(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb, err := makeRawGossipStoreTrace(rawDbs, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	fn := ctx.Args().First()

	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close()

	var (
		reader  io.Reader = fh
		counter int
	)
	if strings.HasSuffix(fn, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return err
		}
		defer reader.(*gzip.Reader).Close()
	}

	log.Info("Importing transaction traces from file", "file", fn)
	start, reported := time.Now(), time.Now()

	stream := rlp.NewStream(reader, 0)
	for {
		select {
		case <-interrupt:
			return fmt.Errorf("interrupted")
		default:
		}
		e := new(TracePayload)
		err = stream.Decode(e)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		} else {
			gdb.TxTraceStore().SetTxTrace(e.Key, e.Traces)
			counter++
			if time.Since(reported) >= statsReportLimit {
				log.Info("Importing transaction traces", "imported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
				reported = time.Now()
			}
		}
	}
	log.Info("Imported transaction traces", "imported", counter, "elapsed", common.PrettyDuration(time.Since(start)))

	return nil
}

// deleteTxTraces removes transaction traces for specified block range
func deleteTxTraces(ctx *cli.Context) error {

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb, err := makeRawGossipStoreTrace(rawDbs, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	from := idx.Block(1)
	if len(ctx.Args()) > 0 {
		n, err := strconv.ParseUint(ctx.Args().Get(0), 10, 64)
		if err != nil {
			return err
		}
		from = idx.Block(n)
	}
	to := gdb.GetLatestBlockIndex()
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}
		to = idx.Block(n)
	}

	log.Info("Deleting transaction traces", "from block", from, "to block", to)

	err = deleteTraces(gdb, from, to)
	if err != nil {
		utils.Fatalf("Deleting traces error: %v\n", err)
	}

	return nil
}

// exportTxTraces exports transaction traces from specified block range
func exportTxTraces(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb, err := makeRawGossipStoreTrace(rawDbs, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	fn := ctx.Args().First()

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

	from := idx.Block(1)
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 64)
		if err != nil {
			return err
		}
		from = idx.Block(n)
	}
	to := gdb.GetLatestBlockIndex()
	if len(ctx.Args()) > 2 {
		n, err := strconv.ParseUint(ctx.Args().Get(2), 10, 64)
		if err != nil {
			return err
		}
		to = idx.Block(n)
	}

	log.Info("Exporting transaction traces to file", "file", fn)

	err = exportTraceTo(writer, gdb, from, to)
	if err != nil {
		utils.Fatalf("Export error: %v\n", err)
	}

	return nil
}

func makeRawGossipStoreTrace(producer kvdb.FlushableDBProducer, cfg *config) (*gossip.Store, error) {
	gdb := makeGossipStore(producer, cfg)

	if gdb.TxTraceStore() == nil {
		return nil, errors.New("transaction traces db store is not initialized")
	}

	return gdb, nil
}

// exportTraceTo writes the active chain
func exportTraceTo(w io.Writer, gdb *gossip.Store, from, to idx.Block) (err error) {

	if from == 1 && to == gdb.GetLatestBlockIndex() {
		exportAllTraceTo(w, gdb)
		return
	}
	start, reported := time.Now(), time.Now()

	var (
		counter int
		block   *inter.Block
	)

	for i := from; i <= to; i++ {
		block = gdb.GetBlock(i)
		for _, tx := range gdb.GetBlockTxs(i, block) {
			traces := gdb.TxTraceStore().GetTx(tx.Hash())
			if len(traces) > 0 {
				counter++
				rlp.Encode(w, TracePayload{tx.Hash(), traces})
			}
			if time.Since(reported) >= statsReportLimit {
				log.Info("Exporting transaction traces", "at block", i, "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
				reported = time.Now()
			}
		}
	}
	log.Info("Exported transaction traces", "from block", from, "to block", to, "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))

	return
}

// exportAllTraceTo writes all transaction traces of the active chain
func exportAllTraceTo(w io.Writer, gdb *gossip.Store) (err error) {
	start, reported := time.Now(), time.Now()
	var counter int

	gdb.TxTraceStore().ForEachTxtrace(func(key common.Hash, traces []byte) bool {
		counter++
		err = rlp.Encode(w, TracePayload{key, traces})
		if err != nil {
			return false
		}
		if time.Since(reported) >= statsReportLimit {
			log.Info("Exporting all transaction traces", "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
		return true
	})
	log.Info("Exported all transaction traces", "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))

	return
}

// deleteTraces removes transaction traces for specified block range
func deleteTraces(gdb *gossip.Store, from, to idx.Block) (err error) {
	start, reported := time.Now(), time.Now()

	var (
		counter int
	)

	for i := from; i <= to; i++ {
		for _, tx := range gdb.GetBlockTxs(i, gdb.GetBlock(i)) {
			ok, err := gdb.TxTraceStore().HasTxTrace(tx.Hash())
			if ok && err == nil {
				counter++
				gdb.TxTraceStore().RemoveTxTrace(tx.Hash())
				if time.Since(reported) >= statsReportLimit {
					log.Info("Deleting traces", "deleted", counter, "elapsed", common.PrettyDuration(time.Since(start)))
					reported = time.Now()
				}
			}
		}
	}
	log.Info("Deleting transaction traces done", "deleted", counter, "from block", from, "to block", to, "elapsed", common.PrettyDuration(time.Since(start)))
	return
}
