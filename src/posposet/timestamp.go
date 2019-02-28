package posposet

import (
	"io"

	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

/*
 * Timestamp:
 */

type Timestamp uint64

// EncodeRLP is a specialized encoder to encode Timestamp into array.
func (t Timestamp) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, uint64(t))
}

// DecodeRLP is a specialized decoder to decode Timestamp from array.
func (t *Timestamp) DecodeRLP(s *rlp.Stream) error {
	return s.Decode((*uint64)(t))

}

/*
 * Event's lamport time validator:
 */

type lamportTimeValidator struct {
	event   *Event
	maxTime Timestamp
}

func newLamportTimeValidator(e *Event) *lamportTimeValidator {
	return &lamportTimeValidator{
		event:   e,
		maxTime: 0,
	}
}

func (v *lamportTimeValidator) IsGreaterThan(time Timestamp) bool {
	if v.event.LamportTime <= time {
		log.Warnf("Event %s has lamport time %d. It isn't next of parents, so rejected",
			v.event.Hash().String(),
			v.event.LamportTime)
		return false
	}
	if v.maxTime < time {
		v.maxTime = time
	}
	return true
}

func (v *lamportTimeValidator) IsSequential() bool {
	if v.event.LamportTime != v.maxTime+1 {
		log.Warnf("Event %s has lamport time %d. It is too far from parents, so rejected",
			v.event.Hash().String(),
			v.event.LamportTime)
		return false
	}
	return true
}

func (v *lamportTimeValidator) GetNext() Timestamp {
	return v.maxTime + 1
}
