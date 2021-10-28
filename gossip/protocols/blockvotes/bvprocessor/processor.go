package bvprocessor

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/utils/datasemaphore"
	"github.com/Fantom-foundation/lachesis-base/utils/workers"

	"github.com/Fantom-foundation/go-opera/inter"
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

	checker *workers.Workers

	itemsSemaphore *datasemaphore.DataSemaphore
}

type ItemCallback struct {
	Process  func(bvs inter.LlrSignedBlockVotes) error
	Released func(bvs inter.LlrSignedBlockVotes, peer string, err error)
	Check    func(bvs inter.LlrSignedBlockVotes, checked func(error))
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
	released := callback.Item.Released
	callback.Item.Released = func(bvs inter.LlrSignedBlockVotes, peer string, err error) {
		f.itemsSemaphore.Release(dag.Metric{Num: 1, Size: uint64(bvs.Size())})
		if released != nil {
			released(bvs, peer, err)
		}
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
	bvs inter.LlrSignedBlockVotes
	err error
	pos idx.Event
}

func (f *Processor) Enqueue(peer string, items []inter.LlrSignedBlockVotes, done func()) error {
	totalSize := uint64(0)
	for _, v := range items {
		totalSize += v.Size()
	}
	if !f.itemsSemaphore.Acquire(dag.Metric{Num: idx.Event(len(items)), Size: totalSize}, f.cfg.SemaphoreTimeout) {
		return ErrBusy
	}

	checkedC := make(chan *checkRes, len(items))
	err := f.checker.Enqueue(func() {
		for i, v := range items {
			pos := idx.Event(i)
			bvs := v
			f.callback.Item.Check(bvs, func(err error) {
				checkedC <- &checkRes{
					bvs: bvs,
					err: err,
					pos: pos,
				}
			})
		}
	})
	if err != nil {
		return err
	}
	itemsLen := len(items)
	return f.inserter.Enqueue(func() {
		if done != nil {
			defer done()
		}

		var orderedResults = make([]*checkRes, itemsLen)
		var processed int
		for processed < itemsLen {
			select {
			case res := <-checkedC:
				orderedResults[res.pos] = res

				for i := processed; processed < len(orderedResults) && orderedResults[i] != nil; i++ {
					f.process(peer, orderedResults[i].bvs, orderedResults[i].err)
					orderedResults[i] = nil // free the memory
					processed++
				}

			case <-f.quit:
				return
			}
		}
	})
}

func (f *Processor) process(peer string, bvs inter.LlrSignedBlockVotes, resErr error) {
	// release item if failed validation
	if resErr != nil {
		f.callback.Item.Released(bvs, peer, resErr)
		return
	}
	// process item
	err := f.callback.Item.Process(bvs)
	f.callback.Item.Released(bvs, peer, err)
}

func (f *Processor) TasksCount() int {
	return f.inserter.TasksCount() + f.checker.TasksCount()
}
