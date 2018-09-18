package proxy

import "github.com/andrecronje/lachesis/poset"

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block poset.Block) ([]byte, error)
	GetSnapshot(blockIndex int) ([]byte, error)
	Restore(snapshot []byte) error
}

type LachesisProxy interface {
	CommitCh() chan poset.Block
	SnapshotRequestCh() chan int
	RestoreCh() chan []byte
	SubmitTx(tx []byte) error
}
