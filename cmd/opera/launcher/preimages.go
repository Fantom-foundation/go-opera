package launcher

import (
	"fmt"
	"os"
	"strings"
	"io"
	"compress/gzip"

	"gopkg.in/urfave/cli.v1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/log"
)



func importPreimagesCmd(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("This command requires an argument.")
	}

	if err := importPreimages(ctx.Args().First()); err != nil {
		return fmt.Errorf("Import error: %q", err)
	}

	return nil
}

func importPreimages(fn string) error {
	log.Info("Importing preimages", "from file file", fn)

	// Open the file handle and potentially unwrap the gzip stream
	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close()

	var reader io.Reader = fh
	if strings.HasSuffix(fn, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return err
		}
	}
	stream := rlp.NewStream(reader, 0)

	// Import the preimages in batches to prevent disk trashing
	preimages := make(map[common.Hash]common.Address)

	i := 0
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
		key, val := crypto.Keccak256Hash(blob), common.BytesToAddress(common.CopyBytes(blob))
		preimages[key] = val
		i += 1
		fmt.Printf("importPreimages %d, address Hash: %s, address: %s\n", i, key.String(), val.Hex())
		if i > 10 {
			break
		}
		/*
		if len(preimages) > 1024 {
			rawdb.WritePreimages(db, preimages)
			preimages = make(map[common.Hash][]byte)
		}
		*/
	}
	// Flush the last batch preimage data
	if len(preimages) == 0 {
		return fmt.Errorf("preimages map is empty")
	}
	return nil
}