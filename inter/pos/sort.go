package pos

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
)

type (
	validator struct {
		Addr  common.Address
		Stake Stake
	}

	validators []validator
)

func (vv validators) Less(i, j int) bool {
	if vv[i].Stake != vv[j].Stake {
		return vv[i].Stake > vv[j].Stake
	}

	return bytes.Compare(vv[i].Addr.Bytes(), vv[j].Addr.Bytes()) < 0
}

func (vv validators) Len() int {
	return len(vv)
}

func (vv validators) Swap(i, j int) {
	vv[i], vv[j] = vv[j], vv[i]
}
