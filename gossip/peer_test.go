package gossip

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeerProgressWithEpoch(t *testing.T) {
	// Increment epoch and see if the peer is progressing
	setProgressThreshold(1 * time.Second)
	newPeer := getPeer()
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	time.Sleep(2 * time.Second) //set the threshold to 2 second
	ep2 := PeerProgress{Epoch: 2}
	newPeer.SetProgress(ep2)
	require.True(t, newPeer.IsPeerProgressing(), "Peer is not progressing")
}

func TestPeerNotProgressWithEpoch(t *testing.T) {
	// Don't Increment epoch and check if the peer is not progressing
	setProgressThreshold(1 * time.Second)
	newPeer := getPeer()
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	time.Sleep(2 * time.Second) //set the threshold to 2 second so that the threshold is expired
	ep2 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep2)
	require.False(t, newPeer.IsPeerProgressing(), "Peer is progressing")
}

func TestPeerNotProgressTimeout(t *testing.T) {
	// Don't Increment epoch and check if the peer is not progressing
	setProgressThreshold(1 * time.Second)
	newPeer := getPeer()
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	time.Sleep(2 * time.Second) //set the threshold to 2 second so that the timer expires
	require.False(t, newPeer.IsPeerProgressing(), "Peer is progressing")
}

func TestApplicationProgressMessage(t *testing.T) {
	// send a valid application message and check if the peer is progressing
	setApplicationThreshold(2 * time.Second)
	newPeer := getPeer()
	newPeer.SetApplicationProgress() // simulate receiving of a valid application message
	time.Sleep(1 * time.Second)      //set the threshold to 1 second
	require.True(t, newPeer.IsApplicationProgressing(), "Application is not progressing")
}

func TestApplicationNotProgressingMessage(t *testing.T) {
	// send a valid application message and check if the peer is progressing
	setApplicationThreshold(1 * time.Second)
	newPeer := getPeer()
	newPeer.SetApplicationProgress() // simulate receiving of a valid application message
	time.Sleep(2 * time.Second)      //set the threshold to 2 second so that the threshold timer expires
	require.False(t, newPeer.IsApplicationProgressing(), "Application is progressing")
}

func getPeer() *peer {
	peer := &peer{
		progressTime: time.Now(),
	}
	return peer
}
