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
		indexes   map[common.Address]int
		list      []Stake
		addresses []common.Address
	}
)

// NewValidators return new pointer of Validators object
func NewValidators() *Validators {
	return &Validators{
		indexes:   make(map[common.Address]int),
		list:      make([]Stake, 0, 200),
		addresses: make([]common.Address, 0, 200),
	}
}

// Len return count of validators in Validators objects
func (vv Validators) Len() int {
	return len(vv.list)
}

// Iterate return chanel of common.Address for get validators in loop
func (vv Validators) Iterate() <-chan common.Address {
	c := make(chan common.Address)
	go func() {
		for _, a := range vv.addresses {
			c <- a
		}
		close(c)
	}()
	return c
}

// Set appends item to Validator object
func (vv *Validators) Set(addr common.Address, stake Stake) {
	if stake != 0 {
		i, ok := vv.indexes[addr]
		if ok {
			vv.list[i] = stake
			return
		}
		vv.list = append(vv.list, stake)
		vv.addresses = append(vv.addresses, addr)
		vv.indexes[addr] = len(vv.list) - 1
	} else {
		i, ok := vv.indexes[addr]
		if ok {
			delete(vv.indexes, addr)
			idxOrig := len(vv.list) - 1
			if i == idxOrig {
				vv.list = vv.list[:idxOrig]
				vv.addresses = vv.addresses[:idxOrig]
			} else {
				// Move last to deleted position + truncate list len
				vv.list[i] = vv.list[idxOrig]
				vv.list = vv.list[:idxOrig]

				vv.indexes[vv.addresses[idxOrig]] = i

				vv.addresses[i] = vv.addresses[idxOrig]
				vv.addresses = vv.addresses[:idxOrig]
			}
		}
	}
}

// Get return stake for validator address
func (vv Validators) Get(addr common.Address) Stake {
	i, ok := vv.indexes[addr]
	if ok {
		return vv.list[i]
	}
	return 0
}

// Exists return boolean true if address exists in Validators object
func (vv Validators) Exists(addr common.Address) bool {
	_, ok := vv.indexes[addr]
	return ok
}

// Addresses returns not sorted addresses.
func (vv Validators) Addresses() []common.Address {
	return vv.addresses
}

// SortedAddresses returns deterministically sorted addresses.
// The order is the same as for Idxs().
func (vv Validators) SortedAddresses() []common.Address {
	array := make([]common.Address, len(vv.list))
	for i, s := range vv.sortedArray() {
		array[i] = s.Addr
	}
	return array
}

// Idxs gets deterministic total order of validators.
func (vv Validators) Idxs() map[common.Address]idx.Validator {
	idxs := make(map[common.Address]idx.Validator, len(vv.list))
	for i, v := range vv.sortedArray() {
		idxs[v.Addr] = idx.Validator(i)
	}
	return idxs
}

func (vv Validators) sortedArray() validators {
	array := make(validators, 0, len(vv.list))
	for addr, i := range vv.indexes {
		s := vv.list[i]
		array = append(array, validator{
			Addr:  addr,
			Stake: s,
		})
	}
	sort.Sort(array)
	return array
}

// Copy constructs a copy.
func (vv Validators) Copy() Validators {
	res := NewValidators()

	if cap(res.list) < len(vv.list) {
		res.list = make([]Stake, len(vv.list))
		res.addresses = make([]common.Address, len(vv.list))
	}
	res.list = res.list[0:len(vv.list)]
	res.addresses = res.addresses[0:len(vv.list)]
	copy(res.list, vv.list)
	copy(res.addresses, vv.addresses)

	for addr, i := range vv.indexes {
		res.indexes[addr] = i
	}

	return *res
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
func (vv Validators) StakeOf(n common.Address) Stake {
	return vv.Get(n)
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
	if vv.addresses == nil {
		vv.addresses = make([]common.Address, 0, 200)
	}
	if vv.indexes == nil {
		vv.indexes = make(map[common.Address]int)
	}
	if vv.list == nil {
		vv.list = make([]Stake, 0, 200)
	}

	var arr []validator
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		vv.Set(w.Addr, w.Stake)
	}

	return nil
}
