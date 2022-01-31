package topicsdb

import (
	"context"
	"fmt"
	"math"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const MaxTopicsCount = math.MaxUint8

var (
	ErrEmptyTopics = fmt.Errorf("Empty topics")
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
	pattern, err := limitPattern(pattern)
	if err != nil {
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

	return tt.searchLazy(ctx, pattern, nil, 0, onMatched)
}

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *Index) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if from > to {
		return nil
	}

	pattern, err := limitPattern(pattern)
	if err != nil {
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

	return tt.searchLazy(ctx, pattern, uintToBytes(uint64(from)), uint64(to), onMatched)
}

func limitPattern(pattern [][]common.Hash) (limited [][]common.Hash, err error) {
	if len(pattern) > MaxTopicsCount {
		limited = make([][]common.Hash, MaxTopicsCount)
	} else {
		limited = make([][]common.Hash, len(pattern))
	}
	copy(limited, pattern)

	ok := false
	for i, variants := range limited {
		ok = ok || len(variants) > 0
		if len(variants) > 1 {
			limited[i] = uniqOnly(variants)
		}
	}
	if !ok {
		err = ErrEmptyTopics
		return
	}

	return
}

func uniqOnly(hh []common.Hash) []common.Hash {
	index := make(map[common.Hash]struct{}, len(hh))
	for _, h := range hh {
		index[h] = struct{}{}
	}

	uniq := make([]common.Hash, 0, len(index))
	for h := range index {
		uniq = append(uniq, h)
	}
	return uniq
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
		var (
			id    = NewID(rec.BlockNumber, rec.TxHash, rec.Index)
			count = posToBytes(uint8(len(rec.Topics)))
			pos   uint8
		)
		pushIndex := func(topic common.Hash) error {
			key := topicKey(topic, pos, id)
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
		for j, topic := range rec.Topics {
			if j >= MaxTopicsCount {
				break // to don't overflow the pos
			}
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
