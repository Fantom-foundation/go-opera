package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type superFrame struct {
	Num idx.SuperFrame

	lasts map[hash.Peer]idx.Event

	sync.RWMutex
}

func (n *Node) initLasts() {
	sf := n.currentSuperFrame()
	if n.superFrame.lasts == nil {
		n.loadLasts(sf)
	}

}

func (n *Node) currentSuperFrame() idx.SuperFrame {
	if n.consensus == nil {
		return idx.SuperFrame(0)
	}

	sf := n.consensus.CurrentSuperFrameN()
	if n.superFrame.Num >= sf {
		return n.superFrame.Num
	}

	n.switchToSF(sf)
	return sf
}

func (n *Node) switchToSF(sf idx.SuperFrame) {
	n.superFrame.Lock()
	defer n.superFrame.Unlock()

	n.superFrame.Num = sf

	n.loadLasts(sf)
	n.loadPotentialParents(sf)
}

func (n *Node) loadLasts(sf idx.SuperFrame) {
	n.superFrame.lasts = n.store.GetAllPeerHeight(sf)
}

func (n *Node) setLast(e *inter.Event) {
	n.superFrame.Lock()
	defer n.superFrame.Unlock()

	sf := n.currentSuperFrame()

	if e.SfNum == sf {
		n.superFrame.lasts[e.Creator] = e.Seq
	}
}
