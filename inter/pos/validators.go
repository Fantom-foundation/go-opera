package pos

import (
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// TODO: move it to config
const (
	// ValidatorsMax in top set.
	ValidatorsMax = 30
	// Qualification is a minimal validator's stake.
	Qualification Stake = 1e6
)

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
// The order is the same as for Idxs().
func (vv Validators) SortedAddresses() []common.Address {
	array := make([]common.Address, len(vv))
	for i, s := range vv.sortedArray() {
		array[i] = s.Addr
	}
	return array
}

// Idxs gets deterministic total order of validators.
func (vv Validators) Idxs() map[common.Address]idx.Validator {
	idxs := make(map[common.Address]idx.Validator, len(vv))
	for i, v := range vv.sortedArray() {
		idxs[v.Addr] = idx.Validator(i)
	}
	return idxs
}

func (vv Validators) sortedArray() validators {
	array := make(validators, 0, len(vv))
	for addr, s := range vv {
		array = append(array, validator{
			Addr:  addr,
			Stake: s,
		})
	}
	sort.Sort(array)
	return array
}

// Top gets top subset.
func (vv Validators) Top() Validators {
	top := vv.sortedArray()

	for i, v := range top {
		if v.Stake < Qualification {
			top = top[:i]
			break
		}
	}

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
