package proxy

import (
	"github.com/andrecronje/lachesis/src/poset"
	bproxy "github.com/andrecronje/lachesis/src/proxy/lachesis"
)

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block poset.Block) ([]byte, error)
	GetSnapshot(blockIndex int) ([]byte, error)
	Restore(snapshot []byte) error
}

type LachesisProxy interface {
	CommitCh() chan bproxy.Commit
	SnapshotRequestCh() chan bproxy.SnapshotRequest
	RestoreCh() chan bproxy.RestoreRequest
	SubmitTx(tx []byte) error
}
