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
	ErrNoOneTopic    = fmt.Errorf("No one topic")
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

	fetchMethod func(topics [][]common.Hash, onLog func(*types.Log) (next bool)) error
}

// New Index instance.
func New(db kvdb.Store) *Index {
	tt := &Index{
		db: db,
	}

	tt.fetchMethod = tt.fetchLazy

	table.MigrateTables(&tt.table, tt.db)

	return tt
}

// ForEach log records by conditions. 1st topics element is an address.
func (tt *Index) ForEach(topics [][]common.Hash, onLog func(*types.Log) (next bool)) error {
	if len(topics) > MaxCount {
		return ErrTooManyTopics
	}

	ok := false
	for _, alternative := range topics {
		if len(alternative) > MaxCount {
			return ErrTooManyTopics
		}
		ok = ok || len(alternative) > 0
	}
	if !ok {
		return ErrNoOneTopic
	}

	return tt.fetchMethod(topics, onLog)
}

// Find log records by conditions. 1st topics element is an address.
func (tt *Index) Find(topics [][]common.Hash) (all []*types.Log, err error) {
	err = tt.ForEach(topics, func(item *types.Log) (next bool) {
		all = append(all, item)
		next = true
		return
	})

	return
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
