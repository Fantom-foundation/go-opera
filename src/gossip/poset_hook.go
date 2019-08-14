package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
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
	}
	return err
}

func (hook *StoreAwareEngine) StakeOf(addr hash.Peer) inter.Stake {
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
