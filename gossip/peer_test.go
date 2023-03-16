package gossip

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeerProgressWithEpoch(t *testing.T) {
	// Increment epoch and see if the peer is progressing
	newPeer := getPeer()
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	time.Sleep(1 * time.Second) //set the threshold to 1 second
	ep2 := PeerProgress{Epoch: 2}
	newPeer.SetProgress(ep2)
	require.True(t, newPeer.IsPeerProgressing(), "Peer is not progressing")
}

func TestPeerNotProgressWithEpoch(t *testing.T) {
	// Don't Increment epoch and check if the peer is not progressing
	newPeer := getPeer()
	newPeer.setProgressThreshold(1 * time.Second)
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	time.Sleep(1 * time.Second) //set the threshold to 1 second
	ep2 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep2)
	require.False(t, newPeer.IsPeerProgressing(), "Peer is progressing")
}

func TestPeerProgressWithApplicationMessage(t *testing.T) {
	// Don't Increment epoch, but add a valid application message and check if the peer is progressing
	newPeer := getPeer()
	newPeer.setProgressThreshold(1 * time.Second)
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	time.Sleep(1 * time.Second) //set the threshold to 1 second
	ep2 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep2)
	newPeer.setPeerAsProgressing() // simulate receiving of a valid application message
	require.True(t, newPeer.IsPeerProgressing(), "Peer is not progressing")
}

func getPeer() *peer {
	peer := &peer{
		lastProgressTime: time.Now(),
	}
	return peer
}
