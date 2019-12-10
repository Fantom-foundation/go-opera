package pos

import (
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

type (
	cache struct {
		indexes    map[idx.StakerID]idx.Validator
		stakes     []Stake
		ids        []idx.StakerID
		totalStake Stake
	}
	// Validators of epoch with stake.
	Validators struct {
		values map[idx.StakerID]Stake
		cache  cache
	}

	ValidatorsBuilder map[idx.StakerID]Stake

	// GenesisValidator is helper structure to define genesis validators
	GenesisValidator struct {
		ID      idx.StakerID
		Address common.Address
		Stake   Stake
	}

	// GValidators defines genesis validators
	GValidators map[idx.StakerID]GenesisValidator
)

func NewBuilder() ValidatorsBuilder {
	return ValidatorsBuilder{}
}

// Set appends item to Validator object
func (vv ValidatorsBuilder) Set(id idx.StakerID, stake Stake) {
	// add
	if stake == 0 {
		delete(vv, id)
	} else {
		vv[id] = stake
	}
}

func (vv ValidatorsBuilder) Build() *Validators {
	return NewValidators(vv)
}

// NewValidators return new pointer of Validators object
func NewValidators(values ValidatorsBuilder) *Validators {
	valuesCopy := make(ValidatorsBuilder)
	for id, s := range values {
		valuesCopy.Set(id, s)
	}

	vv := &Validators{
		values: valuesCopy,
	}
	vv.cache = vv.calcCaches()
	return vv
}

// Len return count of validators in Validators objects
func (vv *Validators) Len() int {
	return len(vv.values)
}

// Iterate return slice of ids for get validators in loop
func (vv *Validators) Iterate() []idx.StakerID {
	return vv.cache.ids
}

func (vv *Validators) calcCaches() cache {
	cache := cache{
		indexes: make(map[idx.StakerID]idx.Validator),
		stakes:  make([]Stake, vv.Len()),
		ids:     make([]idx.StakerID, vv.Len()),
	}

	for i, v := range vv.sortedArray() {
		cache.indexes[v.ID] = idx.Validator(i)
		cache.stakes[i] = v.Stake
		cache.ids[i] = v.ID
		cache.totalStake += v.Stake
	}

	return cache
}

// Get return stake for validator address
func (vv *Validators) Get(id idx.StakerID) Stake {
	return vv.values[id]
}

func (vv *Validators) GetIdx(id idx.StakerID) idx.Validator {
	return vv.cache.indexes[id]
}

// Get return stake for validator address
func (vv *Validators) GetByIdx(i idx.Validator) Stake {
	return vv.cache.stakes[i]
}

// Exists return boolean true if address exists in Validators object
func (vv *Validators) Exists(id idx.StakerID) bool {
	_, ok := vv.values[id]
	return ok
}

// IDs returns not sorted ids.
func (vv *Validators) IDs() []idx.StakerID {
	return vv.cache.ids
}

// SortedIDs returns deterministically sorted ids.
// The order is the same as for Idxs().
func (vv *Validators) SortedIDs() []idx.StakerID {
	return vv.cache.ids
}

// Idxs gets deterministic total order of validators.
func (vv *Validators) Idxs() map[idx.StakerID]idx.Validator {
	return vv.cache.indexes
}

func (vv *Validators) sortedArray() validators {
	array := make(validators, 0, len(vv.values))
	for id, s := range vv.values {
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
	return NewValidators(vv.values)
}

func (vv *Validators) Builder() ValidatorsBuilder {
	return vv.Copy().values
}

// Quorum limit of validators.
func (vv *Validators) Quorum() Stake {
	return vv.TotalStake()*2/3 + 1
}

// TotalStake of validators.
func (vv *Validators) TotalStake() (sum Stake) {
	return vv.cache.totalStake
}

// EncodeRLP is for RLP serialization.
func (vv *Validators) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, vv.sortedArray())
}

// DecodeRLP is for RLP deserialization.
func (vv *Validators) DecodeRLP(s *rlp.Stream) error {
	var arr []validator
	if err := s.Decode(&arr); err != nil {
		return err
	}

	builder := NewBuilder()
	for _, w := range arr {
		builder.Set(w.ID, w.Stake)
	}
	*vv = *builder.Build()

	return nil
}

// Validators converts GValidators to Validators
func (gv GValidators) Validators() *Validators {
	builder := NewBuilder()
	for stakerID, validator := range gv {
		builder.Set(stakerID, validator.Stake)
	}
	return builder.Build()
}

// Addresses returns not sorted genesis addresses
func (gv GValidators) Addresses() []common.Address {
	res := make([]common.Address, 0, len(gv))
	for _, v := range gv {
		res = append(res, v.Address)
	}
	return res
}
