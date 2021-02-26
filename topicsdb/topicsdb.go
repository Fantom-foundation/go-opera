package topicsdb

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const MaxCount = 0xff

var (
	ErrTooManyTopics = fmt.Errorf("Too many topics")
	ErrEmptyTopics   = fmt.Errorf("Empty topics")
)

// Index is a specialized indexes for log records storing and fetching.
type Index struct {
	db    kvdb.Store
	table struct {
		// topic+topicN+(blockN+TxHash+logIndex) -> topic_count
		Topic kvdb.Store `table:"t"`
		// (blockN+TxHash+logIndex) + topicN -> topic (where topicN=0 is for address)
		Other kvdb.Store `table:"o"`
		// (blockN+TxHash+logIndex) -> blockHash, data
		Logrec kvdb.Store `table:"r"`
	}
}

// New Index instance.
func New(db kvdb.Store) *Index {
	tt := &Index{
		db: db,
	}

	table.MigrateTables(&tt.table, tt.db)

	return tt
}

// ForEach matches log records by topics. 1st topics element is an address.
func (tt *Index) ForEach(topics [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if err := checkTopics(topics); err != nil {
		return err
	}

	return tt.fetchLazy(topics, nil, onLog)
}

// ForEachInBlocks matches log records of block range by topics. 1st topics element is an address.
func (tt *Index) ForEachInBlocks(blockFrom, blockTo uint64, topics [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if err := checkTopics(topics); err != nil {
		return err
	}

	bm := blocksMask(blockFrom, blockTo)

	if blockFrom > blockTo {
		blockFrom, blockTo = blockTo, blockFrom
	}
	moreAccurate := func(l *types.Log) (gonext bool) {
		if blockFrom > l.BlockNumber || l.BlockNumber > blockTo {
			return true
		}
		return onLog(l)
	}

	return tt.fetchLazy(topics, bm, moreAccurate)
}

func checkTopics(topics [][]common.Hash) error {
	if len(topics) > MaxCount {
		return ErrTooManyTopics
	}

	ok := false
	for _, variants := range topics {
		if len(variants) > MaxCount {
			return ErrTooManyTopics
		}
		ok = ok || len(variants) > 0
	}
	if !ok {
		return ErrEmptyTopics
	}

	return nil
}

// MustPush calls Write() and panics if error.
func (tt *Index) MustPush(recs ...*types.Log) {
	err := tt.Push(recs...)
	if err != nil {
		panic(err)
	}
}

// Write log record to database.
func (tt *Index) Push(recs ...*types.Log) error {
	for _, rec := range recs {
		if len(rec.Topics) > MaxCount {
			return ErrTooManyTopics
		}
		count := posToBytes(uint8(1 + len(rec.Topics)))

		id := NewID(rec.BlockNumber, rec.TxHash, rec.Index)

		var pos int
		push := func(topic common.Hash) error {
			key := topicKey(topic, uint8(pos), id)
			err := tt.table.Topic.Put(key, count)
			if err != nil {
				return err
			}

			key = otherKey(id, uint8(pos))
			err = tt.table.Other.Put(key, topic.Bytes())
			if err != nil {
				return err
			}

			pos++
			return nil
		}

		if err := push(rec.Address.Hash()); err != nil {
			return err
		}
		for _, topic := range rec.Topics {
			if err := push(topic); err != nil {
				return err
			}
		}

		buf := make([]byte, 0, common.HashLength+len(rec.Data))
		buf = append(buf, rec.BlockHash.Bytes()...)
		buf = append(buf, rec.Data...)

		err := tt.table.Logrec.Put(id.Bytes(), buf)
		if err != nil {
			return err
		}
	}

	return nil
}
