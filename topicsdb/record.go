package topicsdb

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type (
	Topic struct {
		Topic common.Hash
		Data  []byte
	}

	Logrec struct {
		Id     common.Hash
		BlockN uint64
		Topics []*Topic
	}

	logrecBuilder struct {
		conditions2check int
		id               common.Hash
		blockN           uint64
		topicsCount      uint32
		topicsReady      uint32
		topics           []*Topic

		ok    chan struct{}
		ready chan error
	}
)

func newLogrecBuilder(conditions int, id common.Hash, blockN uint64, topicCount uint32) *logrecBuilder {
	return &logrecBuilder{
		conditions2check: conditions,
		id:               id,
		blockN:           blockN,
		topicsCount:      topicCount,
		topics:           make([]*Topic, topicCount),
	}
}

func (rec *logrecBuilder) Build() (r *Logrec, err error) {
	if rec.ready != nil {
		var complete bool
		err, complete = <-rec.ready
		if !complete {
			return nil, nil
		}
	}

	r = &Logrec{
		Id:     rec.id,
		BlockN: rec.blockN,
		Topics: rec.topics,
	}

	return
}

func (rec *logrecBuilder) ConditionOK(cond Condition) {
	rec.conditions2check--
	if rec.conditions2check == 0 && rec.ok != nil {
		rec.ok <- struct{}{}
	}

	if rec.conditions2check < 0 {
		log.Crit("topicsdb.recordBuilder sanity check", "conditions2check", "wrong")
	}
}

func (rec *logrecBuilder) AllConditionsOK() bool {
	return rec.conditions2check == 0
}

func (rec *logrecBuilder) SetParams(blockN uint64, topicCount uint32) {
	if blockN != rec.blockN {
		log.Crit("inconsistent table.Topic", "param", "blockN")
	}
	if topicCount != rec.topicsCount {
		log.Crit("inconsistent table.Topic", "param", "topicCount")
	}
}

func (rec *logrecBuilder) SetTopic(n uint32, raw []byte) {

	time.Sleep(time.Millisecond) // TODO

	if n >= rec.topicsCount {
		log.Crit("inconsistent table.Record", "param", "topicN")
	}

	if rec.topics[n] == nil {
		rec.topicsReady++
	}
	rec.topics[n] = &Topic{
		Topic: common.BytesToHash(raw[:lenHash]),
		Data:  raw[lenHash:],
	}

}

func (t *Topic) Bytes() []byte {
	return append(t.Topic.Bytes(), t.Data...)
}
