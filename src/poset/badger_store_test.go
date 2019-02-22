package poset

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/Fantom-foundation/go-lachesis/src/peers"
)

func initBadgerStore(cacheSize int, t *testing.T) (*BadgerStore, []pub) {
	n := 3
	var participantPubs []pub
	participants := peers.NewPeers()
	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateECDSAKey()
		pubKey := crypto.FromECDSAPub(&key.PublicKey)
		peer := peers.NewPeer(fmt.Sprintf("0x%X", pubKey), "")
		participants.AddPeer(peer)
		participantPubs = append(participantPubs,
			pub{peer.ID, key, pubKey, peer.PubKeyHex})
	}

	if err := os.RemoveAll("test_data"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("test_data", os.ModeDir|0777); err != nil {
		t.Fatal(err)
	}
	dir, err := ioutil.TempDir("test_data", "badger")
	if err != nil {
		t.Fatal(err)
	}

	store, err := NewBadgerStore(participants, cacheSize, dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	return store, participantPubs
}

func removeBadgerStore(store *BadgerStore, t *testing.T) {
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(store.path); err != nil {
		t.Fatal(err)
	}
}

func createTestDB(dir string, t *testing.T) *BadgerStore {
	participants := peers.NewPeersFromSlice([]*peers.Peer{
		peers.NewPeer("0xAA", ""),
		peers.NewPeer("0xBB", ""),
		peers.NewPeer("0xCC", ""),
	})

	cacheSize := 100

	store, err := NewBadgerStore(participants, cacheSize, dir, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return store
}

func TestNewBadgerStore(t *testing.T) {
	if err := os.RemoveAll("test_data"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("test_data", os.ModeDir|0777); err != nil {
		t.Fatal(err)
	}

	dbPath := "test_data/badger"
	store := createTestDB(dbPath, t)
	defer func() {
		if err := os.RemoveAll(store.path); err != nil {
			t.Fatal(err)
		}
	}()

	if store.path != dbPath {
		t.Fatalf("unexpected path %q", store.path)
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("err: %s", err)
	}

	// check roots
	inmemRoots := store.inmemStore.rootsByParticipant

	if len(inmemRoots) != 3 {
		t.Fatalf("DB root should have 3 items, not %d", len(inmemRoots))
	}

	for participant, root := range inmemRoots {
		dbRoot, err := store.dbGetRoot(participant)
		if err != nil {
			t.Fatalf("Error retrieving DB root for participant %s: %s", participant, err)
		}
		if !dbRoot.Equals(&root) {
			t.Fatalf("%s DB root should be %#v, not %#v", participant, root, dbRoot)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestLoadBadgerStore(t *testing.T) {
	if err := os.RemoveAll("test_data"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("test_data", os.ModeDir|0777); err != nil {
		t.Fatal(err)
	}
	dbPath := "test_data/badger"

	// Create the test db
	tempStore := createTestDB(dbPath, t)
	defer func() {
		if err := os.RemoveAll(tempStore.path); err != nil {
			t.Fatal(err)
		}
	}()
	if err := tempStore.Close(); err != nil {
		t.Fatal(err)
	}

	badgerStore, err := LoadBadgerStore(cacheSize, tempStore.path)
	if err != nil {
		t.Fatal(err)
	}

	dbParticipants, err := badgerStore.dbGetParticipants()
	if err != nil {
		t.Fatal(err)
	}

	if badgerStore.participants.Len() != 3 {
		t.Fatalf("store.participants  length should be %d items, not %d", 3, badgerStore.participants.Len())
	}

	if badgerStore.participants.Len() != dbParticipants.Len() {
		t.Fatalf("store.participants should contain %d items, not %d",
			dbParticipants.Len(),
			badgerStore.participants.Len())
	}

	dbParticipants.RLock()
	defer dbParticipants.RUnlock()
	for dbP, dbPeer := range dbParticipants.ByPubKey {
		peer, ok := badgerStore.participants.ReadByPubKey(dbP)
		if !ok {
			t.Fatalf("BadgerStore participants does not contains %s", dbP)
		}
		if peer.ID != dbPeer.ID {
			t.Fatalf("participant %s ID should be %d, not %d", dbP, dbPeer.ID, peer.ID)
		}
	}

}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// Call DB methods directly

func TestDBEventMethods(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	testSize := int64(100)
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	// insert events in db directly
	events := make(map[string][]Event)
	topologicalIndex := int64(0)
	var topologicalEvents []Event
	for _, p := range participants {
		var items []Event
		for k := int64(0); k < testSize; k++ {
			event := NewEvent(
				[][]byte{[]byte(fmt.Sprintf("%s_%d", p.hex[:5], k))},
				[]InternalTransaction{},
				[]BlockSignature{{Validator: []byte("validator"), Index: 0, Signature: "r|s"}},
				make(EventHashes, 2),
				p.pubKey,
				k, nil)
			if err := event.Sign(p.privKey); err != nil {
				t.Fatal(err)
			}
			event.Message.TopologicalIndex = topologicalIndex
			topologicalIndex++
			topologicalEvents = append(topologicalEvents, event)

			items = append(items, event)
			err := store.dbSetEvents([]Event{event})
			if err != nil {
				t.Fatal(err)
			}
		}
		events[p.hex] = items
	}

	// check events where correctly inserted and can be retrieved
	for p, evs := range events {
		for k, ev := range evs {
			rev, err := store.dbGetEventBlock(ev.Hash())
			if err != nil {
				t.Fatal(err)
			}
			if !ev.Message.Body.Equals(rev.Message.Body) {
				t.Fatalf("events[%s][%d].Body should be %#v, not %#v", p, k, ev.Message.Body, rev.Message.Body)
			}
			if !reflect.DeepEqual(ev.Message.Signature, rev.Message.Signature) {
				t.Fatalf("events[%s][%d].Signature should be %#v, not %#v", p, k, ev.Message.Signature, rev.Message.Signature)
			}
			if ver, err := rev.Verify(); err != nil && !ver {
				t.Fatalf("failed to verify signature. err: %s", err)
			}
		}
	}

	// check topological order of events was correctly created
	dbTopologicalEvents, err := store.TopologicalEvents()
	if err != nil {
		t.Fatal(err)
	}
	if len(dbTopologicalEvents) != len(topologicalEvents) {
		t.Fatalf("Length of dbTopologicalEvents should be %d, not %d",
			len(topologicalEvents), len(dbTopologicalEvents))
	}
	for i, dte := range dbTopologicalEvents {
		te := topologicalEvents[i]

		if dte.Hash() != te.Hash() {
			t.Fatalf("dbTopologicalEvents[%d].Hex should be %s, not %s", i,
				te.Hash(),
				dte.Hash())
		}
		if !te.Message.Body.Equals(dte.Message.Body) {
			t.Fatalf("dbTopologicalEvents[%d].Body should be %#v, not %#v", i,
				te.Message.Body,
				dte.Message.Body)
		}
		if !reflect.DeepEqual(te.Message.Signature, dte.Message.Signature) {
			t.Fatalf("dbTopologicalEvents[%d].Signature should be %#v, not %#v", i,
				te.Message.Signature,
				dte.Message.Signature)
		}

		if ver, err := dte.Verify(); err != nil && !ver {
			t.Fatalf("failed to verify signature. err: %s", err)
		}
	}

	// check that participant events where correctly added
	skipIndex := int64(-1) // do not skip any indexes
	for _, p := range participants {
		pEvents, err := store.dbParticipantEvents(p.hex, skipIndex)
		if err != nil {
			t.Fatal(err)
		}
		if l := int64(len(pEvents)); l != testSize {
			t.Fatalf("%s should have %d events, not %d", p.hex, testSize, l)
		}

		expectedEvents := events[p.hex][skipIndex+1:]
		for k, e := range expectedEvents {
			if e.Hash() != pEvents[k] {
				t.Fatalf("ParticipantEvents[%s][%d] should be %s, not %s",
					p.hex, k, e.Hash(), pEvents[k])
			}
		}
	}
}

func TestDBRoundMethods(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	round := NewRoundCreated()
	events := make(map[string]Event)
	for _, p := range participants {
		event := NewEvent([][]byte{},
			[]InternalTransaction{},
			[]BlockSignature{},
			make(EventHashes, 2),
			p.pubKey,
			0, nil)
		events[p.hex] = event
		round.AddEvent(event.Hash(), true)
	}

	if err := store.dbSetRoundCreated(0, *round); err != nil {
		t.Fatal(err)
	}

	storedRound, err := store.dbGetRoundCreated(0)
	if err != nil {
		t.Fatal(err)
	}

	if !round.Equals(&storedRound) {
		t.Fatalf("Round and StoredRound do not match")
	}

	clothos := store.RoundClothos(0)
	expectedClothos := round.Clotho()
	if len(clothos) != len(expectedClothos) {
		t.Fatalf("There should be %d clothos, not %d", len(expectedClothos), len(clothos))
	}
	for _, w := range expectedClothos {
		if !clothos.Contains(w) {
			t.Fatalf("Clothos should contain %s", w)
		}
	}
}

func TestDBParticipantMethods(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, _ := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	if err := store.dbSetParticipants(store.participants); err != nil {
		t.Fatal(err)
	}

	participantsFromDB, err := store.dbGetParticipants()
	if err != nil {
		t.Fatal(err)
	}

	store.participants.RLock()
	defer store.participants.RUnlock()
	for p, peer := range store.participants.ByPubKey {
		dbPeer, ok := participantsFromDB.ReadByPubKey(p)
		if !ok {
			t.Fatalf("DB does not contain participant %s", p)
		}
		if peer.ID != dbPeer.ID {
			t.Fatalf("DB participant %s should have ID %d, not %d", p, peer.ID, dbPeer.ID)
		}
	}
}

func TestDBBlockMethods(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	index := int64(0)
	roundReceived := int64(5)
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
		if err := store.dbSetBlock(block); err != nil {
			t.Fatal(err)
		}

		storedBlock, err := store.dbGetBlock(index)
		if err != nil {
			t.Fatal(err)
		}

		if !storedBlock.Equals(&block) {
			t.Fatalf("Block and StoredBlock do not match")
		}
	})

	t.Run("Check signatures in stored Block", func(t *testing.T) {
		storedBlock, err := store.dbGetBlock(index)
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

func TestDBFrameMethods(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	events := make([]*EventMessage, len(participants))
	roots := make([]*Root, len(participants))
	for id, p := range participants {
		event := NewEvent(
			[][]byte{[]byte(fmt.Sprintf("%s_%d", p.hex[:5], 0))},
			[]InternalTransaction{},
			[]BlockSignature{{Validator: []byte("validator"), Index: 0, Signature: "r|s"}},
			make(EventHashes, 2),
			p.pubKey,
			0, nil)
		if err := event.Sign(p.privKey); err != nil {
			t.Fatal(err)
		}
		events[id] = event.Message

		root := NewBaseRoot(uint64(id))
		roots[id] = &root
	}
	frame := Frame{
		Round:  1,
		Events: events,
		Roots:  roots,
	}

	t.Run("Store Frame", func(t *testing.T) {
		if err := store.dbSetFrame(frame); err != nil {
			t.Fatal(err)
		}

		storedFrame, err := store.dbGetFrame(frame.Round)
		if err != nil {
			t.Fatal(err)
		}

		if !storedFrame.Equals(&frame) {
			t.Fatalf("Frame and StoredFrame do not match")
		}
	})
}

// ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
// Check that the wrapper methods work
// These methods use the inmemStore as a cache on top of the DB

func TestBadgerEvents(t *testing.T) {
	// Insert more events than can fit in cache to test retrieving from db.
	cacheSize := 10
	testSize := int64(100)
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	// insert event
	events := make(map[string][]Event)
	for _, p := range participants {
		var items []Event
		for k := int64(0); k < testSize; k++ {
			event := NewEvent(
				[][]byte{[]byte(fmt.Sprintf("%s_%d", p.hex[:5], k))},
				[]InternalTransaction{},
				[]BlockSignature{{Validator: []byte("validator"), Index: 0, Signature: "r|s"}},
				make(EventHashes, 2),
				p.pubKey,
				k, nil)
			items = append(items, event)
			err := store.SetEvent(event)
			if err != nil {
				t.Fatal(err)
			}
		}
		events[p.hex] = items
	}

	// check that events were correctly inserted
	for p, evs := range events {
		for k, ev := range evs {
			rev, err := store.GetEventBlock(ev.Hash())
			if err != nil {
				t.Fatal(err)
			}
			if !ev.Message.Body.Equals(rev.Message.Body) {
				t.Fatalf("events[%s][%d].Body should be %#v, not %#v", p, k, ev, rev)
			}
			if !reflect.DeepEqual(ev.Message.Signature, rev.Message.Signature) {
				t.Fatalf("events[%s][%d].Signature should be %#v, not %#v", p, k, ev.Message.Signature, rev.Message.Signature)
			}
		}
	}

	// check retrieving events per participant
	skipIndex := int64(-1) // do not skip any indexes
	for _, p := range participants {
		pEvents, err := store.ParticipantEvents(p.hex, skipIndex)
		if err != nil {
			t.Fatal(err)
		}
		if l := int64(len(pEvents)); l != testSize {
			t.Fatalf("%s should have %d events, not %d", p.hex, testSize, l)
		}

		expectedEvents := events[p.hex][skipIndex+1:]
		for k, e := range expectedEvents {
			if e.Hash() != pEvents[k] {
				t.Fatalf("ParticipantEvents[%s][%d] should be %s, not %s",
					p.hex, k, e.Hash(), pEvents[k])
			}
		}
	}

	// check retrieving participant last
	for _, p := range participants {
		last, _, err := store.LastEventFrom(p.hex)
		if err != nil {
			t.Fatal(err)
		}

		evs := events[p.hex]
		expectedLast := evs[len(evs)-1].Hash()
		if last != expectedLast {
			t.Fatalf("%s last should be %s, not %s", p.hex, expectedLast.String(), last.String())
		}
	}

	for _, p := range participants {
		evs := events[p.hex]
		for _, ev := range evs {
			if err := store.AddConsensusEvent(ev); err != nil {
				t.Fatal(err)
			}
		}

	}
}

func TestBadgerRounds(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	round := NewRoundCreated()
	events := make(map[string]Event)
	for _, p := range participants {
		event := NewEvent([][]byte{},
			[]InternalTransaction{},
			[]BlockSignature{},
			make(EventHashes, 2),
			p.pubKey,
			0, nil)
		events[p.hex] = event
		round.AddEvent(event.Hash(), true)
	}

	if err := store.SetRoundCreated(0, *round); err != nil {
		t.Fatal(err)
	}

	if c := store.LastRound(); c != 0 {
		t.Fatalf("Store LastRound should be 0, not %d", c)
	}

	storedRound, err := store.GetRoundCreated(0)
	if err != nil {
		t.Fatal(err)
	}

	if !round.Equals(&storedRound) {
		t.Fatalf("Round and StoredRound do not match")
	}

	clothos := store.RoundClothos(0)
	expectedClothos := round.Clotho()
	if len(clothos) != len(expectedClothos) {
		t.Fatalf("There should be %d clothos, not %d", len(expectedClothos), len(clothos))
	}
	for _, w := range expectedClothos {
		if !clothos.Contains(w) {
			t.Fatalf("Clothos should contain %s", w)
		}
	}
}

func TestBadgerBlocks(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	index := int64(0)
	roundReceived := int64(5)
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

		if !storedBlock.Equals(&block) {
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

func TestBadgerFrames(t *testing.T) {
	cacheSize := 1 // Inmem_store's caches accept positive cacheSize only
	store, participants := initBadgerStore(cacheSize, t)
	defer removeBadgerStore(store, t)

	events := make([]*EventMessage, len(participants))
	roots := make([]*Root, len(participants))
	for id, p := range participants {
		event := NewEvent(
			[][]byte{[]byte(fmt.Sprintf("%s_%d", p.hex[:5], 0))},
			[]InternalTransaction{},
			[]BlockSignature{{Validator: []byte("validator"), Index: 0, Signature: "r|s"}},
			make(EventHashes, 2),
			p.pubKey,
			0, nil)
		if err := event.Sign(p.privKey); err != nil {
			t.Fatal(err)
		}
		events[id] = event.Message

		root := NewBaseRoot(uint64(id))
		roots[id] = &root
	}
	frame := Frame{
		Round:  1,
		Events: events,
		Roots:  roots,
	}

	t.Run("Store Frame", func(t *testing.T) {
		if err := store.SetFrame(frame); err != nil {
			t.Fatal(err)
		}

		storedFrame, err := store.GetFrame(frame.Round)
		if err != nil {
			t.Fatal(err)
		}

		if !storedFrame.Equals(&frame) {
			t.Fatalf("Frame and StoredFrame do not match")
		}
	})
}
