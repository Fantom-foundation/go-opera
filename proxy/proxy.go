package proxy

import "github.com/andrecronje/lachesis/hashgraph"

type AppProxy interface {
	SubmitCh() chan []byte
	CommitBlock(block hashgraph.Block) ([]byte, error)
	GetSnapshot(blockIndex int) ([]byte, error)
	Restore(snapshot []byte) error
}

type LachesisProxy interface {
	CommitCh() chan hashgraph.Block
	SnapshotRequestCh() chan int
	RestoreCh() chan []byte
	SubmitTx(tx []byte) error
}
