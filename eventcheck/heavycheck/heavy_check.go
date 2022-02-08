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

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/validatorpk"
)

var (
	ErrWrongEventSig            = errors.New("event has wrong signature")
	ErrMalformedTxSig           = errors.New("tx has wrong signature")
	ErrWrongPayloadHash         = errors.New("event has wrong payload hash")
	ErrPubkeyChanged            = errors.New("validator pubkey has changed, cannot create BVs/EV for older epochs")
	ErrUnknownEpochEventLocator = errors.New("event locator has unknown epoch")
	ErrImpossibleBVsEpoch       = errors.New("BVs have an impossible epoch")
	ErrUnknownEpochBVs          = errors.New("BVs are unprocessable yet")
	ErrUnknownEpochEV           = errors.New("EV is unprocessable yet")

	errTerminated = errors.New("terminated") // internal err
)

const (
	// MaxBlocksPerEpoch is chosen so that even if validator chooses the latest non-liable epoch for BVs,
	// he still cannot vote for latest blocks (latest = from last 128 epochs), as an epoch has at least one block
	// The value is larger than a maximum possible number of blocks
	// in an epoch where a single validator doesn't have 2/3W+1 weight
	MaxBlocksPerEpoch = idx.Block(basiccheck.MaxLiableEpochs - 128)
)

// Reader is accessed by the validator to get the current state.
type Reader interface {
	GetEpochPubKeys() (map[idx.ValidatorID]validatorpk.PubKey, idx.Epoch)
	GetEpochPubKeysOf(idx.Epoch) map[idx.ValidatorID]validatorpk.PubKey
	GetEpochBlockStart(idx.Epoch) idx.Block
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
	ev    *inter.LlrSignedEpochVote

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

func (v *Checker) EnqueueEV(ev inter.LlrSignedEpochVote, onValidated func(error)) error {
	op := &taskData{
		ev:          &ev,
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

func (v *Checker) ValidateEventLocator(e inter.SignedEventLocator, authEpoch idx.Epoch, authErr error, checkPayload func() bool) error {
	pubkeys := v.reader.GetEpochPubKeysOf(authEpoch)
	if len(pubkeys) == 0 {
		return authErr
	}
	pubkey, ok := pubkeys[e.Locator.Creator]
	if !ok {
		return epochcheck.ErrAuth
	}
	if checkPayload != nil && !checkPayload() {
		return ErrWrongPayloadHash
	}
	if !verifySignature(e.Locator.HashToSign(), e.Sig, pubkey) {
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

func (v *Checker) validateBVsEpoch(bvs inter.LlrBlockVotes) error {
	actualEpochStart := v.reader.GetEpochBlockStart(bvs.Epoch)
	if actualEpochStart == 0 {
		return ErrUnknownEpochBVs
	}
	if bvs.Start < actualEpochStart || bvs.LastBlock() >= actualEpochStart+MaxBlocksPerEpoch {
		return ErrImpossibleBVsEpoch
	}
	return nil
}

func (v *Checker) ValidateBVs(bvs inter.LlrSignedBlockVotes) error {
	if err := v.validateBVsEpoch(bvs.Val); err != nil {
		return err
	}
	return v.ValidateEventLocator(bvs.Signed, bvs.Val.Epoch, ErrUnknownEpochBVs, func() bool {
		return bvs.CalcPayloadHash() == bvs.Signed.Locator.PayloadHash
	})
}

func (v *Checker) ValidateEV(ev inter.LlrSignedEpochVote) error {
	return v.ValidateEventLocator(ev.Signed, ev.Val.Epoch-1, ErrUnknownEpochEV, func() bool {
		return ev.CalcPayloadHash() == ev.Signed.Locator.PayloadHash
	})
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
			for _, vote := range proof.Pair {
				if err := v.ValidateEventLocator(vote, vote.Locator.Epoch, ErrUnknownEpochEventLocator, nil); err != nil {
					return err
				}
			}
		}
		if proof := mp.BlockVoteDoublesign; proof != nil {
			for _, vote := range proof.Pair {
				if err := v.ValidateBVs(vote); err != nil {
					return err
				}
			}
		}
		if proof := mp.WrongBlockVote; proof != nil {
			for _, pal := range proof.Pals {
				if err := v.ValidateBVs(pal); err != nil {
					return err
				}
			}
		}
		if proof := mp.EpochVoteDoublesign; proof != nil {
			for _, vote := range proof.Pair {
				if err := v.ValidateEV(vote); err != nil {
					return err
				}
			}
		}
		if proof := mp.WrongEpochVote; proof != nil {
			for _, pal := range proof.Pals {
				if err := v.ValidateEV(pal); err != nil {
					return err
				}
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
		// ensure that validator's pubkey is the same in both current and vote epochs
		if err := v.matchPubkey(e.Creator(), e.EpochVote().Epoch-1, pubkey.Bytes(), ErrUnknownEpochEV); err != nil {
			return err
		}
	}
	if e.BlockVotes().Epoch != 0 {
		// ensure that validator's BVs epoch passes the check
		if err := v.validateBVsEpoch(e.BlockVotes()); err != nil {
			return err
		}
		// ensure that validator's pubkey is the same in both current and vote epochs
		if e.BlockVotes().Epoch != e.Epoch() {
			if err := v.matchPubkey(e.Creator(), e.BlockVotes().Epoch, pubkey.Bytes(), ErrUnknownEpochBVs); err != nil {
				return err
			}
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
				op.onValidated(v.ValidateEV(*op.ev))
			}

		case <-v.quit:
			return
		}
	}
}
