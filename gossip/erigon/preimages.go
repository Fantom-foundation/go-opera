package erigon

import (
	"compress/gzip"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/crypto"
)

// THis path denotes path to preimages in profiling6 server
const defaultPreimagesPath = "/root/preimages/preimages.gz"

// WritePreimagesToSenders writes preimages to erigon kv.Senders table
func WritePreimagesToSenders(db kv.RwDB) error {

	start := time.Now()
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

	buf := newAppendBuffer(bufferOptimalSize)

	log.Info("Writing preimages to buf")
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

	log.Info("Reading preimages and writing preimages to buf completed",  "elapsed", common.PrettyDuration(time.Since(start)))

	// Sort data in buffer
	log.Info("Buffer length", "is", buf.Len())
	log.Info("Sorting data in buffer started...")
	start = time.Now()
	// there are no duplicates keys in buf
	buf.Sort()
	log.Info("Sorting data in buf completed", "elapsed", common.PrettyDuration(time.Since(start)))
	// sorting items in buffer to write into erigon kv.Senders efficiently
	return buf.writeIntoTable(db, kv.Senders)
}
