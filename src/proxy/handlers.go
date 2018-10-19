package proxy

import (
	"github.com/andrecronje/lachesis/src/poset"
)

/*
These types are exported and need to be implemented and used by the calling
application.
*/

type ProxyHandler interface {
	//CommitHandler is called when Lachesis commits a block to the application. It
	//returns the state hash resulting from applying the block's transactions to the
	//state
	CommitHandler(block poset.Block) (stateHash []byte, err error)

	//SnapshotHandler is called by Lachesis to retrieve a snapshot corresponding to a
	//particular block
	SnapshotHandler(blockIndex int) (snapshot []byte, err error)

	//RestoreHandler is called by Lachesis to restore the application to a specific
	//state
	RestoreHandler(snapshot []byte) (stateHash []byte, err error)
}
