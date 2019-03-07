package bench

import (
	"sort"

	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/posposet"
)

type Event struct {
	EventBody
	Parents posposet.EventHashes
}

func (e *Event) ProtoMarshal() ([]byte, error) {
	parents := e.Parents.Slice()
	sort.Sort(parents)
	e.EventBody.Parents = make([][]byte, len(parents))
	for i := 0; i < len(parents); i++ {
		e.EventBody.Parents[i] = parents[i].Bytes()
	}

	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(&e.EventBody); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (e *Event) ProtoUnmarshal(data []byte) error {
	err := proto.Unmarshal(data, &e.EventBody)
	if err != nil {
		return err
	}

	e.Parents = posposet.EventHashes{}

	for _, buf := range e.EventBody.Parents {
		hash := posposet.EventHash(common.BytesToHash(buf))
		e.Parents.Add(hash)
	}
	return nil
}
