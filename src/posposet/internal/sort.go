package internal

import (
	"bytes"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type (
	member struct {
		Addr  hash.Peer
		Stake inter.Stake
	}

	members []member
)

func (mm members) Less(i, j int) bool {
	if mm[i].Stake != mm[j].Stake {
		return mm[i].Stake > mm[j].Stake
	}

	return bytes.Compare(mm[i].Addr.Bytes(), mm[j].Addr.Bytes()) < 0
}

func (mm members) Len() int {
	return len(mm)
}

func (mm members) Swap(i, j int) {
	mm[i], mm[j] = mm[j], mm[i]
}
