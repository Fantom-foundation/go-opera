package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// stakes is for PoS balances accumulator.
type stakes struct {
	balances       *state.DB
	alreadyCounted map[common.Address]struct{}
	majority       uint64
	total          uint64
}

func (s *stakes) Count(creator common.Address) {
	if _, ok := s.alreadyCounted[creator]; ok {
		return // already counted
	}
	if s.HasMajority() {
		return // no sense to count further
	}
	s.total += s.balances.GetBalance(creator)
	s.alreadyCounted[creator] = struct{}{}
}

func (s *stakes) HasMajority() bool {
	return s.total > s.majority
}

/*
 * Poset's methods:
 */

func (p *Poset) newStakes(f *Frame) *stakes {
	db, err := state.New(f.Balances, p.store.balances)
	if err != nil {
		panic(err)
	}
	return &stakes{
		balances:       db,
		alreadyCounted: make(map[common.Address]struct{}),
		total:          0,
		majority:       p.state.TotalCap * 2 / 3,
	}
}
