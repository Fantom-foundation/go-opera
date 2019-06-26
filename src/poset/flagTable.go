//go:generate protoc --go_out=plugins=grpc:./ flagTableWrapper.proto
package poset

import (
	"github.com/golang/protobuf/proto"
)

// FlagTable is a dedicated type for the Events flags map.
type FlagTable map[EventHash]int64

// NewFlagTable creates new empty FlagTable
func NewFlagTable() FlagTable {
	return FlagTable(make(map[EventHash]int64))
}

// Marshal converts FlagTable to protobuf.
func (ft FlagTable) Marshal() []byte {
	body := make(map[string]int64, len(ft))
	for k, v := range ft {
		body[k.String()] = v
	}

	wrapper := &FlagTableWrapper{Body: body}
	bytes, err := proto.Marshal(wrapper)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Unmarshal reads protobuff into FlagTable.
func (ft FlagTable) Unmarshal(buf []byte) error {
	wrapper := new(FlagTableWrapper)
	err := proto.Unmarshal(buf, wrapper)
	if err != nil {
		return err
	}

	for k, v := range wrapper.Body {
		var hash EventHash
		err = hash.Parse(k)
		if err != nil {
			return err
		}
		ft[hash] = v
	}
	return nil
}
