package posposet

import (
	"io"
	"math"

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
 * timestampsByEvent:
 */

type (
	timestampsByEvent map[EventHash]Timestamp

	// storedEventTime is an internal struct for specialization purpose.
	storedEventTime struct {
		E EventHash
		T Timestamp
	}
)

// EncodeRLP is a specialized encoder to encode index into array.
func (tt timestampsByEvent) EncodeRLP(w io.Writer) error {
	var arr []storedEventTime
	for e, t := range tt {
		arr = append(arr, storedEventTime{e, t})
	}
	return rlp.Encode(w, arr)
}

// DecodeRLP is a specialized decoder to decode index from array.
func (tt *timestampsByEvent) DecodeRLP(s *rlp.Stream) error {
	var arr []storedEventTime
	err := s.Decode(&arr)
	if err != nil {
		return err
	}

	res := make(timestampsByEvent, len(arr))
	for _, pair := range arr {
		res[pair.E] = pair.T
	}

	*tt = res
	return nil
}

/*
 * timeCounter:
 */

type timeCounter map[Timestamp]uint

func (c timeCounter) Add(t Timestamp) {
	c[t] += 1
}

func (c timeCounter) MaxMin() Timestamp {
	var maxs []Timestamp
	freq := uint(0)
	for t, n := range c {
		if n > freq {
			maxs = []Timestamp{t}
			freq = n
		}
		if n == freq {
			maxs = append(maxs, t)
		}
	}

	min := Timestamp(math.MaxUint64)
	for _, t := range maxs {
		if min > t {
			min = t
		}
	}
	return min
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
