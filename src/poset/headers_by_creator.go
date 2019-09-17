package poset

import (
	"bytes"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
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

type pair_headersByCreator struct {
	Creator common.Address
	Header  *inter.EventHeaderData
}

func (hh headersByCreator) EncodeRLP(w io.Writer) error {
	arr := make([]pair_headersByCreator, 0, len(hh))
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
		arr = append(arr, pair_headersByCreator{
			Creator: creator,
			Header:  header,
		})
	}
	return rlp.Encode(w, arr)
}

func (phh *headersByCreator) DecodeRLP(s *rlp.Stream) error {
	if *phh == nil {
		*phh = headersByCreator{}
	}
	hh := *phh
	var arr []pair_headersByCreator
	if err := s.Decode(&arr); err != nil {
		return err
	}

	for _, w := range arr {
		hh[w.Creator] = w.Header
	}

	return nil
}
