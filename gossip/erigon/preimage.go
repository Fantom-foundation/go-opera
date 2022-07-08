package erigon

import (
	//"fmt"
	"time"
	"os"
	"strings"
	"io"
	"compress/gzip"

	//"gopkg.in/urfave/cli.v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
	
	ecommon "github.com/ledgerwatch/erigon/common"
)


const  (
	MainnnetPreimagesCount = 143168825
	defaultPreimagesPath = "/root/preimages/preimages.gz"

)

// importPreimages imports preimages acc.Hash -> account.Address from file and saves it into hashmap
func importPreimages(fn string) (map[common.Hash]ecommon.Address, error) {
	log.Info("Import of preimages started....")
    start := time.Now()
	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	var reader io.Reader = fh
	if strings.HasSuffix(fn, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return nil, err
		}
	}
	stream := rlp.NewStream(reader, 0)

	// Import the preimages in batches to prevent disk trashing
	preimages := make(map[common.Hash]ecommon.Address)
	i := 0
	for {
		
		// Read the next entry and ensure it's not junk
		var blob []byte

		if err := stream.Decode(&blob); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// Accumulate the preimages and flush when enough ws gathered
		key, val := crypto.Keccak256Hash(blob), ecommon.BytesToAddress(common.CopyBytes(blob))
		preimages[key] = val

		/*
		if len(preimages) > 1024 {
			rawdb.WritePreimages(db, preimages)
			preimages = make(map[common.Hash][]byte)
		}
		*/
		i++
	}	
	log.Info("Import preimages is complete", "elapsed", common.PrettyDuration(time.Since(start)))
	log.Info("Total amount of",  "imported preimages", i)
	return preimages, nil

}