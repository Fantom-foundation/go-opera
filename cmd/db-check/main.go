package main

import (
	"os"

	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
)

func main() {
	dir := "~/.lachesis"
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}

	//dir = "/home/fabel/Work/fantom/go-lachesis/build/20200204"

	p := leveldb.NewProducer(dir)
	db := p.OpenDb("gossip-main")
	defer db.Close()

	// checkPacks(db)
	checkEvents(db)
}
