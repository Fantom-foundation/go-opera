package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// StakeOf returns last stake balance of peer.
func (p *Poset) StakeOf(addr hash.Peer) inter.Stake {
	// TODO: implement
	f := p.frameFromStore(0)
	db := p.store.StateDB(f.Balances)
	return db.VoteBalance(addr)
}
