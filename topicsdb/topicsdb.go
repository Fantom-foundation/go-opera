package topicsdb

import (
	"context"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const MaxTopicsCount = 5 // count is limited hard to 5 by EVM (see LOG0...LOG4 ops)

var (
	ErrEmptyTopics  = fmt.Errorf("empty topics")
	ErrTooBigTopics = fmt.Errorf("too many topics")
)

// Index is a specialized indexes for log records storing and fetching.
type Index struct {
	table struct {
		// topic+topicN+(blockN+TxHash+logIndex) -> topic_count (where topicN=0 is for address)
		Topic kvdb.Store `table:"t"`
		// (blockN+TxHash+logIndex) -> ordered topic_count topics, blockHash, address, data
		Logrec kvdb.Store `table:"r"`
	}
}

// New Index instance.
func New(dbs kvdb.DBProducer) *Index {
	tt := &Index{}

	err := table.OpenTables(&tt.table, dbs, "evm-logs")
	if err != nil {
		panic(err)
	}

	return tt
}

func (tt *Index) WrapTablesAsBatched() (unwrap func()) {
	origTables := tt.table
	batchedTopic := batched.Wrap(tt.table.Topic)
	tt.table.Topic = batchedTopic
	batchedLogrec := batched.Wrap(tt.table.Logrec)
	tt.table.Logrec = batchedLogrec
	return func() {
		_ = batchedTopic.Flush()
		_ = batchedLogrec.Flush()
		tt.table = origTables
	}
}

// FindInBlocks returns all log records of block range by pattern. 1st pattern element is an address.
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
	return tt.ForEachInBlocks(ctx, 0, 0, pattern, onLog)
}

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *Index) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
	if 0 < to && to < from {
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

	return tt.searchParallel(ctx, pattern, uint64(from), uint64(to), onMatched)
}

func limitPattern(pattern [][]common.Hash) (limited [][]common.Hash, err error) {
	if len(pattern) > (MaxTopicsCount + 1) {
		limited = make([][]common.Hash, (MaxTopicsCount + 1))
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

// Push log record to database batch
func (tt *Index) Push(recs ...*types.Log) error {
	for _, rec := range recs {
		if len(rec.Topics) > MaxTopicsCount {
			return ErrTooBigTopics
		}

		id := NewID(rec.BlockNumber, rec.TxHash, rec.Index)

		// write data
		buf := make([]byte, 0, common.HashLength*len(rec.Topics)+common.HashLength+common.AddressLength+len(rec.Data))
		for _, topic := range rec.Topics {
			buf = append(buf, topic.Bytes()...)
		}
		buf = append(buf, rec.BlockHash.Bytes()...)
		buf = append(buf, rec.Address.Bytes()...)
		buf = append(buf, rec.Data...)
		if err := tt.table.Logrec.Put(id.Bytes(), buf); err != nil {
			return err
		}

		// write index
		var (
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

		for _, topic := range rec.Topics {
			if err := pushIndex(topic); err != nil {
				return err
			}
		}

	}

	return nil
}

func (tt *Index) Close() {
	_ = tt.table.Topic.Close()
	_ = tt.table.Logrec.Close()
}
