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

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/topicsdb"
)

func init() {
	debug.SetMaxThreads(10000)
}

func main() {
	datadir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}

	cacheRatio := cachescale.Identity
	cfg := integration.Pbl1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles()))
	cacher, err := integration.DbCacheFdlimit(cfg.RuntimeCache)
	if err != nil {
		panic(err)
	}
	dbs := pebble.NewProducer(datadir, cacher)

	// index := topicsdb.New(dbs)
	index := topicsdb.NewWithThreadPool(dbs)
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

	// cd work/fantom/go-opera
	// go run ./cmd/pdb 2>&1 | tee 111.log

	const V = 5000
	// DB WRITE

	var wgWrite sync.WaitGroup
	wgWrite.Add(1)
	go func() {
		defer wgWrite.Done()
		for i := 0; !aborted; i++ {
			addr := FakeAddr(i % V)
			topics := []common.Hash{
				FakeHash(i%V + 1),
				FakeHash(i%V + 2),
				FakeHash(i%V + 3),
			}
			var bn uint64
			if (i % V) == 0 {
				bn = atomic.AddUint64(&blocks, 1)
			} else {
				bn = atomic.LoadUint64(&blocks)
			}
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
	var wgRead sync.WaitGroup
	for i := 0; !aborted; i++ {

		wgRead.Add(1)
		go func(i int) {
			defer wgRead.Done()

			atomic.AddInt64(&readers, 1)
			defer atomic.AddInt64(&readers, -1)

			bnFrom := idx.Block(0)
			bnTo := idx.Block(atomic.AddUint64(&blocks, 1))
			if bnTo > V {
				bnFrom = bnTo - V
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			logs, err := index.FindInBlocks(ctx, bnFrom, bnTo, [][]common.Hash{
				[]common.Hash{FakeAddr(i % V).Hash()},
				[]common.Hash{FakeHash(i%V + 1)},
				[]common.Hash{FakeHash(i%V + 2)},
			})
			if err != nil {
				// panic(err)
				failed++
			}
			success++
			if bnTo > V && len(logs) < 1 && false {
				panic(fmt.Errorf("%d found nothing at block %d - %d", i, bnFrom, bnTo))
			}

		}(i)
	}

	wgWrite.Wait()
	wgRead.Wait()
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
