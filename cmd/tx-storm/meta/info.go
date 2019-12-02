package meta

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type Info struct {
	Created uint64
	From    uint
	To      uint
}

func NewInfo(from, to uint) *Info {
	return &Info{
		Created: uint64(time.Now().UnixNano()),
		From:    from,
		To:      to,
	}
}

func MustParseInfo(bb []byte) *Info {
	m, err := ParseInfo(bb)
	if err != nil {
		panic(err)
	}
	return m
}

func ParseInfo(bb []byte) (*Info, error) {
	m := new(Info)
	err := rlp.DecodeBytes(bb, m)
	return m, err
}

func (m *Info) Nanoseconds() int64 {
	return time.Now().UnixNano() - int64(m.Created)
}

func (m *Info) Bytes() []byte {
	bb, err := rlp.EncodeToBytes(m)
	if err != nil {
		panic(err)
	}
	return bb
}

func (m *Info) String() string {
	return fmt.Sprintf("%d-->%d", m.From, m.To)
}
