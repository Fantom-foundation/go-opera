package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

const (
	lenInt32 = 4
	lenInt64 = 8
	lenHash  = 32

	topicKeySize  = lenHash + lenInt32 + lenInt64 + lenHash
	recordKeySize = lenHash + lenInt32
)

func topicKey(t *Topic, n int, block uint64, id common.Hash) []byte {
	key := make([]byte, 0, topicKeySize)

	key = append(key, t.Topic.Bytes()...)
	key = append(key, bigendian.Int32ToBytes(uint32(n))...)
	key = append(key, bigendian.Int64ToBytes(block)...)
	key = append(key, id.Bytes()...)

	return key
}

func logrecKey(r *Logrec, n int) []byte {
	key := make([]byte, 0, recordKeySize)

	key = append(key, r.Id.Bytes()...)
	key = append(key, bigendian.Int32ToBytes(uint32(n))...)

	return key
}

func extractLogrecId(key []byte) common.Hash {
	switch len(key) {
	case topicKeySize:
		return common.BytesToHash(
			key[topicKeySize-lenHash:])
	case recordKeySize:
		return common.BytesToHash(
			key[:lenHash])
	default:
		panic("unknown key type")
	}
}

func extractBlockN(key []byte) uint64 {
	switch len(key) {
	case topicKeySize:
		return bigendian.BytesToInt64(
			key[lenHash+lenInt32 : lenHash+lenInt32+lenInt64])
	default:
		panic("unknown key type")
	}
}

func extractTopicN(key []byte) uint32 {
	switch len(key) {
	case topicKeySize:
		return bigendian.BytesToInt32(
			key[lenHash : lenHash+lenInt32])
	case recordKeySize:
		return bigendian.BytesToInt32(
			key[lenHash : lenHash+lenInt32])
	default:
		panic("unknown key type")
	}
}
