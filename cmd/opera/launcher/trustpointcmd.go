package launcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/trustpoint"
)

var (
	trustpointCommand = cli.Command{
		Name:        "trustpoint",
		Usage:       "A set of commands based on the trustpoint",
		Category:    "MISCELLANEOUS COMMANDS",
		Description: "",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Prune stale EVM state data and save trustpoint into file",
				ArgsUsage: "<filename>",
				Action:    utils.MigrateFlags(trustpointCreate),
				Category:  "MISCELLANEOUS COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.CacheTrieJournalFlag,
					utils.BloomFilterSizeFlag,
				},
				Description: `
opera trustpoint create

Note: command also prunes EVM state data.
<TODO>
`,
			},
			{
				Name:      "apply",
				Usage:     "Initialize datadir from trustpoint file",
				ArgsUsage: "<filename>",
				Action:    utils.MigrateFlags(trustpointApply),
				Category:  "MISCELLANEOUS COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					utils.CacheTrieJournalFlag,
				},
				Description: `
opera trustpoint apply

Note: datadir shoul be empty.
<TODO>
`,
			},
		},
	}
)

func trustpointCreate(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)
	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	if ctx.NArg() > 1 {
		log.Error("Too many arguments given")
		return errors.New("too many arguments")
	}
	if ctx.NArg() < 1 {
		log.Error("File name argument required")
		return errors.New("file name argument required")
	}
	file := ctx.Args()[0]

	bloomFilterSize := ctx.GlobalUint64(utils.BloomFilterSizeFlag.Name)

	dir, err := ioutil.TempDir("", "trustpoint")
	if err != nil {
		log.Error("Create temporary dir", "err", err)
		return err
	}
	defer os.RemoveAll(dir)
	db, err := leveldb.New(dir, 2*opt.MiB, 0, nil, nil)
	if err != nil {
		return err
	}
	store := trustpoint.NewStore(db)

	err = gossipToTrustpoint(gdb, store, bloomFilterSize)
	if err != nil {
		return err
	}

	return trustpointSaveTo(store, file)
}

func gossipToTrustpoint(gdb *gossip.Store, store *trustpoint.Store, bloomFilterSize uint64) error {
	start, reported := time.Now(), time.Time{}

	// find the last block of prev epoch
	bs, es := gdb.GetBlockEpochState()

	var (
		lastBlockIdx = bs.LastBlock.Idx
		lastBlock    = gdb.GetBlock(lastBlockIdx)
	)
	for lastBlock.Root != es.EpochStateRoot {
		lastBlockIdx--
		lastBlock = gdb.GetBlock(lastBlockIdx)
	}

	_, err := gdb.EvmStore().StateDB(lastBlock.Root)
	if err != nil {
		log.Error("State not found, probably pruned before", "err", err, "root", lastBlock.Root)
		return err
	}

	// prune state db
	err = pruneStateTo(gdb, common.Hash(lastBlock.Root), common.Hash{}, bloomFilterSize)
	if err != nil {
		log.Error("Skip state prunning", "err", err)
		// TODO: why the error
		/* panic(fmt.Errorf("prune state err (epoch=%d, lastblock=%d, currblock=%d)",
			es.Epoch, lastBlockIdx, bs.LastBlock.Idx,
		))
		*/
		return nil
	}

	// export rules
	store.SetRules(es.Rules)

	// EVM needs last 256 blocks only, see core/vm.opBlockhash() instruction
	var firstBlockIdx idx.Block
	const history idx.Block = 256
	if lastBlockIdx > history {
		firstBlockIdx = lastBlockIdx - history
	}
	// export of blocks
	for index := firstBlockIdx; index <= lastBlockIdx; index++ {
		block := gdb.GetBlock(index)
		txs := make([]*types.Transaction, len(block.Txs))
		for i, txid := range block.Txs {
			txs[i] = gdb.EvmStore().GetTx(txid)
		}
		receipts := gdb.EvmStore().GetReceipts(index)
		receiptsForStorage := make([]*types.ReceiptForStorage, len(receipts))
		for i, r := range receipts {
			receiptsForStorage[i] = (*types.ReceiptForStorage)(r)
		}
		store.SetBlock(index, genesis.Block{
			Time:        block.Time,
			Atropos:     block.Atropos,
			Txs:         txs,
			InternalTxs: types.Transactions{},
			Root:        block.Root,
			Receipts:    receiptsForStorage,
		})
		if time.Since(reported) >= statsReportLimit {
			log.Info("Exporting blocks", "last", index, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
	}
	log.Info("Exported blocks", "from", firstBlockIdx, "last", lastBlockIdx, "elapsed", common.PrettyDuration(time.Since(start)))

	// export of EVM state
	log.Info("Exporting EVM state", "root", lastBlock.Root.String())
	it := gdb.EvmStore().EvmDb.NewIterator(nil, nil)
	for it.Next() {
		// TODO: skip unnecessary states after root
		store.SetRawEvmItem(it.Key(), it.Value())
	}
	it.Release()

	// export metadata
	// TODO

	return nil
}

func trustpointSaveTo(store *trustpoint.Store, file string) error {
	log.Info("Encoding go-opera trustpoint", "path", file)
	fh, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fh.Close()

	err = trustpoint.WriteStore(store, fh)
	if err != nil {
		return err
	}

	return nil
}

func trustpointApply(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	if len(rawProducer.Names()) > 0 {
		return fmt.Errorf("datadir is not empty")
	}
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()

	if ctx.NArg() > 1 {
		log.Error("Too many arguments given")
		return errors.New("too many arguments")
	}
	if ctx.NArg() < 1 {
		log.Error("File name argument required")
		return errors.New("file name argument required")
	}
	file := ctx.Args()[0]

	dir, err := ioutil.TempDir("", "trustpoint")
	if err != nil {
		log.Error("Create temporary dir", "err", err)
		return err
	}
	defer os.RemoveAll(dir)
	db, err := leveldb.New(dir, 2*opt.MiB, 0, nil, nil)
	if err != nil {
		return err
	}
	store := trustpoint.NewStore(db)

	err = trustpointReadFrom(file, store)
	if err != nil {
		log.Error("Trustpoint file read", "err", err)
		return err
	}

	err = gossipFromTrustpoint(store, gdb)
	if err != nil {
		return err
	}

	return gdb.Commit()
}

func trustpointReadFrom(file string, store *trustpoint.Store) error {
	log.Info("Decoding go-opera trustpoint", "path", file)
	fh, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fh.Close()

	return trustpoint.ReadStore(fh, store)
}

func gossipFromTrustpoint(store *trustpoint.Store, gdb *gossip.Store) error {
	// start, reported := time.Now(), time.Time{}

	return nil
}
