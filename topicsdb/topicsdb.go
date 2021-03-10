package topicsdb

import (
	"context"
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
		// (blockN+TxHash+logIndex) -> ordered topic_count topics, blockHash, address, data
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

// FindInBlocks returns all log records of block range by pattern. 1st pattern element is an address.
// The same as FindInBlocksAsync but fetches log's body sync.
func (tt *Index) FindInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash) (logs []*types.Log, err error) {
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
func (tt *Index) ForEach(ctx context.Context, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if err := checkPattern(pattern); err != nil {
		return err
	}

	onMatched := func(rec *logrec) (gonext bool, err error) {
		rec.fetch(tt.table.Logrec)
		if rec.err != nil {
			err = rec.err
			return
		}
		gonext = onLog(rec.result)
		return
	}

	return tt.searchLazy(ctx, pattern, nil, onMatched)
}

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *Index) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if from > to {
		return nil
	}

	if err := checkPattern(pattern); err != nil {
		return err
	}

	onMatched := func(rec *logrec) (gonext bool, err error) {
		if rec.ID.BlockNumber() > uint64(to) {
			return
		}

		rec.fetch(tt.table.Logrec)
		if rec.err != nil {
			err = rec.err
			return
		}
		gonext = onLog(rec.result)
		return
	}

	return tt.searchLazy(ctx, pattern, uintToBytes(uint64(from)), onMatched)
}

func checkPattern(pattern [][]common.Hash) error {
	if len(pattern) > MaxTopicsCount {
		return ErrTooManyTopics
	}

	ok := false
	for _, variants := range pattern {
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

		buf := make([]byte, 0, common.HashLength*len(rec.Topics)+common.HashLength+common.AddressLength+len(rec.Data))
		for _, topic := range rec.Topics {
			if err := pushIndex(topic); err != nil {
				return err
			}
			buf = append(buf, topic.Bytes()...)
		}

		buf = append(buf, rec.BlockHash.Bytes()...)
		buf = append(buf, rec.Address.Bytes()...)
		buf = append(buf, rec.Data...)

		if err := tt.table.Logrec.Put(id.Bytes(), buf); err != nil {
			return err
		}
	}

	return nil
}
