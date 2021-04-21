package emitter

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-opera/gossip/emitter/mock"
	"github.com/Fantom-foundation/go-opera/integration/makegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

//go:generate go run github.com/golang/mock/mockgen -package=mock -destination=mock/world.go github.com/Fantom-foundation/go-opera/gossip/emitter External,TxPool,TxSigner,Signer

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
	external := mock.NewMockExternal(ctrl)
	txPool := mock.NewMockTxPool(ctrl)
	signer := mock.NewMockSigner(ctrl)
	txSigner := mock.NewMockTxSigner(ctrl)

	external.EXPECT().Lock().
		AnyTimes()
	external.EXPECT().Unlock().
		AnyTimes()
	external.EXPECT().DagIndex().
		Return((*vecmt.Index)(nil)).
		AnyTimes()
	external.EXPECT().IsSynced().
		Return(true).
		AnyTimes()
	external.EXPECT().PeersNum().
		Return(int(3)).
		AnyTimes()

	em := NewEmitter(cfg, World{
		External: external,
		TxPool:   txPool,
		Signer:   signer,
		TxSigner: txSigner,
	})

	t.Run("init", func(t *testing.T) {
		external.EXPECT().GetRules().
			Return(opera.FakeNetRules()).
			AnyTimes()

		external.EXPECT().GetEpochValidators().
			Return(validators, idx.Epoch(1)).
			AnyTimes()

		external.EXPECT().GetLastEvent(idx.Epoch(1), cfg.Validator.ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		external.EXPECT().GetGenesisTime().
			Return(inter.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

	t.Run("memorizeTxTimes", func(t *testing.T) {
		require := require.New(t)
		tx := types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)

		external.EXPECT().IsBusy().
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

	t.Run("isMyTxTurn", func(t *testing.T) {
		require := require.New(t)
		const accountNonce = 1
		var (
			sender common.Address
			txTime = time.Now()
			tx     = types.NewTransaction(accountNonce, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)

			validators = int(em.validators.Len())
			got        = make(map[idx.ValidatorID]bool, validators)
		)

		for i := 0; i < validators; i++ {
			var (
				onlyOne    bool
				atLeastOne bool
			)
			now := txTime.Add(TxTurnPeriodLatency).Add(TxTurnPeriod * time.Duration(i))
			for _, me := range em.validators.IDs() {
				if em.isMyTxTurn(tx.Hash(), sender, accountNonce, now, em.validators, me, em.epoch) {
					onlyOne = !onlyOne && !atLeastOne
					atLeastOne = true
					got[me] = true
				}
			}
			require.True(atLeastOne, i)
			require.True(onlyOne, i)
		}
		everyOne := len(got) == int(em.validators.Len())
		require.True(everyOne)
	})

	t.Run("tick", func(t *testing.T) {
		require := require.New(t)

		now := time.Now()
		external.EXPECT().IsBusy().
			Return(false).
			Times(1)
		isBusy := em.tick(now)
		require.True(isBusy)

		external.EXPECT().IsBusy().
			Return(true).
			Times(1)
		isBusy = em.tick(now)
		require.True(isBusy)
	})
}

func TestRandomizeEmitTime(t *testing.T) {
	require := require.New(t)

	cfgs := make([]EmitIntervals, 10)
	base := DefaultConfig().EmitIntervals

	for i := 0; i < len(cfgs); i++ {
		r := rand.New(rand.NewSource(time.Now().Add(time.Duration(i) * time.Second).UnixNano()))
		cfgs[i] = base.RandomizeEmitTime(r)
	}

	for i := 0; i < len(cfgs)-1; i++ {
		for j := i + 1; j < len(cfgs)-1; j++ {
			require.NotEqual(cfgs[i], cfgs[j], "%d vs %d", i, j)
		}
	}
}
