package bench

//go:generate protoc --go_out=plugins=grpc:./ ./structs.proto

//go:generate go test -benchmem -bench . -cpuprofile cpu.out
//go:generate go tool pprof -svg -output="cpu_prof.svg" bench.test cpu.out

import (
	"math/rand"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
	"github.com/Fantom-foundation/go-lachesis/src/rlp"
)

func BenchmarkRlp(b *testing.B) {
	e0 := randRlpEvent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := rlp.EncodeToBytes(e0)
		if err != nil {
			b.Fatal(err)
			break
		}

		e1 := &posposet.Event{}
		err = rlp.DecodeBytes(buf, e1)
		if err != nil {
			b.Fatal(err)
			break
		}
	}
}

func BenchmarkProto(b *testing.B) {
	e0 := randProtoEvent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := e0.ProtoMarshal()
		if err != nil {
			b.Fatal(err)
			break
		}

		e1 := &Event{}
		err = e1.ProtoUnmarshal(buf)
		if err != nil {
			b.Fatal(err)
			break
		}
	}
}

func randRlpEvent() *posposet.Event {
	creator := common.FakeAddress()

	return &posposet.Event{
		Index:                rand.Uint64(),
		Creator:              creator,
		Parents:              posposet.FakeEventHashes(2),
		ExternalTransactions: randTxns(),
	}
}

func randProtoEvent() *Event {
	creator := common.FakeAddress()

	return &Event{
		EventBody: EventBody{
			Index:                rand.Uint64(),
			Creator:              creator.Bytes(),
			ExternalTransactions: randTxns(),
		},
		Parents: posposet.FakeEventHashes(2),
	}
}

func randTxns() [][]byte {
	txns := [][]byte{
		make([]byte, 32),
		make([]byte, 32),
		make([]byte, 32),
	}
	for _, tx := range txns {
		rand.Read(tx)
	}

	return txns
}
