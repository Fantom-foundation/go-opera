package erigon

import (
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/crypto"
)

// THis path denotes path to preimages in profiling6 server
const defaultPreimagesPath = "~/preimages/preimages.gz"

// WritePreimagesToSenders writes preimages to erigon kv.Senders table
func WritePreimagesToSenders(db kv.RwDB) error {

	log.Info("Reading preimages", "from file", defaultPreimagesPath)

	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(defaultPreimagesPath)
	if err != nil {
		return err
	}
	defer fh.Close()

	var reader io.Reader = fh
	if strings.HasSuffix(defaultPreimagesPath, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return err
		}
	}
	stream := rlp.NewStream(reader, 0)

	// Import the preimages in batches to prevent disk trashing
	//preimages := make(map[common.Hash][]byte)
	buf := newAppendBuffer(bufferOptimalSize)

	for {
		// Read the next entry and ensure it's not junk
		var blob []byte

		if err := stream.Decode(&blob); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		// Accumulate the preimages and flush when enough ws gathered
		key, val := crypto.Keccak256Hash(blob), common.CopyBytes(blob)
		buf.Put(key.Bytes(), val)
	}

	log.Info("Reading preimages is complete")

	// sorting items in buffer to write into erigon kv.Senders efficiently
	buf.Sort()

	return buf.writeIntoTable(db, kv.Senders)
}
