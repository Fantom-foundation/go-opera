package inter

import (
	"bytes"
	"encoding/json"
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

func emptyEvent(ver uint8) EventPayload {
	empty := MutableEventPayload{}
	empty.SetVersion(ver)
	if ver == 0 {
		empty.SetEpoch(256)
	}
	empty.SetParents(hash.Events{})
	empty.SetExtra([]byte{})
	empty.SetTxs(types.Transactions{})
	empty.SetPayloadHash(EmptyPayloadHash(ver))
	return *empty.Build()
}

func TestEventPayloadSerialization(t *testing.T) {
	max := MutableEventPayload{}
	max.SetEpoch(math.MaxUint32)
	max.SetSeq(idx.Event(math.MaxUint32))
	max.SetLamport(idx.Lamport(math.MaxUint32))
	h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
	max.SetParents(hash.Events{hash.Event(h), hash.Event(h), hash.Event(h)})
	max.SetPayloadHash(hash.Hash(h))
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
		"empty0": emptyEvent(0),
		"empty1": emptyEvent(1),
		"max":    *max.Build(),
		"random": *FakeEvent(12, 1, 1, true),
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
	e := emptyEvent(0)

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
	e := FakeEvent(0, 0, 0, false)

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
	e := FakeEvent(1000, 0, 0, false)

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
	e := emptyEvent(0)
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
	e := FakeEvent(0, 0, 0, false)
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
	e := FakeEvent(1000, 0, 0, false)
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

func TestEventRPCMarshaling(t *testing.T) {
	t.Run("Event", func(t *testing.T) {
		require := require.New(t)
		for i := 0; i < 3; i++ {
			var event0 EventI = &FakeEvent(i, i, i, i != 0).Event
			mapping := RPCMarshalEvent(event0)
			bb, err := json.Marshal(mapping)
			require.NoError(err)

			mapping = make(map[string]interface{})
			err = json.Unmarshal(bb, &mapping)
			require.NoError(err)
			event1 := RPCUnmarshalEvent(mapping)

			require.Equal(event0, event1, i)
		}
	})

	t.Run("EventPayload", func(t *testing.T) {
		require := require.New(t)
		for i := 0; i < 3; i++ {
			var event0 = FakeEvent(i, i, i, i != 0)
			mapping, err := RPCMarshalEventPayload(event0, true, false)
			require.NoError(err)
			bb, err := json.Marshal(mapping)
			require.NoError(err)

			mapping = make(map[string]interface{})
			err = json.Unmarshal(bb, &mapping)

			event1 := RPCUnmarshalEvent(mapping)
			require.Equal(&event0.SignedEvent.Event, event1, i)
		}
	})
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

func randHash(r *rand.Rand) hash.Hash {
	return hash.BytesToHash(randBytes(r, 32))
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
func FakeEvent(txsNum, mpsNum, bvsNum int, ersNum bool) *EventPayload {
	r := rand.New(rand.NewSource(int64(0)))
	random := &MutableEventPayload{}
	random.SetVersion(1)
	random.SetNetForkID(uint16(r.Uint32() >> 16))
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
		if i%3 == 0 {
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
		} else if i%3 == 1 {
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
		} else {
			tx := types.NewTx(&types.DynamicFeeTx{
				ChainID:    randBig(r),
				Nonce:      r.Uint64(),
				GasTipCap:  randBig(r),
				GasFeeCap:  randBig(r),
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
	mps := []MisbehaviourProof{}
	for i := 0; i < mpsNum; i++ {
		// MPs are serialized with RLP, so no need to test extensively
		mps = append(mps, MisbehaviourProof{
			EventsDoublesign: &EventsDoublesign{
				Pair: [2]SignedEventLocator{SignedEventLocator{}, SignedEventLocator{}},
			},
			BlockVoteDoublesign: nil,
			WrongBlockVote:      nil,
			EpochVoteDoublesign: nil,
			WrongEpochVote:      nil,
		})
	}
	bvs := LlrBlockVotes{}
	if bvsNum > 0 {
		bvs.Start = 1 + idx.Block(rand.Intn(1000))
		bvs.Epoch = 1 + idx.Epoch(rand.Intn(1000))
	}
	for i := 0; i < bvsNum; i++ {
		bvs.Votes = append(bvs.Votes, randHash(r))
	}
	ers := LlrEpochVote{}
	if ersNum {
		ers.Epoch = 1 + idx.Epoch(rand.Intn(1000))
		ers.Vote = randHash(r)
	}

	random.SetTxs(txs)
	random.SetMisbehaviourProofs(mps)
	random.SetEpochVote(ers)
	random.SetBlockVotes(bvs)
	random.SetPayloadHash(CalcPayloadHash(random))

	parent := MutableEventPayload{}
	parent.SetVersion(1)
	parent.SetLamport(random.Lamport() - 500)
	parent.SetEpoch(random.Epoch())
	random.SetParents(hash.Events{parent.Build().ID()})

	return random.Build()
}
