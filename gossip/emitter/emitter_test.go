package emitter

import (
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/golang/mock/gomock"

	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
)

//go:generate go run github.com/golang/mock/mockgen -destination=reader_test.go -package=emitter -source=reader.go Reader

func TestEmitter(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockReader(ctrl)
	world := fakeWorld(store)

	cfg := DefaultConfig()

	gValidators := makegenesis.GetFakeValidators(3)
	vv := pos.NewBuilder()
	for _, v := range gValidators {
		vv.Set(v.ID, pos.Weight(1))
	}
	validators := vv.Build()
	cfg.Validator.ID = gValidators[0].ID

	em := NewEmitter(cfg, world)

	t.Run("init", func(t *testing.T) {
		store.EXPECT().GetRules().
			Return(opera.FakeNetRules()).
			AnyTimes()

		store.EXPECT().GetEpochValidators().
			Return(validators, idx.Epoch(1)).
			AnyTimes()

		store.EXPECT().GetLastEvent(idx.Epoch(1), cfg.Validator.ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		store.EXPECT().GetGenesisTime().
			Return(inter.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

}

func fakeWorld(s Reader) World {
	return World{
		Store: s,
	}
}
