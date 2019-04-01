package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

/*
 * Event's parents validator:
 */

// parentsValidator checks parent nodes rule.
type parentsValidator struct {
	event *Event
	nodes map[hash.Peer]struct{}
}

func newParentsValidator(e *Event) *parentsValidator {
	return &parentsValidator{
		event: e,
		nodes: make(map[hash.Peer]struct{}, len(e.Parents)),
	}
}

func (v *parentsValidator) IsParentUnique(node hash.Peer) bool {
	if _, ok := v.nodes[node]; ok {
		log.Warnf("Event %s has double refer to node %s, so rejected",
			v.event.Hash().String(),
			node.String())
		return false
	}
	v.nodes[node] = struct{}{}
	return true

}

func (v *parentsValidator) HasSelfParent() bool {
	if _, ok := v.nodes[v.event.Creator]; !ok {
		log.Warnf("Event %s has no refer to self-node %s, so rejected",
			v.event.Hash().String(),
			v.event.Creator.String())
		return false
	}
	return true
}
