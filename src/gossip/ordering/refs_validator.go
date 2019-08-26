package ordering

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type refsValidator struct {
	event    *inter.Event
	creators map[common.Address]struct{}
}

func newRefsValidator(e *inter.Event) *refsValidator {
	return &refsValidator{
		event:    e,
		creators: make(map[common.Address]struct{}, len(e.Parents)),
	}
}

func (v *refsValidator) AddUniqueParent(node common.Address) error {
	if _, ok := v.creators[node]; ok {
		return fmt.Errorf("event %s has double refer to node %s",
			v.event.Hash().String(),
			node.String())
	}
	v.creators[node] = struct{}{}
	return nil

}

func (v *refsValidator) CheckSelfParent() error {
	if _, ok := v.creators[v.event.Creator]; !ok {
		return fmt.Errorf("event %s has no refer to self-node %s",
			v.event.Hash().String(),
			v.event.Creator.String())
	}
	return nil
}
