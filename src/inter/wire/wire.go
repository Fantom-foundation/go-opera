package wire

//go:generate protoc --go_out=plugins=grpc,paths=source_relative:./ wire.proto

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
	var pbf proto.Buffer
	pbf.SetDeterministic(true)
	if err := pbf.Marshal(t); err != nil {
		return nil, err
	}
	return pbf.Bytes(), nil
}

// ProtoUnmarshal unmarshal from protobuf.
func (t *InternalTransaction) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, t)
}

// ProtoMarshal marshal to protobuf.
func (e *Event) ProtoMarshal() ([]byte, error) {
	var pbf proto.Buffer
	pbf.SetDeterministic(true)
	if err := pbf.Marshal(e); err != nil {
		return nil, err
	}
	return pbf.Bytes(), nil
}

// ProtoUnmarshal unmarshal from protobuf.
func (e *Event) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, e)
}
