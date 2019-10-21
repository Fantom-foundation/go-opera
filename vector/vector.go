package vector

import (
	"encoding/binary"
	"math"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

/*
 * Use binary form for optimization, to avoid serialization. As a result, DB cache works as elements cache.
 */

type (
	// LowestAfterSeq is a vector of lowest events (their Seq) which do observe the source event
	LowestAfterSeq []byte
	// HighestBeforeSeq is a vector of highest events (their Seq + IsForkDetected) which are observed by source event
	HighestBeforeSeq []byte
	// HighestBeforeTime is a vector of highest events (their ClaimedTime) which are observed by source event
	HighestBeforeTime []byte

	// BranchSeq encodes Seq and MinSeq into 8 bytes
	BranchSeq struct {
		Seq    idx.Event // maximum observed e.Seq in the branch
		MinSeq idx.Event // minimum observed e.Seq in the branch
	}

	// allVecs is container of all the vector clocks
	allVecs struct {
		after      LowestAfterSeq
		beforeSeq  HighestBeforeSeq
		beforeTime HighestBeforeTime
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
	return make(HighestBeforeSeq, size*8)
}

// Get i's position in the byte-encoded vector clock
func (b LowestAfterSeq) Get(i idx.Validator) idx.Event {
	for int(i) >= b.Size() {
		return 0
	}
	return idx.Event(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
}

// Size of the vector clock
func (b LowestAfterSeq) Size() int {
	return len(b) / 4
}

// Set i's position in the byte-encoded vector clock
func (b *LowestAfterSeq) Set(i idx.Validator, seq idx.Event) {
	for int(i) >= b.Size() {
		// append zeros if exceeds size
		*b = append(*b, []byte{0, 0, 0, 0}...)
	}

	binary.LittleEndian.PutUint32((*b)[i*4:(i+1)*4], uint32(seq))
}

// Get i's position in the byte-encoded vector clock
func (b HighestBeforeTime) Get(i idx.Validator) inter.Timestamp {
	for int(i) >= b.Size() {
		return 0
	}
	return inter.Timestamp(binary.LittleEndian.Uint64(b[i*8 : (i+1)*8]))
}

// Set i's position in the byte-encoded vector clock
func (b *HighestBeforeTime) Set(i idx.Validator, time inter.Timestamp) {
	for int(i) >= b.Size() {
		// append zeros if exceeds size
		*b = append(*b, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	}
	binary.LittleEndian.PutUint64((*b)[i*8:(i+1)*8], uint64(time))
}

// Size of the vector clock
func (b HighestBeforeTime) Size() int {
	return len(b) / 8
}

// Size of the vector clock
func (b HighestBeforeSeq) Size() int {
	return len(b) / 8
}

// Get i's position in the byte-encoded vector clock
func (b HighestBeforeSeq) Get(i idx.Validator) BranchSeq {
	for int(i) >= b.Size() {
		return BranchSeq{}
	}
	seq1 := binary.LittleEndian.Uint32(b[i*8 : i*8+4])
	seq2 := binary.LittleEndian.Uint32(b[i*8+4 : i*8+8])

	return BranchSeq{
		Seq:    idx.Event(seq1),
		MinSeq: idx.Event(seq2),
	}
}

// Set i's position in the byte-encoded vector clock
func (b *HighestBeforeSeq) Set(i idx.Validator, seq BranchSeq) {
	for int(i) >= b.Size() {
		// append zeros if exceeds size
		*b = append(*b, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	}
	binary.LittleEndian.PutUint32((*b)[i*8:i*8+4], uint32(seq.Seq))
	binary.LittleEndian.PutUint32((*b)[i*8+4:i*8+8], uint32(seq.MinSeq))
}

var (
	// forkDetectedSeq is a special marker of observed fork by a creator
	forkDetectedSeq = BranchSeq{
		Seq:    0,
		MinSeq: idx.Event(math.MaxInt32),
	}
)

// IsForkDetected returns true if observed fork by a creator (in combination of branches)
func (seq BranchSeq) IsForkDetected() bool {
	return seq == forkDetectedSeq
}
