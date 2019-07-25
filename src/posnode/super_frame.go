package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type superFrame struct {
	Num idx.SuperFrame

	lasts map[hash.Peer]idx.Event

	sync.RWMutex
}

func (n *Node) initSuperFrame() {
	if n.superFrame.lasts != nil {
		return
	}

	n.superFrame.Num = n.currentSuperFrame()
	n.superFrame.lasts = n.store.GetAllPeerHeight(n.superFrame.Num)
}

func (n *Node) currentSuperFrame() idx.SuperFrame {
	if n.consensus == nil {
		return idx.SuperFrame(0)
	}

	sf := n.consensus.CurrentSuperFrameN()
	if n.superFrame.Num >= sf {
		return n.superFrame.Num
	}

	n.superFrame.Lock()
	defer n.superFrame.Unlock()

	n.superFrame.Num = sf

	// reset all

	n.superFrame.lasts = make(map[hash.Peer]idx.Event)

	return sf
}
