package topicsdb

import (
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/ethereum/go-ethereum/common"
)

const (
	uint8Size  = 1
	uint64Size = 8
	hashSize   = common.HashLength

	logrecKeySize = uint64Size + hashSize + uint64Size
	topicKeySize  = hashSize + uint8Size + logrecKeySize
	otherKeySize  = logrecKeySize + uint8Size
)

type (
	// ID of log record
	ID [logrecKeySize]byte
)

func NewID(block uint64, tx common.Hash, logIndex uint) (id ID) {
	copy(id[:], uintToBytes(block))
	copy(id[uint64Size:], tx.Bytes())
	copy(id[uint64Size+hashSize:], uintToBytes(uint64(logIndex)))
	return
}

func (id *ID) Bytes() []byte {
	return (*id)[:]
}

func (id *ID) BlockNumber() uint64 {
	return bytesToUint((*id)[:uint64Size])
}

func (id *ID) TxHash() (tx common.Hash) {
	copy(tx[:], (*id)[uint64Size:uint64Size+hashSize])
	return
}

func (id *ID) Index() uint {
	return uint(bytesToUint(
		(*id)[uint64Size+hashSize : uint64Size+hashSize+uint64Size]))
}

func topicKey(topic common.Hash, pos uint8, logrec ID) []byte {
	key := make([]byte, 0, topicKeySize)

	key = append(key, topic.Bytes()...)
	key = append(key, posToBytes(pos)...)
	key = append(key, logrec.Bytes()...)

	return key
}

func posToBytes(pos uint8) []byte {
	return []byte{pos}
}

func bytesToPos(b []byte) uint8 {
	return uint8(b[0])
}

func uintToBytes(n uint64) []byte {
	return bigendian.Uint64ToBytes(n)
}

func bytesToUint(b []byte) uint64 {
	return bigendian.BytesToUint64(b)
}

func extractLogrecID(key []byte) (id ID) {
	switch len(key) {
	case topicKeySize:
		copy(id[:], key[hashSize+uint8Size:])
		return
	default:
		panic("wrong key type")
	}
}
