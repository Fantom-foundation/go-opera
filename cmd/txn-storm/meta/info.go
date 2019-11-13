package meta

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type Info struct {
	Created time.Time
	From    string
	To      uint
}

func NewInfo(from string, to uint) *Info {
	return &Info{
		Created: time.Now(),
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

func (m *Info) Bytes() []byte {
	bb, err := rlp.EncodeToBytes(m)
	if err != nil {
		panic(err)
	}
	return bb
}

func (m *Info) String() string {
	return fmt.Sprintf("%s-->%d", m.From, m.To)
}
