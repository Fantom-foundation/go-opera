package pos

import (
	"bytes"
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// ValidatorsMax in top set.
const ValidatorsMax = 30

type (
	// Validators of epoch with stake.
	Validators map[common.Address]Stake
)

// Set appends item.
func (vv *Validators) Set(addr common.Address, stake Stake) {
	if stake != 0 {
		(*vv)[addr] = stake
	} else {
		delete((*vv), addr)
	}
}

// Addresses returns not sorted addresses.
func (vv Validators) Addresses() []common.Address {
	array := make([]common.Address, 0, len(vv))
	for n := range vv {
		array = append(array, n)
	}
	return array
}

// SortedAddresses returns deterministically sorted addresses.
func (vv Validators) SortedAddresses() []common.Address {
	array := vv.Addresses()
	sort.Slice(array, func(i, j int) bool {
		a, b := array[i], array[j]
		return bytes.Compare(a.Bytes(), b.Bytes()) < 0
	})
	return array
}

func (vv Validators) sortedArray() validators {
	array := make(validators, 0, len(vv))
	for n, s := range vv {
		array = append(array, validator{
			Addr:  n,
			Stake: s,
		})
	}
	sort.Sort(array)
	return array
}

// Top gets top subset.
func (vv Validators) Top() Validators {
	top := vv.sortedArray()

	if len(top) > ValidatorsMax {
		top = top[:ValidatorsMax]
	}

	res := make(Validators)
	for _, v := range top {
		res.Set(v.Addr, v.Stake)
	}

	return res
}

// Copy constructs a copy.
func (vv Validators) Copy() Validators {
	res := make(Validators)
	for addr, stake := range vv {
		res.Set(addr, stake)
	}
	return res
}

// Idxs gets deterministic total order of validators.
func (vv Validators) Idxs() map[common.Address]idx.Validator {
	idxs := make(map[common.Address]idx.Validator, len(vv))
	for i, v := range vv.sortedArray() {
		idxs[v.Addr] = idx.Validator(i)
	}
	return idxs
}

// Quorum limit of validators.
func (vv Validators) Quorum() Stake {
	return vv.TotalStake()*2/3 + 1
}

// TotalStake of validators.
func (vv Validators) TotalStake() (sum Stake) {
	for _, s := range vv {
		sum += s
	}
	return
}

// StakeOf validator.
func (vv Validators) StakeOf(n common.Address) Stake {
	return vv[n]
}

// EncodeRLP is for RLP serialization.
func (vv Validators) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, vv.sortedArray())
}

// DecodeRLP is for RLP deserialization.
func (pp *Validators) DecodeRLP(s *rlp.Stream) error {
	if *pp == nil {
		*pp = Validators{}
	}
	vv := *pp

	var arr []validator
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		vv[w.Addr] = w.Stake
	}

	return nil
}
