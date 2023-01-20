package integration

import (
	"os"
	"path"

	"github.com/Fantom-foundation/go-opera/utils"
)

func isInterrupted(chaindataDir string) bool {
	return utils.FileExists(path.Join(chaindataDir, "unfinished"))
}

func setGenesisProcessing(chaindataDir string) {
	f, _ := os.Create(path.Join(chaindataDir, "unfinished"))
	if f != nil {
		_ = f.Close()
	}
}

func setGenesisComplete(chaindataDir string) {
	_ = os.Remove(path.Join(chaindataDir, "unfinished"))
}
