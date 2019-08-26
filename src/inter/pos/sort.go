package pos

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
)

type (
	member struct {
		Addr  common.Address
		Stake Stake
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
