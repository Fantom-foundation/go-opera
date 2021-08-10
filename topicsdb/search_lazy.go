package topicsdb

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

type logHandler func(rec *logrec) (gonext bool, err error)

func (tt *Index) searchLazy(ctx context.Context, pattern [][]common.Hash, blockStart []byte, blockEnd uint64, onMatched logHandler) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	_, err = tt.walkFirst(ctx, blockStart, blockEnd, pattern, 0, onMatched)
	return
}

// walkFirst for topics recursive.
func (tt *Index) walkFirst(
	ctx context.Context, blockStart []byte, blockEnd uint64, pattern [][]common.Hash, pos uint8, onMatched logHandler,
) (
	gonext bool, err error,
) {
	patternLen := uint8(len(pattern))
	gonext = true
	for {
		if pos >= patternLen {
			return
		}
		if len(pattern[pos]) < 1 {
			pos++
			continue
		}
		break
	}

	var (
		prefix  [topicKeySize]byte
		prefLen int = hashSize
	)
	copy(prefix[prefLen:], posToBytes(pos))
	prefLen += uint8Size

	for _, variant := range pattern[pos] {
		copy(prefix[0:], variant.Bytes())
		it := tt.table.Topic.NewIterator(prefix[:prefLen], blockStart)
		for it.Next() {
			err = ctx.Err()
			if err != nil {
				it.Release()
				return
			}

			topicCount := bytesToPos(it.Value())
			if topicCount < (patternLen - 1) {
				continue
			}
			id := extractLogrecID(it.Key())
			rec := newLogrec(id, topicCount)

			if blockStart != nil && rec.ID.BlockNumber() > blockEnd {
				break
			}

			gonext, err = tt.walkNexts(ctx, rec, pattern, pos+1, onMatched)
			if err != nil || !gonext {
				it.Release()
				return
			}
		}

		err = it.Error()
		it.Release()
		if err != nil {
			return
		}

	}
	return
}

// walkNexts for topics recursive.
func (tt *Index) walkNexts(
	ctx context.Context, rec *logrec, pattern [][]common.Hash, pos uint8, onMatched logHandler,
) (
	gonext bool, err error,
) {
	patternLen := uint8(len(pattern))
	gonext = true
	for {
		// Max recursion depth is equal to len(topics) and limited by MaxCount.
		if pos >= patternLen {
			gonext, err = onMatched(rec)
			return
		}
		if len(pattern[pos]) < 1 {
			pos++
			continue
		}
		break
	}

	var (
		prefix  [topicKeySize]byte
		prefLen int = hashSize
	)
	copy(prefix[prefLen:], posToBytes(pos))
	prefLen += uint8Size
	copy(prefix[prefLen:], rec.ID.Bytes())
	prefLen += logrecKeySize

	for _, variant := range pattern[pos] {
		copy(prefix[0:], variant.Bytes())
		it := tt.table.Topic.NewIterator(prefix[:prefLen], nil)
		for it.Next() {
			err = ctx.Err()
			if err != nil {
				it.Release()
				return
			}

			gonext, err = tt.walkNexts(ctx, rec, pattern, pos+1, onMatched)
			if err != nil || !gonext {
				it.Release()
				return
			}
		}

		err = it.Error()
		it.Release()
		if err != nil {
			return
		}

	}
	return
}
