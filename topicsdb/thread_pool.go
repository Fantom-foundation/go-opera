package topicsdb

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type pool struct {
	mu    sync.Mutex
	sum   int
	queue []int
}

func (p *pool) Lock(want int) (got int, release func()) {
	if want < 1 {
		want = 0
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	got = min(p.sum, want)
	p.sum -= got
	release = func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.sum += got
	}

	return
}

var globalPool = &pool{
	sum: getMaxThreads(),
}

func getMaxThreads() int {
	was := debug.SetMaxThreads(10000)
	debug.SetMaxThreads(was)
	return was
}

type WithThreadPool struct {
	*Index
}

// FindInBlocks returns all log records of block range by pattern. 1st pattern element is an address.
func (tt *WithThreadPool) FindInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash) (logs []*types.Log, err error) {
	err = tt.ForEachInBlocks(
		ctx,
		from, to,
		pattern,
		func(l *types.Log) bool {
			logs = append(logs, l)
			return true
		})

	return
}

// ForEach matches log records by pattern. 1st pattern element is an address.
func (tt *WithThreadPool) ForEach(ctx context.Context, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	return tt.ForEachInBlocks(ctx, 0, 0, pattern, onLog)
}

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *WithThreadPool) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if 0 < to && to < from {
		return nil
	}

	pattern, err := limitPattern(pattern)
	if err != nil {
		return err
	}

	onMatched := func(rec *logrec) (gonext bool, err error) {
		rec.fetch(tt.table.Logrec)
		if rec.err != nil {
			err = rec.err
			return
		}
		gonext = onLog(rec.result)
		return
	}

	splitby := 0
	threads := 0
	for i := range pattern {
		threads += len(pattern[i])
		if len(pattern[splitby]) < len(pattern[i]) {
			splitby = i
		}
	}
	rest := pattern[splitby]
	threads -= len(rest)

	for len(rest) > 0 {
		got, release := globalPool.Lock(threads + len(rest))
		if got <= threads {
			release()
			time.Sleep(time.Microsecond)
			continue
		}

		pattern[splitby] = rest[:got-threads]
		rest = rest[got-threads:]
		err = tt.searchParallel(ctx, pattern, uint64(from), uint64(to), onMatched)
		release()
		if err != nil {
			return err
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
