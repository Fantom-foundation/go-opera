package pos

import (
	"github.com/ethereum/go-ethereum/common"
)

type (
	// Stake amount.
	Stake uint64
)

type (
	// StakeCounterProvider providers stake counter.
	StakeCounterProvider func() *StakeCounter

	// StakeCounter counts stakes.
	StakeCounter struct {
		members Members
		already map[common.Address]struct{}

		quorum Stake
		sum    Stake
	}
)

// NewCounter constructor.
func (mm Members) NewCounter() *StakeCounter {
	return newStakeCounter(mm)
}

func newStakeCounter(mm Members) *StakeCounter {
	return &StakeCounter{
		members: mm,
		quorum:  mm.Quorum(),
		already: make(map[common.Address]struct{}),
		sum:     0,
	}
}

// Count member and return true if it hadn't counted before.
func (s *StakeCounter) Count(node common.Address) bool {
	if _, ok := s.already[node]; ok {
		return false
	}
	s.already[node] = struct{}{}

	s.sum += s.members.StakeOf(node)
	return true
}

// HasQuorum achieved.
func (s *StakeCounter) HasQuorum() bool {
	return s.sum >= s.quorum
}

// Sum of counted stakes.
func (s *StakeCounter) Sum() Stake {
	return s.sum
}
