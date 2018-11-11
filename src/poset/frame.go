package poset

import (
	"github.com/andrecronje/lachesis/src/crypto"
	"github.com/golang/protobuf/proto"
)

//json encoding of Frame
func (f *Frame) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(f); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (f *Frame) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, f)
}

func (f *Frame) Hash() ([]byte, error) {
	hashBytes, err := f.ProtoMarshal()
	if err != nil {
		return nil, err
	}
	return crypto.SHA256(hashBytes), nil
}

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

func (this *Frame) Equals(that *Frame) bool {
	return this.Round == that.Round &&
		RootListEquals(this.Roots, that.Roots) &&
		EventListEquals(this.Events, that.Events)
}
