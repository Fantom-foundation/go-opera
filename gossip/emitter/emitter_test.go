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

	"github.com/Fantom-foundation/go-opera/gossip/emitter/mock"
	"github.com/Fantom-foundation/go-opera/integration/makefakegenesis"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/utils/txtime"
	"github.com/Fantom-foundation/go-opera/vecmt"
)

//go:generate go run github.com/golang/mock/mockgen -package=mock -destination=mock/world.go github.com/Fantom-foundation/go-opera/gossip/emitter External,TxPool,TxSigner,Signer

func TestEmitter(t *testing.T) {
	cfg := DefaultConfig()
	gValidators := makefakegenesis.GetFakeValidators(3)
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
	external.EXPECT().StateDB().
		Return(nil).
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
		txtime.Enabled = true
		require := require.New(t)
		tx1 := types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
		tx2 := types.NewTransaction(2, common.Address{}, big.NewInt(2), 2, big.NewInt(2), nil)

		external.EXPECT().IsBusy().
			Return(true).
			AnyTimes()

		txtime.Saw(tx1.Hash(), time.Unix(1, 0))

		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))
		txtime.Saw(tx1.Hash(), time.Unix(2, 0))
		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))
		txtime.Validated(tx1.Hash(), time.Unix(2, 0))
		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))

		// reversed order
		txtime.Validated(tx2.Hash(), time.Unix(3, 0))
		txtime.Saw(tx2.Hash(), time.Unix(2, 0))

		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
		txtime.Saw(tx2.Hash(), time.Unix(3, 0))
		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
		txtime.Validated(tx2.Hash(), time.Unix(3, 0))
		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
	})

	t.Run("tick", func(t *testing.T) {
		em.tick()
	})
}
