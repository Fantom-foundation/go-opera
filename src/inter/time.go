package inter

import (
	"time"
)

type (
	// Timestamp is a logical time.
	// TODO: replace with time.Time ?
	Timestamp uint64
)

func FromUnix(t int64) Timestamp {
	return Timestamp(int64(t) * int64(time.Second))
}

// Unix returns t as a Unix time, the number of seconds elapsed
// since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t Timestamp) Unix() int64 {
	return int64(t) / int64(time.Second)
}

func (t Timestamp) Time() time.Time {
	return time.Unix(int64(t)/int64(time.Second), int64(t)%int64(time.Second))
}

// MaxTimestamp return max value.
func MaxTimestamp(x, y Timestamp) Timestamp {
	if x > y {
		return x
	}
	return y
}
