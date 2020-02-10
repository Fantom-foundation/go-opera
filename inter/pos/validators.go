package pos

import (
	"io"
	"math/big"
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
	// Validators group of an epoch with stakes.
	// Optimized for BFT algorithm calculations.
	// Read-only.
	Validators struct {
		values map[idx.StakerID]Stake
		cache  cache
	}

	// ValidatorsBuilder is a helper to create Validators object
	ValidatorsBuilder map[idx.StakerID]Stake

	// GenesisValidator is helper structure to define genesis validators
	GenesisValidator struct {
		ID      idx.StakerID
		Address common.Address
		Stake   *big.Int
	}

	// GValidators defines genesis validators
	GValidators []GenesisValidator
)

var (
	// EmptyValidators is empty validators group
	EmptyValidators = NewBuilder().Build()
)

// NewBuilder creates new mutable ValidatorsBuilder
func NewBuilder() ValidatorsBuilder {
	return ValidatorsBuilder{}
}

// Set appends item to ValidatorsBuilder object
func (vv ValidatorsBuilder) Set(id idx.StakerID, stake Stake) {
	if stake == 0 {
		delete(vv, id)
	} else {
		vv[id] = stake
	}
}

// Build new read-only Validators object
func (vv ValidatorsBuilder) Build() *Validators {
	return newValidators(vv)
}

// EqualStakeValidators builds new read-only Validators object with equal stakes (for tests)
func EqualStakeValidators(ids []idx.StakerID, stake Stake) *Validators {
	builder := NewBuilder()
	for _, id := range ids {
		builder.Set(id, stake)
	}
	return builder.Build()
}

// ArrayToValidators builds new read-only Validators object from array
func ArrayToValidators(ids []idx.StakerID, stakes []Stake) *Validators {
	builder := NewBuilder()
	for i, id := range ids {
		builder.Set(id, stakes[i])
	}
	return builder.Build()
}

// newValidators builds new read-only Validators object
func newValidators(values ValidatorsBuilder) *Validators {
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

// Len returns count of validators in Validators objects
func (vv *Validators) Len() int {
	return len(vv.values)
}

// calcCaches calculates internal caches for validators
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

// Get returns stake for validator by ID
func (vv *Validators) Get(id idx.StakerID) Stake {
	return vv.values[id]
}

// GetIdx returns index (offset) of validator in the group
func (vv *Validators) GetIdx(id idx.StakerID) idx.Validator {
	return vv.cache.indexes[id]
}

// GetStakeByIdx returns stake for validator by index
func (vv *Validators) GetStakeByIdx(i idx.Validator) Stake {
	return vv.cache.stakes[i]
}

// Exists returns boolean true if address exists in Validators object
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

// SortedStakes returns deterministically sorted stakes.
// The order is the same as for Idxs().
func (vv *Validators) SortedStakes() []Stake {
	return vv.cache.stakes
}

// Idxs gets deterministic total order of validators.
func (vv *Validators) Idxs() map[idx.StakerID]idx.Validator {
	return vv.cache.indexes
}

// sortedArray is sorted by stake and ID
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
	return newValidators(vv.values)
}

// Builder returns a mutable copy of content
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
	for _, validator := range gv {
		builder.Set(validator.ID, BalanceToStake(validator.Stake))
	}
	return builder.Build()
}

// TotalStake returns sum of stakes
func (gv GValidators) TotalStake() *big.Int {
	totalStake := new(big.Int)
	for _, validator := range gv {
		totalStake.Add(totalStake, validator.Stake)
	}
	return totalStake
}

// Map converts GValidators to map
func (gv GValidators) Map() map[idx.StakerID]GenesisValidator {
	validators := map[idx.StakerID]GenesisValidator{}
	for _, validator := range gv {
		validators[validator.ID] = validator
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
