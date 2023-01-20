package compactdb

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/keycard-go/hexutils"
)

type loggedCompacter struct {
	kvdb.Store
	name string

	currentOp atomic.Value

	wg   sync.WaitGroup
	quit chan struct{}
}

type pair struct {
	p, l []byte
}

func (s *loggedCompacter) Compact(prev []byte, limit []byte) error {
	s.currentOp.Store(pair{prev, limit})
	if err := s.Store.Compact(prev, limit); err != nil {
		log.Error("Compaction error", "name", s.name, "err", err)
		return err
	}
	return nil
}

func (s *loggedCompacter) StartLogging() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ticker.C:
				opI := s.currentOp.Load()
				if opI != nil {
					op := opI.(pair)
					// trim keys for nicer human-readable logging
					op.p, op.l = trimAfterDiff(op.p, op.l, 2)
					untilStr := hexutils.BytesToHex(op.l)
					if len(op.l) == 0 {
						untilStr = "end"
					}
					log.Info("Compacting DB", "name", s.name, "until", untilStr)
				}
			case <-s.quit:
				return
			}
		}
	}()
}

func (s *loggedCompacter) StopLogging() {
	close(s.quit)
	s.wg.Wait()
}
