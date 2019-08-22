package packs_downloader

import (
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/stretchr/testify/assert"
	"testing"

	tree "github.com/emirpasic/gods/maps/treemap"
)

type binarySearchExpected struct {
	requestIndex idx.Pack
	requestFull  bool
	syncedUp     bool
}

type binarySearchTest struct {
	testName string
	infos    map[idx.Pack]bool
	packsNum idx.Pack
	expected binarySearchExpected
	allNums  bool // allPacksNums
}

func TestBinarySearch(t *testing.T) {
	var tests = []binarySearchTest{
		{
			infos: map[idx.Pack]bool{
				1: false,
			},
			allNums:  true,
			expected: binarySearchExpected{1, true, false},
		},
		{
			infos: map[idx.Pack]bool{
				10: false,
			},
			allNums:  true,
			expected: binarySearchExpected{5, false, false},
		},

		{
			infos: map[idx.Pack]bool{
				10: true,
				11: false,
			},
			allNums:  true,
			expected: binarySearchExpected{11, true, false},
		},
		{
			infos: map[idx.Pack]bool{
				10:  true,
				100: false,
			},
			allNums:  true,
			expected: binarySearchExpected{55, false, false},
		},
		{
			infos: map[idx.Pack]bool{
				10:  false,
				100: false,
			},
			allNums:  true,
			expected: binarySearchExpected{5, false, false},
		},
		{
			infos: map[idx.Pack]bool{
				10:  true,
				20:  true,
				60:  false,
				100: false,
			},
			allNums:  true,
			expected: binarySearchExpected{40, false, false},
		},
		{
			infos: map[idx.Pack]bool{
				5:   false, // faulty pack, not known before known
				10:  true,
				20:  true,
				60:  false,
				100: false,
			},
			allNums:  true,
			expected: binarySearchExpected{40, false, false},
		},

		{
			infos:    map[idx.Pack]bool{},
			allNums:  true,
			expected: binarySearchExpected{1, false, false},
		},

		{
			infos: map[idx.Pack]bool{
				1: true,
			},
			packsNum: 1,
			expected: binarySearchExpected{1, false, true},
		},
		{
			infos: map[idx.Pack]bool{
				0: false, // faulty pack, not known before known
				1: true,
			},
			packsNum: 1,
			expected: binarySearchExpected{1, false, true},
		},
		{
			infos: map[idx.Pack]bool{
				29: true,
				30: true,
			},
			packsNum: 30,
			expected: binarySearchExpected{30, false, true},
		},
		{
			infos: map[idx.Pack]bool{
				29: true,
				30: true,
			},
			packsNum: 31,
			expected: binarySearchExpected{31, false, false},
		},
		{
			infos: map[idx.Pack]bool{
				30: true,
				31: false,
			},
			allNums:  true,
			expected: binarySearchExpected{31, true, false},
		},
	}
	for i, test := range tests {
		test.testName = fmt.Sprintf("test#%d", i)
		testBinarySearch(t, &test)
	}
}

func testBinarySearch(t *testing.T, test *binarySearchTest) {
	assertar := assert.New(t)

	d := PeerPacksDownloader{}
	d.packInfos = tree.NewWithIntComparator()

	knownHeads := map[hash.Event]bool{}

	for index, isKnown := range test.infos {
		e := inter.NewEvent()
		e.Extra = index.Bytes() // make each event unique

		// build info with 1 head
		info := &packInfoData{
			index: index,
			heads: hash.Events{e.Hash()},
		}
		d.packInfos.Put(int(index), info)
		if isKnown {
			// memorize pack as known
			knownHeads[e.Hash()] = true
		}
	}

	d.onlyNotConnected = func(ids hash.Events) hash.Events {
		assertar.Equal(1, len(ids), test.testName)
		if knownHeads[ids[0]] {
			return nil
		}
		return ids
	}

	d.packsNum = test.packsNum
	for i := 0; i < 1000; i++ {
		requestIndex, requestFull, syncedUp := d.binarySearchReq()

		assertar.Equal(test.expected.requestIndex, requestIndex, test.testName, test.packsNum)
		assertar.Equal(test.expected.requestFull, requestFull, test.testName, test.packsNum)
		assertar.Equal(test.expected.syncedUp, syncedUp, test.testName, test.packsNum)

		// test with other packsNum
		if !test.allNums {
			break
		}
		d.packsNum = idx.Pack(i)
	}
}
