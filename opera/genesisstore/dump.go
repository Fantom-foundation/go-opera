package genesisstore

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/Fantom-foundation/go-opera/utils/iodb"
	"github.com/Fantom-foundation/go-opera/utils/ioread"
)

var (
	fileHeader  = hexutils.HexToBytes("641b00ac")
	fileVersion = hexutils.HexToBytes("00010001")
)

func (s *Store) Export(writer io.Writer) error {
	return iodb.Write(writer, s.db)
}

func (s *Store) Import(reader io.Reader) error {
	return iodb.Read(reader, s.db.NewBatch())
}

func checkFileHeader(reader io.Reader) error {
	headerAndVersion := make([]byte, len(fileHeader)+len(fileVersion))
	err := ioread.ReadAll(reader, headerAndVersion)
	if err != nil {
		return err
	}
	if bytes.Compare(headerAndVersion[:len(fileHeader)], fileHeader) != 0 {
		return errors.New("expected a genesis file, mismatched file header")
	}
	if bytes.Compare(headerAndVersion[len(fileHeader):], fileVersion) != 0 {
		got := hexutils.BytesToHex(headerAndVersion[len(fileHeader):])
		expected := hexutils.BytesToHex(fileVersion)
		return errors.New(fmt.Sprintf("wrong version of genesis file, got=%s, expected=%s", got, expected))
	}
	return nil
}

func OpenGenesisStore(rawReader io.Reader) (h hash.Hash, readGenesisStore func(*Store) error, err error) {
	err = checkFileHeader(rawReader)
	if err != nil {
		return hash.Zero, nil, err
	}
	err = ioread.ReadAll(rawReader, h[:])
	if err != nil {
		return hash.Zero, nil, err
	}
	readGenesisStore = func(genesisStore *Store) error {
		gzipReader, err := gzip.NewReader(rawReader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		err = genesisStore.Import(gzipReader)
		if err != nil {
			return err
		}
		return nil
	}
	return h, readGenesisStore, nil
}

func WriteGenesisStore(rawWriter io.Writer, genesisStore *Store) error {
	_, err := rawWriter.Write(append(fileHeader, fileVersion...))
	if err != nil {
		return err
	}
	h := genesisStore.Hash()
	_, err = rawWriter.Write(h[:])
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(rawWriter)
	defer gzipWriter.Close()
	err = genesisStore.Export(gzipWriter)
	if err != nil {
		return err
	}
	return nil
}
