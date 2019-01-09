package node

import (
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

/*
 * Diff tool interface implementation (tmp)
 */

// GetLastBlockIndex returns the last block index
func (n *Node) GetLastBlockIndex() int64 {
	return n.core.poset.Store.LastBlockIndex()
}

// RoundClothos returns all clothos in a round
func (n *Node) RoundClothos(i int64) poset.EventHashes {
	return n.core.poset.Store.RoundClothos(i)
}

// GetFrame returns the frame for a given index
func (n *Node) GetFrame(i int64) (poset.Frame, error) {
	return n.core.poset.Store.GetFrame(i)
}

/*
 * Node's method candidates
 */

// PushTx push transactions into the pending pool
func (n *Node) PushTx(tx []byte) error {
	// we do not need coreLock here as n.core.AddTransactions has TransactionPoolLocker
	err := n.core.AddTransactions([][]byte{tx})
	if err != nil {
		n.logger.Errorf("PushTx('%s') %s", tx, err)
	} else {
		n.logger.Debugf("PushTx('%s')", tx)
	}
	return err
}
