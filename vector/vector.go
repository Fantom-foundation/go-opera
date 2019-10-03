package vector

import (
	"encoding/binary"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

/*
 * Use binary form for optimization, to avoid serialization. As a result, DB cache works as elements cache.
 */

type (
	// LowestAfterSeq is vector of lowest events (their Seq) which do observe the source event
	LowestAfterSeq []byte
	// HighestBeforeSeq is vector of highest events (their Seq + IsForkDetected) which are observed by source event
	HighestBeforeSeq []byte
	// HighestBeforeSeq is vector of highest events (their ClaimedTime) which are observed by source event
	HighestBeforeTime []byte

	// ForkSeq encodes IsForkDetected and Seq into 4 bytes
	ForkSeq struct {
		IsForkDetected bool
		Seq            idx.Event
	}

	allVecs struct {
		afterCause  LowestAfterSeq
		beforeCause HighestBeforeSeq
		beforeTime  HighestBeforeTime
	}
)

// NewLowestAfterSeq creates new LowestAfterSeq vector.
func NewLowestAfterSeq(size int) LowestAfterSeq {
	return make(LowestAfterSeq, size*4)
}

// NewHighestBeforeTime creates new HighestBeforeTime vector.
func NewHighestBeforeTime(size int) HighestBeforeTime {
	return make(HighestBeforeTime, size*8)
}

// NewHighestBeforeSeq creates new HighestBeforeSeq vector.
func NewHighestBeforeSeq(size int) HighestBeforeSeq {
	return make(HighestBeforeSeq, size*4)
}

// Get i's position in the byte-encoded vector clock
func (b LowestAfterSeq) Get(i idx.Validator) idx.Event {
	return idx.Event(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
}

// Set i's position in the byte-encoded vector clock
func (b LowestAfterSeq) Set(i idx.Validator, seq idx.Event) {
	binary.LittleEndian.PutUint32(b[i*4:(i+1)*4], uint32(seq))
}

// Get i's position in the byte-encoded vector clock
func (b HighestBeforeTime) Get(i idx.Validator) inter.Timestamp {
	return inter.Timestamp(binary.LittleEndian.Uint64(b[i*8 : (i+1)*8]))
}

// Set i's position in the byte-encoded vector clock
func (b HighestBeforeTime) Set(i idx.Validator, time inter.Timestamp) {
	binary.LittleEndian.PutUint64(b[i*8:(i+1)*8], uint64(time))
}

// ValidatorsNum returns the vector size
func (b HighestBeforeSeq) ValidatorsNum() idx.Validator {
	return idx.Validator(len(b) / 4)
}

// Get i's position in the byte-encoded vector clock
func (b HighestBeforeSeq) Get(i idx.Validator) ForkSeq {
	raw := binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])

	return ForkSeq{
		Seq:            idx.Event(raw >> 1),
		IsForkDetected: (raw & 1) != 0,
	}
}

// Set i's position in the byte-encoded vector clock
func (b HighestBeforeSeq) Set(n idx.Validator, seq ForkSeq) {
	isForkSeen := uint32(0)
	if seq.IsForkDetected {
		isForkSeen = 1
	}
	raw := (uint32(seq.Seq) << 1) | isForkSeen

	binary.LittleEndian.PutUint32(b[n*4:(n+1)*4], raw)
}
