package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
)

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
		nodes: make(map[common.Address]struct{}, len(e.Parents)),
	}
}

func (pi *parentNodesInspector) IsParentUnique(node common.Address) bool {
	if _, ok := pi.nodes[node]; ok {
		log.Warnf("Event %s has double refer to node %s, so rejected",
			pi.event.Hash().String(),
			node.String())
		return false
	}
	pi.nodes[node] = struct{}{}
	return true

}

func (pi *parentNodesInspector) HasSelfParent() bool {
	if _, ok := pi.nodes[pi.event.Creator]; !ok {
		log.Warnf("Event %s has no refer to self-node %s, so rejected",
			pi.event.Hash().String(),
			pi.event.Creator.String())
		return false
	}
	return true
}

/*
 * Event's lamport time inspector:
 */

type parentLamportTimeInspector struct {
	event   *Event
	maxTime uint64
}

func newParentLamportTimeInspector(e *Event) *parentLamportTimeInspector {
	return &parentLamportTimeInspector{
		event:   e,
		maxTime: 0,
	}
}

func (ti *parentLamportTimeInspector) IsGreaterThan(time uint64) bool {
	if ti.event.LamportTime <= time {
		log.Warnf("Event %s has lamport time %d. It isn't next of parents, so rejected",
			ti.event.Hash().String(),
			ti.event.LamportTime)
		return false
	}
	if ti.maxTime < time {
		ti.maxTime = time
	}
	return true
}

func (ti *parentLamportTimeInspector) IsSequential() bool {
	if ti.event.LamportTime != ti.maxTime+1 {
		log.Warnf("Event %s has lamport time %d. It is too far from parents, so rejected",
			ti.event.Hash().String(),
			ti.event.LamportTime)
		return false
	}
	return true
}

func (ti *parentLamportTimeInspector) GetNext() uint64 {
	return ti.maxTime + 1
}
