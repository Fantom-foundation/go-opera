package blockproc

import (
	"io"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/inter/sfctype"
)

type ValidatorProfiles map[idx.ValidatorID]sfctype.SfcValidator

func (vv ValidatorProfiles) SortedArray() []sfctype.SfcValidatorAndID {
	builder := pos.NewBigBuilder()
	for id, profile := range vv {
		builder.Set(id, profile.Weight)
	}
	validators := builder.Build()
	sortedIds := validators.SortedIDs()
	arr := make([]sfctype.SfcValidatorAndID, validators.Len())
	for i, id := range sortedIds {
		arr[i] = sfctype.SfcValidatorAndID{
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
	var arr []sfctype.SfcValidatorAndID
	if err := s.Decode(&arr); err != nil {
		return err
	}

	*vv = make(ValidatorProfiles, len(arr))

	for _, it := range arr {
		(*vv)[it.ValidatorID] = it.Validator
	}

	return nil
}
