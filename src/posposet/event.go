package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/common"
)

// Event is a poset event.
type Event struct {
	Creator common.Address
	Parents EventHashes

	hash    EventHash            // cache for .Hash()
	parents map[EventHash]*Event // temporary cache for internal purpose
}

// Hash calcs hash of event.
func (e *Event) Hash() EventHash {
	if e.hash.IsZero() {
		e.hash = EventHashOf(e)
	}
	return e.hash
}

// String returns string representation.
func (e *Event) String() string {
	return fmt.Sprintf("Event{%s, %s}", e.Hash().ShortString(), e.Parents.ShortString())
}

/*
 * Event's parent inspector:
 */

// parentNodesInspector checks parent nodes rule.
type parentNodesInspector struct {
	event *Event
	nodes map[common.Address]struct{}
}

func newParentNodesInspector(e *Event) *parentNodesInspector {
	return &parentNodesInspector{
		event: e,
		nodes: make(map[common.Address]struct{}, e.Parents.Len()),
	}
}

func (pi *parentNodesInspector) IsParentUnique(node common.Address) bool {
	if _, ok := pi.nodes[node]; ok {
		log.Warnf("Event %s has double refer to node %s, so rejected",
			pi.event.Hash().ShortString(),
			node.String())
		return false
	}
	pi.nodes[node] = struct{}{}
	return true

}

func (pi *parentNodesInspector) HasSelfParent() bool {
	if _, ok := pi.nodes[pi.event.Creator]; !ok {
		log.Warnf("Event %s has no refer to self-node %s, so rejected",
			pi.event.Hash().ShortString(),
			pi.event.Creator.String())
		return false
	}
	return true
}
