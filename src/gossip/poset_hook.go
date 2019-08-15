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
	if hook.engine != nil {
		err := hook.engine.ProcessEvent(e)
		if err != nil { // TODO make it possible to write only on success
			hook.store.DeleteEvent(e.Hash())
			return err
		}
	}
	// set member's last event. we don't care about forks, because this index is used only for emitter
	hook.store.SetLastEvent(e.Creator, e.Hash())

	// track events with no descendants, i.e. heads
	for _, parent := range e.Parents {
		if hook.store.IsHead(parent) {
			hook.store.EraseHead(parent)
		}
	}
	hook.store.AddHead(e.Hash())

	return nil
}

func (hook *StoreAwareEngine) GetVectorIndex() *vector.Index {
	if hook.engine == nil {
		return nil
	}
	return hook.engine.GetVectorIndex()
}

func (hook *StoreAwareEngine) GetGenesisHash() hash.Hash {
	if hook.engine == nil {
		return hash.Hash{}
	}
	return hook.engine.GetGenesisHash()
}

func (hook *StoreAwareEngine) Prepare(e *inter.Event) *inter.Event {
	if hook.engine == nil {
		return e
	}
	return hook.engine.Prepare(e)
}

func (hook *StoreAwareEngine) CurrentSuperFrameN() idx.SuperFrame {
	if hook.engine == nil {
		return 1
	}
	return hook.engine.CurrentSuperFrameN()
}

func (hook *StoreAwareEngine) GetMembers() pos.Members {
	if hook.engine == nil {
		return pos.Members{}
	}
	return hook.engine.GetMembers()
}

func (hook *StoreAwareEngine) Bootstrap(fn inter.ApplyBlockFn) {
	if hook.engine == nil {
		return
	}
	hook.engine.Bootstrap(fn)
}
