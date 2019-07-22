package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// checkpoint is for persistent storing.
type checkpoint struct {
	SuperFrameN       idx.SuperFrame
	LastBlockN        idx.Block
	TotalCap          inter.Stake
	LastConsensusTime inter.Timestamp
}

// ToWire converts to proto.Message.
func (cp *checkpoint) ToWire() *wire.Checkpoint {
	return &wire.Checkpoint{
		SuperFrameN:       uint64(cp.SuperFrameN),
		LastBlockN:        uint64(cp.LastBlockN),
		TotalCap:          uint64(cp.TotalCap),
		LastConsensusTime: uint64(cp.LastConsensusTime),
	}
}

// WireToCheckpoint converts from wire.
func WireToCheckpoint(w *wire.Checkpoint) *checkpoint {
	if w == nil {
		return nil
	}
	return &checkpoint{
		SuperFrameN:       idx.SuperFrame(w.SuperFrameN),
		LastBlockN:        idx.Block(w.LastBlockN),
		TotalCap:          inter.Stake(w.TotalCap),
		LastConsensusTime: inter.Timestamp(w.LastConsensusTime),
	}
}

/*
 * Poset's methods:
 */

// State saves checkpoint.
func (p *Poset) saveCheckpoint() {
	p.store.SetCheckpoint(p.checkpoint)
}

// Bootstrap restores checkpoint from store.
func (p *Poset) Bootstrap() {
	if p.checkpoint != nil {
		return
	}
	// restore checkpoint
	p.checkpoint = p.store.GetCheckpoint()
	if p.checkpoint == nil {
		p.Fatal("Apply genesis for store first")
	}

	// restore genesis
	p.Genesis = p.store.GetSuperFrame(0).balances

	// restore current super-frame
	p.loadSuperFrame()
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	return p.Genesis
}

// GenesisHash calcs hash of genesis balances.
func genesisHash(balances map[hash.Peer]inter.Stake) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances); err != nil {
		logger.Get().Fatal(err)
	}

	return s.GetSuperFrame(0).balances
}
