package gossip

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerProgress(t *testing.T) {
	newPeer := getPeer()

	// Increment epoch and see if the peer is progressing
	ep1 := PeerProgress{Epoch: 1}
	newPeer.SetProgress(ep1)
	ep2 := PeerProgress{Epoch: 2}
	newPeer.SetProgress(ep2)
	ep3 := PeerProgress{Epoch: 3}
	newPeer.SetProgress(ep3)
	ep4 := PeerProgress{Epoch: 4}
	newPeer.SetProgress(ep4)
	require.False(t, newPeer.IsPeerNotProgressing(), "Peer is not progressing")

	// Don't progress consecutively for 3 messages and see if the peer is
	// being progress disconnected
	ep5 := PeerProgress{Epoch: 4}
	newPeer.SetProgress(ep5)
	ep6 := PeerProgress{Epoch: 4}
	newPeer.SetProgress(ep6)
	ep7 := PeerProgress{Epoch: 4}
	newPeer.SetProgress(ep7)
	require.True(t, newPeer.IsPeerNotProgressing(), "Peer is progressing")
}

func getPeer() *peer {
	peer := &peer{
		notProgressingCount: 0,
	}
	return peer
}
