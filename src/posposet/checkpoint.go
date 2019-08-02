package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// checkpoint is for persistent storing.
type checkpoint struct {
	SuperFrameN       idx.SuperFrame
	LastBlockN        idx.Block
	TotalCap          inter.Stake
	LastConsensusTime inter.Timestamp
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
	// TODO store & restore genesis object (p.genesis)
	//p.genesis = p.store.GetSuperFrame(0).balances

	// restore current super-frame
	p.loadSuperFrame()
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	return p.Genesis.Hash()
}

// GenesisHash calcs hash of genesis balances.
func genesisHash(balances map[hash.Peer]inter.Stake) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances); err != nil {
		logger.Get().Fatal(err)
	}

	return s.GetSuperFrame(0).Balances
}
