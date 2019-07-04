package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

// stateGap is a frame-delay to apply new balance.
// TODO: move this magic number to mainnet config
const stateGap = 3

// stakeCounter is for PoS balances accumulator.
type stakeCounter struct {
	balances *state.DB
	amount   uint64
	goal     uint64
}

func (s *stakeCounter) Count(node hash.Peer) {
	if s.IsGoalAchieved() {
		return // no sense to count further
	}
	s.amount += s.balances.VoteBalance(node)
}

func (s *stakeCounter) IsGoalAchieved() bool {
	return s.amount > s.goal
}

/*
 * Poset's methods:
 */

// StakeOf returns last stake balance of peer.
func (p *Poset) StakeOf(addr hash.Peer) uint64 {
	f := p.frameFromStore(p.LastFinishedFrameN() + stateGap)
	db := p.store.StateDB(f.Balances)
	return db.VoteBalance(addr)
}

func (p *Poset) newStakeCounter(frame *Frame, goal uint64) *stakeCounter {
	db := p.store.StateDB(frame.Balances)

	return &stakeCounter{
		balances: db,
		amount:   0,
		goal:     goal,
	}
}

func (p *Poset) hasMajority(frame *Frame, roots EventsByPeer) bool {
	// TODO: set to p.members.TotalStake() * 2 / 3 + 1 with TestPoset{Simple,Far}Root() tests.
	majority := p.members.TotalStake() * 2 / 3 // + 1
	stake := p.newStakeCounter(frame, majority)
	for node := range roots {
		stake.Count(node)
	}
	return stake.IsGoalAchieved()
}
