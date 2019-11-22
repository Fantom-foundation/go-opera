package meta

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type Info struct {
	CreatedUnix uint64
	From        uint
	To          uint
	IsRegular   bool
}

func NewInfo(from, to uint, isRegular bool) *Info {
	return &Info{
		CreatedUnix: uint64(time.Now().Unix()),
		From:        from,
		To:          to,
		IsRegular:   isRegular,
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

func (m *Info) Seconds() int64 {
	return time.Now().Unix() - int64(m.CreatedUnix)
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
