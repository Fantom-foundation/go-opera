package wire

//go:generate protoc --go_out=plugins=grpc:./ event.proto

import (
	"github.com/golang/protobuf/proto"
)

// Equals return true if transactions are equal.
func (t *InternalTransaction) Equals(t2 *InternalTransaction) bool {
	return t.Amount == t2.Amount &&
		t.Receiver == t2.Receiver
}

// ProtoMarshal marshal to protobuf.
func (t *InternalTransaction) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(t); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unmarshal from protobuf.
func (t *InternalTransaction) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, t)
}
