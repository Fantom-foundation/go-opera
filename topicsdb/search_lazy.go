package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
)

func (tt *Index) searchLazy(pattern [][]common.Hash, blockStart []byte, onMatched func(*logrec) (bool, error)) (err error) {
	_, err = tt.walk(nil, blockStart, pattern, 0, onMatched)
	return
}

// walk for topics recursive.
func (tt *Index) walk(
	rec *logrec, blockStart []byte, pattern [][]common.Hash, pos uint8, onMatched func(*logrec) (bool, error),
) (
	gonext bool, err error,
) {
	gonext = true
	for {
		// Max recursion depth is equal to len(topics) and limited by MaxCount.
		if pos >= uint8(len(pattern)) {
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

	for _, variant := range pattern[pos] {
		var (
			prefix  [topicKeySize]byte
			prefLen int
		)
		copy(prefix[prefLen:], variant.Bytes())
		prefLen += common.HashLength
		copy(prefix[prefLen:], posToBytes(pos))
		prefLen += uint8Size
		if rec != nil {
			copy(prefix[prefLen:], rec.ID.Bytes())
			prefLen += logrecKeySize
		}

		it := tt.table.Topic.NewIterator(prefix[:prefLen], blockStart)
		for it.Next() {
			id := extractLogrecID(it.Key())
			topicCount := bytesToPos(it.Value())
			newRec := newLogrec(id, topicCount)
			gonext, err = tt.walk(newRec, nil, pattern, pos+1, onMatched)
			if err != nil || !gonext {
				it.Release()
				return
			}
		}
		it.Release()
		// TODO: why the panic ?
		/*
			err = it.Error()
			if err != nil {
				return
			}
		*/
	}

	return
}
