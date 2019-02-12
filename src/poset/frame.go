package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/crypto"
	"github.com/golang/protobuf/proto"
)

// ProtoMarshal json encoding of Frame
func (f *Frame) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(f); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal converts protobuf to frame
func (f *Frame) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, f)
}

// GetEventBlocks provides alternative for non-existent Protobuf generated function
func (f *Frame) GetEventBlocks() []*EventMessage {
	return f.GetEvents()
}

// Hash returns the Hash of a frame
func (f *Frame) Hash() ([]byte, error) {
	hashBytes, err := f.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(hashBytes), nil
}

// RootListEquals compares the equality of two root lists
func RootListEquals(this []*Root, that []*Root) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if !v.Equals(that[i]) {
			return false
		}
	}
	return true
}

// EventListEquals compares the equality of two event lists
func EventListEquals(this []*EventMessage, that []*EventMessage) bool {
	if len(this) != len(that) {
		return false
	}
	for i, v := range this {
		if !v.Equals(that[i]) {
			return false
		}
	}
	return true
}

// Equals compares the equality of two frames
func (f *Frame) Equals(that *Frame) bool {
	return f.Round == that.Round &&
		RootListEquals(f.Roots, that.Roots) &&
		EventListEquals(f.Events, that.Events)
}
