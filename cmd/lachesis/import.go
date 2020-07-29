package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/keycard-go/hexutils"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-lachesis/eventcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
)

func importChain(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	// avoid P2P interaction, API calls and events emitting
	cfg := makeAllConfigs(ctx)
	cfg.Lachesis.Emitter.Validator = common.Address{}
	cfg.Lachesis.TxPool.Journal = ""
	cfg.Node.IPCPath = ""
	cfg.Node.HTTPHost = ""
	cfg.Node.WSHost = ""
	cfg.Node.NoUSB = true
	cfg.Node.P2P.ListenAddr = ""
	cfg.Node.P2P.NoDiscovery = true
	cfg.Node.P2P.BootstrapNodes = nil
	cfg.Node.P2P.DiscoveryV5 = false
	cfg.Node.P2P.BootstrapNodesV5 = nil
	cfg.Node.P2P.StaticNodes = nil
	cfg.Node.P2P.TrustedNodes = nil

	err := importToNode(ctx, cfg, ctx.Args()...)
	if err != nil {
		return err
	}

	return nil
}

func importToNode(ctx *cli.Context, cfg *config, args ...string) error {
	node := makeNode(ctx, cfg)
	defer node.Close()
	startNode(ctx, node)

	var srv *gossip.Service
	if err := node.Service(&srv); err != nil {
		return err
	}

	check := true
	for _, arg := range ctx.Args() {
		if arg == "check=false" || arg == "check=0" {
			check = false
		}
	}

	for _, fn := range args {
		if strings.HasPrefix(fn, "check=") {
			continue
		}
		if err := importFile(srv, check, fn); err != nil {
			log.Error("Import error", "file", fn, "err", err)
			return err
		}
	}
	return nil
}

func checkEventsFileHeader(reader io.Reader) error {
	headerAndVersion := make([]byte, len(eventsFileHeader)+len(eventsFileVersion))
	n, err := reader.Read(headerAndVersion)
	if err != nil {
		return err
	}
	if n != len(headerAndVersion) {
		return errors.New("expected an events file, the given file is too short")
	}
	if bytes.Compare(headerAndVersion[:len(eventsFileHeader)], eventsFileHeader) != 0 {
		return errors.New("expected an events file, mismatched file header")
	}
	if bytes.Compare(headerAndVersion[len(eventsFileHeader):], eventsFileVersion) != 0 {
		got := hexutils.BytesToHex(headerAndVersion[len(eventsFileHeader):])
		expected := hexutils.BytesToHex(eventsFileVersion)
		return errors.New(fmt.Sprintf("wrong version of events file, got=%s, expected=%s", got, expected))
	}
	return nil
}

func importFile(srv *gossip.Service, check bool, fn string) error {
	// Watch for Ctrl-C while the import is running.
	// If a signal is received, the import will stop.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupt)

	log.Info("Importing events from file", "file", fn, "check", check)

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

	// Check file version and header
	if err := checkEventsFileHeader(reader); err != nil {
		return err
	}

	stream := rlp.NewStream(reader, 0)

	start := time.Now()
	skipped := 0
	imported := 0
	last := hash.Event{}
	for {
		select {
		case <-interrupt:
			return fmt.Errorf("interrupted")
		default:
		}
		e := new(inter.Event)
		err = stream.Decode(e)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		eStart := time.Now()
		if check {
			err := srv.ValidateEvent(e)
			if err != nil && err != epochcheck.ErrNotRelevant && err != eventcheck.ErrAlreadyConnectedEvent {
				return err
			}
		}
		err := srv.ProcessEvent(e)
		if err == epochcheck.ErrNotRelevant || err == eventcheck.ErrAlreadyConnectedEvent {
			skipped++
		} else if err != nil {
			return err
		} else {
			log.Info("New event imported", "id", e.Hash(), "checked", check, "t", time.Since(eStart), "imported", imported, "skipped", skipped, "elapsed", common.PrettyDuration(time.Since(start)))
			last = e.Hash()
			imported++
		}
	}
	log.Info("Events import is finished", "file", fn, "checked", check, "last", last.String(), "imported", imported, "skipped", skipped, "elapsed", common.PrettyDuration(time.Since(start)))

	return nil
}
