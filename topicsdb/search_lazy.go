package topicsdb

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

func (tt *Index) searchLazy(ctx context.Context, pattern [][]common.Hash, blockStart []byte, onMatched func(*logrec) (bool, error)) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	_, err = tt.walk(ctx, nil, blockStart, pattern, 0, onMatched)
	return
}

// walk for topics recursive.
func (tt *Index) walk(
	ctx context.Context, rec *logrec, blockStart []byte, pattern [][]common.Hash, pos uint8, onMatched func(*logrec) (bool, error),
) (
	gonext bool, err error,
) {
	patternLen := uint8(len(pattern))
	gonext = true
	for {
		// Max recursion depth is equal to len(topics) and limited by MaxCount.
		if pos >= patternLen {
			if rec == nil {
				return
			}
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
		prefLen int = common.HashLength + uint8Size
	)
	copy(prefix[common.HashLength:], posToBytes(pos))
	if rec != nil {
		copy(prefix[prefLen:], rec.ID.Bytes())
		prefLen += logrecKeySize
	}
	for _, variant := range pattern[pos] {
		copy(prefix[0:], variant.Bytes())
		it := tt.table.Topic.NewIterator(prefix[:prefLen], blockStart)
		for it.Next() {
			err = ctx.Err()
			if err != nil {
				it.Release()
				return
			}

			var newRec *logrec
			if rec != nil {
				newRec = rec
			} else {
				topicCount := bytesToPos(it.Value())
				if topicCount < (patternLen - 1) {
					continue
				}
				id := extractLogrecID(it.Key())
				newRec = newLogrec(id, topicCount)
			}
			gonext, err = tt.walk(ctx, newRec, nil, pattern, pos+1, onMatched)
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
