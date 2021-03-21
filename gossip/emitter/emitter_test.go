package emitter

import (
	"math/big"
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

//go:generate go run github.com/golang/mock/mockgen -package=emitter -destination=world_test.go github.com/Fantom-foundation/go-opera/gossip/emitter World

func TestEmitter(t *testing.T) {
	cfg := DefaultConfig()
	gValidators := makegenesis.GetFakeValidators(3)
	vv := pos.NewBuilder()
	for _, v := range gValidators {
		vv.Set(v.ID, pos.Weight(1))
	}
	validators := vv.Build()
	cfg.Validator.ID = gValidators[0].ID

	ctrl := gomock.NewController(t)
	world := NewMockWorld(ctrl)

	world.EXPECT().Lock().
		AnyTimes()
	world.EXPECT().Unlock().
		AnyTimes()
	world.EXPECT().DagIndex().
		Return((*vecmt.Index)(nil)).
		AnyTimes()
	world.EXPECT().IsSynced().
		Return(true).
		AnyTimes()
	world.EXPECT().PeersNum().
		Return(int(3)).
		AnyTimes()

	em := NewEmitter(cfg, world)

	t.Run("init", func(t *testing.T) {
		world.EXPECT().GetRules().
			Return(opera.FakeNetRules()).
			AnyTimes()

		world.EXPECT().GetEpochValidators().
			Return(validators, idx.Epoch(1)).
			AnyTimes()

		world.EXPECT().GetLastEvent(idx.Epoch(1), cfg.Validator.ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		world.EXPECT().GetGenesisTime().
			Return(inter.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

	t.Run("memorizeTxTimes", func(t *testing.T) {
		require := require.New(t)
		tx := types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)

		world.EXPECT().IsBusy().
			Return(true).
			AnyTimes()

		_, ok := em.txTime.Get(tx.Hash())
		require.False(ok)

		before := time.Now()
		em.memorizeTxTimes(types.Transactions{tx})
		after := time.Now()

		cached, ok := em.txTime.Get(tx.Hash())
		got := cached.(time.Time)
		require.True(ok)
		require.True(got.After(before))
		require.True(got.Before(after))
	})

	t.Run("tick", func(t *testing.T) {
		em.tick()
	})
}
