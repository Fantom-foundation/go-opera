package inter

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
)

type (
	// Timestamp is a UNIX nanoseconds timestamp
	Timestamp uint64
)

// Bytes gets the byte representation of the index.
func (t Timestamp) Bytes() []byte {
	return bigendian.Uint64ToBytes(uint64(t))
}

// BytesToTimestamp converts bytes to timestamp.
func BytesToTimestamp(b []byte) Timestamp {
	return Timestamp(bigendian.BytesToUint64(b))
}

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
