package epprocessor

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iep"
	"github.com/Fantom-foundation/go-opera/inter/ier"
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
	checker  *workers.Workers

	itemsSemaphore *datasemaphore.DataSemaphore
}

type ItemCallback struct {
	ProcessEV  func(ev inter.LlrSignedEpochVote) error
	ProcessER  func(er ier.LlrIdxFullEpochRecord) error
	ReleasedEV func(ev inter.LlrSignedEpochVote, peer string, err error)
	ReleasedER func(er ier.LlrIdxFullEpochRecord, peer string, err error)
	CheckEV    func(ev inter.LlrSignedEpochVote, checked func(error))
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
		callback:       callback,
	}
	f.callback = callback
	f.inserter = workers.New(&f.wg, f.quit, cfg.MaxTasks)
	f.checker = workers.New(&f.wg, f.quit, cfg.MaxTasks)
	return f
}

// Start boots up the items processor.
func (f *Processor) Start() {
	f.inserter.Start(1)
	f.checker.Start(1)
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
	return f.TasksCount() > f.cfg.MaxTasks*3/4
}

type checkRes struct {
	ev  inter.LlrSignedEpochVote
	err error
	pos idx.Event
}

func (f *Processor) Enqueue(peer string, eps []iep.LlrEpochPack, totalSize uint64, done func()) error {
	if len(eps) == 0 {
		if done != nil {
			done()
		}
		return nil
	}
	metric := dag.Metric{Num: idx.Event(len(eps)), Size: totalSize}
	if !f.itemsSemaphore.Acquire(metric, f.cfg.SemaphoreTimeout) {
		return ErrBusy
	}

	// process each EpochPack as a separate task
	return f.inserter.Enqueue(func() {
		if done != nil {
			defer done()
		}
		defer f.itemsSemaphore.Release(metric)
		for _, ep := range eps {
			items := ep.Votes
			record := ep.Record
			checkedC := make(chan *checkRes, len(items))
			err := f.checker.Enqueue(func() {
				for i, v := range items {
					pos := idx.Event(i)
					ev := v
					f.callback.Item.CheckEV(ev, func(err error) {
						checkedC <- &checkRes{
							ev:  ev,
							err: err,
							pos: pos,
						}
					})
				}
			})
			if err != nil {
				return
			}
			itemsLen := len(items)
			var orderedResults = make([]*checkRes, itemsLen)
			var processed int
			for processed < itemsLen {
				select {
				case res := <-checkedC:
					orderedResults[res.pos] = res

					for i := processed; processed < len(orderedResults) && orderedResults[i] != nil; i++ {
						f.processEV(peer, orderedResults[i].ev, orderedResults[i].err)
						orderedResults[i] = nil // free the memory
						processed++
					}

				case <-f.quit:
					return
				}
			}
			f.processER(peer, record)
		}
	})
}

func (f *Processor) processEV(peer string, ev inter.LlrSignedEpochVote, resErr error) {
	// release item if failed validation
	if resErr != nil {
		f.callback.Item.ReleasedEV(ev, peer, resErr)
		return
	}
	// process item
	err := f.callback.Item.ProcessEV(ev)
	f.callback.Item.ReleasedEV(ev, peer, err)
}

func (f *Processor) processER(peer string, er ier.LlrIdxFullEpochRecord) {
	// process item
	err := f.callback.Item.ProcessER(er)
	f.callback.Item.ReleasedER(er, peer, err)
}

func (f *Processor) TasksCount() int {
	return f.inserter.TasksCount() + f.checker.TasksCount()
}
