package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type (
	Topic struct {
		Topic common.Hash
		Data  []byte
	}

	Logrec struct {
		ID     common.Hash
		BlockN uint64
		Topics []*Topic
	}

	logrecBuilder struct {
		conditions2check uint8
		id               common.Hash
		blockN           uint64
		topicsCount      uint8
		topicsReady      uint8
		topics           []*Topic

		ok    chan struct{}
		ready chan error
	}
)

func newLogrecBuilder(conditions uint8, id common.Hash, blockN uint64, topicCount uint8) *logrecBuilder {
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
		ID:     rec.id,
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
}

func (rec *logrecBuilder) AllConditionsOK() bool {
	return rec.conditions2check == 0
}

func (rec *logrecBuilder) SetParams(blockN uint64, topicCount uint8) {
	if blockN != rec.blockN {
		log.Crit("inconsistent table.Topic", "param", "blockN")
	}
	if topicCount != rec.topicsCount {
		log.Crit("inconsistent table.Topic", "param", "topicCount")
	}
}

func (rec *logrecBuilder) SetTopic(pos uint8, raw []byte) {
	if pos >= rec.topicsCount {
		log.Crit("inconsistent table.Record", "param", "topicN")
	}

	if rec.topics[pos] == nil {
		rec.topicsReady++
	}
	rec.topics[pos] = &Topic{
		Topic: common.BytesToHash(raw[:lenHash]),
		Data:  raw[lenHash:],
	}

}

func (t *Topic) Bytes() []byte {
	return append(t.Topic.Bytes(), t.Data...)
}
