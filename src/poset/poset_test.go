package poset

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/andrecronje/lachesis/src/common"
	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/andrecronje/lachesis/src/peers"
	"github.com/sirupsen/logrus"
)

var (
	cacheSize = 100
	n         = 3
	badgerDir = "test_data/badger"
)

type TestNode struct {
	ID     int
	Pub    []byte
	PubHex string
	Key    *ecdsa.PrivateKey
	Events []Event
}

func NewTestNode(key *ecdsa.PrivateKey, id int) TestNode {
	pub := crypto.FromECDSAPub(&key.PublicKey)
	ID := common.Hash32(pub)
	node := TestNode{
		ID:     ID,
		Key:    key,
		Pub:    pub,
		PubHex: fmt.Sprintf("0x%X", pub),
		Events: []Event{},
	}
	return node
}

func (node *TestNode) signAndAddEvent(event Event, name string, index map[string]string, orderedEvents *[]Event) {
	event.Sign(node.Key)
	node.Events = append(node.Events, event)
	index[name] = event.Hex()
	*orderedEvents = append(*orderedEvents, event)
}

type ancestryItem struct {
	descendant, ancestor string
	val                  bool
	err                  bool
}

type roundItem struct {
	event string
	round int64
}

type play struct {
	to          int
	index       int64
	selfParent  string
	otherParent string
	name        string
	txPayload   [][]byte
	sigPayload  []BlockSignature
}

func testLogger(t testing.TB) *logrus.Entry {
	return common.NewTestLogger(t).WithField("id", "test")
}

/* Initialisation functions */

func initPosetNodes(n int) ([]TestNode, map[string]string, *[]Event, *peers.Peers) {
	index := make(map[string]string)
	var nodes []TestNode
	orderedEvents := &[]Event{}
	keys := map[string]*ecdsa.PrivateKey{}

	participants := peers.NewPeers()

	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateECDSAKey()
		pub := crypto.FromECDSAPub(&key.PublicKey)
		pubHex := fmt.Sprintf("0x%X", pub)
		participants.AddPeer(peers.NewPeer(pubHex, ""))
		keys[pubHex] = key
	}

	for i, peer := range participants.ToPeerSlice() {
		nodes = append(nodes, NewTestNode(keys[peer.PubKeyHex], i))
	}

	return nodes, index, orderedEvents, participants
}

func playEvents(plays []play, nodes []TestNode, index map[string]string, orderedEvents *[]Event) {
	for _, p := range plays {
		e := NewEvent(p.txPayload,
			nil,
			p.sigPayload,
			[]string{index[p.selfParent], index[p.otherParent]},
			nodes[p.to].Pub,
			p.index, nil)

		nodes[p.to].signAndAddEvent(e, p.name, index, orderedEvents)
	}
}

func createPoset(db bool, orderedEvents *[]Event, participants *peers.Peers, logger *logrus.Entry) *Poset {
	var store Store
	if db {
		var err error
		store, err = NewBadgerStore(participants, cacheSize, badgerDir)
		if err != nil {
			logger.Fatal("ERROR creating badger store", err)
		}
	} else {
		store = NewInmemStore(participants, cacheSize)
	}

	poset := NewPoset(participants, store, nil, logger)

	for i, ev := range *orderedEvents {
		if err := poset.InsertEvent(ev, true); err != nil {
			logger.Fatalf("ERROR inserting event %d: %s\n", i, err)
		}
	}

	return poset
}

func initPosetFull(plays []play, db bool, n int, logger *logrus.Entry) (*Poset, map[string]string, *[]Event) {
	nodes, index, orderedEvents, participants := initPosetNodes(n)

	// Needed to have sorted nodes based on participants hash32
	for i, peer := range participants.ToPeerSlice() {
		event := NewEvent(nil, nil, nil, []string{rootSelfParent(peer.ID), ""}, nodes[i].Pub, 0, nil)
		nodes[i].signAndAddEvent(event, fmt.Sprintf("e%d", i), index, orderedEvents)
	}

	playEvents(plays, nodes, index, orderedEvents)

	poset := createPoset(db, orderedEvents, participants, logger)

	// Add reference to each participants' root event
	for i, peer := range participants.ToPeerSlice() {
		root, err := poset.Store.GetRoot(peer.PubKeyHex)
		if err != nil {
			panic(err)
		}
		index["r"+strconv.Itoa(i)] = root.SelfParent.Hash
	}

	return poset, index, orderedEvents
}

/*  */

/*
|  e12  |
|   | \ |
|  s10 e20
|   | / |
|   /   |
| / |   |
s00 |  s20
|   |   |
e01 |   |
| \ |   |
e0  e1  e2
|   |   |
r0  r1  r2
0   1   2
*/
func initPoset(t *testing.T) (*Poset, map[string]string) {
	plays := []play{
		{0, 1, "e0", "e1", "e01", nil, nil},
		{2, 1, "e2", "", "s20", nil, nil},
		{1, 1, "e1", "", "s10", nil, nil},
		{0, 2, "e01", "", "s00", nil, nil},
		{2, 2, "s20", "s00", "e20", nil, nil},
		{1, 2, "s10", "e20", "e12", nil, nil},
	}

	p, index, orderedEvents := initPosetFull(plays, false, n, testLogger(t))

	for i, ev := range *orderedEvents {
		if err := p.Store.SetEvent(ev); err != nil {
			t.Fatalf("%d: %s", i, err)
		}
	}

	return p, index
}

func TestAncestor(t *testing.T) {
	p, index := initPoset(t)

	expected := []ancestryItem{
		//first generation
		{"e01", "e0", true, false},
		{"e01", "e1", true, false},
		{"s00", "e01", true, false},
		{"s20", "e2", true, false},
		{"e20", "s00", true, false},
		{"e20", "s20", true, false},
		{"e12", "e20", true, false},
		{"e12", "s10", true, false},
		//second generation
		{"s00", "e0", true, false},
		{"s00", "e1", true, false},
		{"e20", "e01", true, false},
		{"e20", "e2", true, false},
		{"e12", "e1", true, false},
		{"e12", "s20", true, false},
		//third generation
		{"e20", "e0", true, false},
		{"e20", "e1", true, false},
		{"e20", "e2", true, false},
		{"e12", "e01", true, false},
		{"e12", "e0", true, false},
		{"e12", "e1", true, false},
		{"e12", "e2", true, false},
		//false positive
		{"e01", "e2", false, false},
		{"s00", "e2", false, false},
		{"e0", "", false, true},
		{"s00", "", false, true},
		{"e12", "", false, true},
		//root events
		{"e1", "r1", true, false},
		{"e20", "r1", true, false},
		{"e12", "r0", true, false},
		{"s20", "r1", false, false},
		{"r0", "r1", false, false},
	}

	for _, exp := range expected {
		a, err := p.ancestor(index[exp.descendant], index[exp.ancestor])
		if err != nil && !exp.err {
			t.Fatalf("Error computing ancestor(%s, %s). Err: %v", exp.descendant, exp.ancestor, err)
		}
		if a != exp.val {
			t.Fatalf("ancestor(%s, %s) should be %v, not %v", exp.descendant, exp.ancestor, exp.val, a)
		}
	}
}

func TestSelfAncestor(t *testing.T) {
	p, index := initPoset(t)

	expected := []ancestryItem{
		//1 generation
		{"e01", "e0", true, false},
		{"s00", "e01", true, false},
		//1 generation false negative
		{"e01", "e1", false, false},
		{"e12", "e20", false, false},
		{"s20", "e1", false, false},
		{"s20", "", false, true},
		//2 generations
		{"e20", "e2", true, false},
		{"e12", "e1", true, false},
		//2 generations false negatives
		{"e20", "e0", false, false},
		{"e12", "e2", false, false},
		{"e20", "e01", false, false},
		//roots
		{"e20", "r2", true, false},
		{"e1", "r1", true, false},
		{"e1", "r0", false, false},
		{"r1", "r0", false, false},
	}

	for _, exp := range expected {
		a, err := p.selfAncestor(index[exp.descendant], index[exp.ancestor])
		if err != nil && !exp.err {
			t.Fatalf("Error computing selfAncestor(%s, %s). Err: %v", exp.descendant, exp.ancestor, err)
		}
		if a != exp.val {
			t.Fatalf("selfAncestor(%s, %s) should be %v, not %v", exp.descendant, exp.ancestor, exp.val, a)
		}
	}
}

func TestSee(t *testing.T) {
	p, index := initPoset(t)

	expected := []ancestryItem{
		{"e01", "e0", true, false},
		{"e01", "e1", true, false},
		{"e20", "e0", true, false},
		{"e20", "e01", true, false},
		{"e12", "e01", true, false},
		{"e12", "e0", true, false},
		{"e12", "e1", true, false},
		{"e12", "s20", true, false},
	}

	for _, exp := range expected {
		a, err := p.see(index[exp.descendant], index[exp.ancestor])
		if err != nil && !exp.err {
			t.Fatalf("Error computing see(%s, %s). Err: %v", exp.descendant, exp.ancestor, err)
		}
		if a != exp.val {
			t.Fatalf("see(%s, %s) should be %v, not %v", exp.descendant, exp.ancestor, exp.val, a)
		}
	}
}

func TestLamportTimestamp(t *testing.T) {
	p, index := initPoset(t)

	expectedTimestamps := map[string]int64{
		"e0":  0,
		"e1":  0,
		"e2":  0,
		"e01": 1,
		"s10": 1,
		"s20": 1,
		"s00": 2,
		"e20": 3,
		"e12": 4,
	}

	for e, ets := range expectedTimestamps {
		ts, err := p.lamportTimestamp(index[e])
		if err != nil {
			t.Fatalf("Error computing lamportTimestamp(%s). Err: %s", e, err)
		}
		if ts != ets {
			t.Fatalf("%s LamportTimestamp should be %d, not %d", e, ets, ts)
		}
	}
}

/*
|    |    e20
|    |   / |
|    | /   |
|    /     |
|  / |     |
e01  |     |
| \  |     |
|   \|     |
|    |\    |
|    |  \  |
e0   e1 (a)e2
0    1     2

Node 2 Forks; events a and e2 are both created by node2, they are not self-parents
and yet they are both ancestors of event e20
*/
func TestFork(t *testing.T) {
	index := make(map[string]string)
	var nodes []TestNode
	participants := peers.NewPeers()

	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateECDSAKey()
		node := NewTestNode(key, i)
		nodes = append(nodes, node)
		participants.AddPeer(peers.NewPeer(node.PubHex, ""))
	}

	store := NewInmemStore(participants, cacheSize)
	poset := NewPoset(participants, store, nil, testLogger(t))

	for i, node := range nodes {
		event := NewEvent(nil, nil, nil, []string{"", ""}, node.Pub, 0, nil)
		event.Sign(node.Key)
		index[fmt.Sprintf("e%d", i)] = event.Hex()
		poset.InsertEvent(event, true)
	}

	//a and e2 need to have different hashes
	eventA := NewEvent([][]byte{[]byte("yo")}, nil, nil, []string{"", ""}, nodes[2].Pub, 0, nil)
	eventA.Sign(nodes[2].Key)
	index["a"] = eventA.Hex()
	if err := poset.InsertEvent(eventA, true); err == nil {
		t.Fatal("InsertEvent should return error for 'a'")
	}

	event01 := NewEvent(nil, nil, nil,
		[]string{index["e0"], index["a"]}, //e0 and a
		nodes[0].Pub, 1, nil)
	event01.Sign(nodes[0].Key)
	index["e01"] = event01.Hex()
	if err := poset.InsertEvent(event01, true); err == nil {
		t.Fatal("InsertEvent should return error for e01")
	}

	event20 := NewEvent(nil, nil, nil,
		[]string{index["e2"], index["e01"]}, //e2 and e01
		nodes[2].Pub, 1, nil)
	event20.Sign(nodes[2].Key)
	index["e20"] = event20.Hex()
	if err := poset.InsertEvent(event20, true); err == nil {
		t.Fatal("InsertEvent should return error for e20")
	}
}

/*
|  s11  |
|   |   |
|   f1  |
|  /|   |
| / s10 |
|/  |   |
e02 |   |
| \ |   |
|   \   |
|   | \ |
s00 |  e21
|   | / |
|  e10  s20
| / |   |
e0  e1  e2
0   1    2
*/

func initRoundPoset(t *testing.T) (*Poset, map[string]string) {
	plays := []play{
		{1, 1, "e1", "e0", "e10", nil, nil},
		{2, 1, "e2", "", "s20", nil, nil},
		{0, 1, "e0", "", "s00", nil, nil},
		{2, 2, "s20", "e10", "e21", nil, nil},
		{0, 2, "s00", "e21", "e02", nil, nil},
		{1, 2, "e10", "", "s10", nil, nil},
		{1, 3, "s10", "e02", "f1", nil, nil},
		{1, 4, "f1", "", "s11", [][]byte{[]byte("abc")}, nil},
	}

	p, index, _ := initPosetFull(plays, false, n, testLogger(t))

	return p, index
}

func TestInsertEvent(t *testing.T) {
	p, index := initRoundPoset(t)

	checkParents := func(e, selfAncestor, ancestor string) bool {
		ev, err := p.Store.GetEvent(index[e])
		if err != nil {
			t.Fatal(err)
		}
		return ev.SelfParent() == selfAncestor && ev.OtherParent() == ancestor
	}

	t.Run("Check Event Coordinates", func(t *testing.T) {

		//e0
		e0, err := p.Store.GetEvent(index["e0"])
		if err != nil {
			t.Fatal(err)
		}

		if !(e0.Body.selfParentIndex == -1 &&
			e0.Body.otherParentCreatorID == -1 &&
			e0.Body.otherParentIndex == -1 &&
			e0.Body.creatorID == p.Participants.ByPubKey[e0.Creator()].ID) {
			t.Fatalf("Invalid wire info on e0")
		}

		//e21
		e21, err := p.Store.GetEvent(index["e21"])
		if err != nil {
			t.Fatal(err)
		}

		e10, err := p.Store.GetEvent(index["e10"])
		if err != nil {
			t.Fatal(err)
		}

		if !(e21.Body.selfParentIndex == 1 &&
			e21.Body.otherParentCreatorID == p.Participants.ByPubKey[e10.Creator()].ID &&
			e21.Body.otherParentIndex == 1 &&
			e21.Body.creatorID == p.Participants.ByPubKey[e21.Creator()].ID) {
			t.Fatalf("Invalid wire info on e21")
		}

		//f1
		f1, err := p.Store.GetEvent(index["f1"])
		if err != nil {
			t.Fatal(err)
		}

		if !(f1.Body.selfParentIndex == 2 &&
			f1.Body.otherParentCreatorID == p.Participants.ByPubKey[e0.Creator()].ID &&
			f1.Body.otherParentIndex == 2 &&
			f1.Body.creatorID == p.Participants.ByPubKey[f1.Creator()].ID) {
			t.Fatalf("Invalid wire info on f1")
		}

		e0CreatorID := strconv.FormatInt(p.Participants.ByPubKey[e0.Creator()].ID, 10)

		type Hierarchy struct {
			ev, selfAncestor, ancestor string
		}

		toCheck := []Hierarchy{
			{"e0", "Root" + e0CreatorID, ""},
			{"e10", index["e1"], index["e0"]},
			{"e21", index["s20"], index["e10"]},
			{"e02", index["s00"], index["e21"]},
			{"f1", index["s10"], index["e02"]},
		}

		for _, v := range toCheck {
			if !checkParents(v.ev, v.selfAncestor, v.ancestor) {
				t.Fatal(v.ev + " selfParent not good")
			}
		}
	})

	t.Run("Check UndeterminedEvents", func(t *testing.T) {

		expectedUndeterminedEvents := []string{
			index["e0"],
			index["e1"],
			index["e2"],
			index["e10"],
			index["s20"],
			index["s00"],
			index["e21"],
			index["e02"],
			index["s10"],
			index["f1"],
			index["s11"]}

		for i, eue := range expectedUndeterminedEvents {
			if ue := p.UndeterminedEvents[i]; ue != eue {
				t.Fatalf("UndeterminedEvents[%d] should be %s, not %s", i, eue, ue)
			}
		}

		//Pending loaded Events
		// 3 Events with index 0,
		// 1 Event with non-empty Transactions
		//= 4 Loaded Events
		if ple := p.PendingLoadedEvents; ple != 4 {
			t.Fatalf("PendingLoadedEvents should be 4, not %d", ple)
		}
	})
}

func TestReadWireInfo(t *testing.T) {
	p, index := initRoundPoset(t)

	for k, evh := range index {
		if k[0] == 'r' {
			continue
		}
		ev, err := p.Store.GetEvent(evh)
		if err != nil {
			t.Fatal(err)
		}

		evWire := ev.ToWire()

		evFromWire, err := p.ReadWireInfo(evWire)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(ev.Body.BlockSignatures, evFromWire.Body.BlockSignatures) {
			t.Fatalf("Error converting %s.Body.BlockSignatures from light wire", k)
		}

		if !reflect.DeepEqual(ev.Body, evFromWire.Body) {
			t.Fatalf("Error converting %s.Body from light wire", k)
		}

		if !reflect.DeepEqual(ev.Signature, evFromWire.Signature) {
			t.Fatalf("Error converting %s.Signature from light wire", k)
		}

		ok, err := ev.Verify()
		if !ok {
			t.Fatalf("Error verifying signature for %s from ligh wire: %v", k, err)
		}
	}
}

func TestStronglySee(t *testing.T) {
	p, index := initRoundPoset(t)

	expected := []ancestryItem{
		{"e21", "e0", true, false},
		{"e02", "e10", true, false},
		{"e02", "e0", true, false},
		{"e02", "e1", true, false},
		{"f1", "e21", true, false},
		{"f1", "e10", true, false},
		{"f1", "e0", true, false},
		{"f1", "e1", true, false},
		{"f1", "e2", true, false},
		{"s11", "e2", true, false},
		//false negatives
		{"e10", "e0", false, false},
		{"e21", "e1", false, false},
		{"e21", "e2", false, false},
		{"e02", "e2", false, false},
		{"s11", "e02", false, false},
		{"s11", "", false, true},
		// root events
		{"s11", "r1", true, false},
		{"e21", "r0", true, false},
		{"e21", "r1", false, false},
		{"e10", "r0", false, false},
		{"s20", "r2", false, false},
		{"e02", "r2", false, false},
		{"e21", "r2", false, false},
	}

	for _, exp := range expected {
		a, err := p.stronglySee(index[exp.descendant], index[exp.ancestor])
		if err != nil && !exp.err {
			t.Fatalf("Error computing stronglySee(%s, %s). Err: %v", exp.descendant, exp.ancestor, err)
		}
		if a != exp.val {
			t.Fatalf("stronglySee(%s, %s) should be %v, not %v", exp.descendant, exp.ancestor, exp.val, a)
		}
	}
}

func TestWitness(t *testing.T) {
	p, index := initRoundPoset(t)

	round0Witnesses := make(map[string]RoundEvent)
	round0Witnesses[index["e0"]] = RoundEvent{Witness: true, Famous: Undefined}
	round0Witnesses[index["e1"]] = RoundEvent{Witness: true, Famous: Undefined}
	round0Witnesses[index["e2"]] = RoundEvent{Witness: true, Famous: Undefined}
	p.Store.SetRound(0, RoundInfo{Events: round0Witnesses})

	round1Witnesses := make(map[string]RoundEvent)
	round1Witnesses[index["f1"]] = RoundEvent{Witness: true, Famous: Undefined}
	p.Store.SetRound(1, RoundInfo{Events: round1Witnesses})

	expected := []ancestryItem{
		{"", "e0", true, false},
		{"", "e1", true, false},
		{"", "e2", true, false},
		{"", "f1", true, false},
		{"", "e10", false, false},
		{"", "e21", false, false},
		{"", "e02", false, false},
	}

	for _, exp := range expected {
		a, err := p.witness(index[exp.ancestor])
		if err != nil {
			t.Fatalf("Error computing witness(%s). Err: %v", exp.ancestor, err)
		}
		if a != exp.val {
			t.Fatalf("witness(%s) should be %v, not %v", exp.ancestor, exp.val, a)
		}
	}
}

func TestRound(t *testing.T) {
	p, index := initRoundPoset(t)

	round0Witnesses := make(map[string]RoundEvent)
	round0Witnesses[index["e0"]] = RoundEvent{Witness: true, Famous: Undefined}
	round0Witnesses[index["e1"]] = RoundEvent{Witness: true, Famous: Undefined}
	round0Witnesses[index["e2"]] = RoundEvent{Witness: true, Famous: Undefined}
	p.Store.SetRound(0, RoundInfo{Events: round0Witnesses})

	expected := []roundItem{
		{"e0", 0},
		{"e1", 0},
		{"e2", 0},
		{"s00", 0},
		{"e10", 0},
		{"s20", 0},
		{"e21", 0},
		{"e02", 0},
		{"s10", 0},
		{"f1", 1},
		{"s11", 1},
	}

	for _, exp := range expected {
		r, err := p.round(index[exp.event])
		if err != nil {
			t.Fatalf("Error computing round(%s). Err: %v", exp.event, err)
		}
		if r != exp.round {
			t.Fatalf("round(%s) should be %v, not %v", exp.event, exp.round, r)
		}
	}
}

func TestRoundDiff(t *testing.T) {
	p, index := initRoundPoset(t)

	round0Witnesses := make(map[string]RoundEvent)
	round0Witnesses[index["e0"]] = RoundEvent{Witness: true, Famous: Undefined}
	round0Witnesses[index["e1"]] = RoundEvent{Witness: true, Famous: Undefined}
	round0Witnesses[index["e2"]] = RoundEvent{Witness: true, Famous: Undefined}
	p.Store.SetRound(0, RoundInfo{Events: round0Witnesses})

	if d, err := p.roundDiff(index["f1"], index["e02"]); d != 1 {
		if err != nil {
			t.Fatalf("RoundDiff(f1, e02) returned an error: %s", err)
		}
		t.Fatalf("RoundDiff(f1, e02) should be 1 not %d", d)
	}

	if d, err := p.roundDiff(index["e02"], index["f1"]); d != -1 {
		if err != nil {
			t.Fatalf("RoundDiff(e02, f1) returned an error: %s", err)
		}
		t.Fatalf("RoundDiff(e02, f1) should be -1 not %d", d)
	}
	if d, err := p.roundDiff(index["e02"], index["e21"]); d != 0 {
		if err != nil {
			t.Fatalf("RoundDiff(e20, e21) returned an error: %s", err)
		}
		t.Fatalf("RoundDiff(e20, e21) should be 0 not %d", d)
	}
}

func TestDivideRounds(t *testing.T) {
	p, index := initRoundPoset(t)

	if err := p.DivideRounds(); err != nil {
		t.Fatal(err)
	}

	if l := p.Store.LastRound(); l != 1 {
		t.Fatalf("last round should be 1 not %d", l)
	}

	round0, err := p.Store.GetRound(0)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(round0.Witnesses()); l != 3 {
		t.Fatalf("round 0 should have 3 witnesses, not %d", l)
	}
	if !contains(round0.Witnesses(), index["e0"]) {
		t.Fatalf("round 0 witnesses should contain e0")
	}
	if !contains(round0.Witnesses(), index["e1"]) {
		t.Fatalf("round 0 witnesses should contain e1")
	}
	if !contains(round0.Witnesses(), index["e2"]) {
		t.Fatalf("round 0 witnesses should contain e2")
	}

	round1, err := p.Store.GetRound(1)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(round1.Witnesses()); l != 1 {
		t.Fatalf("round 1 should have 1 witness, not %d", l)
	}
	if !contains(round1.Witnesses(), index["f1"]) {
		t.Fatalf("round 1 witnesses should contain f1")
	}

	expectedPendingRounds := []pendingRound{
		{
			Index:   0,
			Decided: false,
		},
		{
			Index:   1,
			Decided: false,
		},
	}
	for i, pd := range p.PendingRounds {
		if !reflect.DeepEqual(*pd, expectedPendingRounds[i]) {
			t.Fatalf("pendingRounds[%d] should be %v, not %v", i, expectedPendingRounds[i], *pd)
		}
	}

	//[event] => {lamportTimestamp, round}
	type tr struct {
		t, r int64
	}
	expectedTimestamps := map[string]tr{
		"e0":  {0, 0},
		"e1":  {0, 0},
		"e2":  {0, 0},
		"s00": {1, 0},
		"e10": {1, 0},
		"s20": {1, 0},
		"e21": {2, 0},
		"e02": {3, 0},
		"s10": {2, 0},
		"f1":  {4, 1},
		"s11": {5, 1},
	}

	for e, et := range expectedTimestamps {
		ev, err := p.Store.GetEvent(index[e])
		if err != nil {
			t.Fatal(err)
		}
		if r := ev.round; r == nil || *r != et.r {
			t.Fatalf("%s round should be %d, not %d", e, et.r, *r)
		}
		if ts := ev.lamportTimestamp; ts == nil || *ts != et.t {
			t.Fatalf("%s lamportTimestamp should be %d, not %d", e, et.t, *ts)
		}
	}

}

func TestCreateRoot(t *testing.T) {
	p, index := initRoundPoset(t)
	p.DivideRounds()

	participants := p.Participants.ToPeerSlice()

	baseRoot := NewBaseRoot(participants[0].ID)

	expected := map[string]Root{
		"e0": baseRoot,
		"e02": {
			NextRound:  0,
			SelfParent: RootEvent{index["s00"], participants[0].ID, 1, 1, 0},
			Others: map[string]RootEvent{
				index["e02"]: {index["e21"], participants[2].ID, 2, 2, 0},
			},
		},
		"s10": {
			NextRound:  0,
			SelfParent: RootEvent{index["e10"], participants[1].ID, 1, 1, 0},
			Others:     map[string]RootEvent{},
		},
		"f1": {
			NextRound:  1,
			SelfParent: RootEvent{index["s10"], participants[1].ID, 2, 2, 0},
			Others: map[string]RootEvent{
				index["f1"]: {index["e02"], participants[0].ID, 2, 3, 0},
			},
		},
	}

	for evh, expRoot := range expected {
		ev, err := p.Store.GetEvent(index[evh])
		if err != nil {
			t.Fatal(err)
		}
		root, err := p.createRoot(ev)
		if err != nil {
			t.Fatalf("Error creating %s Root: %v", evh, err)
		}
		if !reflect.DeepEqual(expRoot, root) {
			t.Fatalf("%s Root should be %v, not %v", evh, expRoot, root)
		}
	}

}

func contains(s []string, x string) bool {
	for _, e := range s {
		if e == x {
			return true
		}
	}
	return false
}

/*



e01  e12
 |   |  \
 e0  R1  e2
 |       |
 R0      R2

*/
func initDentedPoset(t *testing.T) (*Poset, map[string]string) {
	nodes, index, orderedEvents, participants := initPosetNodes(n)

	orderedPeers := participants.ToPeerSlice()

	for _, peer := range orderedPeers {
		index[rootSelfParent(peer.ID)] = rootSelfParent(peer.ID)
	}

	plays := []play{
		{0, 0, rootSelfParent(orderedPeers[0].ID), "", "e0", nil, nil},
		{2, 0, rootSelfParent(orderedPeers[2].ID), "", "e2", nil, nil},
		{0, 1, "e0", "", "e01", nil, nil},
		{1, 0, rootSelfParent(orderedPeers[1].ID), "e2", "e12", nil, nil},
	}

	playEvents(plays, nodes, index, orderedEvents)

	poset := createPoset(false, orderedEvents, participants, testLogger(t))

	return poset, index
}

func TestCreateRootBis(t *testing.T) {
	p, index := initDentedPoset(t)

	participants := p.Participants.ToPeerSlice()

	expected := map[string]Root{
		"e12": {
			NextRound:  0,
			SelfParent: NewBaseRootEvent(participants[1].ID),
			Others: map[string]RootEvent{
				index["e12"]: {index["e2"], participants[2].ID, 0, 0, 0},
			},
		},
	}

	for evh, expRoot := range expected {
		ev, err := p.Store.GetEvent(index[evh])
		if err != nil {
			t.Fatal(err)
		}
		root, err := p.createRoot(ev)
		if err != nil {
			t.Fatalf("Error creating %s Root: %v", evh, err)
		}
		if !reflect.DeepEqual(expRoot, root) {
			t.Fatalf("%s Root should be %v, not %v", evh, expRoot, root)
		}
	}
}

/*

e0  e1  e2    Block (0, 1)
0   1    2
*/
func initBlockPoset(t *testing.T) (*Poset, []TestNode, map[string]string) {
	nodes, index, orderedEvents, participants := initPosetNodes(n)

	for i, peer := range participants.ToPeerSlice() {
		event := NewEvent(nil, nil, nil, []string{rootSelfParent(peer.ID), ""}, nodes[i].Pub, 0, nil)
		nodes[i].signAndAddEvent(event, fmt.Sprintf("e%d", i), index, orderedEvents)
	}

	poset := NewPoset(participants, NewInmemStore(participants, cacheSize), nil, testLogger(t))

	//create a block and signatures manually
	block := NewBlock(0, 1, []byte("framehash"), [][]byte{[]byte("block tx")})
	err := poset.Store.SetBlock(block)
	if err != nil {
		t.Fatalf("Error setting block. Err: %s", err)
	}

	for i, ev := range *orderedEvents {
		if err := poset.InsertEvent(ev, true); err != nil {
			fmt.Printf("ERROR inserting event %d: %s\n", i, err)
		}
	}

	return poset, nodes, index
}

func TestInsertEventsWithBlockSignatures(t *testing.T) {
	p, nodes, index := initBlockPoset(t)

	block, err := p.Store.GetBlock(0)
	if err != nil {
		t.Fatalf("Error retrieving block 0. %s", err)
	}

	blockSigs := make([]BlockSignature, n)
	for k, n := range nodes {
		blockSigs[k], err = block.Sign(n.Key)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Inserting Events with valid signatures", func(t *testing.T) {

		/*
			s00 |   |
			|   |   |
			|  e10  s20
			| / |   |
			e0  e1  e2
			0   1    2
		*/
		plays := []play{
			{1, 1, "e1", "e0", "e10", nil, []BlockSignature{blockSigs[1]}},
			{2, 1, "e2", "", "s20", nil, []BlockSignature{blockSigs[2]}},
			{0, 1, "e0", "", "s00", nil, []BlockSignature{blockSigs[0]}},
		}

		for _, pl := range plays {
			e := NewEvent(pl.txPayload,
				nil,
				pl.sigPayload,
				[]string{index[pl.selfParent], index[pl.otherParent]},
				nodes[pl.to].Pub,
				pl.index, nil)
			e.Sign(nodes[pl.to].Key)
			index[pl.name] = e.Hex()
			if err := p.InsertEvent(e, true); err != nil {
				t.Fatalf("ERROR inserting event %s: %s\n", pl.name, err)
			}
		}

		//Check SigPool
		if l := len(p.SigPool); l != 3 {
			t.Fatalf("SigPool should contain 3 signatures, not %d", l)
		}

		//Process SigPool
		p.ProcessSigPool()

		//Check that the block contains 3 signatures
		block, _ := p.Store.GetBlock(0)
		if l := len(block.Signatures); l != 2 {
			t.Fatalf("Block 0 should contain 2 signatures, not %d", l)
		}

		//Check that SigPool was cleared
		if l := len(p.SigPool); l != 0 {
			t.Fatalf("SigPool should contain 0 signatures, not %d", l)
		}
	})

	t.Run("Inserting Events with signature of unknown block", func(t *testing.T) {
		//The Event should be inserted
		//The block signature is simply ignored

		block1 := NewBlock(1, 2, []byte("framehash"), [][]byte{})
		sig, _ := block1.Sign(nodes[2].Key)

		//unknown block
		unknownBlockSig := BlockSignature{
			Validator: nodes[2].Pub,
			Index:     1,
			Signature: sig.Signature,
		}
		pl := play{2, 2, "s20", "e10", "e21", nil, []BlockSignature{unknownBlockSig}}

		e := NewEvent(nil,
			nil,
			pl.sigPayload,
			[]string{index[pl.selfParent], index[pl.otherParent]},
			nodes[pl.to].Pub,
			pl.index, nil)
		e.Sign(nodes[pl.to].Key)
		index[pl.name] = e.Hex()
		if err := p.InsertEvent(e, true); err != nil {
			t.Fatalf("ERROR inserting event %s: %s", pl.name, err)
		}

		//check that the event was recorded
		_, err := p.Store.GetEvent(index["e21"])
		if err != nil {
			t.Fatalf("ERROR fetching Event e21: %s", err)
		}

	})

	t.Run("Inserting Events with BlockSignature not from creator", func(t *testing.T) {
		//The Event should be inserted
		//The block signature is simply ignored

		//wrong validator
		//Validator should be same as Event creator (node 0)
		key, _ := crypto.GenerateECDSAKey()
		badNode := NewTestNode(key, 666)
		badNodeSig, _ := block.Sign(badNode.Key)

		pl := play{0, 2, "s00", "e21", "e02", nil, []BlockSignature{badNodeSig}}

		e := NewEvent(nil,
			nil,
			pl.sigPayload,
			[]string{index[pl.selfParent], index[pl.otherParent]},
			nodes[pl.to].Pub,
			pl.index, nil)
		e.Sign(nodes[pl.to].Key)
		index[pl.name] = e.Hex()
		if err := p.InsertEvent(e, true); err != nil {
			t.Fatalf("ERROR inserting event %s: %s\n", pl.name, err)
		}

		//check that the signature was not appended to the block
		block, _ := p.Store.GetBlock(0)
		if l := len(block.Signatures); l > 3 {
			t.Fatalf("Block 0 should contain 3 signatures, not %d", l)
		}
	})

}

/*
                  Round 4
		i0  |   i2
		| \ | / |
		|   i1  |
------- |  /|   | --------------------------------
		h02 |   | Round 3
		| \ |   |
		|   \   |
		|   | \ |
		|   |  h21
		|   | / |
		|  h10  |
		| / |   |
		h0  |   h2
		| \ | / |
		|   h1  |
------- |  /|   | --------------------------------
		g02 |   | Round 2
		| \ |   |
		|   \   |
		|   | \ |
	    |   |  g21
		|   | / |
		|  g10  |
		| / |   |
		g0  |   g2
		| \ | / |
		|   g1  |
------- |  /|   | -------------------------------
		f02b|   |  Round 1           +---------+
		|   |   |                    | Block 1 |
		f02 |   |                    | RR    2 |
		| \ |   |                    | Evs   9 |
		|   \   |                    +---------+
		|   | \ |
	---f0x  |   f21 //f0x's other-parent is e21b. This situation can happen with concurrency
	|	|   | / |
	|	|  f10  |
	|	| / |   |
	|	f0  |   f2
	|	| \ | / |
	|	|  f1b  |
	|	|   |   |
	|	|   f1  |
---	| -	|  /|   | ------------------------------
	|	e02 |   |  Round 0          +---------+
	|	| \ |   |                   | Block 0 |
	|	|   \   |                   | RR    1 |
	|	|   | \ |                   | Evs   7 |
	|   |   | e21b                  +---------+
	|	|   |   |
	---------- e21
		|   | / |
		|  e10  |
	    | / |   |
		e0  e1  e2
		0   1    2
*/
func initConsensusPoset(db bool, t testing.TB) (*Poset, map[string]string) {
	plays := []play{
		{1, 1, "e1", "e0", "e10", nil, nil},
		{2, 1, "e2", "e10", "e21", [][]byte{[]byte("e21")}, nil},
		{2, 2, "e21", "", "e21b", nil, nil},
		{0, 1, "e0", "e21b", "e02", nil, nil},
		{1, 2, "e10", "e02", "f1", nil, nil},
		{1, 3, "f1", "", "f1b", [][]byte{[]byte("f1b")}, nil},
		{0, 2, "e02", "f1b", "f0", nil, nil},
		{2, 3, "e21b", "f1b", "f2", nil, nil},
		{1, 4, "f1b", "f0", "f10", nil, nil},
		{0, 3, "f0", "e21", "f0x", nil, nil},
		{2, 4, "f2", "f10", "f21", nil, nil},
		{0, 4, "f0x", "f21", "f02", nil, nil},
		{0, 5, "f02", "", "f02b", [][]byte{[]byte("f02b")}, nil},
		{1, 5, "f10", "f02b", "g1", nil, nil},
		{0, 6, "f02b", "g1", "g0", nil, nil},
		{2, 5, "f21", "g1", "g2", nil, nil},
		{1, 6, "g1", "g0", "g10", [][]byte{[]byte("g10")}, nil},
		{2, 6, "g2", "g10", "g21", nil, nil},
		{0, 7, "g0", "g21", "g02", [][]byte{[]byte("g02")}, nil},
		{1, 7, "g10", "g02", "h1", nil, nil},
		{0, 8, "g02", "h1", "h0", nil, nil},
		{2, 7, "g21", "h1", "h2", nil, nil},
		{1, 8, "h1", "h0", "h10", nil, nil},
		{2, 8, "h2", "h10", "h21", nil, nil},
		{0, 9, "h0", "h21", "h02", nil, nil},
		{1, 9, "h10", "h02", "i1", nil, nil},
		{0, 10, "h02", "i1", "i0", nil, nil},
		{2, 9, "h21", "i1", "i2", nil, nil},
	}

	poset, index, _ := initPosetFull(plays, db, n, testLogger(t))

	return poset, index
}

func TestDivideRoundsBis(t *testing.T) {
	p, index := initConsensusPoset(false, t)

	if err := p.DivideRounds(); err != nil {
		t.Fatal(err)
	}

	//[event] => {lamportTimestamp, round}
	type tr struct {
		t, r int64
	}
	expectedTimestamps := map[string]tr{
		"e0":   {0, 0},
		"e1":   {0, 0},
		"e2":   {0, 0},
		"e10":  {1, 0},
		"e21":  {2, 0},
		"e21b": {3, 0},
		"e02":  {4, 0},
		"f1":   {5, 1},
		"f1b":  {6, 1},
		"f0":   {7, 1},
		"f2":   {7, 1},
		"f10":  {8, 1},
		"f0x":  {8, 1},
		"f21":  {9, 1},
		"f02":  {10, 1},
		"f02b": {11, 1},
		"g1":   {12, 2},
		"g0":   {13, 2},
		"g2":   {13, 2},
		"g10":  {14, 2},
		"g21":  {15, 2},
		"g02":  {16, 2},
		"h1":   {17, 3},
		"h0":   {18, 3},
		"h2":   {18, 3},
		"h10":  {19, 3},
		"h21":  {20, 3},
		"h02":  {21, 3},
		"i1":   {22, 4},
		"i0":   {23, 4},
		"i2":   {23, 4},
	}

	for e, et := range expectedTimestamps {
		ev, err := p.Store.GetEvent(index[e])
		if err != nil {
			t.Fatal(err)
		}
		if r := ev.round; r == nil || *r != et.r {
			t.Fatalf("%s round should be %d, not %d", e, et.r, *r)
		}
		if ts := ev.lamportTimestamp; ts == nil || *ts != et.t {
			t.Fatalf("%s lamportTimestamp should be %d, not %d", e, et.t, *ts)
		}
	}

}

func TestDecideFame(t *testing.T) {
	p, index := initConsensusPoset(false, t)

	p.DivideRounds()
	if err := p.DecideFame(); err != nil {
		t.Fatal(err)
	}

	round0, err := p.Store.GetRound(0)
	if err != nil {
		t.Fatal(err)
	}
	if f := round0.Events[index["e0"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("e0 should be famous; got %v", f)
	}
	if f := round0.Events[index["e1"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("e1 should be famous; got %v", f)
	}
	if f := round0.Events[index["e2"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("e2 should be famous; got %v", f)
	}

	round1, err := p.Store.GetRound(1)
	if err != nil {
		t.Fatal(err)
	}
	if f := round1.Events[index["f0"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("f0 should be famous; got %v", f)
	}
	if f := round1.Events[index["f1"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("f1 should be famous; got %v", f)
	}
	if f := round1.Events[index["f2"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("f2 should be famous; got %v", f)
	}

	round2, err := p.Store.GetRound(2)
	if err != nil {
		t.Fatal(err)
	}
	if f := round2.Events[index["g0"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("g0 should be famous; got %v", f)
	}
	if f := round2.Events[index["g1"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("g1 should be famous; got %v", f)
	}
	if f := round2.Events[index["g2"]]; !(f.Witness && f.Famous == True) {
		t.Fatalf("g2 should be famous; got %v", f)
	}

	expectedpendingRounds := []pendingRound{
		{
			Index:   0,
			Decided: true,
		},
		{
			Index:   1,
			Decided: true,
		},
		{
			Index:   2,
			Decided: true,
		},
		{
			Index:   3,
			Decided: false,
		},
		{
			Index:   4,
			Decided: false,
		},
	}
	for i, pd := range p.PendingRounds {
		if !reflect.DeepEqual(*pd, expectedpendingRounds[i]) {
			t.Fatalf("pendingRounds[%d] should be %v, not %v", i, expectedpendingRounds[i], *pd)
		}
	}
}

func TestDecideRoundReceived(t *testing.T) {
	p, index := initConsensusPoset(false, t)

	p.DivideRounds()
	p.DecideFame()
	if err := p.DecideRoundReceived(); err != nil {
		t.Fatal(err)
	}

	for name, hash := range index {
		e, _ := p.Store.GetEvent(hash)
		if rune(name[0]) == rune('e') {
			if r := *e.roundReceived; r != 1 {
				t.Fatalf("%s round received should be 1 not %d", name, r)
			}
		} else if rune(name[0]) == rune('f') {
			if r := *e.roundReceived; r != 2 {
				t.Fatalf("%s round received should be 2 not %d", name, r)
			}
		} else if e.roundReceived != nil {
			t.Fatalf("%s round received should be nil not %d", name, *e.roundReceived)
		}
	}

	round0, err := p.Store.GetRound(0)
	if err != nil {
		t.Fatalf("Could not retrieve Round 0. %s", err)
	}
	if ce := len(round0.ConsensusEvents()); ce != 0 {
		t.Fatalf("Round 0 should contain 0 ConsensusEvents, not %d", ce)
	}

	round1, err := p.Store.GetRound(1)
	if err != nil {
		t.Fatalf("Could not retrieve Round 1. %s", err)
	}
	if ce := len(round1.ConsensusEvents()); ce != 7 {
		t.Fatalf("Round 1 should contain 7 ConsensusEvents, not %d", ce)
	}

	round2, err := p.Store.GetRound(2)
	if err != nil {
		t.Fatalf("Could not retrieve Round 2. %s", err)
	}
	if ce := len(round2.ConsensusEvents()); ce != 9 {
		t.Fatalf("Round 1 should contain 9 ConsensusEvents, not %d", ce)
	}

	expectedUndeterminedEvents := []string{
		index["g1"],
		index["g0"],
		index["g2"],
		index["g10"],
		index["g21"],
		index["g02"],
		index["h1"],
		index["h0"],
		index["h2"],
		index["h10"],
		index["h21"],
		index["h02"],
		index["i1"],
		index["i0"],
		index["i2"],
	}

	for i, eue := range expectedUndeterminedEvents {
		if ue := p.UndeterminedEvents[i]; ue != eue {
			t.Fatalf("UndeterminedEvents[%d] should be %s, not %s", i, eue, ue)
		}
	}
}

func TestProcessDecidedRounds(t *testing.T) {
	p, index := initConsensusPoset(false, t)

	p.DivideRounds()
	p.DecideFame()
	p.DecideRoundReceived()
	if err := p.ProcessDecidedRounds(); err != nil {
		t.Fatal(err)
	}

	//--------------------------------------------------------------------------
	consensusEvents := p.Store.ConsensusEvents()

	for i, e := range consensusEvents {
		t.Logf("consensus[%d]: %s\n", i, getName(index, e))
	}

	if l := len(consensusEvents); l != 16 {
		t.Fatalf("length of consensus should be 16 not %d", l)
	}

	if ple := p.PendingLoadedEvents; ple != 2 {
		t.Fatalf("PendingLoadedEvents should be 2, not %d", ple)
	}

	//Block 0 ------------------------------------------------------------------
	block0, err := p.Store.GetBlock(0)
	if err != nil {
		t.Fatalf("Store should contain a block with Index 0: %v", err)
	}

	if ind := block0.Index(); ind != 0 {
		t.Fatalf("Block0's Index should be 0, not %d", ind)
	}

	if rr := block0.RoundReceived(); rr != 1 {
		t.Fatalf("Block0's RoundReceived should be 1, not %d", rr)
	}

	if l := len(block0.Transactions()); l != 1 {
		t.Fatalf("Block0 should contain 1 transaction, not %d", l)
	}
	if tx := block0.Transactions()[0]; !reflect.DeepEqual(tx, []byte("e21")) {
		t.Fatalf("Block0.Transactions[0] should be 'e21', not %s", tx)
	}

	frame1, err := p.GetFrame(block0.RoundReceived())
	frame1Hash, err := frame1.Hash()
	if !reflect.DeepEqual(block0.FrameHash(), frame1Hash) {
		t.Fatalf("Block0.FrameHash should be %v, not %v", frame1Hash, block0.FrameHash())
	}

	//Block 1 ------------------------------------------------------------------
	block1, err := p.Store.GetBlock(1)
	if err != nil {
		t.Fatalf("Store should contain a block with Index 1: %v", err)
	}

	if ind := block1.Index(); ind != 1 {
		t.Fatalf("Block1's Index should be 1, not %d", ind)
	}

	if rr := block1.RoundReceived(); rr != 2 {
		t.Fatalf("Block1's RoundReceived should be 2, not %d", rr)
	}

	if l := len(block1.Transactions()); l != 2 {
		t.Fatalf("Block1 should contain 2 transactions, not %d", l)
	}
	if tx := block1.Transactions()[1]; !reflect.DeepEqual(tx, []byte("f02b")) {
		t.Fatalf("Block1.Transactions[1] should be 'f02b', not %s", tx)
	}

	frame2, err := p.GetFrame(block1.RoundReceived())
	frame2Hash, err := frame2.Hash()
	if !reflect.DeepEqual(block1.FrameHash(), frame2Hash) {
		t.Fatalf("Block1.FrameHash should be %v, not %v", frame2Hash, block1.FrameHash())
	}

	// pendingRounds -----------------------------------------------------------
	expectedpendingRounds := []pendingRound{
		{
			Index:   3,
			Decided: false,
		},
		{
			Index:   4,
			Decided: false,
		},
	}
	for i, pd := range p.PendingRounds {
		if !reflect.DeepEqual(*pd, expectedpendingRounds[i]) {
			t.Fatalf("pendingRounds[%d] should be %v, not %v", i, expectedpendingRounds[i], *pd)
		}
	}

	//Anchor -------------------------------------------------------------------
	if v := p.AnchorBlock; v != nil {
		t.Fatalf("AnchorBlock should be nil, not %v", v)
	}

}

func BenchmarkConsensus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		//we do not want to benchmark the initialization code
		b.StopTimer()
		p, _ := initConsensusPoset(false, b)
		b.StartTimer()

		p.DivideRounds()
		p.DecideFame()
		p.DecideRoundReceived()
		p.ProcessDecidedRounds()
	}
}

func TestKnown(t *testing.T) {
	p, _ := initConsensusPoset(false, t)

	participants := p.Participants.ToPeerSlice()

	expectedKnown := map[int64]int64{
		participants[0].ID: 10,
		participants[1].ID: 9,
		participants[2].ID: 9,
	}

	known := p.Store.KnownEvents()
	for i := range p.Participants.ToIDSlice() {
		if l := known[int64(i)]; l != expectedKnown[int64(i)] {
			t.Fatalf("Known[%d] should be %d, not %d", i, expectedKnown[int64(i)], l)
		}
	}
}

func TestGetFrame(t *testing.T) {
	p, index := initConsensusPoset(false, t)

	participants := p.Participants.ToPeerSlice()

	p.DivideRounds()
	p.DecideFame()
	p.DecideRoundReceived()
	p.ProcessDecidedRounds()

	t.Run("Round 1", func(t *testing.T) {
		expectedRoots := make([]Root, n)
		expectedRoots[0] = NewBaseRoot(participants[0].ID)
		expectedRoots[1] = NewBaseRoot(participants[1].ID)
		expectedRoots[2] = NewBaseRoot(participants[2].ID)

		frame, err := p.GetFrame(1)
		if err != nil {
			t.Fatal(err)
		}

		for p, r := range frame.Roots {
			er := expectedRoots[p]
			if x := r.SelfParent; !reflect.DeepEqual(x, er.SelfParent) {
				t.Fatalf("Roots[%d].SelfParent should be %v, not %v", p, er.SelfParent, x)
			}
			if others := r.Others; !reflect.DeepEqual(others, er.Others) {
				t.Fatalf("Roots[%d].Others should be %v, not %vv", p, er.Others, others)
			}
		}

		expectedEventsHashes := []string{
			index["e0"],
			index["e1"],
			index["e2"],
			index["e10"],
			index["e21"],
			index["e21b"],
			index["e02"]}
		var expectedEvents []Event
		for _, eh := range expectedEventsHashes {
			e, err := p.Store.GetEvent(eh)
			if err != nil {
				t.Fatal(err)
			}
			expectedEvents = append(expectedEvents, e)
		}
		sort.Sort(ByLamportTimestamp(expectedEvents))
		if !reflect.DeepEqual(expectedEvents, frame.Events) {
			t.Fatal("Frame.Events is not good")
		}

		block0, err := p.Store.GetBlock(0)
		if err != nil {
			t.Fatalf("Store should contain a block with Index 1: %v", err)
		}
		frame1Hash, err := frame.Hash()
		if err != nil {
			t.Fatalf("Error computing Frame hash, %v", err)
		}
		if !reflect.DeepEqual(block0.FrameHash(), frame1Hash) {
			t.Fatalf("Block0.FrameHash (%v) and Frame1.Hash (%v) differ", block0.FrameHash(), frame1Hash)
		}
	})

	t.Run("Round 2", func(t *testing.T) {
		expectedRoots := make([]Root, n)
		expectedRoots[0] = Root{
			NextRound:  1,
			SelfParent: RootEvent{index["e02"], participants[0].ID, 1, 4, 0},
			Others: map[string]RootEvent{
				index["f0"]: {
					Hash:             index["f1b"],
					CreatorID:        participants[1].ID,
					Index:            3,
					LamportTimestamp: 6,
					Round:            1,
				},
				index["f0x"]: {
					Hash:             index["e21"],
					CreatorID:        participants[2].ID,
					Index:            1,
					LamportTimestamp: 2,
					Round:            0,
				},
			},
		}
		expectedRoots[1] = Root{
			NextRound:  1,
			SelfParent: RootEvent{index["e10"], participants[1].ID, 1, 1, 0},
			Others: map[string]RootEvent{
				index["f1"]: {
					Hash:             index["e02"],
					CreatorID:        participants[0].ID,
					Index:            1,
					LamportTimestamp: 4,
					Round:            0,
				},
			},
		}
		expectedRoots[2] = Root{
			NextRound:  1,
			SelfParent: RootEvent{index["e21b"], participants[2].ID, 2, 3, 0},
			Others: map[string]RootEvent{
				index["f2"]: {
					Hash:             index["f1b"],
					CreatorID:        participants[1].ID,
					Index:            3,
					LamportTimestamp: 6,
					Round:            1,
				},
			},
		}

		frame, err := p.GetFrame(2)
		if err != nil {
			t.Fatal(err)
		}

		for p, r := range frame.Roots {
			er := expectedRoots[p]
			if x := r.SelfParent; !reflect.DeepEqual(x, er.SelfParent) {
				t.Fatalf("Roots[%d].SelfParent should be %v, not %v", p, er.SelfParent, x)
			}

			if others := r.Others; !reflect.DeepEqual(others, er.Others) {
				t.Fatalf("Roots[%d].Others should be %v, not %v", p, er.Others, others)
			}
		}

		expectedEventsHashes := []string{
			index["f1"],
			index["f1b"],
			index["f0"],
			index["f2"],
			index["f10"],
			index["f0x"],
			index["f21"],
			index["f02"],
			index["f02b"]}
		var expectedEvents []Event
		for _, eh := range expectedEventsHashes {
			e, err := p.Store.GetEvent(eh)
			if err != nil {
				t.Fatal(err)
			}
			expectedEvents = append(expectedEvents, e)
		}
		sort.Sort(ByLamportTimestamp(expectedEvents))
		if !reflect.DeepEqual(expectedEvents, frame.Events) {
			t.Fatal("Frame.Events is not good")
		}
	})

}

func TestResetFromFrame(t *testing.T) {
	p, index := initConsensusPoset(false, t)

	participants := p.Participants.ToPeerSlice()

	p.DivideRounds()
	p.DecideFame()
	p.DecideRoundReceived()
	p.ProcessDecidedRounds()

	block, err := p.Store.GetBlock(1)
	if err != nil {
		t.Fatal(err)
	}

	frame, err := p.GetFrame(block.RoundReceived())
	if err != nil {
		t.Fatal(err)
	}

	//This operation clears the private fields which need to be recomputed
	//in the Events (round, roundReceived,etc)
	marshalledFrame, _ := frame.Marshal()
	unmarshalledFrame := new(Frame)
	unmarshalledFrame.Unmarshal(marshalledFrame)

	p2 := NewPoset(p.Participants,
		NewInmemStore(p.Participants, cacheSize),
		nil,
		testLogger(t))
	err = p2.Reset(block, *unmarshalledFrame)
	if err != nil {
		t.Fatal(err)
	}

	/*
		The poset should now look like this:

		   	   f02b|   |
		   	   |   |   |
		   	   f02 |   |
		   	   | \ |   |
		   	   |   \   |
		   	   |   | \ |
		   +--f0x  |   f21 //f0x's other-parent is e21b; contained in R0
		   |   |   | / |
		   |   |  f10  |
		   |   | / |   |
		   |   f0  |   f2
		   |   | \ | / |
		   |   |  f1b  |
		   |   |   |   |
		   |   |   f1  |
		   |   |   |   |
		   +-- R0  R1  R2
	*/

	//Test Known
	expectedKnown := map[int64]int64{
		participants[0].ID: 5,
		participants[1].ID: 4,
		participants[2].ID: 4,
	}

	known := p2.Store.KnownEvents()
	for _, peer := range p2.Participants.ById {
		if l := known[peer.ID]; l != expectedKnown[peer.ID] {
			t.Fatalf("Known[%d] should be %d, not %d", peer.ID, expectedKnown[peer.ID], l)
		}
	}

	/***************************************************************************
	 Test DivideRounds
	***************************************************************************/
	if err := p2.DivideRounds(); err != nil {
		t.Fatal(err)
	}

	pRound1, err := p.Store.GetRound(1)
	if err != nil {
		t.Fatal(err)
	}
	p2Round1, err := p2.Store.GetRound(1)
	if err != nil {
		t.Fatal(err)
	}

	//Check Round1 Witnesses
	pWitnesses := pRound1.Witnesses()
	p2Witnesses := p2Round1.Witnesses()
	sort.Strings(pWitnesses)
	sort.Strings(p2Witnesses)
	if !reflect.DeepEqual(pWitnesses, p2Witnesses) {
		t.Fatalf("Reset Hg Round 1 witnesses should be %v, not %v", pWitnesses, p2Witnesses)
	}

	//check Event Rounds and LamportTimestamps
	for _, ev := range frame.Events {
		p2r, err := p2.round(ev.Hex())
		if err != nil {
			t.Fatalf("Error computing %s Round: %d", getName(index, ev.Hex()), p2r)
		}
		hr, _ := p.round(ev.Hex())
		if p2r != hr {

			t.Fatalf("p2[%v].Round should be %d, not %d", getName(index, ev.Hex()), hr, p2r)
		}

		p2s, err := p2.lamportTimestamp(ev.Hex())
		if err != nil {
			t.Fatalf("Error computing %s LamportTimestamp: %d", getName(index, ev.Hex()), p2s)
		}
		hs, _ := p.lamportTimestamp(ev.Hex())
		if p2s != hs {
			t.Fatalf("p2[%v].LamportTimestamp should be %d, not %d", getName(index, ev.Hex()), hs, p2s)
		}
	}

	/***************************************************************************
	Test Consensus
	***************************************************************************/
	p2.DecideFame()
	p2.DecideRoundReceived()
	p2.ProcessDecidedRounds()

	if lbi := p2.Store.LastBlockIndex(); lbi != block.Index() {
		t.Fatalf("LastBlockIndex should be %d, not %d", block.Index(), lbi)
	}

	if r := p2.LastConsensusRound; r == nil || *r != block.RoundReceived() {
		t.Fatalf("LastConsensusRound should be %d, not %d", block.RoundReceived(), *r)
	}

	if v := p2.AnchorBlock; v != nil {
		t.Fatalf("AnchorBlock should be nil, not %v", v)
	}

	/***************************************************************************
	Test continue after Reset
	***************************************************************************/
	//Insert remaining Events into the Reset poset
	for r := int64(2); r <= int64(4); r++ {
		round, err := p.Store.GetRound(r)
		if err != nil {
			t.Fatal(err)
		}

		var events []Event
		for _, e := range round.RoundEvents() {
			ev, err := p.Store.GetEvent(e)
			if err != nil {
				t.Fatal(err)
			}
			events = append(events, ev)
			t.Logf("R%d %s", r, getName(index, e))
		}

		sort.Sort(ByTopologicalOrder(events))

		for _, ev := range events {

			marshalledEv, _ := ev.Marshal()
			unmarshalledEv := new(Event)
			unmarshalledEv.Unmarshal(marshalledEv)

			err = p2.InsertEvent(*unmarshalledEv, true)
			if err != nil {
				t.Fatalf("ERR Inserting Event %s: %v", getName(index, ev.Hex()), err)
			}
		}
	}

	p2.DivideRounds()
	p2.DecideFame()
	p2.DecideRoundReceived()
	p2.ProcessDecidedRounds()

	for r := int64(1); r <= 4; r++ {
		pRound, err := p.Store.GetRound(r)
		if err != nil {
			t.Fatal(err)
		}
		p2Round, err := p2.Store.GetRound(r)
		if err != nil {
			t.Fatal(err)
		}

		pWitnesses := pRound.Witnesses()
		p2Witnesses := p2Round.Witnesses()
		sort.Strings(pWitnesses)
		sort.Strings(p2Witnesses)

		if !reflect.DeepEqual(pWitnesses, p2Witnesses) {
			t.Fatalf("Reset Hg Round %d witnesses should be %v, not %v", r, pWitnesses, p2Witnesses)
		}
	}
}

func TestBootstrap(t *testing.T) {

	//Initialize a first Poset with a DB backend
	//Add events and run consensus methods on it
	p, _ := initConsensusPoset(true, t)
	p.DivideRounds()
	p.DecideFame()
	p.DecideRoundReceived()
	p.ProcessDecidedRounds()

	p.Store.Close()
	defer os.RemoveAll(badgerDir)

	//Now we want to create a new Poset based on the database of the previous
	//Poset and see if we can boostrap it to the same state.
	recycledStore, err := LoadBadgerStore(cacheSize, badgerDir)
	np := NewPoset(recycledStore.participants,
		recycledStore,
		nil,
		logrus.New().WithField("id", "bootstrapped"))
	err = np.Bootstrap()
	if err != nil {
		t.Fatal(err)
	}

	hConsensusEvents := p.Store.ConsensusEvents()
	nhConsensusEvents := np.Store.ConsensusEvents()
	if len(hConsensusEvents) != len(nhConsensusEvents) {
		t.Fatalf("Bootstrapped poset should contain %d consensus events,not %d",
			len(hConsensusEvents), len(nhConsensusEvents))
	}

	hKnown := p.Store.KnownEvents()
	nhKnown := np.Store.KnownEvents()
	if !reflect.DeepEqual(hKnown, nhKnown) {
		t.Fatalf("Bootstrapped poset's Known should be %#v, not %#v",
			hKnown, nhKnown)
	}

	if *p.LastConsensusRound != *np.LastConsensusRound {
		t.Fatalf("Bootstrapped poset's LastConsensusRound should be %#v, not %#v",
			*p.LastConsensusRound, *np.LastConsensusRound)
	}

	if p.LastCommitedRoundEvents != np.LastCommitedRoundEvents {
		t.Fatalf("Bootstrapped poset's LastCommitedRoundEvents should be %#v, not %#v",
			p.LastCommitedRoundEvents, np.LastCommitedRoundEvents)
	}

	if p.ConsensusTransactions != np.ConsensusTransactions {
		t.Fatalf("Bootstrapped poset's ConsensusTransactions should be %#v, not %#v",
			p.ConsensusTransactions, np.ConsensusTransactions)
	}

	if p.PendingLoadedEvents != np.PendingLoadedEvents {
		t.Fatalf("Bootstrapped poset's PendingLoadedEvents should be %#v, not %#v",
			p.PendingLoadedEvents, np.PendingLoadedEvents)
	}
}

/*

	This example demonstrates that a Round can be 'decided' before an earlier
	round. Here, rounds 1 and 2 are decided before round 0 because the fame of
	witness w00 is only decided at round 5.

--------------------------------------------------------------------------------
	|   w51   |    | This section is only added in 'full' mode
    |    |  \ |    | w51 collects votes from w40, w41, w42, and w43. It DECIDES
	|    |   e23   | yes.
----------------\---------------------------------------------------------------
	|    |    |   w43
	|    |    | /  | Round 4 is a Coin Round [(4 -0) mod 4 = 0].
    |    |   w42   | No decision will be made.
    |    | /  |    | w40 collects votes from w33, w32 and w31. It votes yes.
    |   w41   |    | w41 collects votes from w33, w32 and w31. It votes yes.
	| /  |    |    | w42 collects votes from w30, w31, w32 and w33. It votes yes.
   w40   |    |    | w43 collects votes from w30, w31, w32 and w33. It votes yes.
    | \  |    |    |------------------------
    |   d13   |    | w30 collects votes from w20, w21, w22 and w23. It votes yes
    |    |  \ |    | w31 collects votes from w21, w22 and w23. It votes no
   w30   |    \    | w32 collects votes from w20, w21, w22 and w23. It votes yes
    | \  |    | \  | w33 collects votes from w20, w21, w22 and w23. It votes yes
    |   \     |   w33
    |    | \  |  / |Again, none of the witnesses in round 3 are able to decide.
    |    |   w32   |However, a strong majority votes yes
    |    |  / |    |
	|   w31   |    |
    |  / |    |    |--------------------------
   w20   |    |    | w23 collects votes from w11, w12 and w13. It votes no
    |  \ |    |    | w21 collects votes from w11, w12, and w13. It votes no
    |    \    |    | w22 collects votes from w11, w12, w13 and w14. It votes yes
    |    | \  |    | w20 collects votes from w11, w12, w13 and w14. It votes yes
    |    |   w22   |
    |    | /  |    | None of the witnesses in round 2 were able to decide.
    |   c10   |    | They voted according to the majority of votes they observed
    | /  |    |    | in round 1. The vote is split 2-2
   b00  w21   |    |
    |    |  \ |    |
    |    |    \    |
    |    |    | \  |
    |    |    |   w23
    |    |    | /  |------------------------
   w10   |   b21   |
	| \  | /  |    | w10 votes yes (it can see w00)
    |   w11   |    | w11 votes yes
    |    |  \ |    | w12 votes no  (it cannot see w00)
	|    |   w12   | w13 votes no
    |    |    | \  |
    |    |    |   w13
    |    |    | /  |------------------------
    |   a10  a21   | We want to decide the fame of w00
    |  / |  / |    |
    |/  a12   |    |
   a00   |  \ |    |
	|    |   a23   |
    |    |    | \  |
   w00  w01  w02  w03
	0	 1	  2	   3
*/

func initFunkyPoset(logger *logrus.Logger, full bool) (*Poset, map[string]string) {
	nodes, index, orderedEvents, participants := initPosetNodes(4)

	for i, peer := range participants.ToPeerSlice() {
		name := fmt.Sprintf("w0%d", i)
		event := NewEvent([][]byte{[]byte(name)}, nil, nil, []string{rootSelfParent(peer.ID), ""}, nodes[i].Pub, 0, nil)
		nodes[i].signAndAddEvent(event, name, index, orderedEvents)
	}

	plays := []play{
		{2, 1, "w02", "w03", "a23", [][]byte{[]byte("a23")}, nil},
		{1, 1, "w01", "a23", "a12", [][]byte{[]byte("a12")}, nil},
		{0, 1, "w00", "", "a00", [][]byte{[]byte("a00")}, nil},
		{1, 2, "a12", "a00", "a10", [][]byte{[]byte("a10")}, nil},
		{2, 2, "a23", "a12", "a21", [][]byte{[]byte("a21")}, nil},
		{3, 1, "w03", "a21", "w13", [][]byte{[]byte("w13")}, nil},
		{2, 3, "a21", "w13", "w12", [][]byte{[]byte("w12")}, nil},
		{1, 3, "a10", "w12", "w11", [][]byte{[]byte("w11")}, nil},
		{0, 2, "a00", "w11", "w10", [][]byte{[]byte("w10")}, nil},
		{2, 4, "w12", "w11", "b21", [][]byte{[]byte("b21")}, nil},
		{3, 2, "w13", "b21", "w23", [][]byte{[]byte("w23")}, nil},
		{1, 4, "w11", "w23", "w21", [][]byte{[]byte("w21")}, nil},
		{0, 3, "w10", "", "b00", [][]byte{[]byte("b00")}, nil},
		{1, 5, "w21", "b00", "c10", [][]byte{[]byte("c10")}, nil},
		{2, 5, "b21", "c10", "w22", [][]byte{[]byte("w22")}, nil},
		{0, 4, "b00", "w22", "w20", [][]byte{[]byte("w20")}, nil},
		{1, 6, "c10", "w20", "w31", [][]byte{[]byte("w31")}, nil},
		{2, 6, "w22", "w31", "w32", [][]byte{[]byte("w32")}, nil},
		{0, 5, "w20", "w32", "w30", [][]byte{[]byte("w30")}, nil},
		{3, 3, "w23", "w32", "w33", [][]byte{[]byte("w33")}, nil},
		{1, 7, "w31", "w33", "d13", [][]byte{[]byte("d13")}, nil},
		{0, 6, "w30", "d13", "w40", [][]byte{[]byte("w40")}, nil},
		{1, 8, "d13", "w40", "w41", [][]byte{[]byte("w41")}, nil},
		{2, 7, "w32", "w41", "w42", [][]byte{[]byte("w42")}, nil},
		{3, 4, "w33", "w42", "w43", [][]byte{[]byte("w43")}, nil},
	}
	if full {
		newPlays := []play{
			{2, 8, "w42", "w43", "e23", [][]byte{[]byte("e23")}, nil},
			{1, 9, "w41", "e23", "w51", [][]byte{[]byte("w51")}, nil},
		}
		plays = append(plays, newPlays...)
	}

	playEvents(plays, nodes, index, orderedEvents)

	poset := createPoset(false, orderedEvents, participants, logger.WithField("test", 6))

	return poset, index
}

func TestFunkyPosetFame(t *testing.T) {
	p, index := initFunkyPoset(common.NewTestLogger(t), false)

	if err := p.DivideRounds(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideFame(); err != nil {
		t.Fatal(err)
	}

	if l := p.Store.LastRound(); l != 4 {
		t.Fatalf("last round should be 4 not %d", l)
	}

	for r := int64(0); r < 5; r++ {
		round, err := p.Store.GetRound(r)
		if err != nil {
			t.Fatal(err)
		}
		var witnessNames []string
		for _, w := range round.Witnesses() {
			witnessNames = append(witnessNames, getName(index, w))
		}
		t.Logf("Round %d witnesses: %v", r, witnessNames)
	}

	//Rounds 1 and 2 should get decided BEFORE round 0
	expectedpendingRounds := []pendingRound{
		{
			Index:   0,
			Decided: false,
		},
		{
			Index:   1,
			Decided: true,
		},
		{
			Index:   2,
			Decided: true,
		},
		{
			Index:   3,
			Decided: false,
		},
		{
			Index:   4,
			Decided: false,
		},
	}

	for i, pd := range p.PendingRounds {
		if !reflect.DeepEqual(*pd, expectedpendingRounds[i]) {
			t.Fatalf("pendingRounds[%d] should be %v, not %v", i, expectedpendingRounds[i], *pd)
		}
	}

	if err := p.DecideRoundReceived(); err != nil {
		t.Fatal(err)
	}
	if err := p.ProcessDecidedRounds(); err != nil {
		t.Fatal(err)
	}

	//But a dicided round should never be processed until all previous rounds
	//are decided. So the PendingQueue should remain the same after calling
	//ProcessDecidedRounds()

	for i, pd := range p.PendingRounds {
		if !reflect.DeepEqual(*pd, expectedpendingRounds[i]) {
			t.Fatalf("pendingRounds[%d] should be %v, not %v", i, expectedpendingRounds[i], *pd)
		}
	}
}

func TestFunkyPosetBlocks(t *testing.T) {
	p, index := initFunkyPoset(common.NewTestLogger(t), true)

	if err := p.DivideRounds(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideFame(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideRoundReceived(); err != nil {
		t.Fatal(err)
	}
	if err := p.ProcessDecidedRounds(); err != nil {
		t.Fatal(err)
	}

	if l := p.Store.LastRound(); l != 5 {
		t.Fatalf("last round should be 5 not %d", l)
	}

	for r := int64(0); r < 6; r++ {
		round, err := p.Store.GetRound(r)
		if err != nil {
			t.Fatal(err)
		}
		var witnessNames []string
		for _, w := range round.Witnesses() {
			witnessNames = append(witnessNames, getName(index, w))
		}
		t.Logf("Round %d witnesses: %v", r, witnessNames)
	}

	//rounds 0,1, 2 and 3 should be decided
	expectedpendingRounds := []pendingRound{
		{
			Index:   4,
			Decided: false,
		},
		{
			Index:   5,
			Decided: false,
		},
	}
	for i, pd := range p.PendingRounds {
		if !reflect.DeepEqual(*pd, expectedpendingRounds[i]) {
			t.Fatalf("pendingRounds[%d] should be %v, not %v", i, expectedpendingRounds[i], *pd)
		}
	}

	expectedBlockTxCounts := map[int64]int64{
		0: 6,
		1: 7,
		2: 7,
	}

	for bi := int64(0); bi < 3; bi++ {
		b, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}
		for i, tx := range b.Transactions() {
			t.Logf("block %d, tx %d: %s", bi, i, string(tx))
		}
		if txs := int64(len(b.Transactions())); txs != expectedBlockTxCounts[bi] {
			t.Fatalf("Blocks[%d] should contain %d transactions, not %d", bi,
				expectedBlockTxCounts[bi], txs)
		}
	}
}

func TestFunkyPosetFrames(t *testing.T) {
	p, index := initFunkyPoset(common.NewTestLogger(t), true)

	participants := p.Participants.ToPeerSlice()

	if err := p.DivideRounds(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideFame(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideRoundReceived(); err != nil {
		t.Fatal(err)
	}
	if err := p.ProcessDecidedRounds(); err != nil {
		t.Fatal(err)
	}

	t.Logf("------------------------------------------------------------------")
	for bi := int64(0); bi < 3; bi++ {
		block, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}

		frame, err := p.GetFrame(block.RoundReceived())
		for k, ev := range frame.Events {
			r, _ := p.round(ev.Hex())
			t.Logf("frame[%d].Events[%d]: %s, round %d", frame.Round, k, getName(index, ev.Hex()), r)
		}
		for k, r := range frame.Roots {
			t.Logf("frame[%d].Roots[%d]: SelfParent: %v, Others: %v",
				frame.Round, k, r.SelfParent, r.Others)
		}
	}
	t.Logf("------------------------------------------------------------------")

	expectedFrameRoots := map[int64][]Root{
		1: {
			NewBaseRoot(participants[0].ID),
			NewBaseRoot(participants[1].ID),
			NewBaseRoot(participants[2].ID),
			NewBaseRoot(participants[3].ID),
		},
		2: {
			NewBaseRoot(participants[0].ID),
			{
				NextRound:  0,
				SelfParent: RootEvent{index["a12"], participants[1].ID, 1, 2, 0},
				Others: map[string]RootEvent{
					index["a10"]: {index["a00"], participants[0].ID, 1, 1, 0},
				},
			},
			{
				NextRound:  1,
				SelfParent: RootEvent{index["a21"], participants[2].ID, 2, 3, 0},
				Others: map[string]RootEvent{
					index["w12"]: {index["w13"], participants[3].ID, 1, 4, 1},
				},
			},
			{
				NextRound:  1,
				SelfParent: RootEvent{index["w03"], participants[3].ID, 0, 0, 0},
				Others: map[string]RootEvent{
					index["w13"]: {index["a21"], participants[2].ID, 2, 3, 0},
				},
			},
		},
		3: {
			{
				NextRound:  1,
				SelfParent: RootEvent{index["a00"], participants[0].ID, 1, 1, 0},
				Others: map[string]RootEvent{
					index["w10"]: {index["w11"], participants[1].ID, 3, 6, 1},
				},
			},
			{
				NextRound:  2,
				SelfParent: RootEvent{index["w11"], participants[1].ID, 3, 6, 1},
				Others: map[string]RootEvent{
					index["w21"]: {index["w23"], participants[3].ID, 2, 8, 2},
				},
			},
			{
				NextRound:  2,
				SelfParent: RootEvent{index["b21"], participants[2].ID, 4, 7, 1},
				Others: map[string]RootEvent{
					index["w22"]: {index["c10"], participants[1].ID, 5, 10, 2},
				},
			},
			{
				NextRound:  2,
				SelfParent: RootEvent{index["w13"], participants[3].ID, 1, 4, 1},
				Others: map[string]RootEvent{
					index["w23"]: {index["b21"], participants[2].ID, 4, 7, 1},
				},
			},
		},
	}

	for bi := int64(0); bi < 3; bi++ {
		block, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}

		frame, err := p.GetFrame(block.RoundReceived())
		if err != nil {
			t.Fatal(err)
		}

		for k, r := range frame.Roots {
			if !reflect.DeepEqual(expectedFrameRoots[frame.Round][k], r) {
				t.Fatalf("frame[%d].Roots[%d] should be %v, not %v", frame.Round, k, expectedFrameRoots[frame.Round][k], r)
			}
		}
	}
}

func TestFunkyPosetReset(t *testing.T) {
	p, index := initFunkyPoset(common.NewTestLogger(t), true)

	p.DivideRounds()
	p.DecideFame()
	p.DecideRoundReceived()
	p.ProcessDecidedRounds()

	for bi := int64(0); bi < 3; bi++ {
		t.Logf("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		t.Logf("RESETTING FROM BLOCK %d", bi)
		t.Logf("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

		block, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}

		frame, err := p.GetFrame(block.RoundReceived())
		if err != nil {
			t.Fatal(err)
		}

		//This operation clears the private fields which need to be recomputed
		//in the Events (round, roundReceived,etc)
		marshalledFrame, _ := frame.Marshal()
		unmarshalledFrame := new(Frame)
		unmarshalledFrame.Unmarshal(marshalledFrame)

		p2 := NewPoset(p.Participants,
			NewInmemStore(p.Participants, cacheSize),
			nil,
			testLogger(t))
		err = p2.Reset(block, *unmarshalledFrame)
		if err != nil {
			t.Fatal(err)
		}

		/***********************************************************************
		Test continue after Reset
		***********************************************************************/

		//Compute diff
		p2Known := p2.Store.KnownEvents()
		diff := getDiff(p, p2Known, t)

		wireDiff := make([]WireEvent, len(diff), len(diff))
		for i, e := range diff {
			wireDiff[i] = e.ToWire()
		}

		//Insert remaining Events into the Reset poset
		for i, wev := range wireDiff {
			ev, err := p2.ReadWireInfo(wev)
			if err != nil {
				t.Fatalf("Reading WireInfo for %s: %s", getName(index, diff[i].Hex()), err)
			}
			err = p2.InsertEvent(*ev, false)
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Logf("RUN CONSENSUS METHODS*****************************************")
		p2.DivideRounds()
		p2.DecideFame()
		p2.DecideRoundReceived()
		p2.ProcessDecidedRounds()
		t.Logf("**************************************************************")

		compareRoundWitnesses(p, p2, index, bi, true, t)
	}

}

/*


ATTENTION: Look at roots in Rounds 1 and 2

    |   w51   |    |
  ----------- \  ---------------------
	|    |      \  |  Round 4
    |    |    |   i32
    |    |    | /  |
    |    |   w42   |
    |    |  / |    |
    |   w41   |    |
	|    |   \     |
	|    |    | \  |
    |    |    |   w43
--------------------------------------
	|    |    | /  |  Round 3
    |    |   h21   |
    |    | /  |    |
    |   w31   |    |
	|    |   \     |
	|    |    | \  |
    |    |    |   w33
    |    |    | /  |
    |    |   w32   |
--------------------------------------
	|    | /  |    |  Round 2
    |   g13   |    |  ConsensusRound 3
	|    |   \     |
	|    |    | \  |  Frame {
    |    |    |   w23 	Evs  : w21, w22, w23, g13
    |    |    | /  |	Roots: [w00, e32], [w11, w13], [w12, w21], [w13, w22]
    |    |   w22   |  }
    |    | /  |    |
    |   w21   |    |
 ----------- \  ---------------------
	|    |      \  |  Round 1
    |    |    |   w13 ConsensusRound 2
    |    |    | /  |
    |    |   w12   |  Frame {
    |     /   |    |	Evs  : w10, w11, f01, w12, w13
	|  / |    |    |	Roots: [w00,e32], [w10, e10], [w11, e21], [w12, e32]
   f01   |    |    |  }
	| \  |    |    |
    |   w11   |    |
    | /  |    |    |
   w10   |    |    |
-------------------------------------
    |    \    |    |  Round 0
    |    |    \    |  ConsensusRound 1
    |    |    |   e32
    |    |    | /  |  Frame {
    |    |   e21   |     Evs  : w00, w01, w02, w03, e10, e21, e32
    |    | /  |    |     Roots: R0, R1, R2, R3
    |   e10   |    |  }
    |  / |    |    |
   w00  w01  w02  w03
	|    |    |    |
    R0   R1   R2   R3
	0	 1	  2	   3
*/

func initSparsePoset(logger *logrus.Logger) (*Poset, map[string]string) {
	nodes, index, orderedEvents, participants := initPosetNodes(4)

	for i, peer := range participants.ToPeerSlice() {
		name := fmt.Sprintf("w0%d", i)
		event := NewEvent([][]byte{[]byte(name)}, nil, nil, []string{rootSelfParent(peer.ID), ""}, nodes[i].Pub, 0, nil)
		nodes[i].signAndAddEvent(event, name, index, orderedEvents)
	}

	plays := []play{
		{1, 1, "w01", "w00", "e10", [][]byte{[]byte("e10")}, nil},
		{2, 1, "w02", "e10", "e21", [][]byte{[]byte("e21")}, nil},
		{3, 1, "w03", "e21", "e32", [][]byte{[]byte("e32")}, nil},
		{0, 1, "w00", "e32", "w10", [][]byte{[]byte("w10")}, nil},
		{1, 2, "e10", "w10", "w11", [][]byte{[]byte("w11")}, nil},
		{0, 2, "w10", "w11", "f01", [][]byte{[]byte("f01")}, nil},
		{2, 2, "e21", "f01", "w12", [][]byte{[]byte("w12")}, nil},
		{3, 2, "e32", "w12", "w13", [][]byte{[]byte("w13")}, nil},
		{1, 3, "w11", "w13", "w21", [][]byte{[]byte("w21")}, nil},
		{2, 3, "w12", "w21", "w22", [][]byte{[]byte("w22")}, nil},
		{3, 3, "w13", "w22", "w23", [][]byte{[]byte("w23")}, nil},
		{1, 4, "w21", "w23", "g13", [][]byte{[]byte("g13")}, nil},
		{2, 4, "w22", "g13", "w32", [][]byte{[]byte("w32")}, nil},
		{3, 4, "w23", "w32", "w33", [][]byte{[]byte("w33")}, nil},
		{1, 5, "g13", "w33", "w31", [][]byte{[]byte("w31")}, nil},
		{2, 5, "w32", "w31", "h21", [][]byte{[]byte("h21")}, nil},
		{3, 5, "w33", "h21", "w43", [][]byte{[]byte("w43")}, nil},
		{1, 6, "w31", "w43", "w41", [][]byte{[]byte("w41")}, nil},
		{2, 6, "h21", "w41", "w42", [][]byte{[]byte("w42")}, nil},
		{3, 6, "w43", "w42", "i32", [][]byte{[]byte("i32")}, nil},
		{1, 7, "w41", "i32", "w51", [][]byte{[]byte("w51")}, nil},
	}

	for _, p := range plays {
		e := NewEvent(p.txPayload,
			nil,
			p.sigPayload,
			[]string{index[p.selfParent], index[p.otherParent]},
			nodes[p.to].Pub,
			p.index, nil)
		nodes[p.to].signAndAddEvent(e, p.name, index, orderedEvents)
	}

	playEvents(plays, nodes, index, orderedEvents)

	poset := createPoset(false, orderedEvents, participants, logger.WithField("test", 6))

	return poset, index
}

/*
func TestSparsePosetFrames(t *testing.T) {
	p, index := initSparsePoset(common.NewTestLogger(t))

	participants := p.Participants.ToPeerSlice()

	if err := p.DivideRounds(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideFame(); err != nil {
		t.Fatal(err)
	}
	if err := p.DecideRoundReceived(); err != nil {
		t.Fatal(err)
	}
	if err := p.ProcessDecidedRounds(); err != nil {
		t.Fatal(err)
	}

	t.Logf("------------------------------------------------------------------")
	for bi := int64(0); bi < 3; bi++ {
		block, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}

		frame, err := p.GetFrame(block.RoundReceived())
		for k, ev := range frame.Events {
			r, _ := p.round(ev.Hex())
			t.Logf("frame[%d].Events[%d]: %s, round %d", frame.Round, k, getName(index, ev.Hex()), r)
		}
		for k, r := range frame.Roots {
			t.Logf("frame[%d].Roots[%d]: SelfParent: %v, Others: %v",
				frame.Round, k, r.SelfParent, r.Others)
		}
	}
	t.Logf("------------------------------------------------------------------")

	expectedFrameRoots := map[int64][]Root{
		1: {
			NewBaseRoot(participants[0].ID),
			NewBaseRoot(participants[1].ID),
			NewBaseRoot(participants[2].ID),
			NewBaseRoot(participants[3].ID),
		},
		2: {
			{
				NextRound:  1,
				SelfParent: RootEvent{index["w00"], participants[0].ID, 0, 0, 0},
				Others: map[string]RootEvent{
					index["w10"]: {index["e32"], participants[3].ID, 1, 3, 0},
				},
			},
			{
				NextRound:  1,
				SelfParent: RootEvent{index["e10"], participants[1].ID, 1, 1, 0},
				Others: map[string]RootEvent{
					index["w11"]: {index["w10"], participants[0].ID, 1, 4, 1},
				},
			},
			{
				NextRound:  1,
				SelfParent: RootEvent{index["e21"], participants[2].ID, 1, 2, 0},
				Others: map[string]RootEvent{
					index["w12"]: {index["f01"], participants[0].ID, 2, 6, 1},
				},
			},
			{
				NextRound:  1,
				SelfParent: RootEvent{index["e32"], participants[3].ID, 1, 3, 0},
				Others: map[string]RootEvent{
					index["w13"]: {index["w12"], participants[2].ID, 2, 7, 1},
				},
			},
		},
		3: {
			{
				NextRound:  1,
				SelfParent: RootEvent{index["w10"], participants[0].ID, 1, 4, 1},
				Others: map[string]RootEvent{
					index["f01"]: {index["w11"], participants[1].ID, 2, 5, 1},
				},
			},
			{
				NextRound:  2,
				SelfParent: RootEvent{index["w11"], participants[1].ID, 2, 5, 1},
				Others: map[string]RootEvent{
					index["w21"]: {index["w13"], participants[3].ID, 2, 8, 1},
				},
			},
			{
				NextRound:  2,
				SelfParent: RootEvent{index["w12"], participants[2].ID, 2, 7, 1},
				Others: map[string]RootEvent{
					index["w22"]: {index["w21"], participants[1].ID, 3, 9, 2},
				},
			},
			{
				NextRound:  2,
				SelfParent: RootEvent{index["w13"], participants[3].ID, 2, 8, 1},
				Others: map[string]RootEvent{
					index["w23"]: {index["w22"], participants[2].ID, 3, 10, 2},
				},
			},
		},
	}

	for bi := int64(0); bi < 3; bi++ {
		block, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}

		frame, err := p.GetFrame(block.RoundReceived())
		if err != nil {
			t.Fatal(err)
		}

		for k, r := range frame.Roots {
			if !reflect.DeepEqual(expectedFrameRoots[frame.Round][k], r) {
				t.Fatalf("frame[%d].Roots[%d] should be %v, not %v", frame.Round, k, expectedFrameRoots[frame.Round][k], r)
			}
		}
	}
}
*/

/*
func TestSparsePosetReset(t *testing.T) {
	p, index := initSparsePoset(common.NewTestLogger(t))

	p.DivideRounds()
	p.DecideFame()
	p.DecideRoundReceived()
	p.ProcessDecidedRounds()

	for bi := int64(0); bi < 3; bi++ {
		t.Logf("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		t.Logf("RESETTING FROM BLOCK %d", bi)
		t.Logf("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

		block, err := p.Store.GetBlock(bi)
		if err != nil {
			t.Fatal(err)
		}

		frame, err := p.GetFrame(block.RoundReceived())
		if err != nil {
			t.Fatal(err)
		}

		//This operation clears the private fields which need to be recomputed
		//in the Events (round, roundReceived,etc)
		marshalledFrame, _ := frame.Marshal()
		unmarshalledFrame := new(Frame)
		unmarshalledFrame.Unmarshal(marshalledFrame)

		p2 := NewPoset(p.Participants,
			NewInmemStore(p.Participants, cacheSize),
			nil,
			testLogger(t))
		err = p2.Reset(block, *unmarshalledFrame)
		if err != nil {
			t.Fatal(err)
		}


		//Test continue after Reset


		//Compute diff
		p2Known := p2.Store.KnownEvents()
		diff := getDiff(p, p2Known, t)

		t.Logf("p2.Known: %v", p2Known)
		t.Logf("diff: %v", len(diff))

		wireDiff := make([]WireEvent, len(diff), len(diff))
		for i, e := range diff {
			wireDiff[i] = e.ToWire()
		}

		//Insert remaining Events into the Reset poset
		for i, wev := range wireDiff {
			eventName := getName(index, diff[i].Hex())
			ev, err := p2.ReadWireInfo(wev)
			if err != nil {
				t.Fatalf("ReadWireInfo(%s): %s", eventName, err)
			}
			if !reflect.DeepEqual(ev.Message.Body, diff[i].Message.Body) {
				t.Fatalf("%s from WireInfo should be %#v, not %#v", eventName, diff[i].Message.Body, ev.Message.Body)
			}
			err = p2.InsertEvent(*ev, false)
			if err != nil {
				t.Fatalf("InsertEvent(%s): %s", eventName, err)
			}
		}

		t.Logf("RUN CONSENSUS METHODS*****************************************")
		p2.DivideRounds()
		p2.DecideFame()
		p2.DecideRoundReceived()
		p2.ProcessDecidedRounds()
		t.Logf("**************************************************************")

		compareRoundWitnesses(p, p2, index, bi, true, t)
	}

}
*/

func compareRoundWitnesses(p, p2 *Poset, index map[string]string, round int64, check bool, t *testing.T) {

	for i := round; i <= 5; i++ {
		pRound, err := p.Store.GetRound(i)
		if err != nil {
			t.Fatal(err)
		}
		p2Round, err := p2.Store.GetRound(i)
		if err != nil {
			t.Fatal(err)
		}

		//Check Round1 Witnesses
		pWitnesses := pRound.Witnesses()
		p2Witnesses := p2Round.Witnesses()
		sort.Strings(pWitnesses)
		sort.Strings(p2Witnesses)
		hwn := make([]string, len(pWitnesses))
		p2wn := make([]string, len(p2Witnesses))
		for _, w := range pWitnesses {
			hwn = append(hwn, getName(index, w))
		}
		for _, w := range p2Witnesses {
			p2wn = append(p2wn, getName(index, w))
		}

		t.Logf("h Round%d witnesses: %v", i, hwn)
		t.Logf("p2 Round%d witnesses: %v", i, p2wn)

		if check && !reflect.DeepEqual(hwn, p2wn) {
			t.Fatalf("Reset Hg Round %d witnesses should be %v, not %v", i, hwn, p2wn)
		}
	}

}

func getDiff(p *Poset, known map[int64]int64, t *testing.T) []Event {
	var diff []Event
	for id, ct := range known {
		pk := p.Participants.ById[id].PubKeyHex
		//get participant Events with index > ct
		participantEvents, err := p.Store.ParticipantEvents(pk, ct)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range participantEvents {
			ev, err := p.Store.GetEvent(e)
			if err != nil {
				t.Fatal(err)
			}
			diff = append(diff, ev)
		}
	}
	sort.Sort(ByTopologicalOrder(diff))
	return diff
}

func getName(index map[string]string, hash string) string {
	for name, h := range index {
		if h == hash {
			return name
		}
	}
	return ""
}

func disp(index map[string]string, events []string) string {
	var names []string
	for _, h := range events {
		names = append(names, getName(index, h))
	}
	return fmt.Sprintf("[%s]", strings.Join(names, " "))
}

func create(x int) *int {
	return &x
}
