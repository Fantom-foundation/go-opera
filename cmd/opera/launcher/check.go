package launcher

import (
	"bytes"
	"path"
	"time"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/simplewlru"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter"
)

var (
	emptyCodeHash = common.BytesToHash(crypto.Keccak256(nil))
	emptyHash     = common.Hash{}
)

func checkEvm(ctx *cli.Context) error {
	if len(ctx.Args()) != 0 {
		utils.Fatalf("This command doesn't require an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	defer gdb.Close()
	evms := gdb.EvmStore()
	evmd := evms.EvmDatabase()
	evmt := evms.EvmKvdbTable()

	start, reported := time.Now(), time.Now()

	checkedCache, _ := simplewlru.New(1000000, 1000000)
	cached := func(h common.Hash) bool {
		_, ok := checkedCache.Get(h)
		return ok
	}

	log.Info("Checking every node hash")
	nodeIt := evmt.NewIterator(nil, nil)
	defer nodeIt.Release()
	for nodeIt.Next() {
		if len(nodeIt.Key()) != 32 {
			continue
		}
		calcHash := crypto.Keccak256(nodeIt.Value())
		if !bytes.Equal(nodeIt.Key(), calcHash) {
			log.Crit("Malformed node record", "exp", common.Bytes2Hex(calcHash), "got", common.Bytes2Hex(nodeIt.Key()))
		}
	}

	log.Info("Checking every code hash")
	codeIt := table.New(evmt, []byte("c")).NewIterator(nil, nil)
	defer codeIt.Release()
	for codeIt.Next() {
		if len(codeIt.Key()) != 32 {
			continue
		}
		calcHash := crypto.Keccak256(codeIt.Value())
		if !bytes.Equal(codeIt.Key(), calcHash) {
			log.Crit("Malformed code record", "exp", common.Bytes2Hex(calcHash), "got", common.Bytes2Hex(codeIt.Key()))
		}
	}

	log.Info("Checking every preimage")
	preimageIt := table.New(evmt, []byte("secure-key-")).NewIterator(nil, nil)
	defer preimageIt.Release()
	for preimageIt.Next() {
		if len(preimageIt.Key()) != 32 {
			continue
		}
		calcHash := crypto.Keccak256(preimageIt.Value())
		if !bytes.Equal(preimageIt.Key(), calcHash) {
			log.Crit("Malformed preimage record", "exp", common.Bytes2Hex(calcHash), "got", common.Bytes2Hex(preimageIt.Key()))
		}
	}

	log.Info("Checking presence of every node")
	prevPoint := idx.Block(0)
	prevPointRootExist := false
	lastIdx := gdb.GetLatestBlockIndex()
	visitedHashes := make([]common.Hash, 0, 100000)
	gdb.ForEachBlock(func(index idx.Block, block *inter.Block) {
		stateTrie, err := evmd.OpenTrie(common.Hash(block.Root))
		ok := stateTrie != nil && err == nil
		if ok != prevPointRootExist {
			if index > 0 && ok {
				log.Warn("EVM history is pruned", "fromBlock", prevPoint, "toBlock", index-1)
			}
			prevPointRootExist = ok
			prevPoint = index
		}
		if index == lastIdx && !ok {
			log.Crit("State trie for the latest block is not found", "block", index)
		}
		if !ok {
			return
		}

		// check existence of every code hash and root of every storage trie
		stateIt := stateTrie.NodeIterator(nil)
		for stateItSkip := false; stateIt.Next(!stateItSkip); {
			stateItSkip = false
			if stateIt.Hash() != emptyHash {
				if cached(stateIt.Hash()) {
					stateItSkip = true
					continue
				}
				visitedHashes = append(visitedHashes, stateIt.Hash())
			}
			if stateIt.Leaf() {
				addrHash := common.BytesToHash(stateIt.LeafKey())

				var account state.Account
				if err := rlp.Decode(bytes.NewReader(stateIt.LeafBlob()), &account); err != nil {
					log.Crit("Failed to decode account", "err", err, "block", index)
				}

				codeHash := common.BytesToHash(account.CodeHash)
				if codeHash != emptyCodeHash && !cached(codeHash) {
					code, _ := evmd.ContractCode(addrHash, codeHash)
					if code == nil {
						log.Crit("failed to get code", "addrHash", addrHash.String(), "codeHash", codeHash.String(), "block", index)
					}
					checkedCache.Add(codeHash, true, 1)
				}

				if account.Root != emptyRoot && !cached(account.Root) {
					storageTrie, err := evmd.OpenStorageTrie(addrHash, account.Root)
					if err != nil {
						log.Crit("failed to open storage trie", "err", err, "addrHash", addrHash.String(), "storageRoot", account.Root.String(), "block", index)
					}
					storageIt := storageTrie.NodeIterator(nil)
					for storageItSkip := false; storageIt.Next(!storageItSkip); {
						storageItSkip = false
						if storageIt.Hash() != emptyHash {
							if cached(storageIt.Hash()) {
								storageItSkip = true
								continue
							}
							visitedHashes = append(visitedHashes, storageIt.Hash())
						}
					}
					if storageIt.Error() != nil {
						log.Crit("EVM storage trie iteration error", "addrHash", addrHash.String(), "storageRoot", account.Root.String(), "block", index, "err", storageIt.Error())
					}
				}
			}
		}
		if stateIt.Error() != nil {
			log.Crit("EVM state trie iteration error", "root", block.Root.String(), "block", index, "err", stateIt.Error())
		}
		for _, h := range visitedHashes {
			checkedCache.Add(h, true, 1)
		}
		visitedHashes = visitedHashes[:0]
		if time.Since(reported) >= statsReportLimit {
			log.Info("Checking presence of every node", "last", index, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
	})
	log.Info("EVM storage is verified", "elapsed", common.PrettyDuration(time.Since(start)))

	return nil
}
