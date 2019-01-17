package fakenet_test

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/net/fakenet"
)

func TestNetworkConnRefused(t *testing.T) {
	address := "localhost:1234"

	network := fakenet.NewNetwork()
	_, err := network.CreateNetConn("tcp", address, time.Second)
	if err == nil {
		t.Fatal("error should not be null")
	}

	listener, err := network.CreateListener("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			_, err := listener.Accept()
			if err != nil {
				return
			}
		}
	}()

	conn, err := network.CreateNetConn("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	listener.Close()

	_, err = network.CreateNetConn("tcp", address, time.Second)
	if err == nil {
		t.Fatal("error should not be null")
	}
}

func TestNetworkAddrAlreadyInUse(t *testing.T) {
	address := "localhost:1234"

	network := fakenet.NewNetwork()
	listener, err := network.CreateListener("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	_, err = network.CreateListener("tcp", address)
	if err != fakenet.ErrAddressAlreadyInUse {
		t.Fatal(err)
	}
}
