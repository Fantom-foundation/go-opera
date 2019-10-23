package poset

import (
	"errors"
	"io"
	"math"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// TODO: make FrameInfo internal

// FrameInfo stores persistent data, associated with a frame
type FrameInfo struct {
	TimeOffset        int64 // may be negative
	TimeRatio         inter.Timestamp
	LastConsensusTime inter.Timestamp
}

type frameInfoMarshaling struct {
	TimeOffset        uint64
	TimeRatio         uint64
	LastConsensusTime uint64
}

func (f *FrameInfo) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &frameInfoMarshaling{
		TimeOffset:        uint64(f.TimeOffset + math.MaxInt64/2),
		TimeRatio:         uint64(f.TimeRatio),
		LastConsensusTime: uint64(f.LastConsensusTime),
	})
}

func (f *FrameInfo) DecodeRLP(st *rlp.Stream) error {
	m := &frameInfoMarshaling{}
	if err := st.Decode(m); err != nil {
		return err
	}
	f.TimeOffset = int64(m.TimeOffset) - math.MaxInt64/2
	f.TimeRatio = inter.Timestamp(m.TimeRatio)
	f.LastConsensusTime = inter.Timestamp(m.LastConsensusTime)

	return nil
}

// CalcConsensusTime calcs consensus timestamp, using frame's time offset/ratio
func (f *FrameInfo) CalcConsensusTime(lamport idx.Lamport) inter.Timestamp {
	return inter.Timestamp(int64(lamport)*int64(f.TimeRatio) + f.TimeOffset)
}

// GetConsensusTime calc consensus timestamp for given event, if event is confirmed.
func (p *Poset) GetConsensusTime(id hash.Event) (inter.Timestamp, error) {
	f := p.store.GetEventConfirmedOn(id)
	if f == 0 {
		return 0, errors.New("event wasn't confirmed/found")
	}
	finfo := p.store.GetFrameInfo(id.Epoch(), f)
	return finfo.CalcConsensusTime(id.Lamport()), nil
}
