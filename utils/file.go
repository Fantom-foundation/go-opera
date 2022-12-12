package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
)

func OpenFile(path string, isSyncMode bool) *os.File {
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		log.Crit("Failed to create file dir", "file", path, "err", err)
	}
	sync := 0
	if isSyncMode {
		sync = os.O_SYNC
	}
	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|sync, 0666)
	if err != nil {
		log.Crit("Failed to open file", "file", path, "err", err)
	}
	return fh
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FilePut(path string, content []byte, isSyncMode bool) {
	fh := OpenFile(path, isSyncMode)
	defer fh.Close()
	if err := fh.Truncate(0); err != nil {
		log.Crit("Failed to truncate file", "file", path, "err", err)
	}
	if _, err := fh.Write(content); err != nil {
		log.Crit("Failed to write to file", "file", path, "err", err)
	}
}

func FileGet(path string) []byte {
	if !FileExists(path) {
		return nil
	}
	fh, err := os.Open(path)
	if err != nil {
		log.Crit("Failed to open file", "file", path, "err", err)
	}
	defer fh.Close()
	res, _ := ioutil.ReadAll(fh)
	return res
}
