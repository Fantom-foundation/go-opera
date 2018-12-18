package node

import (
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

/*
 * Diff tool interface implementation (tmp)
 */

func (n *Node) GetLastBlockIndex() int64 {
	return n.core.poset.Store.LastBlockIndex()
}

func (n *Node) RoundClothos(i int64) []string {
	return n.core.poset.Store.RoundClothos(i)
}

func (n *Node) GetFrame(i int64) (poset.Frame, error) {
	return n.core.poset.Store.GetFrame(i)
}

/*
 * Node's method candidates
 */

func (n *Node) PushTx(tx []byte) {
	// we do not need coreLock here as n.core.AddTransactions has TransactionPoolLocker
	n.core.AddTransactions([][]byte{tx})
	n.logger.Debugf("PushTx('%s')", tx)
}
