package members

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

const MembersCount = 30

type (
	Member struct {
		Addr  hash.Peer
		Stake uint64
	}

	Members []Member
)

func (mm *Members) Add(addr hash.Peer, stake uint64) {
	*mm = append(*mm, Member{
		Addr:  addr,
		Stake: stake,
	})
}

func (mm Members) Top() Members {
	top := make(Members, len(mm))
	copy(top, mm)

	sort.Sort(top)
	if len(top) > MembersCount {
		top = top[:MembersCount]
	}

	return top
}

func (mm Members) TotalStake() (sum uint64) {
	for _, m := range mm {
		sum += m.Stake
	}
	return
}

func (mm Members) ToWire() *wire.Members {
	w := &wire.Members{}

	for _, m := range mm {
		w.Members = append(w.Members, m.ToWire())
	}

	return w
}

func (m *Member) ToWire() *wire.Member {
	return &wire.Member{
		Addr:  m.Addr.Bytes(),
		Stake: m.Stake,
	}
}

func WireToMembers(w *wire.Members) Members {
	if w == nil {
		return nil
	}

	mm := make(Members, len(w.Members))
	for i, m := range w.Members {
		mm[i] = Member{
			Addr:  hash.BytesToPeer(m.Addr),
			Stake: m.Stake,
		}
	}

	return mm
}

/*
 * sort interface:
 */

func (mm Members) Less(i, j int) bool {
	if mm[i].Stake != mm[j].Stake {
		return mm[i].Stake > mm[j].Stake
	}

	return bytes.Compare(mm[i].Addr.Bytes(), mm[j].Addr.Bytes()) < 0
}

func (mm Members) Len() int {
	return len(mm)
}

func (mm Members) Swap(i, j int) {
	mm[i], mm[j] = mm[j], mm[i]
}
