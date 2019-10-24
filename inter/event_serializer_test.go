package inter

import (
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
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

	// Only for go1.13+
	buf, _ := header.MarshalBinary()
	b.ReportMetric(float64(len(buf)), "Bytes")

	for i := 0; i < b.N; i++ {
		_, _ = header.MarshalBinary()
	}
}

func BenchmarkEventHeaderData_EncodeRLP(b *testing.B) {
	events := FakeEventWithOneEpoch()
	header := events[len(events)-1].EventHeaderData

	// Only for go1.13+
	buf, err := rlp.EncodeToBytes(&header)
	if err != nil {
		b.Fatalf("Error rlp serialization: %s", err)
	}
	b.ReportMetric(float64(len(buf)), "Bytes")

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

func BenchmarkEventHeaderData_DecodeRLP(b *testing.B) {
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
