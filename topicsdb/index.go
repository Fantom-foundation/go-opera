package topicsdb

import (
	"context"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// index is a specialized indexes for log records storing and fetching.
type index struct {
	table struct {
		// topic+topicN+(blockN+TxHash+logIndex) -> topic_count (where topicN=0 is for address)
		Topic kvdb.Store `table:"t"`
		// (blockN+TxHash+logIndex) -> ordered topic_count topics, blockHash, address, data
		Logrec kvdb.Store `table:"r"`
	}
}

func newIndex(dbs kvdb.DBProducer) *index {
	tt := &index{}

	err := table.OpenTables(&tt.table, dbs, "evm-logs")
	if err != nil {
		panic(err)
	}

	return tt
}

func (tt *index) WrapTablesAsBatched() (unwrap func()) {
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
func (tt *index) FindInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash) (logs []*types.Log, err error) {
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

// ForEachInBlocks matches log records of block range by pattern. 1st pattern element is an address.
func (tt *index) ForEachInBlocks(ctx context.Context, from, to idx.Block, pattern [][]common.Hash, onLog func(*types.Log) (gonext bool)) error {
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

// Push log record to database batch
func (tt *index) Push(recs ...*types.Log) error {
	for _, rec := range recs {
		if len(rec.Topics) > maxTopicsCount {
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

func (tt *index) Close() {
	_ = tt.table.Topic.Close()
	_ = tt.table.Logrec.Close()
}
