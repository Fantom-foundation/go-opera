package gossip

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type testWrsItem struct {
	idx  int
	widx *int
}

func (t *testWrsItem) Weight() int64 {
	w := *t.widx
	if w == -1 || w == t.idx {
		return int64(t.idx + 1)
	}
	return 0
}

func TestWeightedRandomSelect(t *testing.T) {
	logger.SetTestMode(t)

	testFn := func(cnt int) {
		s := newWeightedRandomSelect()
		w := -1
		list := make([]testWrsItem, cnt)
		for i := range list {
			list[i] = testWrsItem{idx: i, widx: &w}
			s.update(&list[i])
		}
		w = rand.Intn(cnt)
		c := s.choose()
		if c == nil {
			t.Errorf("expected item, got nil")
		} else {
			if c.(*testWrsItem).idx != w {
				t.Errorf("expected another item")
			}
		}
		w = -2
		if s.choose() != nil {
			t.Errorf("expected nil, got item")
		}
	}
	testFn(1)
	testFn(10)
	testFn(100)
	testFn(1000)
	testFn(10000)
	testFn(100000)
	testFn(1000000)
}
