package topicsdb

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

const MaxCount = 0xff

var ErrTooManyTopics = fmt.Errorf("Too many topics")

// TopicsDb is a specialized indexes for log records storing and fetching.
type TopicsDb struct {
	db    kvdb.KeyValueStore
	table struct {
		// topic+topicN+(blockN+TxHash+logIndex) -> topic_count
		Topic kvdb.KeyValueStore `table:"topic"`
		// (blockN+TxHash+logIndex) + topicN -> topic
		Other kvdb.KeyValueStore `table:"other"`
		// (blockN+TxHash+logIndex) -> data
		Logrec kvdb.KeyValueStore `table:"logrec"`
	}

	fetchMethod func(...Condition) ([]*types.Log, error)
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
func (tt *TopicsDb) Find(cc ...Condition) ([]*types.Log, error) {
	return tt.fetchMethod(cc...)
}

// Push log record to database.
func (tt *TopicsDb) Push(rec *types.Log) error {
	if len(rec.Topics) > MaxCount {
		return ErrTooManyTopics
	}
	count := posToBytes(uint8(len(rec.Topics)))

	//	fmt.Println("PUSH", recToString(rec))

	id := NewID(rec.BlockNumber, rec.TxHash, rec.Index)

	for pos, topic := range rec.Topics {
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
	}

	err := tt.table.Logrec.Put(id.Bytes(), rec.Data)
	if err != nil {
		return err
	}

	return nil
}

/*
func recToString(rec *types.Log) string {
	return fmt.Sprintf("{%d,%s,%d,[topics]}", rec.BlockNumber, rec.TxHash.String(), rec.Index)
}
*/
