package inter

import (
	"io"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter/drivertype"
)

type ValidatorProfiles map[idx.ValidatorID]drivertype.Validator

func (vv ValidatorProfiles) Copy() ValidatorProfiles {
	cp := make(ValidatorProfiles, len(vv))
	for k, v := range vv {
		cpv := v
		cpv.Weight = new(big.Int).Set(cpv.Weight)
		cpv.PubKey = cpv.PubKey.Copy()
		cp[k] = cpv
	}
	return cp
}

func (vv ValidatorProfiles) SortedArray() []drivertype.ValidatorAndID {
	builder := pos.NewBigBuilder()
	for id, profile := range vv {
		builder.Set(id, profile.Weight)
	}
	validators := builder.Build()
	sortedIds := validators.SortedIDs()
	arr := make([]drivertype.ValidatorAndID, validators.Len())
	for i, id := range sortedIds {
		arr[i] = drivertype.ValidatorAndID{
			ValidatorID: id,
			Validator:   vv[id],
		}
	}
	return arr
}

// EncodeRLP is for RLP serialization.
func (vv ValidatorProfiles) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, vv.SortedArray())
}

// DecodeRLP is for RLP deserialization.
func (vv *ValidatorProfiles) DecodeRLP(s *rlp.Stream) error {
	var arr []drivertype.ValidatorAndID
	if err := s.Decode(&arr); err != nil {
		return err
	}

	*vv = make(ValidatorProfiles, len(arr))

	for _, it := range arr {
		(*vv)[it.ValidatorID] = it.Validator
	}

	return nil
}
