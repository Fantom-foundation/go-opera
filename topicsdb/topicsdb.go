package topicsdb

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

const MaxCount = 0xff

var ErrTooManyTopics = fmt.Errorf("Too many topics")

// TopicsDb is a specialized indexes for log records storing and fetching.
type TopicsDb struct {
	db    kvdb.KeyValueStore
	table struct {
		// topic+topicN+blockN+logrecID -> pair_count
		Topic kvdb.KeyValueStore `table:"t"`
		// logrecID+topicN -> topic, data
		Logrec kvdb.KeyValueStore `table:"l"`
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
	if len(rec.Topics) > MaxCount {
		return ErrTooManyTopics
	}
	count := posToBytes(uint8(len(rec.Topics)))

	for pos, topic := range rec.Topics {
		key := topicKey(topic, uint8(pos), rec.BlockN, rec.ID)
		err := tt.table.Topic.Put(key, count)
		if err != nil {
			return err
		}
	}

	for pos, topic := range rec.Topics {
		key := logrecKey(rec, uint8(pos))
		err := tt.table.Logrec.Put(key, topic.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}
