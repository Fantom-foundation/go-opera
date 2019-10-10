package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type (
	Topic struct {
		Val  common.Hash
		Data []byte
	}

	Record struct {
		Id     common.Hash
		BlockN uint64
		Topics []*Topic
	}

	recordBuilder struct {
		conditions2check int
		id               common.Hash
		blockN           uint64
		topicCount       uint32
		topics           []*Topic
	}
)

func newRecordBuilder(conditions int, id common.Hash, blockN uint64, topicCount uint32) *recordBuilder {
	return &recordBuilder{
		conditions2check: conditions,
		id:               id,
		blockN:           blockN,
		topicCount:       topicCount,
		topics:           make([]*Topic, topicCount),
	}
}

func (b *recordBuilder) Build() *Record {
	return &Record{
		Id:     b.id,
		BlockN: b.blockN,
		Topics: b.topics,
	}
}

func (b *recordBuilder) ConditionOK(cond Condition) bool {
	b.conditions2check--
	return b.conditions2check == 0
}

func (b *recordBuilder) AllConditionsOK() bool {
	return b.conditions2check == 0
}

func (b *recordBuilder) SetParams(blockN uint64, topicCount uint32) {
	if blockN != b.blockN {
		log.Crit("inconsistent table.Topic", "param", "blockN")
	}
	if topicCount != b.topicCount {
		log.Crit("inconsistent table.Topic", "param", "topicCount")
	}
}

func (b *recordBuilder) SetTopic(n uint32, raw []byte) {
	if n >= b.topicCount {
		log.Crit("inconsistent table.Record", "param", "topicN")
	}

	b.topics[n] = &Topic{
		Val:  common.BytesToHash(raw[:lenHash]),
		Data: raw[lenHash:],
	}
}

func (t *Topic) Bytes() []byte {
	return append(t.Val.Bytes(), t.Data...)
}
