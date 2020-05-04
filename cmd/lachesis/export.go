package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func exportChain(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	_, _, _, _, gdb := makeFNode(ctx, false)
	defer gdb.Close()

	start := time.Now()

	var err error
	fp := ctx.Args().First()

	err = ExportChain(gdb, fp)
	if err != nil {
		utils.Fatalf("Export error: %v\n", err)
	}
	fmt.Printf("Export done in %v\n", time.Since(start))
	return nil
}

// ExportChain exports a events into the specified file, truncating any data
// already present in the file.
func ExportChain(gdb *gossip.Store, fn string) error {
	log.Info("Exporting events", "file", fn)

	// Open the file handle and potentially wrap with a gzip stream
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	var writer io.Writer = fh
	if strings.HasSuffix(fn, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}
	// Iterate over the blocks and export them
	if err := Export(gdb, writer); err != nil {
		return err
	}
	log.Info("Exported events", "file", fn)

	return nil
}

// Export writes the active chain to the given writer.
func Export(gdb *gossip.Store, w io.Writer) error {
	log.Info("Exporting batch of events")
	var err error
	start, reported := time.Now(), time.Now()

	var (
		events      inter.Events
		sealedEpoch idx.Epoch
		prevEvent   *inter.Event
	)
	gdb.ForEachEventWithoutEpoch(func(event *inter.Event) bool {
		if event == nil {
			err = errors.New("export failed, event not found")
			return false
		}
		events = append(events, event)
		if prevEvent == nil {
			prevEvent = event
		}
		if len(event.Parents) == 0 && prevEvent.Epoch != event.Epoch {
			sealedEpoch = prevEvent.Epoch
		}
		prevEvent = event
		return true
	})

	for _, event := range events {
		if event.Epoch > sealedEpoch {
			break
		}
		log.Debug("exported", "event", event.String())
		err := event.EncodeRLP(w)
		if err != nil {
			err = fmt.Errorf("export failed, error: %v", err)
			return err
		}
		if time.Since(reported) >= statsReportLimit {
			log.Info("Exporting events", "exported", event.String(), "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
	}

	return nil
}
