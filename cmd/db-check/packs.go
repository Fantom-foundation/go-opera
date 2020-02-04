package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

func checkPacks(db kvdb.KeyValueStore) {
	t := table.New(db, []byte("p"))

	rmPrefix(t, "serverPool")

	it := t.NewIterator()
	defer it.Release()

	for it.Next() {
		buf := it.Key()
		w := it.Value()

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

func rmPrefix(t kvdb.KeyValueStore, prefix string) {
	t1 := table.New(t, []byte(prefix))
	it := t1.NewIterator()
	for it.Next() {
		if err := t1.Delete(it.Key()); err != nil {
			panic(err)
		}
	}
	it.Release()
}
