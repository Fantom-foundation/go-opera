package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// stakeCounter is for PoS balances accumulator.
type stakeCounter struct {
	balances *state.DB
	amount   inter.Stake
	goal     inter.Stake
}

func (s *stakeCounter) Count(node hash.Peer) {
	if s.IsGoalAchieved() {
		return // no sense to count further
	}
	s.amount += s.balances.VoteBalance(node)
}

func (s *stakeCounter) IsGoalAchieved() bool {
	return s.amount >= s.goal
}

/*
 * Poset's methods:
 */

// StakeOf returns last stake balance of peer.
func (p *Poset) StakeOf(addr hash.Peer) inter.Stake {
	f := p.frameFromStore(p.LastFinishedFrameN() + stateGap)
	db := p.store.StateDB(f.Balances)
	return db.VoteBalance(addr)
}

func (p *Poset) newStakeCounter(frame *Frame, goal inter.Stake) *stakeCounter {
	db := p.store.StateDB(frame.Balances)

	return &stakeCounter{
		balances: db,
		amount:   0,
		goal:     goal,
	}
}

// NOTE: deprecated, use members
func (p *Poset) getSuperMajority() inter.Stake {
	return p.TotalCap*2/3 + 1
}

// NOTE: deprecated, use members
func (p *Poset) hasMajority(frame *Frame, roots EventsByPeer) bool {
	stake := p.newStakeCounter(frame, p.getSuperMajority())
	for node := range roots {
		stake.Count(node)
	}
	return stake.IsGoalAchieved()
}
