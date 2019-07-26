package inter

type (
	// Timestamp is a logical time.
	Timestamp uint64
)

// MaxTimestamp return max value
func MaxTimestamp(x, y Timestamp) Timestamp {
    if x > y {
        return x
    }
    return y
}
