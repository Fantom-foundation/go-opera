package proxy

import (
	"github.com/andrecronje/lachesis/src/poset"
	"github.com/andrecronje/lachesis/src/proxy/proto"
)

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block poset.Block) ([]byte, error)
	GetSnapshot(blockIndex int) ([]byte, error)
	Restore(snapshot []byte) error
}

type LachesisProxy interface {
	CommitCh() chan proto.Commit
	SnapshotRequestCh() chan proto.SnapshotRequest
	RestoreCh() chan proto.RestoreRequest
	SubmitTx(tx []byte) error
}
