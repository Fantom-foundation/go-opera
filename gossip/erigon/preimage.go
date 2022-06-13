package erigon

import (
	//"fmt"
	"os"
	"strings"
	"io"
	"compress/gzip"

	//"gopkg.in/urfave/cli.v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
)

const preimagesPath = "/root/preimages/preimages.gz"

func importPreimages(fn string) (map[common.Hash]common.Address, error) {
	log.Info("Importing preimages", "from file file", fn)

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
	preimages := make(map[common.Hash]common.Address)

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
		key, val := crypto.Keccak256Hash(blob), common.BytesToAddress(common.CopyBytes(blob))
		preimages[key] = val

		/*
		if len(preimages) > 1024 {
			rawdb.WritePreimages(db, preimages)
			preimages = make(map[common.Hash][]byte)
		}
		*/
	}	
	return preimages, nil
}