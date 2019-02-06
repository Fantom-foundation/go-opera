package posposet

// Frame
type Frame struct {
	Index uint64
	Roots EventHashes
}

func StartNewFrame(prev uint64) *Frame {
	return &Frame{
		Index: prev + 1,
		Roots: EventHashes{},
	}
}
