package internal

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// stakeCounter counts stakes.
type stakeCounter struct {
	members Members
	already map[hash.Peer]struct{}

	quorum inter.Stake
	sum    inter.Stake
}

func (mm Members) NewCounter() *stakeCounter {
	return newStakeCounter(mm)
}

func newStakeCounter(mm Members) *stakeCounter {
	return &stakeCounter{
		members: mm,
		quorum:  mm.Quorum(),
		already: make(map[hash.Peer]struct{}),
		sum:     0,
	}
}

func (s *stakeCounter) Count(node hash.Peer) bool {
	if _, ok := s.already[node]; ok {
		return false
	}
	s.already[node] = struct{}{}

	s.sum += s.members.StakeOf(node)
	return true
}

func (s *stakeCounter) HasQuorum() bool {
	return s.sum >= s.quorum
}

func (s *stakeCounter) Sum() inter.Stake {
	return s.sum
}
