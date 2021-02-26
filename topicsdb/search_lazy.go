package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (tt *Index) fetchLazy(topics [][]common.Hash, blocksMask []byte, onLog func(*types.Log) bool) (err error) {
	_, err = tt.walk(nil, blocksMask, topics, 0, onLog)
	return
}

func (tt *Index) walk(
	rec *logrec, blocksMask []byte, topics [][]common.Hash, pos uint8, onLog func(*types.Log) bool,
) (
	gonext bool, err error,
) {
	gonext = true
	for {
		if pos >= uint8(len(topics)) {
			if rec == nil {
				return
			}

			var r *types.Log
			r, err = rec.FetchLog(tt.table.Other, tt.table.Logrec)
			if err != nil {
				return
			}
			gonext = onLog(r)
			return
		}
		if len(topics[pos]) < 1 {
			pos++
			continue
		}
		break
	}

	for _, variant := range topics[pos] {
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
		} else {
			copy(prefix[prefLen:], blocksMask)
			prefLen += len(blocksMask)
		}

		it := tt.table.Topic.NewIterator(prefix[:prefLen], nil)
		for it.Next() {
			id := extractLogrecID(it.Key())
			topicCount := bytesToPos(it.Value())
			newRec := newLogrec(id, topicCount)
			gonext, err = tt.walk(newRec, nil, topics, pos+1, onLog)
			if err != nil || !gonext {
				it.Release()
				return
			}
		}
		it.Release()
	}

	return
}
