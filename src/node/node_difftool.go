package node

import (
	"github.com/andrecronje/lachesis/src/poset"
)

/*
 * Diff tool interface implementation (tmp)
 */

func (n *Node) GetLastBlockIndex() int {
	return n.core.poset.Store.LastBlockIndex()
}

func (n *Node) RoundWitnesses(i int) []string {
	return n.core.poset.Store.RoundWitnesses(i)
}

func (n *Node) GetFrame(i int) (poset.Frame, error) {
	return n.core.poset.Store.GetFrame(i)
}

/*
 * Node's method candidates
 */

func (n *Node) PushTx(tx []byte) {
	n.coreLock.Lock()
	defer n.coreLock.Unlock()
	n.core.AddTransactions([][]byte{tx})
	n.logger.Debugf("PushTx('%s')", tx)
}
