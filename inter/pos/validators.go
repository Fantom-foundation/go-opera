package pos

import (
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type (
	// Validators of epoch with stake.
	Validators struct {
		indexes map[idx.StakerID]int
		list    []Stake
		ids     []idx.StakerID
	}

	// GenesisValidator is helper structure to define genesis validators
	GenesisValidator struct {
		ID      idx.StakerID
		Address common.Address
		Stake   Stake
	}

	// GValidators defines genesis validators
	GValidators map[idx.StakerID]GenesisValidator
)

// NewValidators return new pointer of Validators object
func NewValidators() *Validators {
	return &Validators{
		indexes: make(map[idx.StakerID]int),
		list:    make([]Stake, 0, 200),
		ids:     make([]idx.StakerID, 0, 200),
	}
}

// Len return count of validators in Validators objects
func (vv Validators) Len() int {
	return len(vv.list)
}

// Set appends item to Validator object
func (vv *Validators) Set(id idx.StakerID, stake Stake) {
	if stake != 0 {
		i, ok := vv.indexes[id]
		if ok {
			vv.list[i] = stake
			return
		}
		vv.list = append(vv.list, stake)
		vv.ids = append(vv.ids, id)
		vv.indexes[id] = len(vv.list) - 1
	} else {
		i, ok := vv.indexes[id]
		if ok {
			delete(vv.indexes, id)
			idxOrig := len(vv.list) - 1
			if i == idxOrig {
				vv.list = vv.list[:idxOrig]
				vv.ids = vv.ids[:idxOrig]
			} else {
				// Move last to deleted position + truncate list len
				vv.list[i] = vv.list[idxOrig]
				vv.list = vv.list[:idxOrig]

				vv.indexes[vv.ids[idxOrig]] = i

				vv.ids[i] = vv.ids[idxOrig]
				vv.ids = vv.ids[:idxOrig]
			}
		}
	}
}

// Get return stake for validator address
func (vv Validators) Get(id idx.StakerID) Stake {
	i, ok := vv.indexes[id]
	if ok {
		return vv.list[i]
	}
	return 0
}

// Exists return boolean true if address exists in Validators object
func (vv Validators) Exists(id idx.StakerID) bool {
	_, ok := vv.indexes[id]
	return ok
}

// IDs returns not sorted ids.
func (vv Validators) IDs() []idx.StakerID {
	return vv.ids
}

// SortedIDs returns deterministically sorted ids.
// The order is the same as for Idxs().
func (vv Validators) SortedIDs() []idx.StakerID {
	array := make([]idx.StakerID, len(vv.list))
	for i, s := range vv.sortedArray() {
		array[i] = s.ID
	}
	return array
}

// Idxs gets deterministic total order of validators.
func (vv Validators) Idxs() map[idx.StakerID]idx.Validator {
	idxs := make(map[idx.StakerID]idx.Validator, len(vv.list))
	for i, v := range vv.sortedArray() {
		idxs[v.ID] = idx.Validator(i)
	}
	return idxs
}

func (vv Validators) sortedArray() validators {
	array := make(validators, 0, len(vv.list))
	for id, i := range vv.indexes {
		s := vv.list[i]
		array = append(array, validator{
			ID:    id,
			Stake: s,
		})
	}
	sort.Sort(array)
	return array
}

// Copy constructs a copy.
func (vv *Validators) Copy() *Validators {
	res := NewValidators()

	if cap(res.list) < len(vv.list) {
		res.list = make([]Stake, len(vv.list))
		res.ids = make([]idx.StakerID, len(vv.list))
	}
	res.list = res.list[0:len(vv.list)]
	res.ids = res.ids[0:len(vv.list)]
	copy(res.list, vv.list)
	copy(res.ids, vv.ids)

	for id, i := range vv.indexes {
		res.indexes[id] = i
	}

	return res
}

// Quorum limit of validators.
func (vv Validators) Quorum() Stake {
	return vv.TotalStake()*2/3 + 1
}

// TotalStake of validators.
func (vv Validators) TotalStake() (sum Stake) {
	for _, s := range vv.list {
		sum += s
	}
	return
}

// StakeOf validator.
func (vv Validators) StakeOf(id idx.StakerID) Stake {
	return vv.Get(id)
}

// EncodeRLP is for RLP serialization.
func (vv Validators) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, vv.sortedArray())
}

// DecodeRLP is for RLP deserialization.
func (vv *Validators) DecodeRLP(s *rlp.Stream) error {
	if vv == nil {
		vv = NewValidators()
	}
	if vv.ids == nil {
		vv.ids = make([]idx.StakerID, 0, 200)
	}
	if vv.indexes == nil {
		vv.indexes = make(map[idx.StakerID]int)
	}
	if vv.list == nil {
		vv.list = make([]Stake, 0, 200)
	}

	var arr []validator
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		vv.Set(w.ID, w.Stake)
	}

	return nil
}

// Validators converts GValidators to Validators
func (gv GValidators) Validators() *Validators {
	validators := NewValidators()
	for stakerID, validator := range gv {
		validators.Set(stakerID, validator.Stake)
	}
	return validators
}

// Addresses returns not sorted genesis addresses
func (gv GValidators) Addresses() []common.Address {
	res := make([]common.Address, 0, len(gv))
	for _, v := range gv {
		res = append(res, v.Address)
	}
	return res
}
