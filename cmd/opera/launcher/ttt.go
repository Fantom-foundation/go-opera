package launcher

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime/pprof"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip/emitter"
)

var (
	tttCommand = cli.Command{
		Name:        "ttt",
		Usage:       "debugging",
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `(temporary code)`,
		Action:      utils.MigrateFlags(ttt),
		Flags: []cli.Flag{
			DataDirFlag,
		},
	}
)

func ttt(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		utils.Fatalf("This command requires an argument (AND, OR).")
	}
	// avoid P2P interaction, API calls and events emitting
	genesis := getOperaGenesis(ctx)
	cfg := makeAllConfigs(ctx)
	cfg.Opera.Protocol.EventsSemaphoreLimit.Size = math.MaxUint32
	cfg.Opera.Protocol.EventsSemaphoreLimit.Num = math.MaxUint32
	cfg.Opera.Emitter.Validator = emitter.ValidatorConfig{}
	cfg.Opera.TxPool.Journal = ""
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

	node, svc, nodeClose := makeNode(ctx, cfg, genesis)
	defer nodeClose()
	startNode(ctx, node)

	var (
		fn string
		f  *os.File

		pattern [][]common.Hash
		ll      []*types.Log
		err     error
	)

	switch x := ctx.Args().First(); x {
	case "AND":
		fn = "AND.pprof"
		pattern = [][]common.Hash{
			{},
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
			{common.HexToHash("0x000000000000000000000000e66693d9eb8562c8d7e294f792f45f767490f375")},
		}
	case "AND1":
		fn = "AND1.pprof"
		pattern = [][]common.Hash{
			{},
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		}
	case "OR":
		fn = "OR.pprof"
		pattern = [][]common.Hash{
			{},
			{common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
				common.HexToHash("0x000000000000000000000000e66693d9eb8562c8d7e294f792f45f767490f375")},
		}
	case "OR1":
		fn = "OR1.pprof"
		pattern = [][]common.Hash{
			{},
			{common.HexToHash("0x000000000000000000000000e66693d9eb8562c8d7e294f792f45f767490f375"),
				common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
		}
	default:
		panic("Unknown arg:" + x)
	}

	// AND
	f, err = os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	ll, err = svc.EthAPI.EvmLogIndex().FindInBlocks(context.Background(),
		0x4D50AC, 0x4D50B6, pattern)
	pprof.StopCPUProfile()

	fmt.Printf("Found %d logs.\n", len(ll))
	return err

}
