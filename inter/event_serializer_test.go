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
	"github.com/stretchr/testify/assert"

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
	tx1 := types.NewRawTransaction(math.MaxUint64, nil, h.Big(), math.MaxUint64, h.Big(), []byte{}, big.NewInt(0xff), h.Big(), h.Big())
	tx2 := types.NewRawTransaction(math.MaxUint64, &common.Address{}, h.Big(), math.MaxUint64, h.Big(), max.extra, big.NewInt(0xff), h.Big(), h.Big())
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
		assertar := assert.New(t)

		for name, header0 := range ee {
			buf, err := rlp.EncodeToBytes(&header0)
			if !assertar.NoError(err) {
				return
			}

			var header1 EventPayload
			err = rlp.DecodeBytes(buf, &header1)
			if !assertar.NoError(err, name) {
				return
			}

			if !assert.EqualValues(t, header0.extEventData, header1.extEventData, name) {
				return
			}
			if !assert.EqualValues(t, header0.sigData, header1.sigData, name) {
				return
			}
			for i := range header0.payloadData.txs {
				if !assert.EqualValues(t, header0.payloadData.txs[i].Hash(), header1.payloadData.txs[i].Hash(), name) {
					return
				}
			}
			if !assert.EqualValues(t, header0.baseEvent, header1.baseEvent, name) {
				return
			}
			if !assert.EqualValues(t, header0.ID(), header1.ID(), name) {
				return
			}
			if !assert.EqualValues(t, header0.HashToSign(), header1.HashToSign(), name) {
				return
			}
			if !assert.EqualValues(t, header0.Size(), header1.Size(), name) {
				return
			}
		}
	})

	t.Run("err", func(t *testing.T) {
		assertar := assert.New(t)

		for name, header0 := range ee {
			bin, err := header0.MarshalBinary()
			if !assertar.NoError(err, name) {
				return
			}

			n := rand.Intn(len(bin) - len(header0.Extra()) - 1)
			bin = bin[0:n]

			buf, err := rlp.EncodeToBytes(bin)
			if !assertar.NoError(err, name) {
				return
			}

			var header1 Event
			err = rlp.DecodeBytes(buf, &header1)
			if !assertar.Error(err, name) {
				return
			}
			//t.Log(err)
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
		h := hash.BytesToHash(bytes.Repeat([]byte{math.MaxUint8}, 32))
		addr := common.BytesToAddress(h.Bytes()[:20])
		if i%2 == 0 {
			tx := types.NewRawTransaction(r.Uint64(), nil, randBig(r), r.Uint64(), randBig(r), []byte{}, big.NewInt(int64(r.Intn(0xffffffff))), h.Big(), h.Big())
			txs = append(txs, tx)
		} else {
			tx := types.NewRawTransaction(r.Uint64(), &addr, randBig(r), r.Uint64(), randBig(r), []byte{}, new(big.Int), new(big.Int), new(big.Int))
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
