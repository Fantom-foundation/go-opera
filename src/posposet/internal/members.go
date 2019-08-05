package internal

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// MembersCount in top set.
const MembersCount = 30

type (
	// Members of super-frame with stake.
	Members map[hash.Peer]inter.Stake
)

// Add appends item.
func (mm *Members) Add(addr hash.Peer, stake inter.Stake) {
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
		res.Add(m.Addr, m.Stake)
	}

	return res
}

// Deterministic total order of members.
func (mm Members) Idxs() map[hash.Peer]idx.Member {
	idxs := make(map[hash.Peer]idx.Member, len(mm))
	for i, m := range mm.sortedArray() {
		idxs[m.Addr] = idx.Member(i)
	}
	return idxs
}

// Quorum limit of members.
func (mm Members) Quorum() inter.Stake {
	return mm.TotalStake()*2/3 + 1
}

// TotalStake of members.
func (mm Members) TotalStake() (sum inter.Stake) {
	for _, s := range mm {
		sum += s
	}
	return
}

// StakeOf member.
func (mm Members) StakeOf(n hash.Peer) inter.Stake {
	return mm[n]
}

func (mm Members) EncodeRLP(w io.Writer) error {
	var arr []member
	for addr, stake := range mm {
		arr = append(arr, member{
			Addr:  addr,
			Stake: stake,
		})
	}
	return rlp.Encode(w, arr)
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
