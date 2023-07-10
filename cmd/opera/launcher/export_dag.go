package launcher

import (
	"io"
	"os"
	"strconv"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/utils/dag"
)

func exportDAGgraph(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	defer gdb.Close()

	fn := ctx.Args().First()

	// Open the file handle and potentially wrap with a gzip stream
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	var writer io.Writer = fh

	from := idx.Epoch(1)
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 32)
		if err != nil {
			return err
		}
		from = idx.Epoch(n)
	}
	to := idx.Epoch(0)
	if len(ctx.Args()) > 2 {
		n, err := strconv.ParseUint(ctx.Args().Get(2), 10, 32)
		if err != nil {
			return err
		}
		to = idx.Epoch(n)
	}

	graph := dag.Graph(gdb, from, to)
	buf, err := dot.Marshal(graph, "DAG", "", "\t")
	if err != nil {
		utils.Fatalf("Export error: %v\n", err)
	}

	log.Info("Exporting events DAG to file", "file", fn)
	_, err = writer.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
