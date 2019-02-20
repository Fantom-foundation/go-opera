package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// stakeCounter is for PoS balances accumulator.
type stakeCounter struct {
	balances       *state.DB
	alreadyCounted map[common.Address]struct{}
	amount         uint64
	majoritySum    uint64
	trustSum       uint64
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
	s.amount += s.balances.GetBalance(creator)
	s.alreadyCounted[creator] = struct{}{}
}

func (s *stakeCounter) HasMajority() bool {
	return s.amount > s.majoritySum
}

func (s *stakeCounter) HasTrust() bool {
	return s.amount > s.trustSum
}

/*
 * Poset's methods:
 */

func (p *Poset) newStakeCounter() *stakeCounter {
	frame := p.frame(p.state.LastFinishedFrameN, false)
	db, err := state.New(frame.Balances, p.store.balances)
	if err != nil {
		panic(err)
	}
	return &stakeCounter{
		balances:       db,
		alreadyCounted: make(map[common.Address]struct{}),
		amount:         0,
		majoritySum:    p.state.TotalCap * 2 / 3,
		trustSum:       p.state.TotalCap * 1 / 3,
	}
}

func (p *Poset) hasMajority(roots eventsByNode) bool {
	stake := p.newStakeCounter()
	for node := range roots {
		stake.Count(node)
	}
	return stake.HasMajority()
}
