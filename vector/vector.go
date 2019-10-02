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
	LowestAfterSeq    []byte
	HighestBeforeSeq  []byte
	HighestBeforeTime []byte

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

func NewLowestAfterSeq(size int) LowestAfterSeq {
	return make(LowestAfterSeq, size*4)
}

func NewHighestBeforeTime(size int) HighestBeforeTime {
	return make(HighestBeforeTime, size*8)
}

func NewHighestBeforeSeq(size int) HighestBeforeSeq {
	return make(HighestBeforeSeq, size*4)
}

func (b LowestAfterSeq) Get(n idx.Validator) idx.Event {
	return idx.Event(binary.LittleEndian.Uint32(b[n*4 : (n+1)*4]))
}

func (b LowestAfterSeq) Set(n idx.Validator, seq idx.Event) {
	binary.LittleEndian.PutUint32(b[n*4:(n+1)*4], uint32(seq))
}

func (b HighestBeforeTime) Get(n idx.Validator) inter.Timestamp {
	return inter.Timestamp(binary.LittleEndian.Uint64(b[n*8 : (n+1)*8]))
}

func (b HighestBeforeTime) Set(n idx.Validator, time inter.Timestamp) {
	binary.LittleEndian.PutUint64(b[n*8:(n+1)*8], uint64(time))
}

func (b HighestBeforeSeq) ValidatorsNum() idx.Validator {
	return idx.Validator(len(b) / 4)
}

func (b HighestBeforeSeq) Get(n idx.Validator) ForkSeq {
	raw := binary.LittleEndian.Uint32(b[n*4 : (n+1)*4])

	return ForkSeq{
		Seq:            idx.Event(raw >> 1),
		IsForkDetected: (raw & 1) != 0,
	}
}

func (b HighestBeforeSeq) Set(n idx.Validator, seq ForkSeq) {
	isForkSeen := uint32(0)
	if seq.IsForkDetected {
		isForkSeen = 1
	}
	raw := (uint32(seq.Seq) << 1) | isForkSeen

	binary.LittleEndian.PutUint32(b[n*4:(n+1)*4], raw)
}
