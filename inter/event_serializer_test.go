package inter

import (
	"bytes"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

func emptyEvent() EventPayload {
	empty := MutableEventPayload{}
	empty.SetParents(hash.Events{})
	empty.SetExtra([]byte{})
	empty.SetTxs(types.Transactions{})
	empty.SetTxHash(EmptyTxHash)
	return *empty.Build()
}

func TestEventPayloadSerialization(t *testing.T) {
	max := MutableEventPayload{}
	max.SetEpoch(math.MaxUint32)
	max.SetSeq(idx.Event(math.MaxUint32))
	max.SetLamport(idx.Lamport(math.MaxUint32))
	h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
	max.SetParents(hash.Events{hash.Event(h), hash.Event(h), hash.Event(h)})
	max.SetTxHash(hash.Hash(h))
	max.SetSig(BytesToSignature(bytes.Repeat([]byte{math.MaxUint8}, SigSize)))
	max.SetExtra(bytes.Repeat([]byte{math.MaxUint8}, 100))
	max.SetCreationTime(math.MaxUint64)
	max.SetMedianTime(math.MaxUint64)
	tx1 := types.NewTx(&types.LegacyTx{
		Nonce:    math.MaxUint64,
		GasPrice: h.Big(),
		Gas:      math.MaxUint64,
		To:       nil,
		Value:    h.Big(),
		Data:     []byte{},
		V:        big.NewInt(0xff),
		R:        h.Big(),
		S:        h.Big(),
	})
	tx2 := types.NewTx(&types.LegacyTx{
		Nonce:    math.MaxUint64,
		GasPrice: h.Big(),
		Gas:      math.MaxUint64,
		To:       &common.Address{},
		Value:    h.Big(),
		Data:     max.extra,
		V:        big.NewInt(0xff),
		R:        h.Big(),
		S:        h.Big(),
	})
	txs := types.Transactions{}
	for i := 0; i < 200; i++ {
		txs = append(txs, tx1)
		txs = append(txs, tx2)
	}
	max.SetTxs(txs)

	ee := map[string]EventPayload{
		"empty":  emptyEvent(),
		"max":    *max.Build(),
		"random": *FakeEvent(2),
	}

	t.Run("ok", func(t *testing.T) {
		require := require.New(t)

		for name, header0 := range ee {
			buf, err := rlp.EncodeToBytes(&header0)
			require.NoError(err)

			var header1 EventPayload
			err = rlp.DecodeBytes(buf, &header1)
			require.NoError(err, name)

			require.EqualValues(header0.extEventData, header1.extEventData, name)
			require.EqualValues(header0.sigData, header1.sigData, name)
			for i := range header0.payloadData.txs {
				require.EqualValues(header0.payloadData.txs[i].Hash(), header1.payloadData.txs[i].Hash(), name)
			}
			require.EqualValues(header0.baseEvent, header1.baseEvent, name)
			require.EqualValues(header0.ID(), header1.ID(), name)
			require.EqualValues(header0.HashToSign(), header1.HashToSign(), name)
			require.EqualValues(header0.Size(), header1.Size(), name)
		}
	})

	t.Run("err", func(t *testing.T) {
		require := require.New(t)

		for name, header0 := range ee {
			bin, err := header0.MarshalBinary()
			require.NoError(err, name)

			n := rand.Intn(len(bin) - len(header0.Extra()) - 1)
			bin = bin[0:n]

			buf, err := rlp.EncodeToBytes(bin)
			require.NoError(err, name)

			var header1 Event
			err = rlp.DecodeBytes(buf, &header1)
			require.Error(err, name)
		}
	})
}

func BenchmarkEventPayload_EncodeRLP_empty(b *testing.B) {
	e := emptyEvent()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, err := rlp.EncodeToBytes(&e)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(buf)), "size")
	}
}

func BenchmarkEventPayload_EncodeRLP_NoPayload(b *testing.B) {
	e := FakeEvent(0)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, err := rlp.EncodeToBytes(&e)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(buf)), "size")
	}
}

func BenchmarkEventPayload_EncodeRLP(b *testing.B) {
	e := FakeEvent(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf, err := rlp.EncodeToBytes(&e)
		if err != nil {
			b.Fatal(err)
		}
		b.ReportMetric(float64(len(buf)), "size")
	}
}

func BenchmarkEventPayload_DecodeRLP_empty(b *testing.B) {
	e := emptyEvent()
	me := MutableEventPayload{}

	buf, err := rlp.EncodeToBytes(&e)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = rlp.DecodeBytes(buf, &me)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventPayload_DecodeRLP_NoPayload(b *testing.B) {
	e := FakeEvent(0)
	me := MutableEventPayload{}

	buf, err := rlp.EncodeToBytes(&e)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = rlp.DecodeBytes(buf, &me)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEventPayload_DecodeRLP(b *testing.B) {
	e := FakeEvent(1000)
	me := MutableEventPayload{}

	buf, err := rlp.EncodeToBytes(&e)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = rlp.DecodeBytes(buf, &me)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func randBig(r *rand.Rand) *big.Int {
	b := make([]byte, r.Intn(8))
	_, _ = r.Read(b)
	if len(b) == 0 {
		b = []byte{0}
	}
	return new(big.Int).SetBytes(b)
}

func randAddr(r *rand.Rand) common.Address {
	addr := common.Address{}
	r.Read(addr[:])
	return addr
}

func randBytes(r *rand.Rand, size int) []byte {
	b := make([]byte, size)
	r.Read(b)
	return b
}

func randAddrPtr(r *rand.Rand) *common.Address {
	addr := randAddr(r)
	return &addr
}

func randAccessList(r *rand.Rand, maxAddrs, maxKeys int) types.AccessList {
	accessList := make(types.AccessList, r.Intn(maxAddrs))
	for i := range accessList {
		accessList[i].Address = randAddr(r)
		accessList[i].StorageKeys = make([]common.Hash, r.Intn(maxKeys))
		for j := range accessList[i].StorageKeys {
			r.Read(accessList[i].StorageKeys[j][:])
		}
	}
	return accessList
}

// FakeEvent generates random event for testing purpose.
func FakeEvent(txsNum int) *EventPayload {

	r := rand.New(rand.NewSource(int64(0)))
	random := MutableEventPayload{}
	random.SetLamport(1000)
	random.SetExtra([]byte{byte(r.Uint32())})
	random.SetSeq(idx.Event(r.Uint32() >> 8))
	random.SetCreator(idx.ValidatorID(r.Uint32()))
	random.SetFrame(idx.Frame(r.Uint32() >> 16))
	random.SetCreationTime(Timestamp(r.Uint64()))
	random.SetMedianTime(Timestamp(r.Uint64()))
	random.SetGasPowerUsed(r.Uint64())
	random.SetGasPowerLeft(GasPowerLeft{[2]uint64{r.Uint64(), r.Uint64()}})
	txs := types.Transactions{}
	for i := 0; i < txsNum; i++ {
		h := hash.Hash{}
		r.Read(h[:])
		if i%2 == 0 {
			tx := types.NewTx(&types.LegacyTx{
				Nonce:    r.Uint64(),
				GasPrice: randBig(r),
				Gas:      257 + r.Uint64(),
				To:       nil,
				Value:    randBig(r),
				Data:     randBytes(r, r.Intn(300)),
				V:        big.NewInt(int64(r.Intn(0xffffffff))),
				R:        h.Big(),
				S:        h.Big(),
			})
			txs = append(txs, tx)
		} else {
			tx := types.NewTx(&types.AccessListTx{
				ChainID:    randBig(r),
				Nonce:      r.Uint64(),
				GasPrice:   randBig(r),
				Gas:        r.Uint64(),
				To:         randAddrPtr(r),
				Value:      randBig(r),
				Data:       randBytes(r, r.Intn(300)),
				AccessList: randAccessList(r, 300, 300),
				V:          big.NewInt(int64(r.Intn(0xffffffff))),
				R:          h.Big(),
				S:          h.Big(),
			})
			txs = append(txs, tx)
		}
	}
	if txs.Len() == 0 {
		random.SetTxHash(EmptyTxHash)
	}
	random.SetTxs(txs)

	parent := MutableEventPayload{}
	parent.SetLamport(random.Lamport() - 500)
	parent.SetEpoch(random.Epoch())
	random.SetParents(hash.Events{parent.Build().ID()})

	return random.Build()
}
