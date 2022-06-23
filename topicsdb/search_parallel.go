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
		flag              sync.Mutex
		wgStart, wgFinish sync.WaitGroup
		count             int
		result            = make(map[ID]*logrec)
	)

	wgStart.Add(1)
	for pos := range pattern {
		if len(pattern[pos]) > 0 {
			count++
		}
		for _, variant := range pattern[pos] {
			wgFinish.Add(1)
			go tt.scanPatternVariant(
				uint8(pos), variant, blockStart,
				func(rec *logrec) (gonext bool, err error) {
					wgStart.Wait()

					if rec == nil {
						wgFinish.Done()
						return
					}

					err = ctx.Err()
					gonext = (err == nil)

					if rec.topicsCount < uint8(len(pattern)-1) {
						return
					}
					if blockEnd > 0 && rec.ID.BlockNumber() > blockEnd {
						gonext = false
						return
					}

					flag.Lock()
					defer flag.Unlock()

					if prev, ok := result[rec.ID]; ok {
						prev.matched++
					} else {
						rec.matched++
						result[rec.ID] = rec
					}

					return
				})
		}
	}
	wgStart.Done()

	wgFinish.Wait()
	var gonext bool
	for _, rec := range result {
		if rec.matched != count {
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
