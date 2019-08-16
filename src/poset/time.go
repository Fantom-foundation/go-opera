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
		hash_ := hash.HexToEventHash(hex)
		res[hash_] = inter.Timestamp(t)
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
