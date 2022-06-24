package topicsdb

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

func (tt *Index) searchParallel(ctx context.Context, pattern [][]common.Hash, blockStart, blockEnd uint64, onMatched logHandler) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var (
		threads   sync.WaitGroup
		positions = make([]int, 0, len(pattern))
		result    = make(map[ID]*logrec)
	)

	var mu sync.Mutex
	aggregator := func(pos, num int) logHandler {
		return func(rec *logrec) (gonext bool, err error) {
			if rec == nil {
				threads.Done()
				return
			}

			err = ctx.Err()
			if err != nil {
				return
			}

			if blockEnd > 0 && rec.ID.BlockNumber() > blockEnd {
				return
			}

			gonext = true

			if rec.topicsCount < uint8(len(pattern)-1) {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if prev, ok := result[rec.ID]; ok {
				prev.matched++
			} else {
				rec.matched++
				result[rec.ID] = rec
			}

			return
		}
	}

	// start the threads
	var preparing sync.WaitGroup
	preparing.Add(1)
	for pos := range pattern {
		if len(pattern[pos]) == 0 {
			continue
		}
		positions = append(positions, pos)
		for i, variant := range pattern[pos] {
			threads.Add(1)
			go func(pos, i int, variant common.Hash) {
				onMatched := aggregator(pos, i)
				preparing.Wait()
				tt.scanPatternVariant(uint8(pos), variant, blockStart, onMatched)
			}(pos, i, variant)
		}
	}
	preparing.Done()

	threads.Wait()
	var gonext bool
	for _, rec := range result {
		if rec.matched != len(positions) {
			continue
		}
		gonext, err = onMatched(rec)
		if !gonext {
			break
		}
	}

	return
}

func (tt *Index) scanPatternVariant(pos uint8, variant common.Hash, start uint64, onMatched logHandler) {
	prefix := append(variant.Bytes(), posToBytes(pos)...)

	it := tt.table.Topic.NewIterator(prefix, uintToBytes(start))
	defer it.Release()
	for it.Next() {
		id := extractLogrecID(it.Key())
		topicCount := bytesToPos(it.Value())
		rec := newLogrec(id, topicCount)

		gonext, _ := onMatched(rec)
		if !gonext {
			break
		}
	}
	onMatched(nil)
}
