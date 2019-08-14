package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/vector"
)

type StoreAwareEngine struct {
	engine Consensus
	store  *Store
}

// not safe for concurrent use
func (hook *StoreAwareEngine) ProcessEvent(e *inter.Event) error {
	hook.store.SetEvent(e)
	err := hook.engine.ProcessEvent(e)
	if err != nil { // TODO make it possible to write only on success
		hook.store.DeleteEvent(e.Hash())
		return err
	}
	// set member's last event. we don't care about forks, because this index is used only for emitter
	hook.store.SetLastEvent(e.Creator, e.Hash())
	return err
}

func (hook *StoreAwareEngine) GetHeads() hash.Events {
	return hook.engine.GetHeads()
}

func (hook *StoreAwareEngine) GetVectorIndex() *vector.Index {
	return hook.engine.GetVectorIndex()
}

func (hook *StoreAwareEngine) StakeOf(addr hash.Peer) pos.Stake {
	return hook.engine.StakeOf(addr)
}

func (hook *StoreAwareEngine) GetGenesisHash() hash.Hash {
	return hook.engine.GetGenesisHash()
}

func (hook *StoreAwareEngine) Prepare(e *inter.Event) *inter.Event {
	return hook.engine.Prepare(e)
}

func (hook *StoreAwareEngine) CurrentSuperFrameN() idx.SuperFrame {
	return hook.engine.CurrentSuperFrameN()
}

func (hook *StoreAwareEngine) SuperFrameMembers() []hash.Peer {
	return hook.engine.SuperFrameMembers()
}
