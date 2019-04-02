package posposet

import (
	"math"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type (
	// TimestampsByEvent is a timestamps by event index.
	TimestampsByEvent map[hash.Event]inter.Timestamp
)

// ToWire converts to simple slice.
func (tt TimestampsByEvent) ToWire() map[string]uint64 {
	res := make(map[string]uint64, len(tt))

	for e, t := range tt {
		res[e.Hex()] = uint64(t)
	}

	return res
}

// WireToTimestampsByEvent converts from wire.
func WireToTimestampsByEvent(arr map[string]uint64) TimestampsByEvent {
	res := make(TimestampsByEvent, len(arr))

	for hex, t := range arr {
		hash := hash.HexToEventHash(hex)
		res[hash] = inter.Timestamp(t)
	}

	return res
}

/*
 * timeCounter:
 */

type timeCounter map[inter.Timestamp]uint

func (c timeCounter) Add(t inter.Timestamp) {
	c[t]++
}

func (c timeCounter) MaxMin() inter.Timestamp {
	var maxs []inter.Timestamp
	freq := uint(0)
	for t, n := range c {
		if n > freq {
			maxs = []inter.Timestamp{t}
			freq = n
		}
		if n == freq {
			maxs = append(maxs, t)
		}
	}

	min := inter.Timestamp(math.MaxUint64)
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
	maxTime inter.Timestamp
}

func newLamportTimeValidator(e *Event) *lamportTimeValidator {
	return &lamportTimeValidator{
		event:   e,
		maxTime: 0,
	}
}

func (v *lamportTimeValidator) IsGreaterThan(time inter.Timestamp) bool {
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

func (v *lamportTimeValidator) GetNext() inter.Timestamp {
	return v.maxTime + 1
}
