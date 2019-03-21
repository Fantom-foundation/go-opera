package posnode

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/posnode/network"
)

func TestNode(t *testing.T) {
	key, err := crypto.GenerateECDSAKey()
	if err != nil {
		t.Fatal(err)
	}

	// TODO: use New(key, ConsensusMock, network.FakeConnect)
	n := New(key, nil, nil)
	defer n.Shutdown()

	// TODO: use network.FakeListener("")
	listener := network.TcpListener("")
	n.StartService(listener)
	defer n.StopService()

	t.Logf("node listen at %v", listener.Addr())
}
