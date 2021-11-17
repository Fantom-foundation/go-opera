package concurrent

import (
	"sync"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type ValidatorEpochsSet struct {
	sync.RWMutex
	Val map[idx.ValidatorID]idx.Epoch
}

func WrapValidatorEpochsSet(v map[idx.ValidatorID]idx.Epoch) *ValidatorEpochsSet {
	return &ValidatorEpochsSet{
		RWMutex: sync.RWMutex{},
		Val:     v,
	}
}
