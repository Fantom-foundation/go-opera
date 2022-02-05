package launcher

import (
	"path"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter"
)

func checkEvm(ctx *cli.Context) error {
	if len(ctx.Args()) != 0 {
		utils.Fatalf("This command doesn't require an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cfg.Cachescale)
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()
	evms := gdb.EvmStore()

	start, reported := time.Now(), time.Now()
	defer log.Info("EVM storage is verified", "elapsed", common.PrettyDuration(time.Since(start)))

	checkBlocks := func(stateOK func(root common.Hash) (bool, error)) {
		var (
			lastIdx            = gdb.GetLatestBlockIndex()
			prevPoint          idx.Block
			prevPointRootExist bool
		)
		gdb.ForEachBlock(func(index idx.Block, block *inter.Block) {
			if time.Since(reported) >= statsReportLimit {
				log.Info("Checking presence of every node", "last", index, "elapsed", common.PrettyDuration(time.Since(start)))
				reported = time.Now()
			}

			found, err := stateOK(common.Hash(block.Root))
			if found != prevPointRootExist {
				if index > 0 && found {
					log.Warn("EVM history is pruned", "fromBlock", prevPoint, "toBlock", index-1)
				}
				prevPointRootExist = found
				prevPoint = index
			}
			if index == lastIdx && !found {
				log.Crit("State trie for the latest block is not found", "block", index)
			}
			if !found {
				return
			}
			if err != nil {
				log.Crit("State trie error", "err", err, "block", index)
			}

		})
	}

	return evms.CheckEvm(checkBlocks)
}
