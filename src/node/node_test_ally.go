// +build test

package node

// This funcs use only for test purposes. Don't use it for debug or product build.

// SubmitCh func for test
func (n *Node) SubmitCh(tx []byte) error {
	n.proxy.SubmitCh() <- []byte(tx)
	return nil
}

// GetState func for test
func (n *Node) GetState() state {
	return n.getState()
}

// GetFirstConsensusRound func for test
func (n *Node) GetFirstConsensusRound() *int64 {
	return n.core.poset.FirstConsensusRound
}