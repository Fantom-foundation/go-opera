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

func TestEventHeaderData_Serialize(t *testing.T) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.Serialize()
	newHeader := EventHeaderData{}
	newHeader.Deserialize(buf)

	assert.EqualValues(t, header, newHeader)
}

func TestEventHeaderData_SerializeA(t *testing.T) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeA()
	newHeader := EventHeaderData{}
	newHeader.DeserializeA(buf)

	assert.EqualValues(t, header, newHeader)
}

func TestEventHeaderData_SerializeB(t *testing.T) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeB()
	newHeader := EventHeaderData{}
	newHeader.DeserializeB(buf)

	assert.EqualValues(t, header, newHeader)
}

func TestEventHeaderData_SerializeC(t *testing.T) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeC()
	newHeader := EventHeaderData{}
	newHeader.DeserializeC(buf)

	assert.EqualValues(t, header, newHeader)
}

func BenchmarkEventHeaderData_Serialize(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.Serialize()
	b.ReportMetric(float64(len(*buf)), "Bytes")

	/*
	if b.N == 1 {
		fmt.Printf("Serialized: %X\n", buf)

		tmp, err := rlp.EncodeToBytes(header)
		if err != nil {
			b.Fatalf("Error rlp serialization: %s", err)
		}

		fmt.Printf("OldSerialized: %X\n", tmp)
	}
	*/

	for i := 0; i < b.N; i++ {
		_ = header.Serialize()
	}
}

func BenchmarkEventHeaderData_SerializeA(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeA()
	b.ReportMetric(float64(len(*buf)), "Bytes")

	/*
	if b.N == 1 {
		fmt.Printf("Serialized: %X\n", buf)

		tmp, err := rlp.EncodeToBytes(header)
		if err != nil {
			b.Fatalf("Error rlp serialization: %s", err)
		}

		fmt.Printf("OldSerialized: %X\n", tmp)
	}
	*/

	for i := 0; i < b.N; i++ {
		_ = header.SerializeA()
	}
}

func BenchmarkEventHeaderData_SerializeB(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeB()
	b.ReportMetric(float64(len(*buf)), "Bytes")

	/*
		if b.N == 1 {
			fmt.Printf("Serialized: %X\n", buf)

			tmp, err := rlp.EncodeToBytes(header)
			if err != nil {
				b.Fatalf("Error rlp serialization: %s", err)
			}

			fmt.Printf("OldSerialized: %X\n", tmp)
		}
	*/

	for i := 0; i < b.N; i++ {
		_ = header.SerializeB()
	}
}

func BenchmarkEventHeaderData_SerializeC(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeC()
	b.ReportMetric(float64(len(*buf)), "Bytes")

	for i := 0; i < b.N; i++ {
		_ = header.SerializeC()
	}
}

func BenchmarkEventHeaderData_OldSerialize(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf, err := rlp.EncodeToBytes(header)
	if err != nil {
		b.Fatalf("Error rlp serialization: %s", err)
	}
	b.ReportMetric(float64(len(buf)), "Bytes")

	for i := 0; i < b.N; i++ {
		_, _ = rlp.EncodeToBytes(header)
	}
}

func BenchmarkEventHeaderData_Deserialize(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.Serialize()

	for i := 0; i < b.N; i++ {
		header.Deserialize(buf)
	}
}

func BenchmarkEventHeaderData_DeserializeA(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeA()

	for i := 0; i < b.N; i++ {
		header.DeserializeA(buf)
	}
}

func BenchmarkEventHeaderData_DeserializeB(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeB()

	for i := 0; i < b.N; i++ {
		header.DeserializeB(buf)
	}
}

func BenchmarkEventHeaderData_DeserializeC(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf := header.SerializeC()

	for i := 0; i < b.N; i++ {
		header.DeserializeC(buf)
	}
}
func BenchmarkEventHeaderData_OldDeserialize(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	buf, err := rlp.EncodeToBytes(header)
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
	copy(h[0:4], []byte{1,0,1,0})
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
