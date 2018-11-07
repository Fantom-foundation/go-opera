package node

/*
 * Diff tool interface implementation
 */

func (n *Node) GetCore() *Core {
	return n.core
}

/*
 * Node's method candidates
 */

func (n *Node) GetLastBlockIndex() int {
	return n.core.GetLastBlockIndex()
}

func (n *Node) PushTx(tx []byte) {
	n.coreLock.Lock()
	defer n.coreLock.Unlock()
	n.core.AddTransactions([][]byte{tx})
	n.logger.Debugf("PushTx('%s')", tx)
}
