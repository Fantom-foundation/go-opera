package proxy

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

/*
These types are exported and need to be implemented and used by the calling
application.
*/

// App provides an interface for the application to set handlers for
// committing, retrieving and restoring state and transactions.
type App interface {
	// CommitHandler is called when Lachesis commits a block to the DAG. It returns
	// the state hash resulting from applying the block's transactions to the state.
	CommitHandler(block inter.Block) (stateHash []byte, err error)

	// SnapshotHandler is called by Lachesis to retrieve a snapshot
	// corresponding to a particular block.
	SnapshotHandler(blockIndex int64) (snapshot []byte, err error)

	// RestoreHandler is called by Lachesis to restore the application
	// to a specific state.
	RestoreHandler(snapshot []byte) (stateHash []byte, err error)
}

// Node is a set of node handlers.
type Node interface {
	GetID() common.Address
	AddInternalTxn(inter.InternalTransaction) (hash.Transaction, error)
	GetInternalTxn(hash.Transaction) (*inter.InternalTransaction, *inter.Event)
}

// Consensus is a set of consensus handlers.
type Consensus interface {
	StakeOf(peer common.Address) pos.Stake
	GetEventBlock(hash.Event) *inter.Block
}
