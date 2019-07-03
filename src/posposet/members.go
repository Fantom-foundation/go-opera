package posposet

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

const MembersCount = 30

type (
	member struct {
		Addr  hash.Peer
		Stake uint64
	}

	members []member
)

func (mm *members) Add(addr hash.Peer, stake uint64) {
	*mm = append(*mm, member{
		Addr:  addr,
		Stake: stake,
	})
}

func (mm members) Top() members {
	top := make(members, len(mm))
	copy(top, mm)

	sort.Sort(top)
	if len(top) > MembersCount {
		top = top[:MembersCount]
	}

	return top
}

func (mm members) ToWire() *wire.Members {
	w := &wire.Members{}

	for _, m := range mm {
		w.Members = append(w.Members, m.ToWire())
	}

	return w
}

func (m *member) ToWire() *wire.Member {
	return &wire.Member{
		Addr:  m.Addr.Bytes(),
		Stake: m.Stake,
	}
}

func WireToMembers(w *wire.Members) members {
	if w == nil {
		return nil
	}

	mm := make(members, len(w.Members))
	for i, m := range w.Members {
		mm[i] = member{
			Addr:  hash.BytesToPeer(m.Addr),
			Stake: m.Stake,
		}
	}

	return mm
}

/*
 * sort interface:
 */

func (mm members) Less(i, j int) bool {
	if mm[i].Stake != mm[j].Stake {
		return mm[i].Stake > mm[j].Stake
	}

	return bytes.Compare(mm[i].Addr.Bytes(), mm[j].Addr.Bytes()) < 0
}

func (mm members) Len() int {
	return len(mm)
}

func (mm members) Swap(i, j int) {
	mm[i], mm[j] = mm[j], mm[i]
}
