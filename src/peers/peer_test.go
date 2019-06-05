package peers

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"reflect"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

func TestJSONPeers(t *testing.T) {
	// Create a test dir
	dir, err := ioutil.TempDir("", "lachesis")
	if err != nil {
		t.Fatalf("err: %v ", err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()

	// Create the store
	store := NewJSONPeers(dir)

	// Try a read, should get nothing
	peers, err := store.Peers()
	if err == nil {
		t.Fatalf("store.Peers() should generate an error")
	}
	if peers != nil {
		t.Fatalf("peers: %v", peers)
	}

	keys := map[string]*crypto.PrivateKey{}
	newPeers := NewPeers()
	for i := 0; i < 3; i++ {
		key, _ := crypto.GenerateKey()
		peer := Peer{
			NetAddr:   fmt.Sprintf("addr%d", i),
			PubKeyHex: fmt.Sprintf("0x%X", key.Public().Bytes()),
		}
		newPeers.AddPeer(&peer)
		keys[peer.NetAddr] = key
	}

	newPeersSlice := newPeers.ToPeerSlice()

	if err := store.SetPeers(newPeersSlice); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Try a read, should find 3 peers
	peers, err = store.Peers()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if peers.Len() != 3 {
		t.Fatalf("peers: %v", peers)
	}

	peersSlice := peers.ToPeerSlice()

	for i := 0; i < 3; i++ {
		if peersSlice[i].NetAddr != newPeersSlice[i].NetAddr {
			t.Fatalf("peers[%d] NetAddr should be %s, not %s", i,
				newPeersSlice[i].NetAddr, peersSlice[i].NetAddr)
		}
		if peersSlice[i].PubKeyHex != newPeersSlice[i].PubKeyHex {
			t.Fatalf("peers[%d] PubKeyHex should be %s, not %s", i,
				newPeersSlice[i].PubKeyHex, peersSlice[i].PubKeyHex)
		}
		pubKeyBytes, err := peersSlice[i].PubKeyBytes()
		if err != nil {
			t.Fatal(err)
		}
		pubKey := crypto.BytesToPubKey(pubKeyBytes)
		if !reflect.DeepEqual(pubKey, keys[peersSlice[i].NetAddr].Public()) {
			t.Fatalf("peers[%d] PublicKey not parsed correctly", i)
		}
	}
}
