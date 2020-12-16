package genesisstore

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/status-im/keycard-go/hexutils"
)

func (s *Store) Export(writer io.Writer) error {
	it := s.db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		_, err := writer.Write(bigendian.Uint32ToBytes(uint32(len(it.Key()))))
		if err != nil {
			return err
		}
		_, err = writer.Write(it.Key())
		if err != nil {
			return err
		}
		_, err = writer.Write(bigendian.Uint32ToBytes(uint32(len(it.Value()))))
		if err != nil {
			return err
		}
		_, err = writer.Write(it.Value())
		if err != nil {
			return err
		}
	}
	return nil
}

func readAll(reader io.Reader, buf []byte) error {
	consumed := 0
	for {
		n, err := reader.Read(buf[consumed:])
		consumed += n
		if consumed == len(buf) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (s *Store) Import(reader io.Reader) error {
	batch := s.db.NewBatch()
	defer batch.Reset()
	var lenB [4]byte
	for {
		err := readAll(reader, lenB[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		lenKey := bigendian.BytesToUint32(lenB[:])
		key := make([]byte, lenKey)
		err = readAll(reader, key)
		if err != nil {
			return err
		}

		err = readAll(reader, lenB[:])
		if err != nil {
			return err
		}

		lenValue := bigendian.BytesToUint32(lenB[:])
		value := make([]byte, lenValue)
		err = readAll(reader, value)
		if err != nil {
			return err
		}

		err = batch.Put(key, value)
		if err != nil {
			return err
		}
		if batch.ValueSize() > kvdb.IdealBatchSize {
			err = batch.Write()
			if err != nil {
				return err
			}
			batch.Reset()
		}
	}
	return batch.Write()
}

var (
	fileHeader  = hexutils.HexToBytes("641b00ac")
	fileVersion = hexutils.HexToBytes("00010001")
)

func checkFileHeader(reader io.Reader) error {
	headerAndVersion := make([]byte, len(fileHeader)+len(fileVersion))
	n, err := reader.Read(headerAndVersion)
	if err != nil {
		return err
	}
	if n != len(headerAndVersion) {
		return errors.New("expected a genesis file, the given file is too short")
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
	err = readAll(rawReader, h[:])
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
