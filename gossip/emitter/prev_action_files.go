package emitter

import (
	"io"
	"os"
	"path/filepath"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/log"
)

func openPrevActionFile(path string, isSyncMode bool) *os.File {
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
	v := hash.BytesToEvent(buf)
	return &v
}

func (em *Emitter) writeLastEmittedBlockVotes(b idx.Block) {
	if em.emittedBvsFile == nil {
		return
	}
	_, err := em.emittedBvsFile.WriteAt(b.Bytes(), 0)
	if err != nil {
		log.Crit("Failed to write BVs file", "file", em.config.PrevBlockVotesFile.Path, "err", err)
	}
}

func (em *Emitter) readLastBlockVotes() *idx.Block {
	if em.emittedBvsFile == nil {
		return nil
	}
	buf := make([]byte, 8)
	_, err := em.emittedBvsFile.ReadAt(buf, 0)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Crit("Failed to read BVs file", "file", em.config.PrevBlockVotesFile.Path, "err", err)
	}
	v := idx.BytesToBlock(buf)
	return &v
}

func (em *Emitter) writeLastEmittedEpochVote(e idx.Epoch) {
	if em.emittedEvFile == nil {
		return
	}
	_, err := em.emittedEvFile.WriteAt(e.Bytes(), 0)
	if err != nil {
		log.Crit("Failed to write BVs file", "file", em.config.PrevEpochVoteFile.Path, "err", err)
	}
}

func (em *Emitter) readLastEpochVote() *idx.Epoch {
	if em.emittedEvFile == nil {
		return nil
	}
	buf := make([]byte, 4)
	_, err := em.emittedEvFile.ReadAt(buf, 0)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		log.Crit("Failed to read EV file", "file", em.config.PrevEpochVoteFile.Path, "err", err)
	}
	v := idx.BytesToEpoch(buf)
	return &v
}
