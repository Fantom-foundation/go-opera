package topicsdb

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

// TopicsDb is a specialized indexes for log records storing and fetching.
type TopicsDb struct {
	db    kvdb.KeyValueStore
	table struct {
		// topic+topicN+blockN+logrecID -> pair_count
		Topic kvdb.KeyValueStore `table:"topic"`
		// logrecID+N -> topic, data
		Logrec kvdb.KeyValueStore `table:"logrec"`
	}

	fetchMethod func(cc ...Condition) (res []*Logrec, err error)
}

// New TopicsDb instance.
func New(db kvdb.KeyValueStore) *TopicsDb {
	tt := &TopicsDb{
		db: db,
	}

	tt.fetchMethod = tt.fetchAsync

	table.MigrateTables(&tt.table, tt.db)

	return tt
}

// Find log records by conditions.
func (tt *TopicsDb) Find(cc ...Condition) (res []*Logrec, err error) {
	return tt.fetchMethod(cc...)
}

// Push log record to database.
func (tt *TopicsDb) Push(rec *Logrec) error {
	count := bigendian.Int32ToBytes(uint32(len(
		rec.Topics)))

	for n, topic := range rec.Topics {
		key := topicKey(topic, n, rec.BlockN, rec.Id)
		err := tt.table.Topic.Put(key, count)
		if err != nil {
			return err
		}
	}

	for n, topic := range rec.Topics {
		key := logrecKey(rec, n)
		err := tt.table.Logrec.Put(key, topic.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}
