package internal

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// stakeCounter counts stakes.
type stakeCounter struct {
	members Members
	already map[hash.Peer]struct{}

	majority inter.Stake
	sum      inter.Stake
	dirtySum inter.Stake
}

func (mm Members) NewCounter() *stakeCounter {
	return newStakeCounter(mm)
}

func newStakeCounter(mm Members) *stakeCounter {
	return &stakeCounter{
		members:  mm,
		majority: mm.Majority(),
		already:  make(map[hash.Peer]struct{}),
	}
}

func (s *stakeCounter) Count(node hash.Peer) bool {
	stake := s.members.StakeOf(node)
	s.dirtySum += stake

	if _, ok := s.already[node]; ok {
		return false
	}
	s.already[node] = struct{}{}

	s.sum += stake
	return true
}

func (s *stakeCounter) HasMajority() bool {
	return s.sum >= s.majority
}
