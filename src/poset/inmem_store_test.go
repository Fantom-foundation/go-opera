package poset

import (
	"crypto/ecdsa"
	"fmt"
	"reflect"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

type pub struct {
	id      uint64
	privKey *ecdsa.PrivateKey
	pubKey  []byte
	hex     string
}

func initInmemStore(cacheSize int) (*InmemStore, []pub) {
	n := uint64(3)
	var participantPubs []pub
	participants := peers.NewPeers()
	for i := uint64(0); i < n; i++ {
		key, _ := crypto.GenerateECDSAKey()
		pubKey := crypto.FromECDSAPub(&key.PublicKey)
		peer := peers.NewPeer(fmt.Sprintf("0x%X", pubKey), "")
		participantPubs = append(participantPubs,
			pub{i, key, pubKey, peer.PubKeyHex})
		participants.AddPeer(peer)
		participantPubs[len(participantPubs)-1].id = peer.ID
	}

	store := NewInmemStore(participants, cacheSize, nil)
	return store, participantPubs
}

func TestInmemEvents(t *testing.T) {
	cacheSize := 100
	testSize := int64(15)
	store, participants := initInmemStore(cacheSize)

	events := make(map[string][]Event)

	t.Run("Store Events", func(t *testing.T) {
		for _, p := range participants {
			var items []Event
			for k := int64(0); k < testSize; k++ {
				event := NewEvent([][]byte{[]byte(fmt.Sprintf("%s_%d", p.hex[:5], k))},
					nil,
					[]BlockSignature{{Validator: []byte("validator"), Index: 0, Signature: "r|s"}},
					make(EventHashes, 2),
					p.pubKey,
					k, nil)
				_ = event.Hash() // just to set private variables
				items = append(items, event)
				err := store.SetEvent(event)
				if err != nil {
					t.Fatal(err)
				}
			}
			events[p.hex] = items
		}

		for p, evs := range events {
			for k, ev := range evs {
				rev, err := store.GetEventBlock(ev.Hash())
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(ev.Message.Body, rev.Message.Body) {
					t.Fatalf("events[%s][%d] should be %#v, not %#v", p, k, ev, rev)
				}
			}
		}
	})

	t.Run("Check ParticipantEventsCache", func(t *testing.T) {
		skipIndex := int64(-1) // do not skip any indexes
		for _, p := range participants {
			pEvents, err := store.ParticipantEvents(p.hex, skipIndex)
			if err != nil {
				t.Fatal(err)
			}
			if l := int64(len(pEvents)); l != testSize {
				t.Fatalf("%s should have %d Events, not %d", p.hex, testSize, l)
			}

			expectedEvents := events[p.hex][skipIndex+1:]
			for k, e := range expectedEvents {
				if e.Hash() != pEvents[k] {
					t.Fatalf("ParticipantEvents[%s][%d] should be %s, not %s",
						p.hex, k, e.Hash(), pEvents[k])
				}
			}
		}
	})

	t.Run("Add ConsensusEvents", func(t *testing.T) {
		for _, p := range participants {
			evs := events[p.hex]
			for _, ev := range evs {
				if err := store.AddConsensusEvent(ev); err != nil {
					t.Fatal(err)
				}
			}
		}
	})

}

func TestInmemRounds(t *testing.T) {
	store, participants := initInmemStore(10)

	round := NewRoundCreated()
	events := make(map[string]Event)
	for _, p := range participants {
		event := NewEvent([][]byte{},
			nil,
			[]BlockSignature{},
			make(EventHashes, 2),
			p.pubKey,
			0, nil)
		events[p.hex] = event
		round.AddEvent(event.Hash(), true)
	}

	t.Run("Store Round", func(t *testing.T) {
		if err := store.SetRoundCreated(0, *round); err != nil {
			t.Fatal(err)
		}
		storedRound, err := store.GetRoundCreated(0)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(*round, storedRound) {
			t.Fatalf("Round and StoredRound do not match")
		}
	})

	t.Run("Check LastRound", func(t *testing.T) {
		if c := store.LastRound(); c != 0 {
			t.Fatalf("Store LastRound should be 0, not %d", c)
		}
	})

	t.Run("Check clothos", func(t *testing.T) {
		clothos := store.RoundClothos(0)
		expectedClotho := round.Clotho()
		if len(clothos) != len(expectedClotho) {
			t.Fatalf("There should be %d clothos, not %d", len(expectedClotho), len(clothos))
		}
		for _, w := range expectedClotho {
			if !clothos.Contains(w) {
				t.Fatalf("Clotho should contain %s", w)
			}
		}
	})
}

func TestInmemBlocks(t *testing.T) {
	store, participants := initInmemStore(10)

	index := int64(0)
	roundReceived := int64(7)
	transactions := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
		[]byte("tx4"),
		[]byte("tx5"),
	}
	frameHash := []byte("this is the frame hash")
	block := NewBlock(index, roundReceived, frameHash, transactions)

	sig1, err := block.Sign(participants[0].privKey)
	if err != nil {
		t.Fatal(err)
	}

	sig2, err := block.Sign(participants[1].privKey)
	if err != nil {
		t.Fatal(err)
	}

	if err := block.SetSignature(sig1); err != nil {
		t.Fatal(err)
	}
	if err := block.SetSignature(sig2); err != nil {
		t.Fatal(err)
	}

	t.Run("Store Block", func(t *testing.T) {
		if err := store.SetBlock(block); err != nil {
			t.Fatal(err)
		}

		storedBlock, err := store.GetBlock(index)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(storedBlock, block) {
			t.Fatalf("Block and StoredBlock do not match")
		}
	})

	t.Run("Check signatures in stored Block", func(t *testing.T) {
		storedBlock, err := store.GetBlock(index)
		if err != nil {
			t.Fatal(err)
		}

		val1Sig, ok := storedBlock.Signatures[participants[0].hex]
		if !ok {
			t.Fatalf("Validator1 signature not stored in block")
		}
		if val1Sig != sig1.Signature {
			t.Fatal("Validator1 block signatures differ")
		}

		val2Sig, ok := storedBlock.Signatures[participants[1].hex]
		if !ok {
			t.Fatalf("Validator2 signature not stored in block")
		}
		if val2Sig != sig2.Signature {
			t.Fatal("Validator2 block signatures differ")
		}
	})
}
