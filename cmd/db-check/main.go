package main

import (
	"os"
	"path/filepath"

	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, ".lachesis")

	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}

	p := leveldb.NewProducer(dir)
	db := p.OpenDb("gossip-main")
	defer db.Close()

	//checkPacks(db)
	checkEvents(db)
}
