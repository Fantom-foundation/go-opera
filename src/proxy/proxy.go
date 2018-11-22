package proxy

import (
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

// AppProxy provides an interface for lachesis to communicate
// with the application.
type AppProxy interface {
	SubmitCh() chan []byte
	SubmitInternalCh() chan poset.InternalTransaction
	CommitBlock(block poset.Block) ([]byte, error)
	GetSnapshot(blockIndex int64) ([]byte, error)
	Restore(snapshot []byte) error
}

// LachesisProxy provides an interface for the application to
// submit transactions to the lachesis node.
type LachesisProxy interface {
	CommitCh() chan proto.Commit
	SnapshotRequestCh() chan proto.SnapshotRequest
	RestoreCh() chan proto.RestoreRequest
	SubmitTx(tx []byte) error
}
