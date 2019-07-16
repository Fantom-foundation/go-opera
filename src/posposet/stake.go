package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// StakeOf returns last stake balance of peer.
func (p *Poset) StakeOf(addr hash.Peer) inter.Stake {
	db := p.store.StateDB(p.superFrame.balances)
	return db.VoteBalance(addr)
}
