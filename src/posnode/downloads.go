package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type (
	downloads struct {
		heights    map[idx.SuperFrame]heights // TODO: clean old SuperFrame
		hashes     hash.Events
		sync.Mutex // TODO: split to separates for heights and hashes
	}
)

func (n *Node) initDownloads() {
	if n.downloads.heights != nil {
		return
	}
	n.downloads.heights = make(map[idx.SuperFrame]heights)
	n.downloads.hashes = hash.Events{}
}

// lockFreeHeights returns start indexes of height free intervals and reserves their.
func (n *Node) lockFreeHeights(sf idx.SuperFrame, wants heights) heights {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	if _, ok := n.downloads.heights[sf]; !ok {
		n.downloads.heights[sf] = make(heights)
	}

	res := make(heights, len(wants))

	for creator, want := range wants {
		locked := n.downloads.heights[sf][creator]
		if locked.to == 0 {
			locked.to = n.store.GetPeerHeight(creator, sf)
		}
		if want.to <= locked.to {
			continue
		}

		res[creator] = interval{
			from: max(locked.to+1, want.from),
			to:   want.to,
		}
		n.downloads.heights[sf][creator] = want
	}

	return res
}

// unlockFreeHeights known peer height.
func (n *Node) unlockFreeHeights(sf idx.SuperFrame, hh heights) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	if _, ok := n.downloads.heights[sf]; !ok {
		return
	}

	for creator, interval := range hh {
		locked := n.downloads.heights[sf][creator]
		if locked.to <= interval.to {
			delete(n.downloads.heights[sf], creator)
		}
	}
}

// lockNotDownloaded returns not downloaded yet from event list and reserves their.
func (n *Node) lockNotDownloaded(events hash.Events) hash.Events {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	res := hash.Events{}

	for e := range events {
		if n.store.GetEvent(e) != nil {
			continue
		}

		if n.downloads.hashes.Contains(e) {
			continue
		}

		n.downloads.hashes.Add(e)
		res.Add(e)
	}

	return res
}

// unlockDownloaded marks events are not downloading.
func (n *Node) unlockDownloaded(events hash.Events) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	for e := range events {
		delete(n.downloads.hashes, e)
	}
}
