package heavycheck

import (
	"bytes"
	"errors"
	"runtime"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
)

var (
	ErrWrongEventSig            = errors.New("event has wrong signature")
	ErrMalformedTxSig           = errors.New("tx has wrong signature")
	ErrWrongPayloadHash         = errors.New("event has wrong txs payload hash")
	ErrPubkeyChanged            = errors.New("validator pubkey has changed, cannot create BVs/EV for older epochs")
	ErrUnknownEpochEventLocator = errors.New("event locator has unknown epoch")
	ErrUnknownEpochBVs          = errors.New("BVs is unprocessable yet")
	ErrUnknownEpochEV           = errors.New("EV is unprocessable yet")

	errTerminated = errors.New("terminated") // internal err
)

// Reader is accessed by the validator to get the current state.
type Reader interface {
	GetEpochPubKeys() (map[idx.ValidatorID]validatorpk.PubKey, idx.Epoch)
	GetEpochPubKeysOf(idx.Epoch) map[idx.ValidatorID]validatorpk.PubKey
}

// Checker which requires only parents list + current epoch info
type Checker struct {
	config   Config
	txSigner types.Signer
	reader   Reader

	tasksQ chan *taskData
	quit   chan struct{}
	wg     sync.WaitGroup
}

type taskData struct {
	event inter.EventPayloadI
	bvs   *inter.LlrSignedBlockVotes
	ers   *inter.LlrSignedEpochVote

	onValidated func(error)
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
		tasksQ:   make(chan *taskData, config.MaxQueuedTasks),
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
	return len(v.tasksQ) > v.config.MaxQueuedTasks/2
}

func (v *Checker) EnqueueEvent(e inter.EventPayloadI, onValidated func(error)) error {
	op := &taskData{
		event:       e,
		onValidated: onValidated,
	}
	select {
	case v.tasksQ <- op:
		return nil
	case <-v.quit:
		return errTerminated
	}
}

func (v *Checker) EnqueueBVs(bvs inter.LlrSignedBlockVotes, onValidated func(error)) error {
	op := &taskData{
		bvs:         &bvs,
		onValidated: onValidated,
	}
	select {
	case v.tasksQ <- op:
		return nil
	case <-v.quit:
		return errTerminated
	}
}

func (v *Checker) EnqueueEV(ers inter.LlrSignedEpochVote, onValidated func(error)) error {
	op := &taskData{
		ers:         &ers,
		onValidated: onValidated,
	}
	select {
	case v.tasksQ <- op:
		return nil
	case <-v.quit:
		return errTerminated
	}
}

// verifySignature checks the signature against e.Creator.
func verifySignature(signedHash hash.Hash, sig inter.Signature, pubkey validatorpk.PubKey) bool {
	if pubkey.Type != validatorpk.Types.Secp256k1 {
		return false
	}
	return crypto.VerifySignature(pubkey.Raw, signedHash.Bytes(), sig.Bytes())
}

func (v *Checker) ValidateEventLocator(e inter.SignedEventLocator, authEpoch idx.Epoch, authErr error) error {
	pubkeys := v.reader.GetEpochPubKeysOf(authEpoch)
	if len(pubkeys) == 0 {
		return authErr
	}
	pubkey, ok := pubkeys[e.Creator]
	if !ok {
		return epochcheck.ErrAuth
	}
	if !verifySignature(e.HashToSign(), e.Sig, pubkey) {
		return ErrWrongEventSig
	}
	return nil
}

func (v *Checker) matchPubkey(creator idx.ValidatorID, epoch idx.Epoch, want []byte, authErr error) error {
	pubkeys := v.reader.GetEpochPubKeysOf(epoch)
	if len(pubkeys) == 0 {
		return authErr
	}
	pubkey, ok := pubkeys[creator]
	if !ok {
		return epochcheck.ErrAuth
	}
	if bytes.Compare(pubkey.Bytes(), want) != 0 {
		return ErrPubkeyChanged
	}
	return nil
}

func (v *Checker) ValidateBVs(bvs inter.LlrSignedBlockVotes) error {
	if bvs.CalcPayloadHash() != bvs.EventLocator.PayloadHash {
		return ErrWrongPayloadHash
	}
	return v.ValidateEventLocator(bvs.EventLocator, bvs.Epoch, ErrUnknownEpochBVs)
}

func (v *Checker) ValidateEV(ers inter.LlrSignedEpochVote) error {
	if ers.CalcPayloadHash() != ers.EventLocator.PayloadHash {
		return ErrWrongPayloadHash
	}
	return v.ValidateEventLocator(ers.EventLocator, ers.Epoch-1, ErrUnknownEpochEV)
}

// ValidateEvent runs heavy checks for event
func (v *Checker) ValidateEvent(e inter.EventPayloadI) error {
	pubkeys, epoch := v.reader.GetEpochPubKeys()
	if e.Epoch() != epoch {
		return epochcheck.ErrNotRelevant
	}
	// validatorID
	pubkey, ok := pubkeys[e.Creator()]
	if !ok {
		return epochcheck.ErrAuth
	}
	// event sig
	if !verifySignature(e.HashToSign(), e.Sig(), pubkey) {
		return ErrWrongEventSig
	}
	// MPs
	for _, mp := range e.MisbehaviourProofs() {
		if proof := mp.EventsDoublesign; proof != nil {
			if err := v.ValidateEventLocator(proof.Pair[0], proof.Pair[0].Epoch, ErrUnknownEpochEventLocator); err != nil {
				return err
			}
			if err := v.ValidateEventLocator(proof.Pair[1], proof.Pair[1].Epoch, ErrUnknownEpochEventLocator); err != nil {
				return err
			}
		}
		if proof := mp.BlockVoteDoublesign; proof != nil {
			if err := v.ValidateBVs(proof.Pair[0]); err != nil {
				return err
			}
			if err := v.ValidateBVs(proof.Pair[1]); err != nil {
				return err
			}
		}
		if proof := mp.WrongBlockVote; proof != nil {
			if err := v.ValidateBVs(proof.Votes); err != nil {
				return err
			}
		}
		if proof := mp.EpochVoteDoublesign; proof != nil {
			if err := v.ValidateEV(proof.Pair[0]); err != nil {
				return err
			}
			if err := v.ValidateEV(proof.Pair[1]); err != nil {
				return err
			}
		}
		if proof := mp.WrongEpochVote; proof != nil {
			if err := v.ValidateEV(proof.Votes); err != nil {
				return err
			}
		}
	}
	// pre-cache tx sig
	for _, tx := range e.Txs() {
		_, err := types.Sender(v.txSigner, tx)
		if err != nil {
			return ErrMalformedTxSig
		}
	}
	// Payload hash
	if e.PayloadHash() != inter.CalcPayloadHash(e) {
		return ErrWrongPayloadHash
	}
	// Epochs of BVs and EV
	if e.EpochVote().Epoch != 0 {
		if err := v.matchPubkey(e.Creator(), e.EpochVote().Epoch-1, pubkey.Bytes(), ErrUnknownEpochEV); err != nil {
			return err
		}
	}
	if e.BlockVotes().Epoch != 0 {
		if err := v.matchPubkey(e.Creator(), e.BlockVotes().Epoch, pubkey.Bytes(), ErrUnknownEpochBVs); err != nil {
			return err
		}
	}

	return nil
}

func (v *Checker) loop() {
	defer v.wg.Done()
	for {
		select {
		case op := <-v.tasksQ:
			if op.event != nil {
				op.onValidated(v.ValidateEvent(op.event))
			} else if op.bvs != nil {
				op.onValidated(v.ValidateBVs(*op.bvs))
			} else {
				op.onValidated(v.ValidateEV(*op.ers))
			}

		case <-v.quit:
			return
		}
	}
}
