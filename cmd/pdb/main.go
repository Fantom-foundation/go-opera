package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/topicsdb"
)

// go run ./cmd/pdb 2>&1 | tee 111.log

func init() {
	debug.SetMaxThreads(60) // got 'thread exhaustion' when consumes ~20Gb mem
}

func main() {
	datadir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	fmt.Printf("-- DataDir: %s\n", datadir)

	cacheRatio := cachescale.Identity
	cfg := integration.Pbl1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles()))
	cacher, err := integration.DbCacheFdlimit(cfg.RuntimeCache)
	if err != nil {
		panic(err)
	}
	dbs := pebble.NewProducer(datadir, cacher)

	index := topicsdb.New(dbs) // origin 'thread exhaustion'
	// index := topicsdb.NewWithThreadPool(dbs) // threads limit fix
	defer index.Close()

	// WATCHDOG
	var (
		blocks  uint64
		readers int64
		abort   = make(chan os.Signal, 1)
		aborted bool

		success, failed uint64
	)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(abort)
	go func() {
		for {
			select {
			case <-abort:
				fmt.Println("-- ABORT --")
				aborted = true
				return
			case <-time.After(3 * time.Second):
				fmt.Printf("\treader count: %d,\tblocks: %d\n", readers, blocks)
				fmt.Printf("\tsuccess: %d,\tfailed: %d\n", success, failed)
			}
		}
	}()

	const V = 500
	// DB WRITE
	var wgWrite sync.WaitGroup
	wgWrite.Add(1)
	go func() {
		defer wgWrite.Done()
		for i := 0; !aborted; i++ {
			time.Sleep(1 * time.Microsecond)
			var bn uint64
			if (i % V) == 0 {
				bn = atomic.AddUint64(&blocks, 1)
			} else {
				bn = atomic.LoadUint64(&blocks)
			}
			addr, topics := FakeLog(i % V)
			index.Push(
				&types.Log{
					BlockNumber: bn,
					Address:     addr,
					Topics:      topics,
				},
			)
		}
	}()

	// DB READ
	var pattern = make([][]common.Hash, 4)
	for i := range pattern {
		pattern[i] = make([]common.Hash, V)
	}
	for i := 0; i < V; i++ {
		addr, topics := FakeLog(i)
		pattern[0][i] = addr.Hash()
		pattern[1][i] = topics[0]
		pattern[2][i] = topics[1]
		pattern[3][i] = topics[2]
	}

	var wgRead sync.WaitGroup
	for i := 0; !aborted; i++ {
		time.Sleep(1 * time.Microsecond)
		wgRead.Add(1)
		go func(i int) {
			defer wgRead.Done()

			atomic.AddInt64(&readers, 1)
			defer atomic.AddInt64(&readers, -1)

			bnFrom := idx.Block(0)
			bnTo := idx.Block(atomic.AddUint64(&blocks, 1))
			if bnTo > 10 {
				bnFrom = bnTo - 10
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := index.FindInBlocks(ctx, bnFrom, bnTo, pattern)
			if err != nil {
				failed++
			} else {
				success++
			}

		}(i % V)
	}

	wgWrite.Wait()
	wgRead.Wait()
}

func FakeLog(n int) (addr common.Address, topics []common.Hash) {
	addr = FakeAddr(n)
	topics = []common.Hash{
		FakeHash(n + 1),
		FakeHash(n + 2),
		FakeHash(n + 3),
	}
	return
}

func FakeAddr(n int) (a common.Address) {
	reader := rand.New(rand.NewSource(int64(n)))
	reader.Read(a[:])

	return
}

// FakeKey gets n-th fake private key.
func FakeHash(n int) (h common.Hash) {
	reader := rand.New(rand.NewSource(int64(n)))
	reader.Read(h[:])

	return
}
