package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type downloads struct {
	heights map[string]bool

	sync.Mutex
}

type parentDownloads struct {
	hashes map[hash.Event]bool

	sync.Mutex
}

func initDownloads() downloads {
	return downloads{
		heights: map[string]bool{},
	}
}

func initParentDownloads() parentDownloads {
	return parentDownloads{
		hashes: map[hash.Event]bool{},
	}
}

// Main Downloads

// addToDownloads known peer height
func (n *Node) addToDownloads(data *map[string]uint64) (*map[string]bool, *map[string]uint64) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	toDelete := map[string]bool{}
	toDownload := map[string]uint64{}

	for hex, height := range *data {
		creator := hash.HexToPeer(hex)
		last := n.store.GetPeerHeight(creator)
		for i := last + 1; i <= height; i++ {
			key := hex + string(i)

			if _, ok := n.downloads.heights[key]; ok {
				continue
			}

			n.downloads.heights[key] = true

			toDelete[key] = true

			toDownload[hex] = i
		}
	}

	return &toDelete, &toDownload
}

// deleteFromDownloads known peer height
func (n *Node) deleteFromDownloads(data *map[string]bool) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	for key := range *data {
		delete(n.downloads.heights, key)
	}
}

// Parent Downloads

// addParentToDownloads add parent event hash to downloads
func (n *Node) addParentToDownloads(parents *hash.Events) *map[hash.Event]bool {
	n.parentDownloads.Lock()
	defer n.parentDownloads.Unlock()

	toDownload := map[hash.Event]bool{}

	for p := range *parents {
		// Check parent in store
		if n.store.GetEvent(p) != nil {
			continue
		}

		if _, ok := n.parentDownloads.hashes[p]; ok {
			continue
		}

		n.parentDownloads.hashes[p] = true

		toDownload[p] = true
	}

	return &toDownload
}

// deleteParentFromDownloads delete parent event hash from downloads
func (n *Node) deleteParentFromDownloads(data *map[hash.Event]bool) {
	n.parentDownloads.Lock()
	defer n.parentDownloads.Unlock()

	for key := range *data {
		delete(n.parentDownloads.hashes, key)
	}
}
