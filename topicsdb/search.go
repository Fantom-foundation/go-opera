package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

const conditionSize = hashSize + uint8Size

type (
	// Condition to search log records by.
	Condition [conditionSize]byte
)

// NewCondition from topic and its position in log record.
func NewCondition(topic common.Hash, position uint8) (c Condition) {
	copy(c[:], topic.Bytes())
	copy(c[hashSize:], posToBytes(position))

	return
}

// Topic getter.
func (cond Condition) Topic() common.Hash {
	return common.BytesToHash(
		cond[:hashSize])
}

// Position getter.
func (cond Condition) Position() uint8 {
	return uint8(bigendian.BytesToInt32(
		cond[hashSize:]))
}
