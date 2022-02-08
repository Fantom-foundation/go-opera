package brprocessor

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"

	"github.com/Fantom-foundation/go-opera/inter/ibr"
)

var (
	ErrBusy = errors.New("failed to acquire events semaphore")
)

// Processor is responsible for processing incoming events
type Processor struct {
	cfg Config

	quit chan struct{}
	wg   sync.WaitGroup

	callback Callback

	inserter *workers.Workers

	itemsSemaphore *datasemaphore.DataSemaphore
}

type ItemCallback struct {
	Process  func(br ibr.LlrIdxFullBlockRecord) error
	Released func(br ibr.LlrIdxFullBlockRecord, peer string, err error)
}

type Callback struct {
	Item ItemCallback
}

// New creates an event processor
func New(itemsSemaphore *datasemaphore.DataSemaphore, cfg Config, callback Callback) *Processor {
	f := &Processor{
		cfg:            cfg,
		quit:           make(chan struct{}),
		itemsSemaphore: itemsSemaphore,
	}
	f.callback = callback
	f.inserter = workers.New(&f.wg, f.quit, cfg.MaxTasks)
	return f
}

// Start boots up the items processor.
func (f *Processor) Start() {
	f.inserter.Start(1)
}

// Stop interrupts the processor, canceling all the pending operations.
// Stop waits until all the internal goroutines have finished.
func (f *Processor) Stop() {
	close(f.quit)
	f.itemsSemaphore.Terminate()
	f.wg.Wait()
}

// Overloaded returns true if too much items are being processed
func (f *Processor) Overloaded() bool {
	return f.inserter.TasksCount() > f.cfg.MaxTasks*3/4
}

func (f *Processor) Enqueue(peer string, items []ibr.LlrIdxFullBlockRecord, totalSize uint64, done func()) error {
	metric := dag.Metric{Num: idx.Event(len(items)), Size: totalSize}
	if !f.itemsSemaphore.Acquire(metric, f.cfg.SemaphoreTimeout) {
		return ErrBusy
	}

	return f.inserter.Enqueue(func() {
		if done != nil {
			defer done()
		}
		defer f.itemsSemaphore.Release(metric)
		for i, item := range items {
			// process item
			err := f.callback.Item.Process(item)
			items[i].Txs = nil
			items[i].Receipts = nil
			f.callback.Item.Released(item, peer, err)
		}
	})
}

func (f *Processor) TasksCount() int {
	return f.inserter.TasksCount()
}
