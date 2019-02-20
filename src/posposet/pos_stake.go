package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// stakeCounter is for PoS balances accumulator.
type stakeCounter struct {
	balances *state.DB
	amount   uint64
	goal     uint64
}

func (s *stakeCounter) Count(node common.Address) {
	if s.IsGoalAchieved() {
		return // no sense to count further
	}
	s.amount += s.balances.GetBalance(node)
}

func (s *stakeCounter) IsGoalAchieved() bool {
	return s.amount > s.goal
}

/*
 * Poset's methods:
 */

func (p *Poset) newStakeCounter(goal uint64) *stakeCounter {
	frame := p.frame(p.state.LastFinishedFrameN, false)
	db, err := state.New(frame.Balances, p.store.balances)
	if err != nil {
		panic(err)
	}
	return &stakeCounter{
		balances: db,
		amount:   0,
		goal:     goal,
	}
}

func (p *Poset) hasMajority(roots eventsByNode) bool {
	stake := p.newStakeCounter(
		p.state.TotalCap * 2 / 3)
	for node := range roots {
		stake.Count(node)
	}
	return stake.IsGoalAchieved()
}

func (p *Poset) hasTrust(roots eventsByNode) bool {
	stake := p.newStakeCounter(
		p.state.TotalCap * 1 / 3)
	for node := range roots {
		stake.Count(node)
	}
	return stake.IsGoalAchieved()
}
