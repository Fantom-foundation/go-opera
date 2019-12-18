package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"
)

const conditionSize = hashSize + uint8Size

type (
	// Condition to search log records by.
	Condition [][conditionSize]byte
)

// NewCondition from topic and its position in log record.
func NewCondition(position uint8, topics ...common.Hash) (c Condition) {
	c = make(Condition, len(topics))
	for i, topic := range topics {
		copy(c[i][:], topic.Bytes())
		copy(c[i][hashSize:], posToBytes(position))
	}

	return
}
