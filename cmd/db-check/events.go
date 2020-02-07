package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

func checkEvents(db kvdb.KeyValueStore) {
	t := table.New(db, []byte("e"))

	it := t.NewIterator()
	defer it.Release()

	found := make(map[hash.Event]struct{})
	losts := make(map[hash.Event]struct{})

	exists := func(e hash.Event) bool {
		_, ok := found[e]
		return ok
	}

	for it.Next() {
		event := &inter.Event{}
		err := rlp.DecodeBytes(it.Value(), event)
		if err != nil {
			panic(err)
		}

		found[event.Hash()] = struct{}{}

		for _, p := range event.Parents {
			if exists(p) {
				continue
			}
			losts[p] = struct{}{}
		}
	}

	// sanity check
	for p := range losts {
		if exists(p) {
			panic("event order")
		}
		fmt.Printf("%v\n", p)
	}
}
