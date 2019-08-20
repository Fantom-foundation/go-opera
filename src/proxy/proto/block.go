package proto

import (
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/wire"
)

func MarshalBlock(b *inter.Block) ([]byte, error) {
	w := b.ToWire()
	return proto.Marshal(w)
}

func UnmarshalBlock(buf []byte) (*inter.Block, error) {
	w := new(wire.Block)
	err := proto.Unmarshal(buf, w)
	if err != nil {
		return nil, err
	}

	return inter.WireToBlock(w), nil
}
