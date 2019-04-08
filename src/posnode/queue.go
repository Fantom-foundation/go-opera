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
func (n *Node) addToDownloads(key string) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	n.downloads.heights[key] = true
}

// deleteFromDownloads known peer height
func (n *Node) deleteFromDownloads(key string) {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	delete(n.downloads.heights, key)
}

// checkDownloads known peer height
func (n *Node) checkDownloads(key string) bool {
	n.downloads.Lock()
	defer n.downloads.Unlock()

	_, ok := n.downloads.heights[key]

	return ok
}

// Parent Downloads

// addParentToDownloads add parent event hash to downloads
func (n *Node) addParentToDownloads(key hash.Event) {
	n.parentDownloads.Lock()
	defer n.parentDownloads.Unlock()

	n.parentDownloads.hashes[key] = true
}

// deleteParentFromDownloads delete parent event hash from downloads
func (n *Node) deleteParentFromDownloads(key hash.Event) {
	n.parentDownloads.Lock()
	defer n.parentDownloads.Unlock()

	delete(n.parentDownloads.hashes, key)
}

// checkParentDownloads check parent event hash in downloads
func (n *Node) checkParentDownloads(key hash.Event) bool {
	n.parentDownloads.Lock()
	defer n.parentDownloads.Unlock()

	_, ok := n.parentDownloads.hashes[key]

	return ok
}
