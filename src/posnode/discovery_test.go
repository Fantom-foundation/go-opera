package posnode

import (
	"net"
	"testing"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/pkg/errors"
)

func TestNodeCheckPeerIsKnown(t *testing.T) {
	tt := []struct {
		name           string
		beforeFunc     func(*Store) error
		shouldDiscover bool
	}{
		{
			name: "id known",
			beforeFunc: func(store *Store) error {
				key, err := crypto.GenerateECDSAKey()
				if err != nil {
					return errors.Wrap(err, "generate ecdsa key")
				}
				pubKey := key.PublicKey
				id := "known"
				netAddr := "8.8.8.8:8083"

				peer := Peer{
					ID:      common.HexToAddress(id),
					PubKey:  &pubKey,
					NetAddr: netAddr,
				}

				if err := store.SetPeer(&peer); err != nil {
					return errors.Wrap(err, "set peer")
				}

				return nil

			},
		},
		{
			name: "id unknown",
			beforeFunc: func(store *Store) error {
				return nil
			},
			shouldDiscover: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := NewMemStore()
			key, err := crypto.GenerateECDSAKey()
			if err != nil {
				t.Fatalf("failed to generate ecdsa key: %v", err)
			}

			n := NewWithName("node001", key, store, nil)
			n.discovery.tasks = make(chan discoveryTask)
			defer close(n.discovery.tasks)

			if err := tc.beforeFunc(store); err != nil {
				t.Errorf("before func failed: %v", err)
			}

			go n.CheckPeerIsKnown("any", common.HexToAddress("known"))

			if tc.shouldDiscover {
				select {
				case <-n.discovery.tasks:
				case <-time.After(time.Second):
					t.Error("expected to create discovery task")
				}
				return
			}

			if len(n.discovery.tasks) > 0 {
				t.Error("unexpected discovery tasks")
			}
		})
	}
}

func TestNodeAskPeerInfo(t *testing.T) {
	tt := []struct {
		name           string
		beforeFunc     func(*Store) error
		shouldDiscover bool
	}{
		{
			name: "id known",
			beforeFunc: func(store *Store) error {
				key, err := crypto.GenerateECDSAKey()
				if err != nil {
					return errors.Wrap(err, "generate ecdsa key")
				}
				pubKey := key.PublicKey
				id := "known"
				netAddr := "8.8.8.8:8083"

				peer := Peer{
					ID:      common.HexToAddress(id),
					PubKey:  &pubKey,
					NetAddr: netAddr,
				}

				if err := store.SetPeer(&peer); err != nil {
					return errors.Wrap(err, "set peer")
				}

				return nil

			},
		},
		{
			name: "id unknown",
			beforeFunc: func(store *Store) error {
				return nil
			},
			shouldDiscover: true,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := NewMemStore()
			key, err := crypto.GenerateECDSAKey()
			if err != nil {
				t.Fatalf("failed to generate ecdsa key: %v", err)
			}

			n := NewWithName("node001", key, store, nil)
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("failed to start listener: %v", err)
			}
			go n.StartService(listener)
			defer n.StopService()

			if err := tc.beforeFunc(store); err != nil {
				t.Errorf("before func failed: %v", err)
			}

			n.AskPeerInfo(listener.Addr().String(), common.HexToAddress("known"))
		})
	}
}
