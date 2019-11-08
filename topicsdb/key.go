package topicsdb

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

const (
	lenUint8 = 1
	lenInt64 = 8
	lenHash  = 32

	topicKeySize  = lenHash + lenUint8 + lenInt64 + lenHash
	recordKeySize = lenHash + lenUint8
)

func topicKey(t *Topic, pos uint8, block uint64, id common.Hash) []byte {
	key := make([]byte, 0, topicKeySize)

	key = append(key, t.Topic.Bytes()...)
	key = append(key, posToBytes(pos)...)
	key = append(key, blockToBytes(block)...)
	key = append(key, id.Bytes()...)

	return key
}

func logrecKey(r *Logrec, pos uint8) []byte {
	key := make([]byte, 0, recordKeySize)

	key = append(key, r.ID.Bytes()...)
	key = append(key, posToBytes(pos)...)

	return key
}

func posToBytes(pos uint8) []byte {
	return []byte{pos}
}

func bytesToPos(b []byte) uint8 {
	return uint8(b[0])
}

func blockToBytes(n uint64) []byte {
	return bigendian.Int64ToBytes(n)
}

func bytesToBlock(b []byte) uint64 {
	return bigendian.BytesToInt64(b)
}

func extractLogrecID(key []byte) common.Hash {
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
		return bytesToBlock(
			key[lenHash+lenUint8 : lenHash+lenUint8+lenInt64])
	default:
		panic("unknown key type")
	}
}

func extractTopicPos(key []byte) uint8 {
	switch len(key) {
	case topicKeySize:
		return bytesToPos(
			key[lenHash : lenHash+lenUint8])
	case recordKeySize:
		return bytesToPos(
			key[lenHash : lenHash+lenUint8])
	default:
		panic("unknown key type")
	}
}
