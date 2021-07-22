package heavycheck

import (
	"errors"
	"runtime"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/eventcheck/epochcheck"
	"github.com/Fantom-foundation/lachesis-base/eventcheck/queuedcheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
)

var (
	ErrWrongEventSig  = errors.New("event has wrong signature")
	ErrMalformedTxSig = errors.New("tx has wrong signature")
	ErrWrongTxHash    = errors.New("tx has wrong txs Merkle tree root")

	errTerminated = errors.New("terminated") // internal err
)

// Reader is accessed by the validator to get the current state.
type Reader interface {
	GetEpochPubKeys() (map[idx.ValidatorID]validatorpk.PubKey, idx.Epoch)
}

// Checker which requires only parents list + current epoch info
type Checker struct {
	config   Config
	txSigner types.Signer
	reader   Reader

	tasksQ chan *TasksData
	quit   chan struct{}
	wg     sync.WaitGroup
}

type TasksData struct {
	Tasks []queuedcheck.EventTask // events to validate

	onValidated func([]queuedcheck.EventTask)
}

// New validator which performs heavy checks, related to signatures validation and Merkle tree validation
func New(config Config, reader Reader, txSigner types.Signer) *Checker {
	if config.Threads == 0 {
		config.Threads = runtime.NumCPU()
		if config.Threads > 1 {
			config.Threads--
		}
		if config.Threads < 1 {
			config.Threads = 1
		}
	}
	return &Checker{
		config:   config,
		txSigner: txSigner,
		reader:   reader,
		tasksQ:   make(chan *TasksData, config.MaxQueuedBatches),
		quit:     make(chan struct{}),
	}
}

func (v *Checker) Start() {
	for i := 0; i < v.config.Threads; i++ {
		v.wg.Add(1)
		go v.loop()
	}
}

func (v *Checker) Stop() {
	close(v.quit)
	v.wg.Wait()
}

func (v *Checker) Overloaded() bool {
	return len(v.tasksQ) > v.config.MaxQueuedBatches/2
}

func (v *Checker) Enqueue(tasks []queuedcheck.EventTask, onValidated func([]queuedcheck.EventTask)) error {
	// divide big batch into smaller ones
	for start := 0; start < len(tasks); start += v.config.MaxBatch {
		end := len(tasks)
		if end > start+v.config.MaxBatch {
			end = start + v.config.MaxBatch
		}
		op := &TasksData{
			Tasks:       tasks[start:end],
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
func verifySignature(e inter.EventPayloadI, pubkey validatorpk.PubKey) bool {
	if pubkey.Type != validatorpk.Types.Secp256k1 {
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
		case op := <-v.tasksQ:
			for _, t := range op.Tasks {
				t.SetResult(v.Validate(t.Event()))
			}
			op.onValidated(op.Tasks)

		case <-v.quit:
			return
		}
	}
}
