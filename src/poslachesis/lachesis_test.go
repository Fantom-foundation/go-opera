package lachesis

import (
	"testing"
	"time"

	"github.com/dgraph-io/badger"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/network"
	"github.com/Fantom-foundation/go-lachesis/src/posnode"
)

func TestRing(t *testing.T) {
	ll := LachesisNetworkRing(5)

	time.Sleep(1 * time.Second)

	for _, l := range ll {
		st := l.consensusStore.GetState()
		t.Logf("%s: frame %d, block %d", l.node.Host(), st.LastFinishedFrameN, st.LastBlockN)
		l.Stop()
	}
}

func TestStar(t *testing.T) {
	ll := LachesisNetworkStar(5)

	time.Sleep(1 * time.Second)

	for _, l := range ll {
		st := l.consensusStore.GetState()
		t.Logf("%s: frame %d, block %d", l.node.Host(), st.LastFinishedFrameN, st.LastBlockN)
		l.Stop()
	}
}

/*
 * Utils:
 */

// NewForTests makes lachesis node with fake network.
// It does not start any process.
func NewForTests(
	db *badger.DB,
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
