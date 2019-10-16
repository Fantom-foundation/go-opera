package inter

import (
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func TestEventHeaderData_MarshalBinary(t *testing.T) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf, _ := header.MarshalBinary()
	newHeader := EventHeaderData{}
	_ = newHeader.UnmarshalBinary(buf)

	assert.EqualValues(t, header, newHeader)
}

func TestEventHeaderData_EncodeRLP(t *testing.T) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf, err := rlp.EncodeToBytes(&header)

	assert.Equal(t, nil, err)

	newHeader := EventHeaderData{}
	_ = rlp.DecodeBytes(buf, &newHeader)

	assert.EqualValues(t, header, newHeader)
}

func BenchmarkEventHeaderData_MarshalBinary(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	/*
	// Only for go1.13+
	buf, _ := header.MarshalBinary()
	b.ReportMetric(float64(len(buf)), "Bytes")
	*/

	for i := 0; i < b.N; i++ {
		_, _ = header.MarshalBinary()
	}
}

func BenchmarkEventHeaderData_RLPMarshal(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	/*
	// Only for go1.13+
	buf, err := rlp.EncodeToBytes(header)
	if err != nil {
		b.Fatalf("Error rlp serialization: %s", err)
	}
	b.ReportMetric(float64(len(buf)), "Bytes")
	*/

	for i := 0; i < b.N; i++ {
		_, _ = rlp.EncodeToBytes(header)
	}
}

func BenchmarkEventHeaderData_UnmarshalBinary(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf, _ := header.MarshalBinary()

	for i := 0; i < b.N; i++ {
		_ = header.UnmarshalBinary(buf)
	}
}

func BenchmarkEventHeaderData_RLPUnmarshal(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf, err := rlp.EncodeToBytes(&header)
	if err != nil {
		b.Fatalf("Error rlp serialization: %s", err)
	}

	for i := 0; i < b.N; i++ {
		err = rlp.DecodeBytes(buf, &header)
		if err != nil {
			b.Fatalf("Error rlp deserialization: %s", err)
		}
	}
}

func FakeEventWithOneEpoch() (res []*Event) {
	creators := []common.Address{
		{},
		hash.FakePeer(),
		hash.FakePeer(),
		hash.FakePeer(),
	}
	parents := []hash.Events{
		FakeEventsEpoch(1),
		FakeEventsEpoch(2),
		FakeEventsEpoch(8),
	}
	i := 0
	for c := 0; c < len(creators); c++ {
		for p := 0; p < len(parents); p++ {
			e := NewEvent()
			e.Seq = idx.Event(p)
			e.Creator = creators[c]
			e.Parents = parents[p]
			e.Extra = []byte{}
			e.Sig = []byte{}

			res = append(res, e)
			i++
		}
	}
	return
}

// FakeEvent generates random fake event hash with one epoch for testing purpose.
func FakeEventEpoch() (h hash.Event) {
	_, err := rand.Read(h[:])
	if err != nil {
		panic(err)
	}
	copy(h[0:4], []byte{1, 0, 1, 0})
	return
}

// FakeEvents generates random fake event hashes with one epoch for testing purpose.
func FakeEventsEpoch(n int) hash.Events {
	res := hash.Events{}
	for i := 0; i < n; i++ {
		res.Add(FakeEventEpoch())
	}
	return res
}
