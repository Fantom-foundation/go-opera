package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// stakes is for PoS balances accumulation
type stakes struct {
	balances common.Hash
	items    map[common.Address]struct{}
	total    uint64
}

func (p *Poset) newStakes(f *Frame) *stakes {
	return &stakes{
		balances: f.Balances,
		items:    make(map[common.Address]struct{}),
		total:    0,
	}
}

func (s *stakes) Count(creator common.Address) {
	if _, ok := s.items[creator]; ok {
		return // already counted
	}
	if s.IsMajority() {
		return // no sense to count further
	}
	balance := uint64(1) // TODO: take it from PoS state
	s.items[creator] = struct{}{}
	s.total += balance
}

func (s *stakes) IsMajority() bool {
	const totalCap = uint64(4) // TODO: take it from genesis
	return s.total > totalCap*2/3
}
