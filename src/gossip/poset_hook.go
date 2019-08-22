package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/vector"
)

type HookedEngine struct {
	engine Consensus

	processEvent func(realEngine Consensus, e *inter.Event) error
}

// not safe for concurrent use
func (hook *HookedEngine) ProcessEvent(e *inter.Event) error {
	return hook.processEvent(hook.engine, e)
}

func (hook *HookedEngine) GetVectorIndex() *vector.Index {
	if hook.engine == nil {
		return nil
	}
	return hook.engine.GetVectorIndex()
}

func (hook *HookedEngine) GetGenesisHash() hash.Hash {
	if hook.engine == nil {
		return hash.Hash{}
	}
	return hook.engine.GetGenesisHash()
}

func (hook *HookedEngine) Prepare(e *inter.Event) *inter.Event {
	if hook.engine == nil {
		return e
	}
	return hook.engine.Prepare(e)
}

func (hook *HookedEngine) CurrentSuperFrameN() idx.SuperFrame {
	if hook.engine == nil {
		return 1
	}
	return hook.engine.CurrentSuperFrameN()
}

func (hook *HookedEngine) LastBlock() (idx.Block, hash.Event) {
	if hook.engine == nil {
		return idx.Block(1), hash.ZeroEvent
	}
	return hook.engine.LastBlock()
}

func (hook *HookedEngine) GetMembers() pos.Members {
	if hook.engine == nil {
		return pos.Members{}
	}
	return hook.engine.GetMembers()
}

func (hook *HookedEngine) Bootstrap(fn inter.ApplyBlockFn) {
	if hook.engine == nil {
		return
	}
	hook.engine.Bootstrap(fn)
}
