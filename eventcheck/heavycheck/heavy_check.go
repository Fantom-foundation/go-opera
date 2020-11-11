package heavycheck

import (
	"errors"
	"runtime"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/validator"
	"github.com/Fantom-foundation/go-opera/opera"
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

// Reader is accessed by the validator to get the current state.
type Reader interface {
	GetEpochPubKeys() (map[idx.ValidatorID]validator.PubKey, idx.Epoch)
}

// Check which require only parents list + current epoch info
type Checker struct {
	config   *opera.DagConfig
	txSigner types.Signer
	reader   Reader

	numOfThreads int

	tasksQ chan *TaskData
	quit   chan struct{}
	wg     sync.WaitGroup
}

type TaskData struct {
	Events dag.Events // events to validate

	onValidated func(dag.Events, []error)
}

// NewDefault uses N-1 threads
func NewDefault(config *opera.DagConfig, reader Reader, txSigner types.Signer) *Checker {
	threads := runtime.NumCPU()
	if threads > 1 {
		threads--
	}
	if threads < 1 {
		threads = 1
	}
	return New(config, reader, txSigner, threads)
}

// New validator which performs heavy checks, related to signatures validation and Merkle tree validation
func New(config *opera.DagConfig, reader Reader, txSigner types.Signer, numOfThreads int) *Checker {
	return &Checker{
		config:       config,
		txSigner:     txSigner,
		reader:       reader,
		numOfThreads: numOfThreads,
		tasksQ:       make(chan *TaskData, maxQueuedTasks),
		quit:         make(chan struct{}),
	}
}

func (v *Checker) Start() {
	for i := 0; i < v.numOfThreads; i++ {
		v.wg.Add(1)
		go v.loop()
	}
}

func (v *Checker) Stop() {
	close(v.quit)
	v.wg.Wait()
}

func (v *Checker) Overloaded() bool {
	return len(v.tasksQ) > maxQueuedTasks/2
}

func (v *Checker) Enqueue(events dag.Events, onValidated func(dag.Events, []error)) error {
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

// verifySignature checks the signature against e.Creator.
func verifySignature(e inter.EventPayloadI, pubkey validator.PubKey) bool {
	if pubkey.Type != "secp256k1" {
		return false
	}
	signedHash := e.HashToSign().Bytes()
	sig := e.Sig()
	return crypto.VerifySignature(pubkey.Raw, signedHash, sig.Bytes())
}

// Validate event
func (v *Checker) Validate(de dag.Event) error {
	e := de.(inter.EventPayloadI)
	addrs, epoch := v.reader.GetEpochPubKeys()
	if e.Epoch() != epoch {
		return epochcheck.ErrNotRelevant
	}
	// validatorID
	addr, ok := addrs[e.Creator()]
	if !ok {
		return epochcheck.ErrAuth
	}
	// event sig
	if !verifySignature(e, addr) {
		return ErrWrongEventSig
	}
	// pre-cache tx sig
	for _, tx := range e.Txs() {
		_, err := types.Sender(v.txSigner, tx)
		if err != nil {
			return ErrMalformedTxSig
		}
	}
	// Merkle tree
	if e.TxHash() != hash.Hash(types.DeriveSha(e.Txs(), new(trie.Trie))) {
		return ErrWrongTxHash
	}

	return nil
}

func (v *Checker) loop() {
	defer v.wg.Done()
	for {
		select {
		case <-v.quit:
			return

		case op := <-v.tasksQ:
			result := make([]error, len(op.Events))
			for i, e := range op.Events {
				result[i] = v.Validate(e)
			}
			op.onValidated(op.Events, result)
		}
	}
}
