package poset

import (
	"bytes"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"sort"
)

type (
	// headersByCreator is a event hashes grouped by creator.
	// ( creator --> event header )
	headersByCreator map[common.Address]*inter.EventHeaderData
)

type headersByCreatorPair struct {
	Creator common.Address
	Header  *inter.EventHeaderData
}

func (hh headersByCreator) EncodeRLP(w io.Writer) error {
	arr := make([]headersByCreatorPair, 0, len(hh))
	creators := make([]common.Address, 0, len(hh))
	for creator := range hh {
		creators = append(creators, creator)
	}
	// for determinism
	sort.Slice(creators, func(i, j int) bool {
		a, b := creators[i], creators[j]
		return bytes.Compare(a.Bytes(), b.Bytes()) < 0
	})

	for _, creator := range creators {
		header := hh[creator]
		arr = append(arr, headersByCreatorPair{
			Creator: creator,
			Header:  header,
		})
	}
	return rlp.Encode(w, arr)
}

func (hh *headersByCreator) DecodeRLP(s *rlp.Stream) error {
	if *hh == nil {
		*hh = headersByCreator{}
	}
	var arr []headersByCreatorPair
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		(*hh)[w.Creator] = w.Header
	}

	return nil
}
