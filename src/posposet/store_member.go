package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/wire"
)

// SetMember stores members by index
func (s *Store) SetMember(n uint64, members []hash.Peer) {
	key := common.IntToBytes(n)
	msg := &wire.Members{
		Members: make([][]byte, 0, len(members)),
	}

	for _, addr := range members {
		msg.Members = append(msg.Members, addr.Bytes())
	}

	s.set(s.table.Members, key, msg)
}

// GetMembers returns stored members by index
func (s *Store) GetMembers(n uint64) []hash.Peer {
	key := common.IntToBytes(n)

	msg := s.get(s.table.Members, key, &wire.Members{}).(*wire.Members)

	members := make([]hash.Peer, 0, len(msg.Members))

	for _, member := range msg.Members {
		members = append(members, hash.BytesToPeer(member))
	}

	return members
}
