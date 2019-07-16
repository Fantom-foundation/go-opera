package internal

import (
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

const MembersCount = 30

type (
	Members map[hash.Peer]inter.Stake
)

func (mm *Members) Add(addr hash.Peer, stake inter.Stake) {
	(*mm)[addr] = stake
}

func (mm Members) Top() Members {
	top := make(members, 0, len(mm))
	for n, s := range mm {
		top = append(top, member{
			Addr:  n,
			Stake: s,
		})
	}

	sort.Sort(top)
	if len(top) > MembersCount {
		top = top[:MembersCount]
	}

	res := make(Members)
	for _, m := range top {
		res.Add(m.Addr, m.Stake)
	}

	return res
}

func (mm Members) Quorum() inter.Stake {
	return mm.TotalStake()*2/3 + 1
}

func (mm Members) TotalStake() (sum inter.Stake) {
	for _, s := range mm {
		sum += s
	}
	return
}

func (mm Members) StakeOf(n hash.Peer) inter.Stake {
	return mm[n]
}

func (mm Members) ToWire() map[string]uint64 {
	w := make(map[string]uint64)

	for n, s := range mm {
		w[n.Hex()] = uint64(s)
	}

	return w
}

func WireToMembers(w map[string]uint64) Members {
	if w == nil {
		return nil
	}

	mm := make(Members, len(w))
	for hex, amount := range w {
		addr := hash.HexToPeer(hex)
		mm[addr] = inter.Stake(amount)
	}

	return mm
}
