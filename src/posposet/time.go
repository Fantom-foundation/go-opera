package posposet

import (
	"math"
)

type (
	Timestamp uint64

	timestampsByEvent map[EventHash]Timestamp
)

// ToWire converts to simple slice.
func (tt timestampsByEvent) ToWire() map[string]uint64 {
	res := make(map[string]uint64, len(tt))

	for e, t := range tt {
		res[e.Hex()] = uint64(t)
	}

	return res
}

// WireToTimestampsByEvent converts from wire.
func WireToTimestampsByEvent(arr map[string]uint64) timestampsByEvent {
	res := make(timestampsByEvent, len(arr))

	for hex, t := range arr {
		hash := HexToEventHash(hex)
		res[hash] = Timestamp(t)
	}

	return res
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
