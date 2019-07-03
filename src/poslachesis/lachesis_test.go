package lachesis

import (
	"testing"
	"time"

	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode"
)

func TestRing(t *testing.T) {
	logger.SetTestMode(t)

	ll := LachesisNetworkRing(5)

	time.Sleep(1 * time.Second)

	for _, l := range ll {
		cp := l.consensusStore.GetCheckpoint()
		t.Logf("%s: frame %d, block %d", l.node.Host(), cp.LastFinishedFrameN(), cp.LastBlockN)
		l.Stop()
	}
}

func TestStar(t *testing.T) {
	logger.SetTestMode(t)

	ll := LachesisNetworkStar(5)

	time.Sleep(1 * time.Second)

	for _, l := range ll {
		cp := l.consensusStore.GetCheckpoint()
		t.Logf("%s: frame %d, block %d", l.node.Host(), cp.LastFinishedFrameN(), cp.LastBlockN)
		l.Stop()
	}
}

/*
 * Utils:
 */

// NewForTests makes lachesis node with fake network.
// It does not start any process.
func NewForTests(
	db *bbolt.DB,
	host string,
	key *crypto.PrivateKey,
	conf *Config,
) *Lachesis {
	l := makeLachesis(db, host, key, conf, network.FakeListener, posnode.FakeClient(host))

	l.node.SetName(host)
	l.nodeStore.SetName(host)
	l.consensus.SetName(host)
	l.consensusStore.SetName(host)

	return l
}
