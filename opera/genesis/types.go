package genesis

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
)

type (
	Hashes map[string]hash.Hash
	Header struct {
		GenesisID   hash.Hash
		NetworkID   uint64
		NetworkName string
	}
	Blocks interface {
		ForEach(fn func(ibr.LlrIdxFullBlockRecord) bool)
	}
	Epochs interface {
		ForEach(fn func(ier.LlrIdxFullEpochRecord) bool)
	}
	EvmItems interface {
		ForEach(fn func(key, value []byte) bool)
	}
	Genesis struct {
		Header

		Blocks      Blocks
		Epochs      Epochs
		RawEvmItems EvmItems
	}
)

func (hh Hashes) Includes(hh2 Hashes) bool {
	for n, h := range hh {
		if hh2[n] != h {
			return false
		}
	}
	return true
}

func (hh Hashes) Equal(hh2 Hashes) bool {
	return hh.Includes(hh2) && hh2.Includes(hh)
}

func (hh Hashes) String() string {
	bb := make([]string, 0, len(hh))
	for n, h := range hh {
		bb = append(bb, fmt.Sprintf("%s: %s", n, h.String()))
	}
	sort.Strings(bb)
	return "{" + strings.Join(bb, ", ") + "}"
}

func (h Header) Equal(h2 Header) bool {
	return h == h2
}

func (h Header) String() string {
	return fmt.Sprintf("{%d, net:%s, id:%s}", h.NetworkID, h.NetworkName, h.GenesisID.String())
}
