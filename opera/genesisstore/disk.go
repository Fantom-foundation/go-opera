package genesisstore

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore/filelog"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore/fileshash"
	"github.com/Fantom-foundation/go-opera/opera/genesisstore/fileszip"
	"github.com/Fantom-foundation/go-opera/utils/ioread"
)

var (
	FileHeader  = hexutils.HexToBytes("641b00ac")
	FileVersion = hexutils.HexToBytes("00020001")
)

const (
	FilesHashMaxMemUsage = 256 * opt.MiB
	FilesHashPieceSize   = 64 * opt.MiB
)

type dummyByteReader struct {
	io.Reader
}

func (r dummyByteReader) ReadByte() (byte, error) {
	b := make([]byte, 1)
	err := ioread.ReadAll(r.Reader, b)
	return b[0], err
}

func checkFileHeader(reader io.Reader) error {
	headerAndVersion := make([]byte, len(FileHeader)+len(FileVersion))
	err := ioread.ReadAll(reader, headerAndVersion)
	if err != nil {
		return err
	}
	if bytes.Compare(headerAndVersion[:len(FileHeader)], FileHeader) != 0 {
		return errors.New("expected a genesis file, mismatched file header")
	}
	if bytes.Compare(headerAndVersion[len(FileHeader):], FileVersion) != 0 {
		got := hexutils.BytesToHex(headerAndVersion[len(FileHeader):])
		expected := hexutils.BytesToHex(FileVersion)
		return errors.New(fmt.Sprintf("wrong version of genesis file, got=%s, expected=%s", got, expected))
	}
	return nil
}

func OpenGenesisStore(rawReaders []fileszip.Reader, close func() error) (*Store, genesis.Hashes, error) {
	header := genesis.Header{}
	hashes := genesis.Hashes{}
	for i, rawReader := range rawReaders {
		reader := io.NewSectionReader(rawReader.Reader, 0, rawReader.Size)
		err := checkFileHeader(reader)
		if err != nil {
			return nil, hashes, err
		}
		_header := genesis.Header{}
		err = rlp.Decode(dummyByteReader{reader}, &_header)
		if err != nil {
			return nil, hashes, err
		}
		if i == 0 {
			header = _header
		} else {
			if !header.Equal(_header) {
				return nil, hashes, errors.New("subsequent genesis header doesn't match to the first header")
			}
		}
		_hashes := genesis.Hashes{}
		err = rlp.Decode(dummyByteReader{reader}, &_hashes)
		if err != nil {
			return nil, hashes, err
		}
		hashes.Add(_hashes)
		consumed, err := reader.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, hashes, err
		}
		// move reader past consumed data
		rawReaders[i].Reader = io.NewSectionReader(rawReader.Reader, consumed, rawReader.Size-consumed)
	}

	zipMap, err := fileszip.Open(rawReaders)
	if err != nil {
		return nil, hashes, err
	}

	hashesMap := make(map[string]hash.Hash)
	for i, h := range hashes.Blocks {
		hashesMap[getSectionName(BlocksSection, i)] = h
	}
	for i, h := range hashes.Epochs {
		hashesMap[getSectionName(EpochsSection, i)] = h
	}
	for i, h := range hashes.RawEvmItems {
		hashesMap[getSectionName(EvmSection, i)] = h
	}
	hashedMap := fileshash.Wrap(func(name string) (io.ReadCloser, error) {
		// wrap with a logger
		f, size, err := zipMap.Open(name)
		if err != nil {
			return nil, err
		}
		// human-readable name
		if name == BlocksSection {
			name = "blocks"
		}
		if name == EpochsSection {
			name = "epochs"
		}
		if name == EvmSection {
			name = "EVM data"
		}
		return filelog.Wrap(f, name, size, time.Minute), nil
	}, FilesHashMaxMemUsage, hashesMap)

	return NewStore(hashedMap, header, close), hashes, nil
}
