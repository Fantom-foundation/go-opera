package filelog

import (
	"fmt"
	"io"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/utils"
)

type Filelog struct {
	io.ReadCloser
	name     string
	size     uint64
	period   time.Duration
	consumed uint64
	prevLog  time.Time
	start    time.Time
}

func (f *Filelog) Read(p []byte) (n int, err error) {
	n, err = f.ReadCloser.Read(p)
	f.consumed += uint64(n)
	if f.prevLog.IsZero() {
		log.Info(fmt.Sprintf("- Reading %s", f.name))
		f.prevLog = time.Now()
		f.start = time.Now()
	} else if f.consumed > 0 && f.consumed < f.size && time.Since(f.prevLog) >= f.period {
		elapsed := time.Since(f.start)
		eta := float64(f.size-f.consumed) / float64(f.consumed) * float64(elapsed)
		eta *= 1.2 // show slightly higher ETA as performance degrades over larger volumes of data
		progress := fmt.Sprintf("%.2f%%", 100*float64(f.consumed)/float64(f.size))
		log.Info(fmt.Sprintf("- Reading %s", f.name), "progress", progress, "elapsed", utils.PrettyDuration(elapsed), "eta", utils.PrettyDuration(eta))
		f.prevLog = time.Now()
	}
	return
}

func Wrap(r io.ReadCloser, name string, size uint64, period time.Duration) *Filelog {
	return &Filelog{
		ReadCloser: r,
		name:       name,
		size:       size,
		period:     period,
	}
}
