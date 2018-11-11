package node

/*
 * Diff tool interface implementation (tmp)
 */

func (n *Node) GetLastBlockIndex() int {
	return n.core.poset.Store.LastBlockIndex()
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
