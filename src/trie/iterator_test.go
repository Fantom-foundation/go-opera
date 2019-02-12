package trie

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

func TestIterator(t *testing.T) {
	trie := newEmpty()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"dog", "puppy"},
		{"somethingveryoddindeedthis is", "myothernodedata"},
	}
	all := make(map[string]string)
	for _, val := range vals {
		all[val.k] = val.v
		trie.Update([]byte(val.k), []byte(val.v))
	}
	_, err := trie.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]string)
	it := NewIterator(trie.NodeIterator(nil))
	for it.Next() {
		found[string(it.Key)] = string(it.Value)
	}

	for k, v := range all {
		if found[k] != v {
			t.Errorf("iterator value mismatch for %s: got %q want %q", k, found[k], v)
		}
	}
}

type kv struct {
	k, v []byte
	t    bool
}

func TestIteratorLargeData(t *testing.T) {
	trie := newEmpty()
	vals := make(map[string]*kv)

	for i := byte(0); i < 255; i++ {
		value := &kv{common.LeftPadBytes([]byte{i}, 32), []byte{i}, false}
		value2 := &kv{common.LeftPadBytes([]byte{10, i}, 32), []byte{i}, false}
		trie.Update(value.k, value.v)
		trie.Update(value2.k, value2.v)
		vals[string(value.k)] = value
		vals[string(value2.k)] = value2
	}

	it := NewIterator(trie.NodeIterator(nil))
	for it.Next() {
		vals[string(it.Key)].t = true
	}

	var untouched []*kv
	for _, value := range vals {
		if !value.t {
			untouched = append(untouched, value)
		}
	}

	if len(untouched) > 0 {
		t.Errorf("Missed %d nodes", len(untouched))
		for _, value := range untouched {
			t.Error(value)
		}
	}
}

// Tests that the node iterator indeed walks over the entire database contents.
func TestNodeIteratorCoverage(t *testing.T) {
	// Create some arbitrary test trie to iterate
	db, trie, _ := makeTestTrie(t)

	// Gather all the node hashes found by the iterator
	hashes := make(map[common.Hash]struct{})
	for it := trie.NodeIterator(nil); it.Next(true); {
		if it.Hash() != (common.Hash{}) {
			hashes[it.Hash()] = struct{}{}
		}
	}
	// Cross check the hashes and the database itself
	for hash := range hashes {
		if _, err := db.Node(hash); err != nil {
			t.Errorf("failed to retrieve reported node %x: %v", hash, err)
		}
	}
	for hash, obj := range db.dirties {
		if obj != nil && hash != (common.Hash{}) {
			if _, ok := hashes[hash]; !ok {
				t.Errorf("state entry not reported %x", hash)
			}
		}
	}
	for _, key := range db.diskdb.(*kvdb.MemDatabase).Keys() {
		if _, ok := hashes[common.BytesToHash(key)]; !ok {
			t.Errorf("state entry not reported %x", key)
		}
	}
}

type kvs struct{ k, v string }

var testdata1 = []kvs{
	{"barb", "ba"},
	{"bard", "bc"},
	{"bars", "bb"},
	{"bar", "b"},
	{"fab", "z"},
	{"food", "ab"},
	{"foos", "aa"},
	{"foo", "a"},
}

var testdata2 = []kvs{
	{"aardvark", "c"},
	{"bar", "b"},
	{"barb", "bd"},
	{"bars", "be"},
	{"fab", "z"},
	{"foo", "a"},
	{"foos", "aa"},
	{"food", "ab"},
	{"jars", "d"},
}

func TestIteratorSeek(t *testing.T) {
	trie := newEmpty()
	for _, val := range testdata1 {
		trie.Update([]byte(val.k), []byte(val.v))
	}

	// Seek to the middle.
	it := NewIterator(trie.NodeIterator([]byte("fab")))
	if err := checkIteratorOrder(testdata1[4:], it); err != nil {
		t.Fatal(err)
	}

	// Seek to a non-existent key.
	it = NewIterator(trie.NodeIterator([]byte("barc")))
	if err := checkIteratorOrder(testdata1[1:], it); err != nil {
		t.Fatal(err)
	}

	// Seek beyond the end.
	it = NewIterator(trie.NodeIterator([]byte("z")))
	if err := checkIteratorOrder(nil, it); err != nil {
		t.Fatal(err)
	}
}

func checkIteratorOrder(want []kvs, it *Iterator) error {
	for it.Next() {
		if len(want) == 0 {
			return fmt.Errorf("didn't expect any more values, got key %q", it.Key)
		}
		if !bytes.Equal(it.Key, []byte(want[0].k)) {
			return fmt.Errorf("wrong key: got %q, want %q", it.Key, want[0].k)
		}
		want = want[1:]
	}
	if len(want) > 0 {
		return fmt.Errorf("iterator ended early, want key %q", want[0])
	}
	return nil
}

func TestDifferenceIterator(t *testing.T) {
	trieA := newEmpty()
	for _, val := range testdata1 {
		trieA.Update([]byte(val.k), []byte(val.v))
	}
	_, err := trieA.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}

	trieB := newEmpty()
	for _, val := range testdata2 {
		trieB.Update([]byte(val.k), []byte(val.v))
	}
	_, err = trieB.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]string)
	di, _ := NewDifferenceIterator(trieA.NodeIterator(nil), trieB.NodeIterator(nil))
	it := NewIterator(di)
	for it.Next() {
		found[string(it.Key)] = string(it.Value)
	}

	all := []struct{ k, v string }{
		{"aardvark", "c"},
		{"barb", "bd"},
		{"bars", "be"},
		{"jars", "d"},
	}
	for _, item := range all {
		if found[item.k] != item.v {
			t.Errorf("iterator value mismatch for %s: got %v want %v", item.k, found[item.k], item.v)
		}
	}
	if len(found) != len(all) {
		t.Errorf("iterator count mismatch: got %d values, want %d", len(found), len(all))
	}
}

func TestUnionIterator(t *testing.T) {
	trieA := newEmpty()
	for _, val := range testdata1 {
		trieA.Update([]byte(val.k), []byte(val.v))
	}
	_, err := trieA.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}

	trieB := newEmpty()
	for _, val := range testdata2 {
		trieB.Update([]byte(val.k), []byte(val.v))
	}
	_, err = trieB.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}

	di, _ := NewUnionIterator([]NodeIterator{trieA.NodeIterator(nil), trieB.NodeIterator(nil)})
	it := NewIterator(di)

	all := []struct{ k, v string }{
		{"aardvark", "c"},
		{"barb", "ba"},
		{"barb", "bd"},
		{"bard", "bc"},
		{"bars", "bb"},
		{"bars", "be"},
		{"bar", "b"},
		{"fab", "z"},
		{"food", "ab"},
		{"foos", "aa"},
		{"foo", "a"},
		{"jars", "d"},
	}

	for i, kv := range all {
		if !it.Next() {
			t.Errorf("Iterator ends prematurely at element %d", i)
		}
		if kv.k != string(it.Key) {
			t.Errorf("iterator value mismatch for element %d: got key %s want %s", i, it.Key, kv.k)
		}
		if kv.v != string(it.Value) {
			t.Errorf("iterator value mismatch for element %d: got value %s want %s", i, it.Value, kv.v)
		}
	}
	if it.Next() {
		t.Errorf("Iterator returned extra values.")
	}
}

func TestIteratorNoDuplicates(t *testing.T) {
	var tr Trie
	for _, val := range testdata1 {
		tr.Update([]byte(val.k), []byte(val.v))
	}
	checkIteratorNoDuplicates(t, tr.NodeIterator(nil), nil)
}

// This test checks that nodeIterator.Next can be retried after inserting missing trie nodes.
func TestIteratorContinueAfterErrorDisk(t *testing.T)       { testIteratorContinueAfterError(t, false) }
func TestIteratorContinueAfterErrorMemoryOnly(t *testing.T) { testIteratorContinueAfterError(t, true) }

func testIteratorContinueAfterError(t *testing.T, memoryOnly bool) {
	diskdb := kvdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)

	tr, _ := New(common.Hash{}, triedb)
	for _, val := range testdata1 {
		tr.Update([]byte(val.k), []byte(val.v))
	}
	_, err := tr.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !memoryOnly {
		err = triedb.Commit(tr.Hash(), true)
		if err != nil {
			t.Fatal(err)
		}
	}
	wantNodeCount := checkIteratorNoDuplicates(t, tr.NodeIterator(nil), nil)

	var (
		diskKeys [][]byte
		memKeys  []common.Hash
	)
	if memoryOnly {
		memKeys = triedb.Nodes()
	} else {
		diskKeys = diskdb.Keys()
	}
	for i := 0; i < 20; i++ {
		// Create trie that will load all nodes from DB.
		tr, _ := New(tr.Hash(), triedb)

		// Remove a random node from the database. It can't be the root node
		// because that one is already loaded.
		var (
			rkey common.Hash
			rval []byte
			robj *cachedNode
		)
		for {
			if memoryOnly {
				rkey = memKeys[rand.Intn(len(memKeys))]
			} else {
				copy(rkey[:], diskKeys[rand.Intn(len(diskKeys))])
			}
			if rkey != tr.Hash() {
				break
			}
		}
		if memoryOnly {
			robj = triedb.dirties[rkey]
			delete(triedb.dirties, rkey)
		} else {
			rval, _ = diskdb.Get(rkey[:])
			err = diskdb.Delete(rkey[:])
			if err != nil {
				t.Fatal(err)
			}
		}
		// Iterate until the error is hit.
		seen := make(map[string]bool)
		it := tr.NodeIterator(nil)
		checkIteratorNoDuplicates(t, it, seen)
		missing, ok := it.Error().(*MissingNodeError)
		if !ok || missing.NodeHash != rkey {
			t.Fatal("didn't hit missing node, got", it.Error())
		}

		// Add the node back and continue iteration.
		if memoryOnly {
			triedb.dirties[rkey] = robj
		} else {
			err = diskdb.Put(rkey[:], rval)
			if err != nil {
				t.Fatal(err)
			}
		}
		checkIteratorNoDuplicates(t, it, seen)
		if it.Error() != nil {
			t.Fatal("unexpected error", it.Error())
		}
		if len(seen) != wantNodeCount {
			t.Fatal("wrong node iteration count, got", len(seen), "want", wantNodeCount)
		}
	}
}

// Similar to the test above, this one checks that failure to create nodeIterator at a
// certain key prefix behaves correctly when Next is called. The expectation is that Next
// should retry seeking before returning true for the first time.
func TestIteratorContinueAfterSeekErrorDisk(t *testing.T) {
	testIteratorContinueAfterSeekError(t, false)
}
func TestIteratorContinueAfterSeekErrorMemoryOnly(t *testing.T) {
	testIteratorContinueAfterSeekError(t, true)
}

func testIteratorContinueAfterSeekError(t *testing.T, memoryOnly bool) {
	// Commit test trie to db, then remove the node containing "bars".
	diskdb := kvdb.NewMemDatabase()
	triedb := NewDatabase(diskdb)

	ctr, _ := New(common.Hash{}, triedb)
	for _, val := range testdata1 {
		ctr.Update([]byte(val.k), []byte(val.v))
	}
	root, err := ctr.Commit(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !memoryOnly {
		err := triedb.Commit(root, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	barNodeHash := common.HexToHash("05041990364eb72fcb1127652ce40d8bab765f2bfe53225b1170d276cc101c2e")
	var (
		barNodeBlob []byte
		barNodeObj  *cachedNode
	)
	if memoryOnly {
		barNodeObj = triedb.dirties[barNodeHash]
		delete(triedb.dirties, barNodeHash)
	} else {
		barNodeBlob, _ = diskdb.Get(barNodeHash[:])
		err = diskdb.Delete(barNodeHash[:])
		if err != nil {
			t.Fatal(err)
		}
	}
	// Create a new iterator that seeks to "bars". Seeking can't proceed because
	// the node is missing.
	tr, _ := New(root, triedb)
	it := tr.NodeIterator([]byte("bars"))
	missing, ok := it.Error().(*MissingNodeError)
	if !ok {
		t.Fatal("want MissingNodeError, got", it.Error())
	} else if missing.NodeHash != barNodeHash {
		t.Fatal("wrong node missing")
	}
	// Reinsert the missing node.
	if memoryOnly {
		triedb.dirties[barNodeHash] = barNodeObj
	} else {
		err = diskdb.Put(barNodeHash[:], barNodeBlob)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Check that iteration produces the right set of values.
	if err := checkIteratorOrder(testdata1[2:], NewIterator(it)); err != nil {
		t.Fatal(err)
	}
}

func checkIteratorNoDuplicates(t *testing.T, it NodeIterator, seen map[string]bool) int {
	if seen == nil {
		seen = make(map[string]bool)
	}
	for it.Next(true) {
		if seen[string(it.Path())] {
			t.Fatalf("iterator visited node path %x twice", it.Path())
		}
		seen[string(it.Path())] = true
	}
	return len(seen)
}
