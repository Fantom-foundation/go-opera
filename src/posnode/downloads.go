package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type (
	downloads struct {
		heights    map[hash.Peer]uint64
		hashes     hash.Events
		sync.Mutex // TODO: split to separates for heights and hashes
	}

	interval struct {
		from, to uint64
	}
)

func (n *Node) initDownloads() {
	if n.downloads.heights != nil {
		return
	}
	n.downloads.heights = make(map[hash.Peer]uint64)
	n.downloads.hashes = hash.Events{}
}

// lockFreeHeights returns start indexes of height free intervals and reserves their.
func (n *Node) lockFreeHeights(want map[hash.Peer]uint64) map[hash.Peer]interval {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	res := make(map[hash.Peer]interval, len(want))

	for creator, height := range want {
		locked := n.downloads.heights[creator]
		if locked == 0 {
			locked = n.store.GetPeerHeight(creator)
		}
		if height <= locked {
			continue
		}

		res[creator] = interval{locked + 1, height}
		n.downloads.heights[creator] = height
	}

	return res
}

// unlockFreeHeights known peer height.
func (n *Node) unlockFreeHeights(hh map[hash.Peer]interval) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	for creator, interval := range hh {
		locked := n.downloads.heights[creator]
		if locked <= interval.to {
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
