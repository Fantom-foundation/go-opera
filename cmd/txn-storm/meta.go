package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
)

type metaInfo struct {
	Created time.Time
	From    uint
	To      uint
}

func makeInfo(from, to uint) *metaInfo {
	return &metaInfo{
		Created: time.Now(),
		From:    from,
		To:      to,
	}
}

func mustParseInfo(bb []byte) *metaInfo {
	m, err := parseInfo(bb)
	if err != nil {
		panic(err)
	}
	return m
}

func parseInfo(bb []byte) (*metaInfo, error) {
	m := new(metaInfo)
	err := rlp.DecodeBytes(bb, m)
	return m, err
}

func (m *metaInfo) Bytes() []byte {
	bb, err := rlp.EncodeToBytes(m)
	if err != nil {
		panic(err)
	}
	return bb
}

func (m *metaInfo) String() string {
	return fmt.Sprintf("%d-->%d", m.From, m.To)
}
