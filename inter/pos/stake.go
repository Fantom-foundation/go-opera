package pos

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
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
		validators Validators
		already    map[common.Address]struct{}

		quorum Stake
		sum    Stake
	}
)

var (
	balanceToStakeRatio = new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil) // 10^12
)

func BalanceToStake(balance *big.Int) Stake {
	stakeBig := new(big.Int).Div(balance, balanceToStakeRatio)
	if stakeBig.Sign() < 0 || stakeBig.BitLen() >= 64 {
		log.Error("Too big stake amount!", "balance", balance.String())
		return 0
	}
	return Stake(stakeBig.Uint64())
}

// Warning: for tests only!
func StakeToBalance(stake Stake) *big.Int {
	return new(big.Int).Mul(big.NewInt(int64(stake)), balanceToStakeRatio)
}

// NewCounter constructor.
func (vv Validators) NewCounter() *StakeCounter {
	return newStakeCounter(vv)
}

func newStakeCounter(vv Validators) *StakeCounter {
	return &StakeCounter{
		validators: vv,
		quorum:     vv.Quorum(),
		already:    make(map[common.Address]struct{}),
		sum:        0,
	}
}

// Count validator and return true if it hadn't counted before.
func (s *StakeCounter) Count(node common.Address) bool {
	if _, ok := s.already[node]; ok {
		return false
	}
	s.already[node] = struct{}{}

	s.sum += s.validators.StakeOf(node)
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
