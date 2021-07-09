package concurrent

import (
	"sync"

	"github.com/Fantom-foundation/go-opera/utils/hash"
)

type ValidatorEventsSet struct {
	sync.RWMutex
	hash.ValidatorEventsSet
}

func WrapValidatorEventsSet(v hash.ValidatorEventsSet) *ValidatorEventsSet {
	return &ValidatorEventsSet{
		RWMutex:            sync.RWMutex{},
		ValidatorEventsSet: v,
	}
}
