package internal

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type (
	// StakeCounterProvider providers stake counter.
	StakeCounterProvider func() *StakeCounter

	// StakeCounter counts stakes.
	StakeCounter struct {
		members Members
		already map[hash.Peer]struct{}

		quorum inter.Stake
		sum    inter.Stake
	}
)

func (mm Members) NewCounter() *StakeCounter {
	return newStakeCounter(mm)
}

func newStakeCounter(mm Members) *StakeCounter {
	return &StakeCounter{
		members: mm,
		quorum:  mm.Quorum(),
		already: make(map[hash.Peer]struct{}),
		sum:     0,
	}
}

func (s *StakeCounter) Count(node hash.Peer) bool {
	if _, ok := s.already[node]; ok {
		return false
	}
	s.already[node] = struct{}{}

	s.sum += s.members.StakeOf(node)
	return true
}

func (s *StakeCounter) HasQuorum() bool {
	return s.sum >= s.quorum
}

func (s *StakeCounter) Sum() inter.Stake {
	return s.sum
}
