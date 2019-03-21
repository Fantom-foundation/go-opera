package posnode

import (
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
)

func TestNode(t *testing.T) {
	key, err := crypto.GenerateECDSAKey()
	if err != nil {
		t.Fatal(err)
	}

	n := New(key, nil)
	defer n.Shutdown()

	t.Log("Hello World!")
}
