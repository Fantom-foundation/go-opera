package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type (
	downloads struct {
		heights    heights
		hashes     hash.Events
		sync.Mutex // TODO: split to separates for heights and hashes
	}
)

func (n *Node) initDownloads() {
	if n.downloads.heights != nil {
		return
	}
	n.downloads.heights = make(heights)
	n.downloads.hashes = hash.Events{}
}

// lockFreeHeights returns start indexes of height free intervals and reserves their.
func (n *Node) lockFreeHeights(wants heights) heights {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	res := make(heights, len(wants))

	for creator, want := range wants {
		locked := n.downloads.heights[creator]
		if locked.to == 0 {
			locked.to = n.store.GetPeerHeight(creator)
		}
		if want.to <= locked.to {
			continue
		}

		res[creator] = interval{
			from: max(locked.to+1, want.from),
			to:   want.to,
		}
		n.downloads.heights[creator] = want
	}

	return res
}

// unlockFreeHeights known peer height.
func (n *Node) unlockFreeHeights(hh heights) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	for creator, interval := range hh {
		locked := n.downloads.heights[creator]
		if locked.to <= interval.to {
			delete(n.downloads.heights, creator)
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
