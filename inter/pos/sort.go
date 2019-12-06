package pos

import (
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type (
	validator struct {
		ID    idx.StakerID
		Stake Stake
	}

	validators []validator
)

func (vv validators) Less(i, j int) bool {
	if vv[i].Stake != vv[j].Stake {
		return vv[i].Stake > vv[j].Stake
	}

	return vv[i].ID < vv[j].ID
}

func (vv validators) Len() int {
	return len(vv)
}

func (vv validators) Swap(i, j int) {
	vv[i], vv[j] = vv[j], vv[i]
}
