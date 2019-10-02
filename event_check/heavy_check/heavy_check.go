package heavy_check

import (
	"errors"
	"runtime"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

var (
	ErrWrongEventSig  = errors.New("event has wrong signature")
	ErrMalformedTxSig = errors.New("tx has wrong signature")
	ErrWrongTxHash    = errors.New("tx has wrong txs Merkle tree root")

	errTerminated = errors.New("terminated") // internal err
)

const (
	maxQueuedTasks = 128 // the maximum number of events to queue up
	maxBatch       = 4   // Maximum number of events in an task batch (batch is divided if exceeded)
)

// OnValidatedFn is a callback type for notifying about validation result.
type OnValidatedFn func(*TaskData)

// Check which require only parents list + current epoch info
type Validator struct {
	config   *lachesis.DagConfig
	txSigner types.Signer

	numOfThreads int

	tasksQ chan *TaskData
	quit   chan struct{}
	wg     sync.WaitGroup
}

type TaskData struct {
	Events inter.Events // events to validate
	Result []error      // resulting errors of events, nil if ok

	onValidated OnValidatedFn
}

// Uses N-1 threads
func NewDefault(config *lachesis.DagConfig, txSigner types.Signer) *Validator {
	threads := runtime.NumCPU()
	if threads > 1 {
		threads--
	}
	if threads < 1 {
		threads = 1
	}
	return New(config, txSigner, threads)
}

// Performs heavy checks, related to signatures validation and Merkle tree validation
func New(config *lachesis.DagConfig, txSigner types.Signer, numOfThreads int) *Validator {
	return &Validator{
		config:       config,
		txSigner:     txSigner,
		numOfThreads: numOfThreads,
		tasksQ:       make(chan *TaskData, maxQueuedTasks),
		quit:         make(chan struct{}),
	}
}

func (v *Validator) Start() {
	for i := 0; i < v.numOfThreads; i++ {
		v.wg.Add(1)
		go v.loop()
	}
}

func (v *Validator) Stop() {
	close(v.quit)
	v.wg.Wait()
}

func (v *Validator) Overloaded() bool {
	return len(v.tasksQ) > maxQueuedTasks/2
}

func (v *Validator) Enqueue(events inter.Events, onValidated OnValidatedFn) error {
	// divide big batch into smaller ones
	for start := 0; start < len(events); start += maxBatch {
		end := len(events)
		if end > start+maxBatch {
			end = start + maxBatch
		}
		op := &TaskData{
			Events:      events[start:end],
			onValidated: onValidated,
		}
		select {
		case v.tasksQ <- op:
			continue
		case <-v.quit:
			return errTerminated
		}
	}
	return nil
}

func (v *Validator) Validate(e *inter.Event) error {
	// event sig
	if !e.VerifySignature() {
		return ErrWrongEventSig
	}
	// pre-cache tx sig
	for _, tx := range e.Transactions {
		// TODO ensure that tx won't get serialized/deserialized before tx execution, because it'll invalidate the cache
		_, err := types.Sender(v.txSigner, tx)
		if err != nil {
			return ErrMalformedTxSig
		}
	}
	// Merkle tree
	if e.TxHash != types.DeriveSha(e.Transactions) {
		return ErrWrongTxHash
	}

	return nil
}

func (v *Validator) loop() {
	defer v.wg.Done()
	for {
		select {
		case <-v.quit:
			return

		case op := <-v.tasksQ:
			op.Result = make([]error, len(op.Events))
			for i, e := range op.Events {
				op.Result[i] = v.Validate(e)
			}
			op.onValidated(op)
		}
	}
}
