package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// StakeOf returns last stake balance of peer.
func (p *Poset) StakeOf(addr hash.Peer) pos.Stake {
	db := p.store.StateDB(p.checkpoint.Balances)
	return db.VoteBalance(addr)
}
