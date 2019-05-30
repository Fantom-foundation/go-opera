package proxy

//go:generate protoc --go_out=plugins=grpc:. ./internal/app.proto
//go:generate protoc --go_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/golang/protobuf/ptypes/empty:. ./internal/ctrl.proto
// Install before go generate:
//  wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
//  unzip protoc-3.6.1-linux-x86_64.zip -x readme.txt -d /usr/local/
//  go get -u github.com/golang/protobuf/protoc-gen-go

//go:generate mockgen -package=proxy -source=handlers.go -destination=mock_test.go App,Node,Consensus

/* Terms:

App  <===> (LachesisProxy ==grpc/inmem==> AppProxy ) <===> LachesisNode
Ctrl	 <===> (NodeProxy     ==grpc/inmem==> CtrlProxy) <===> LachesisNode

*/

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
	"github.com/Fantom-foundation/go-lachesis/src/proxy/proto"
)

const (
	connectTimeout = 3 * time.Second
)

// AppProxy is an interface for lachesis to communicate with the application.
type AppProxy interface {
	// SubmitCh returns the channel of app transactions.
	SubmitCh() chan []byte
	// SubmitInternalCh returns the channel of stake transactions.
	SubmitInternalCh() chan inter.InternalTransaction
	CommitBlock(block poset.Block) ([]byte, error)
	GetSnapshot(blockIndex int64) ([]byte, error)
	Restore(snapshot []byte) error
	Close()
}

// LachesisProxy is an interface for application to communicate with the lachesis.
type LachesisProxy interface {
	CommitCh() chan proto.Commit
	SnapshotRequestCh() chan proto.SnapshotRequest
	RestoreCh() chan proto.RestoreRequest
	SubmitTx(tx []byte) error
	Close()
}

// NodeProxy is an interface for remote node controlling.
type NodeProxy interface {
	// GetSelfID returns node id.
	GetSelfID() (hash.Peer, error)
	// StakeOf returns stake balance of peer.
	StakeOf(hash.Peer) (uint64, error)
	// SendTo makes stake transfer transaction.
	SendTo(receiver hash.Peer, index, amount, until uint64) (hash.Transaction, error)
	// GetTransaction returns information about transaction.
	GetTransaction(hash.Transaction) (*inter.InternalTransaction, error)
	// Close stops proxy.
	Close()
}

// CtrlProxy is a control proxy.
type CtrlProxy interface {
	// Set ...
	Set()
	// Close stops proxy.
	Close()
}
