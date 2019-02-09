package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// stakeCounter is for PoS balances accumulator.
type stakeCounter struct {
	balances       *state.DB
	alreadyCounted map[common.Address]struct{}
	majority       uint64
	total          uint64
}

func (s *stakeCounter) Count(creator common.Address) {
	if _, ok := s.alreadyCounted[creator]; ok {
		// log.WithField("node", creator.String()).Debug("already counted")
		return // already counted
	}
	if s.HasMajority() {
		// log.WithField("node", creator.String()).Debug("no sense to count further")
		return // no sense to count further
	}
	s.total += s.balances.GetBalance(creator)
	s.alreadyCounted[creator] = struct{}{}
}

func (s *stakeCounter) HasMajority() bool {
	return s.total > s.majority
}

/*
 * Poset's methods:
 */

func (p *Poset) newStakeCounter() *stakeCounter {
	frame := p.frame(p.state.LastFinishedFrameN)
	db, err := state.New(frame.Balances, p.store.balances)
	if err != nil {
		panic(err)
	}
	return &stakeCounter{
		balances:       db,
		alreadyCounted: make(map[common.Address]struct{}),
		total:          0,
		majority:       p.state.TotalCap * 2 / 3,
	}
}
