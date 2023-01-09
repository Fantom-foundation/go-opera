package topicsdb

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type WithThreadPool struct {
	*Index
}

func getMaxThreads() int {
	was := debug.SetMaxThreads(10000)
	debug.SetMaxThreads(was)
	return was
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

	fmt.Printf("getMaxThreads() == %d\n", getMaxThreads())

	onMatched := func(rec *logrec) (gonext bool, err error) {
		rec.fetch(tt.table.Logrec)
		if rec.err != nil {
			err = rec.err
			return
		}
		gonext = onLog(rec.result)
		return
	}

	return tt.searchParallel(ctx, pattern, uint64(from), uint64(to), onMatched)
}
