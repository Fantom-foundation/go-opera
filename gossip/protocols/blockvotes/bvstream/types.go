package bvstream

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream"
)

type Request struct {
	Session   Session
	Limit     Metric
	Type      basestream.RequestType
	MaxChunks uint32
}

type Response struct {
	SessionID uint32
	Done      bool
	Payload   []rlp.RawValue
}

type Session struct {
	ID    uint32
	Start Locator
	Stop  Locator
}

type Locator []byte

func (l Locator) Compare(b basestream.Locator) int {
	return bytes.Compare(l, b.(Locator))
}

func (l Locator) Inc() basestream.Locator {
	nextBn := new(big.Int).SetBytes(l)
	nextBn.Add(nextBn, common.Big1)
	return Locator(common.LeftPadBytes(nextBn.Bytes(), len(l)))
}

type Payload struct {
	Items []rlp.RawValue
	Keys  []Locator
	Size  uint64
}

func (p *Payload) AddSignedBlockVotes(id Locator, bvsB rlp.RawValue) {
	p.Items = append(p.Items, bvsB)
	p.Keys = append(p.Keys, id)
	p.Size += uint64(len(bvsB))
}

func (p Payload) Len() int {
	return len(p.Keys)
}

func (p Payload) TotalSize() uint64 {
	return p.Size
}

func (p Payload) TotalMemSize() int {
	return int(p.Size) + len(p.Keys)*128
}

type Metric struct {
	Num  idx.Block
	Size uint64
}

func (m Metric) String() string {
	return fmt.Sprintf("{Num=%d,Size=%d}", m.Num, m.Size)
}
