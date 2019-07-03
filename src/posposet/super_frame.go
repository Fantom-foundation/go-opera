package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type superFrame struct {
	frames  map[uint64]*Frame
	members map[hash.Peer]uint64
}

func newSuperFrame() *superFrame {
	return &superFrame{
		frames:  make(map[uint64]*Frame),
		members: make(map[hash.Peer]uint64),
	}
}

func (f superFrame) sumOfStakes() uint64 {
	sum := uint64(0)
	for _, v := range f.members {
		sum += v
	}
	return sum
}
