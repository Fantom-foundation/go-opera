package topicsdb

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const MaxTopicsCount = 0xff

var (
	ErrTooManyTopics = fmt.Errorf("Too many topics")
	ErrEmptyTopics   = fmt.Errorf("Empty topics")
)

// Index is a specialized indexes for log records storing and fetching.
type Index struct {
	db    kvdb.Store
	table struct {
		// topic+topicN+(blockN+TxHash+logIndex) -> topic_count (where topicN=0 is for address)
		Topic kvdb.Store `table:"t"`
		// (blockN+TxHash+logIndex) -> ordered topics
		Other kvdb.Store `table:"o"`
		// (blockN+TxHash+logIndex) -> blockHash, address, data
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
func (tt *Index) ForEachInBlocks(from, to idx.Block, topics [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if from > to {
		return nil
	}

	if err := checkTopics(topics); err != nil {
		return err
	}

	moreAccurate := func(l *types.Log) (gonext bool) {
		if uint64(from) > l.BlockNumber || l.BlockNumber > uint64(to) {
			return true
		}
		return onLog(l)
	}

	bm := blocksMask(from, to)
	return tt.fetchLazy(topics, bm, moreAccurate)
}

func checkTopics(topics [][]common.Hash) error {
	if len(topics) > MaxTopicsCount {
		return ErrTooManyTopics
	}

	ok := false
	for _, variants := range topics {
		if len(variants) > MaxTopicsCount {
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
		if len(rec.Topics) > MaxTopicsCount {
			return ErrTooManyTopics
		}

		var (
			buf   []byte
			id    = NewID(rec.BlockNumber, rec.TxHash, rec.Index)
			count = posToBytes(uint8(len(rec.Topics)))
			pos   int
		)
		pushIndex := func(topic common.Hash) error {
			key := topicKey(topic, uint8(pos), id)
			if err := tt.table.Topic.Put(key, count); err != nil {
				return err
			}
			pos++
			return nil
		}

		if err := pushIndex(rec.Address.Hash()); err != nil {
			return err
		}

		buf = make([]byte, 0, common.HashLength*len(rec.Topics))
		for _, topic := range rec.Topics {
			if err := pushIndex(topic); err != nil {
				return err
			}
			buf = append(buf, topic.Bytes()...)
		}
		if err := tt.table.Other.Put(id.Bytes(), buf); err != nil {
			return err
		}

		buf = make([]byte, 0, common.HashLength+common.AddressLength+len(rec.Data))
		buf = append(buf, rec.BlockHash.Bytes()...)
		buf = append(buf, rec.Address.Bytes()...)
		buf = append(buf, rec.Data...)

		if err := tt.table.Logrec.Put(id.Bytes(), buf); err != nil {
			return err
		}
	}

	return nil
}
