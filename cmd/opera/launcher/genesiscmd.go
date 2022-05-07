package launcher

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strconv"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/integration"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore/fileshash"
	"github.com/Fantom-foundation/go-opera/utils/iodb"
)

type dropableFile struct {
	io.ReadWriteSeeker
	io.Closer
	path string
}

func (f dropableFile) Drop() error {
	return os.Remove(f.path)
}

type mptIterator struct {
	kvdb.Iterator
}

func (it mptIterator) Next() bool {
	for it.Iterator.Next() {
		if evmstore.IsMptKey(it.Key()) {
			return true
		}
	}
	return false
}

type mptAndPreimageIterator struct {
	kvdb.Iterator
}

func (it mptAndPreimageIterator) Next() bool {
	for it.Iterator.Next() {
		if evmstore.IsMptKey(it.Key()) || evmstore.IsPreimageKey(it.Key()) {
			return true
		}
	}
	return false
}

func wrapIntoHashFile(backend *zip.Writer, tmpDirPath, name string) *fileshash.Writer {
	zWriter, err := backend.Create(name)
	if err != nil {
		log.Crit("Zip file creation error", "err", err)
	}
	tmpI := 0
	return fileshash.WrapWriter(zWriter, genesisstore.FilesHashPieceSize, 64*opt.GiB, func() fileshash.TmpWriter {
		tmpI++
		tmpPath := path.Join(tmpDirPath, fmt.Sprintf("genesis-%s-tmp-%d", name, tmpI))
		_ = os.MkdirAll(tmpDirPath, os.ModePerm)
		tmpFh, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			log.Crit("File opening error", "path", tmpPath, "err", err)
		}
		return dropableFile{
			ReadWriteSeeker: tmpFh,
			Closer:          tmpFh,
			path:            tmpPath,
		}
	})
}

func exportGenesis(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	from := idx.Epoch(1)
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 32)
		if err != nil {
			return err
		}
		from = idx.Epoch(n)
	}
	to := idx.Epoch(math.MaxUint32)
	if len(ctx.Args()) > 2 {
		n, err := strconv.ParseUint(ctx.Args().Get(2), 10, 32)
		if err != nil {
			return err
		}
		to = idx.Epoch(n)
	}
	mode := ctx.String(EvmExportMode.Name)
	if mode != "full" && mode != "ext-mpt" && mode != "mpt" && mode != "none" {
		return errors.New("--export.evm.mode must be one of {full, ext-mpt, mpt, none}")
	}

	cfg := makeAllConfigs(ctx)
	tmpPath := path.Join(cfg.Node.DataDir, "tmp")
	_ = os.RemoveAll(tmpPath)
	defer os.RemoveAll(tmpPath)

	rawProducer := integration.DBProducer(path.Join(cfg.Node.DataDir, "chaindata"), cacheScaler(ctx))
	gdb, err := makeRawGossipStore(rawProducer, cfg)
	if err != nil {
		log.Crit("DB opening error", "datadir", cfg.Node.DataDir, "err", err)
	}
	if gdb.GetHighestLamport() != 0 {
		log.Warn("Attempting genesis export not in a beginning of an epoch. Genesis file output may contain excessive data.")
	}
	defer gdb.Close()

	fn := ctx.Args().First()

	// Open the file handle
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	// Write file header and version
	_, err = fh.Write(append(genesisstore.FileHeader, genesisstore.FileVersion...))
	if err != nil {
		return err
	}

	log.Info("Exporting genesis header")
	err = rlp.Encode(fh, genesis.Header{
		GenesisID:   *gdb.GetGenesisID(),
		NetworkID:   gdb.GetEpochState().Rules.NetworkID,
		NetworkName: gdb.GetEpochState().Rules.Name,
	})
	if err != nil {
		return err
	}
	// write dummy genesis hashes
	hashesFilePos, err := fh.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	dummy := genesis.Hashes{
		Blocks:      hash.Hashes{hash.Zero},
		Epochs:      hash.Hashes{hash.Zero},
		RawEvmItems: hash.Hashes{hash.Zero},
	}
	if mode == "none" {
		dummy.RawEvmItems = hash.Hashes{}
	}
	b, _ := rlp.EncodeToBytes(dummy)
	hashesFileLen := len(b)
	_, err = fh.Write(b)
	if err != nil {
		return err
	}
	hashes := genesis.Hashes{}

	// Create the zip archive
	z := zip.NewWriter(fh)
	defer z.Close()

	if from < 1 {
		// avoid underflow
		from = 1
	}
	if to > gdb.GetEpoch() {
		to = gdb.GetEpoch()
	}
	toBlock := idx.Block(0)
	fromBlock := idx.Block(0)
	{
		log.Info("Exporting epochs", "from", from, "to", to)
		writer := wrapIntoHashFile(z, tmpPath, genesisstore.EpochsSection)
		for i := to; i >= from; i-- {
			er := gdb.GetFullEpochRecord(i)
			if er == nil {
				log.Warn("No epoch record", "epoch", i)
				break
			}
			b, _ := rlp.EncodeToBytes(ier.LlrIdxFullEpochRecord{
				LlrFullEpochRecord: *er,
				Idx:                i,
			})
			_, err := writer.Write(b)
			if err != nil {
				return err
			}
			if i == from {
				fromBlock = er.BlockState.LastBlock.Idx
			}
			if i == to {
				toBlock = er.BlockState.LastBlock.Idx
			}
		}
		sectionRoot, err := writer.Flush()
		if err != nil {
			return err
		}
		hashes.Epochs.Add(sectionRoot)
		err = z.Flush()
		if err != nil {
			return err
		}
	}

	if fromBlock < 1 {
		// avoid underflow
		fromBlock = 1
	}
	{
		log.Info("Exporting blocks", "from", fromBlock, "to", toBlock)
		writer := wrapIntoHashFile(z, tmpPath, genesisstore.BlocksSection)
		for i := toBlock; i >= fromBlock; i-- {
			br := gdb.GetFullBlockRecord(i)
			if br == nil {
				log.Warn("No block record", "block", i)
				break
			}
			if i%200000 == 0 {
				log.Info("Exporting blocks", "last", i)
			}
			b, _ := rlp.EncodeToBytes(ibr.LlrIdxFullBlockRecord{
				LlrFullBlockRecord: *br,
				Idx:                i,
			})
			_, err := writer.Write(b)
			if err != nil {
				return err
			}
		}
		sectionRoot, err := writer.Flush()
		if err != nil {
			return err
		}
		hashes.Blocks.Add(sectionRoot)
		err = z.Flush()
		if err != nil {
			return err
		}
	}

	if mode != "none" {
		log.Info("Exporting EVM storage")
		writer := wrapIntoHashFile(z, tmpPath, genesisstore.EvmSection)
		it := gdb.EvmStore().EvmDb.NewIterator(nil, nil)
		if mode == "mpt" {
			// iterate only over MPT data
			it = mptIterator{it}
		} else if mode == "ext-mpt" {
			// iterate only over MPT data and preimages
			it = mptAndPreimageIterator{it}
		}
		defer it.Release()
		err = iodb.Write(writer, it)
		if err != nil {
			return err
		}
		sectionRoot, err := writer.Flush()
		if err != nil {
			return err
		}
		hashes.RawEvmItems.Add(sectionRoot)
		err = z.Flush()
		if err != nil {
			return err
		}
	}

	// write real file hashes after they were calculated
	_, err = fh.Seek(hashesFilePos, io.SeekStart)
	if err != nil {
		return err
	}
	b, _ = rlp.EncodeToBytes(hashes)
	if len(b) != hashesFileLen {
		return fmt.Errorf("real hashes length doesn't match to dummy hashes length: %d!=%d", len(b), hashesFileLen)
	}
	_, err = fh.Write(b)
	if err != nil {
		return err
	}
	_, err = fh.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	fmt.Printf("- Epochs hashes: %v \n", hashes.Epochs)
	fmt.Printf("- Blocks hashes: %v \n", hashes.Blocks)
	fmt.Printf("- EVM hashes: %v \n", hashes.RawEvmItems)

	return nil
}
