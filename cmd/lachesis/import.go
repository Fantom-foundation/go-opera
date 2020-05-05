package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func importChain(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}
	_, stack, engine, db, gdb := makeFNode(ctx, true)
	utils.StartNode(stack)

	// Start periodically gathering memory profiles
	var peakMemAlloc, peakMemSys uint64
	go func() {
		stats := new(runtime.MemStats)
		for {
			runtime.ReadMemStats(stats)
			if atomic.LoadUint64(&peakMemAlloc) < stats.Alloc {
				atomic.StoreUint64(&peakMemAlloc, stats.Alloc)
			}
			if atomic.LoadUint64(&peakMemSys) < stats.Sys {
				atomic.StoreUint64(&peakMemSys, stats.Sys)
			}
			time.Sleep(5 * time.Second)
		}
	}()
	// Import the chain
	start := time.Now()

	if len(ctx.Args()) == 1 {
		if err := ImportChain(engine, gdb, ctx.Args().First()); err != nil {
			log.Error("Import error", "err", err)
		}
	} else {
		for _, arg := range ctx.Args() {
			if err := ImportChain(engine, gdb, arg); err != nil {
				log.Error("Import error", "file", arg, "err", err)
			}
		}
	}

	fmt.Printf("Import done in %v.\n\n", time.Since(start))

	// Output pre-compaction stats mostly to see the import trashing
	stats, err := db.Stat("leveldb.stats")
	if err != nil {
		utils.Fatalf("Failed to read database stats: %v", err)
	}
	fmt.Println(stats)

	ioStats, err := db.Stat("leveldb.iostats")
	if err != nil {
		utils.Fatalf("Failed to read database iostats: %v", err)
	}
	fmt.Println(ioStats)

	// Print the memory statistics used by the importing
	mem := new(runtime.MemStats)
	runtime.ReadMemStats(mem)

	fmt.Printf("Object memory: %.3f MB current, %.3f MB peak\n", float64(mem.Alloc)/1024/1024, float64(atomic.LoadUint64(&peakMemAlloc))/1024/1024)
	fmt.Printf("System memory: %.3f MB current, %.3f MB peak\n", float64(mem.Sys)/1024/1024, float64(atomic.LoadUint64(&peakMemSys))/1024/1024)
	fmt.Printf("Allocations:   %.3f million\n", float64(mem.Mallocs)/1000000)
	fmt.Printf("GC pause:      %v\n\n", time.Duration(mem.PauseTotalNs))

	if ctx.GlobalBool(utils.NoCompactionFlag.Name) {
		stack.Stop()
		return nil
	}

	// Compact the entire database to more accurately measure disk io and print the stats
	start = time.Now()
	fmt.Println("Compacting entire database...")
	if err = db.Compact(nil, nil); err != nil {
		utils.Fatalf("Compaction failed: %v", err)
	}
	fmt.Printf("Compaction done in %v.\n\n", time.Since(start))

	stats, err = db.Stat("leveldb.stats")
	if err != nil {
		utils.Fatalf("Failed to read database stats: %v", err)
	}
	fmt.Println(stats)

	ioStats, err = db.Stat("leveldb.iostats")
	if err != nil {
		utils.Fatalf("Failed to read database iostats: %v", err)
	}
	fmt.Println(ioStats)

	var (
		networkMsgFlag string
		hint           = " flag when starting a node"
		fakenet        = ctx.GlobalString(FakeNetFlag.Name)
		testnet        = ctx.GlobalString(utils.TestnetFlag.Name)
	)
	if fakenet != "" {
		networkMsgFlag = "--" + FakeNetFlag.Name + " " + fakenet
		hint = "Use " + networkMsgFlag + hint
	} else if testnet != "" {
		networkMsgFlag = "--" + utils.TestnetFlag.Name
		hint = "Use " + networkMsgFlag + hint
	} else {
		networkMsgFlag = "mainnet"
		hint = "Mainnet installed by default" + hint
	}

	log.Warn("Import was made for "+networkMsgFlag+" network, imported events apply only to this network", "hint", hint)
	stack.Stop()
	return nil
}

func ImportChain(engine gossip.Consensus, gdb *gossip.Store, fn string) error {
	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop at the next batch.
	interrupt := make(chan os.Signal, 1)
	stop := make(chan struct{})
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)
	defer close(interrupt)
	go func() {
		if _, ok := <-interrupt; ok {
			log.Info("Interrupted during import, stopping at next batch")
		}
		close(stop)
	}()
	checkInterrupt := func() bool {
		select {
		case <-stop:
			return true
		default:
			return false
		}
	}

	log.Info("Importing events", "file", fn)

	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close()

	var reader io.Reader = fh
	if strings.HasSuffix(fn, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return err
		}
	}
	stream := rlp.NewStream(reader, 0)
	// Run actual the import.
	for batch := 0; ; batch++ {
		// Load a batch of RLP events.
		if checkInterrupt() {
			return fmt.Errorf("interrupted")
		}
		eventsBatch := make(inter.Events, 0, importBatchSize)
		i := 0
		for ; i < importBatchSize; i++ {
			var e inter.Event
			if err = stream.Decode(&e); err == io.EOF {
				break
			} else if err != nil {
				return fmt.Errorf("at event decode error: %v", err)
			}
			eventsBatch = append(eventsBatch, &e)
		}
		if len(eventsBatch) == 0 {
			break
		}
		// Import the batch.
		if checkInterrupt() {
			return fmt.Errorf("interrupted")
		}

		missing := missingEvents(gdb, eventsBatch)
		if len(missing) == 0 {
			log.Info(
				"Skipping batch as all events present", "batch", batch,
				"first", eventsBatch[0].Hash(), "last", eventsBatch[len(eventsBatch)-1].Hash())
			continue
		}
		if err := insertEvents(engine, gdb, missing); err != nil {
			return fmt.Errorf("insert events error: %v", err)
		}
		log.Debug(
			"events batch inserted", "batch", batch,
			"first", eventsBatch[0].Hash(), "last", eventsBatch[len(eventsBatch)-1].Hash())
	}
	return nil
}

func missingEvents(gdb *gossip.Store, events []*inter.Event) []*inter.Event {
	var missingEvents []*inter.Event
	for _, event := range events {
		if gdb.HasEvent(event.Hash()) {
			log.Debug("event exist", "event", event)
			continue
		}
		missingEvents = append(missingEvents, event)
	}
	return missingEvents
}

func insertEvents(engine gossip.Consensus, gdb *gossip.Store, events []*inter.Event) error {
	if len(events) == 0 {
		return nil
	}
	var (
		current     int
		newEpoch    idx.Epoch
		sealedEpoch idx.Epoch
	)
	for _, event := range events {
		if newEpoch < event.Epoch {
			sealedEpoch = newEpoch
			newEpoch = event.Epoch
		}
		if current == importFlushBatch {
			current = 0
			err := gdb.Commit(nil, true)
			if err != nil {
				return err
			}
		}
		current++
		if gdb.HasEventHeader(event.Hash()) {
			continue
		}

		ev := engine.Prepare(event)
		if ev == nil {
			continue
		}
		gdb.SetEvent(ev)
		err := engine.ProcessEvent(ev)
		if err != nil {
			gdb.DeleteEvent(ev.Epoch, ev.Hash())
			return err
		}
		gdb.EpochDbs.Del(uint64(sealedEpoch))
	}
	err := gdb.Commit(nil, true)
	if err != nil {
		return err
	}
	return nil
}
