package launcher

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip/erigon"
	"github.com/Fantom-foundation/go-opera/logger"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ledgerwatch/erigon/common"

	"github.com/ledgerwatch/erigon-lib/kv"
)

func writePreimagesCmd(_ *cli.Context) error {
	start := time.Now()

	db := erigon.MakeChainDatabase(logger.New("mdbx"))
	defer db.Close()

	if err := erigon.WritePreimagesToSenders(db); err != nil {
		return fmt.Errorf("Import error: %q", err)
	}

	log.Info("Writing preimages is complete", "elapsed", common.PrettyDuration(time.Since(start)))

	return nil
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
