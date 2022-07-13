package launcher

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	//"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip/erigon"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/crypto"
	"github.com/ledgerwatch/erigon/common"

)

const defaultPreimagesPath = "/root/preimages/preimages.gz"

func writePreimagesCmd(_ *cli.Context) error {
	/*
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("This command requires an argument.")
	}
	*/
	start := time.Now()

	db := erigon.MakeChainDatabase(logger.New("mdbx"))
	defer db.Close()

	tx, err := db.BeginRw(context.Background())
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := writePreimages(tx); err != nil {
		return fmt.Errorf("Import error: %q", err)
	}

	log.Info("Writing preimages is complete", "elapsed", common.PrettyDuration(time.Since(start)))

	return nil
}

func writePreimages(tx kv.RwTx) error {

	log.Info("Writing preimages", "from file", defaultPreimagesPath)
	
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
	preimages := make(map[common.Hash][]byte)

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
		preimages[key] = val
	
		//fmt.Printf("importPreimages %d, address Hash: %s, address: %s\n", i, key.String(), val.Hex())
		
		if len(preimages) > 1024 {
			if err := erigon.WriteSenders(tx, preimages); err != nil {
				return err
			}
			preimages = make(map[common.Hash][]byte)
		}
		
	}

	if len(preimages) > 0 {
		if err := erigon.WriteSenders(tx, preimages); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func readPreimagesCmd(_ *cli.Context) error {

	db := erigon.MakeChainDatabase(logger.New("mdbx"))
	defer db.Close()

	tx, err := db.BeginRo(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback()


	if err := erigon.ReadErigonTableNoDups(kv.Senders, tx); err != nil {
		return err
	}

	// TODO handle flags
	return nil
}