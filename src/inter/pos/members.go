package pos

import (
	"io"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// MembersCount in top set.
const MembersCount = 30

type (
	// Members of super-frame with stake.
	Members map[common.Address]Stake
)

// Set appends item.
func (mm *Members) Set(addr common.Address, stake Stake) {
	if stake != 0 {
		(*mm)[addr] = stake
	} else {
		delete((*mm), addr)
	}
}

func (mm Members) sortedArray() members {
	array := make(members, 0, len(mm))
	for n, s := range mm {
		array = append(array, member{
			Addr:  n,
			Stake: s,
		})
	}
	sort.Sort(array)
	return array
}

// Top gets top subset.
func (mm Members) Top() Members {
	top := mm.sortedArray()

	if len(top) > MembersCount {
		top = top[:MembersCount]
	}

	res := make(Members)
	for _, m := range top {
		res.Set(m.Addr, m.Stake)
	}

	return res
}

// Copy constructs a copy.
func (mm Members) Copy() Members {
	res := make(Members)
	for addr, stake := range mm {
		res.Set(addr, stake)
	}
	return res
}

// Idxs gets deterministic total order of members.
func (mm Members) Idxs() map[common.Address]idx.Member {
	idxs := make(map[common.Address]idx.Member, len(mm))
	for i, m := range mm.sortedArray() {
		idxs[m.Addr] = idx.Member(i)
	}
	return idxs
}

// Quorum limit of members.
func (mm Members) Quorum() Stake {
	return mm.TotalStake()*2/3 + 1
}

// TotalStake of members.
func (mm Members) TotalStake() (sum Stake) {
	for _, s := range mm {
		sum += s
	}
	return
}

// StakeOf member.
func (mm Members) StakeOf(n common.Address) Stake {
	return mm[n]
}

func (mm Members) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, mm.sortedArray())
}

func (pp *Members) DecodeRLP(s *rlp.Stream) error {
	if *pp == nil {
		*pp = Members{}
	}
	mm := *pp

	var arr []member
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		mm[w.Addr] = w.Stake
	}

	return nil
}
