package node

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

/*
 * stuff
 */

func clonePeers(src *peers.Peers) *peers.Peers {
	dst := peers.NewPeers()
	for _, p0 := range src.ToPeerSlice() {
		p1 := *p0
		dst.AddPeer(&p1)
	}
	return dst
}

func fakePeers(n int) *peers.Peers {
	participants := peers.NewPeers()
	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateECDSAKey()
		peer := peers.Peer{
			NetAddr:   fakeAddr(i),
			PubKeyHex: fmt.Sprintf("0x%X", crypto.FromECDSAPub(&key.PublicKey)),
		}
		participants.AddPeer(&peer)
	}
	return participants
}

func fakeAddr(i int) string {
	return fmt.Sprintf("addr%d", i)
}
