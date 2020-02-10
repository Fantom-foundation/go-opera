package main

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

func checkPacks(db kvdb.KeyValueStore) {
	t := table.New(db, []byte("p"))

	it := t.NewIterator()
	defer it.Release()

	for it.Next() {
		buf := it.Key()
		w := it.Value()

		if strings.HasPrefix(string(buf), "serverPool") {
			fmt.Printf("skip %s key\n", string(buf))
			continue
		}

		var info gossip.PackInfo
		err := rlp.DecodeBytes(w, &info)
		if err != nil {
			fmt.Printf(">>> %s\n ", string(buf))
			continue
		}

		epoch := idx.BytesToEpoch(buf[0:4])
		pack := idx.BytesToEpoch(buf[4:8])
		fmt.Printf("%d:%d %+v\n", epoch, pack, info)
	}
}
