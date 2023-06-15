package launcher

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore/fileshash"
	"github.com/Fantom-foundation/go-opera/utils/devnullfile"
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

type excludingIterator struct {
	kvdb.Iterator
	exclude kvdb.Reader
}

func (it excludingIterator) Next() bool {
	for it.Iterator.Next() {
		if ok, _ := it.exclude.Has(it.Key()); !ok {
			return true
		}
	}
	return false
}

type unitWriter struct {
	plain            io.WriteSeeker
	gziper           *gzip.Writer
	fileshasher      *fileshash.Writer
	dataStartPos     int64
	uncompressedSize uint64
}

func newUnitWriter(plain io.WriteSeeker) *unitWriter {
	return &unitWriter{
		plain: plain,
	}
}

func (w *unitWriter) Start(header genesis.Header, name, tmpDirPath string) error {
	if w.plain == nil {
		// dry run
		w.fileshasher = fileshash.WrapWriter(nil, genesisstore.FilesHashPieceSize, func(int) fileshash.TmpWriter {
			return devnullfile.DevNull{}
		})
		return nil
	}
	// Write unit marker and version
	_, err := w.plain.Write(append(genesisstore.FileHeader, genesisstore.FileVersion...))
	if err != nil {
		return err
	}

	// write genesis header
	err = rlp.Encode(w.plain, genesisstore.Unit{
		UnitName: name,
		Header:   header,
	})
	if err != nil {
		return err
	}

	w.dataStartPos, err = w.plain.Seek(8+8+32, io.SeekCurrent)
	if err != nil {
		return err
	}

	w.gziper, _ = gzip.NewWriterLevel(w.plain, gzip.BestCompression)

	w.fileshasher = fileshash.WrapWriter(w.gziper, genesisstore.FilesHashPieceSize, func(tmpI int) fileshash.TmpWriter {
		tmpI++
		tmpPath := path.Join(tmpDirPath, fmt.Sprintf("genesis-%s-tmp-%d", name, tmpI))
		_ = os.MkdirAll(tmpDirPath, os.ModePerm)
		tmpFh, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			log.Crit("File opening error", "path", tmpPath, "err", err)
		}
		return dropableFile{
			ReadWriteSeeker: tmpFh,
			Closer:          tmpFh,
			path:            tmpPath,
		}
	})
	return nil
}

func (w *unitWriter) Flush() (hash.Hash, error) {
	if w.plain == nil {
		return w.fileshasher.Root(), nil
	}
	h, err := w.fileshasher.Flush()
	if err != nil {
		return hash.Hash{}, err
	}

	err = w.gziper.Close()
	if err != nil {
		return hash.Hash{}, err
	}

	endPos, err := w.plain.Seek(0, io.SeekCurrent)
	if err != nil {
		return hash.Hash{}, err
	}

	_, err = w.plain.Seek(w.dataStartPos-(8+8+32), io.SeekStart)
	if err != nil {
		return hash.Hash{}, err
	}

	_, err = w.plain.Write(h.Bytes())
	if err != nil {
		return hash.Hash{}, err
	}
	_, err = w.plain.Write(bigendian.Uint64ToBytes(uint64(endPos - w.dataStartPos)))
	if err != nil {
		return hash.Hash{}, err
	}
	_, err = w.plain.Write(bigendian.Uint64ToBytes(w.uncompressedSize))
	if err != nil {
		return hash.Hash{}, err
	}

	_, err = w.plain.Seek(0, io.SeekEnd)
	if err != nil {
		return hash.Hash{}, err
	}
	return h, nil
}

func (w *unitWriter) Write(b []byte) (n int, err error) {
	n, err = w.fileshasher.Write(b)
	w.uncompressedSize += uint64(n)
	return
}

func getEpochBlock(epoch idx.Epoch, store *gossip.Store) idx.Block {
	bs, _ := store.GetHistoryBlockEpochState(epoch)
	if bs == nil {
		return 0
	}
	return bs.LastBlock.Idx
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
	if mode != "full" && mode != "ext-mpt" && mode != "mpt" {
		return errors.New("--export.evm.mode must be one of {full, ext-mpt, mpt}")
	}

	var excludeEvmDB kvdb.Store
	if excludeEvmDBPath := ctx.String(EvmExportExclude.Name); len(excludeEvmDBPath) > 0 {
		db, err := pebble.New(excludeEvmDBPath, 1024*opt.MiB, utils.MakeDatabaseHandles()/2, nil, nil)
		if err != nil {
			return err
		}
		excludeEvmDB = db
	}

	sectionsStr := ctx.String(GenesisExportSections.Name)
	sections := map[string]string{}
	for _, str := range strings.Split(sectionsStr, ",") {
		before := len(sections)
		if strings.HasPrefix(str, "brs") {
			sections["brs"] = str
		} else if strings.HasPrefix(str, "ers") {
			sections["ers"] = str
		} else if strings.HasPrefix(str, "evm") {
			sections["evm"] = str
		} else {
			return fmt.Errorf("unknown section '%s': has to start with either 'brs' or 'ers' or 'evm'", str)
		}
		if len(sections) == before {
			return fmt.Errorf("duplicate section: '%s'", str)
		}
	}

	cfg := makeAllConfigs(ctx)
	tmpPath := path.Join(cfg.Node.DataDir, "tmp")
	_ = os.RemoveAll(tmpPath)
	defer os.RemoveAll(tmpPath)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	if gdb.GetHighestLamport() != 0 {
		log.Warn("Attempting genesis export not in a beginning of an epoch. Genesis file output may contain excessive data.")
	}
	defer gdb.Close()

	fn := ctx.Args().First()

	// Open the file handle
	var plain io.WriteSeeker
	if fn != "dry-run" {
		fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		defer fh.Close()
		plain = fh
	}

	header := genesis.Header{
		GenesisID:   *gdb.GetGenesisID(),
		NetworkID:   gdb.GetEpochState().Rules.NetworkID,
		NetworkName: gdb.GetEpochState().Rules.Name,
	}
	var epochsHash hash.Hash
	var blocksHash hash.Hash
	var evmHash hash.Hash

	if from < 1 {
		// avoid underflow
		from = 1
	}
	if to > gdb.GetEpoch() {
		to = gdb.GetEpoch()
	}
	if len(sections["ers"]) > 0 {
		log.Info("Exporting epochs", "from", from, "to", to)
		writer := newUnitWriter(plain)
		err := writer.Start(header, sections["ers"], tmpPath)
		if err != nil {
			return err
		}
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
		}
		epochsHash, err = writer.Flush()
		if err != nil {
			return err
		}
		log.Info("Exported epochs")
		fmt.Printf("- Epochs hash: %v \n", epochsHash.String())
	}

	if len(sections["brs"]) > 0 {
		toBlock := getEpochBlock(to, gdb)
		fromBlock := getEpochBlock(from, gdb)
		if sections["brs"] != "brs" {
			// to continue prev section, include blocks of prev epochs too, excluding first blocks of prev epoch (which is last block if prev section)
			fromBlock = getEpochBlock(from-1, gdb) + 1
		}
		if fromBlock < 1 {
			// avoid underflow
			fromBlock = 1
		}
		log.Info("Exporting blocks", "from", fromBlock, "to", toBlock)
		writer := newUnitWriter(plain)
		err := writer.Start(header, sections["brs"], tmpPath)
		if err != nil {
			return err
		}
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
		blocksHash, err = writer.Flush()
		if err != nil {
			return err
		}
		log.Info("Exported blocks")
		fmt.Printf("- Blocks hash: %v \n", blocksHash.String())
	}

	if len(sections["evm"]) > 0 {
		log.Info("Exporting EVM data")
		writer := newUnitWriter(plain)
		err := writer.Start(header, sections["evm"], tmpPath)
		if err != nil {
			return err
		}
		it := gdb.EvmStore().EvmDb.NewIterator(nil, nil)
		if mode == "mpt" {
			// iterate only over MPT data
			it = mptIterator{it}
		} else if mode == "ext-mpt" {
			// iterate only over MPT data and preimages
			it = mptAndPreimageIterator{it}
		}
		if excludeEvmDB != nil {
			it = excludingIterator{it, excludeEvmDB}
		}
		defer it.Release()
		err = iodb.Write(writer, it)
		if err != nil {
			return err
		}
		evmHash, err = writer.Flush()
		if err != nil {
			return err
		}
		log.Info("Exported EVM data")
		fmt.Printf("- EVM hash: %v \n", evmHash.String())
	}

	return nil
}
