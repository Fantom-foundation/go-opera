package errlock

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/cmd/utils"
)

// Check if errlock is written
func Check() {
	locked, reason, eLockPath, _ := read(datadir)
	if locked {
		utils.Fatalf("Node isn't allowed to start due to a previous error. Please fix the issue and then delete file \"%s\". Error message:\n%s", eLockPath, reason)
	}
}

var (
	datadir string
)

// SetDefaultDatadir for errlock files
func SetDefaultDatadir(dir string) {
	datadir = dir
}

// Permanent error
func Permanent(err error) {
	eLockPath, _ := write(datadir, err.Error())
	utils.Fatalf("Node is permanently stopping due to an issue. Please fix the issue and then delete file \"%s\". Error message:\n%s", eLockPath, err.Error())
}

// read errlock file
func read(dir string) (bool, string, string, error) {
	eLockPath := path.Join(dir, "errlock")

	data, err := os.Open(eLockPath)
	if err != nil {
		return false, "", eLockPath, err
	}
	defer data.Close()

	// read no more than N bytes
	maxFileLen := 5000
	eLockBytes := make([]byte, maxFileLen)
	n, err := data.Read(eLockBytes)
	if err != nil {
		return true, "", eLockPath, err
	}
	return true, string(eLockBytes[:n]), eLockPath, nil
}

// write errlock file
func write(dir string, eLockStr string) (string, error) {
	eLockPath := path.Join(dir, "errlock")

	return eLockPath, ioutil.WriteFile(eLockPath, []byte(eLockStr), 0666) // assume no custom encoding needed
}
