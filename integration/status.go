package integration

import (
	"os"
	"path"
)

func isInterrupted(chaindataDir string) bool {
	_, err := os.Stat(path.Join(chaindataDir, "unfinished"))
	return err == nil
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
