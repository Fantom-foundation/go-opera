package emitter

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/ethereum/go-ethereum/log"
)

func openEventFile(path string, isSyncMode bool) *os.File {
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		log.Crit("Failed to create open event file", "file", path, "err", err)
	}
	sync := 0
	if isSyncMode {
		sync = os.O_SYNC
	}
	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|sync, 0666)
	if err != nil {
		log.Crit("Failed to open event file", "file", path, "err", err)
	}
	return fh
}

func (em *Emitter) writeLastEmittedEventID(id hash.Event) {
	if em.emittedEventFile == nil {
		return
	}
	_, err := em.emittedEventFile.WriteAt(id.Bytes(), 0)
	if err != nil {
		log.Crit("Failed to write event file", "file", em.config.PrevEmittedEventFile.Path, "err", err)
	}
}

func (em *Emitter) readLastEmittedEventID() *hash.Event {
	if em.emittedEventFile == nil {
		return nil
	}
	buf := make([]byte, 32)
	_, err := em.emittedEventFile.ReadAt(buf, 0)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Crit("Failed to read event file", "file", em.config.PrevEmittedEventFile.Path, "err", err)
	}
	id := hash.BytesToEvent(buf)
	return &id
}
