package basiccheck

import (
	"errors"
	"math"

	base "github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/inter"
)

var (
	ErrZeroTime      = errors.New("event has zero timestamp")
	ErrNegativeValue = errors.New("negative value")
	ErrIntrinsicGas  = errors.New("intrinsic gas too low")
	// ErrTipAboveFeeCap is a sanity error to ensure no one is able to specify a
	// transaction with a tip higher than the total fee cap.
	ErrTipAboveFeeCap = errors.New("max priority fee per gas higher than max fee per gas")
	ErrWrongMP        = errors.New("inconsistent misbehaviour proof")
	ErrMalformedMP    = errors.New("malformed MP union struct")
	FutureBVsEpoch    = errors.New("future block votes epoch")
	FutureEVEpoch     = errors.New("future epoch vote")
	MalformedBVs      = errors.New("malformed BVs")
	MalformedEV       = errors.New("malformed EV")
	TooManyBVs        = errors.New("too many BVs")
	EmptyEV           = errors.New("empty EV")
	EmptyBVs          = errors.New("empty BVs")
)

const (
	MaxBlockVotesPerEvent = 64
)

type Checker struct {
	base base.Checker
}

// New validator which performs checks which don't require anything except event
func New() *Checker {
	return &Checker{
		base: base.Checker{},
	}
}

// validateTx checks whether a transaction is valid according to the consensus
// rules
func validateTx(tx *types.Transaction) error {
	// Transactions can't be negative. This may never happen using RLP decoded
	// transactions but may occur if you create a transaction using the RPC.
	if tx.Value().Sign() < 0 || tx.GasPrice().Sign() < 0 {
		return ErrNegativeValue
	}
	// Ensure the transaction has more gas than the basic tx fee.
	intrGas, err := evmcore.IntrinsicGas(tx.Data(), tx.AccessList(), tx.To() == nil)
	if err != nil {
		return err
	}
	if tx.Gas() < intrGas {
		return ErrIntrinsicGas
	}
	if tx.GasFeeCapIntCmp(tx.GasTipCap()) < 0 {
		return ErrTipAboveFeeCap
	}
	return nil
}

func (v *Checker) validateMP(mp inter.MisbehaviourProof) error {
	count := 0
	if proof := mp.EventsDoublesign; proof != nil {
		count++
		if proof.Pair[0].Creator != proof.Pair[1].Creator {
			return ErrWrongMP
		}
		if proof.Pair[0].Epoch != proof.Pair[1].Epoch {
			return ErrWrongMP
		}
		if proof.Pair[0].Seq != proof.Pair[1].Seq {
			return ErrWrongMP
		}
		if proof.Pair[0].HashToSign() == proof.Pair[1].HashToSign() {
			return ErrWrongMP
		}
	}
	if proof := mp.BlockVoteDoublesign; proof != nil {
		count++
		if err := v.ValidateBVs(proof.Pair[0]); err != nil {
			return ErrWrongMP
		}
		if err := v.ValidateBVs(proof.Pair[1]); err != nil {
			return ErrWrongMP
		}
		if proof.Block < proof.Pair[0].Start || proof.Block >= proof.Pair[0].Start+idx.Block(len(proof.Pair[0].Votes)) {
			return ErrWrongMP
		}
		if proof.Block < proof.Pair[1].Start || proof.Block >= proof.Pair[1].Start+idx.Block(len(proof.Pair[1].Votes)) {
			return ErrWrongMP
		}
		if proof.GetVote(0) == proof.GetVote(1) {
			return ErrWrongMP
		}
	}
	if proof := mp.WrongBlockVote; proof != nil {
		count++
		if err := v.ValidateBVs(proof.Votes); err != nil {
			return ErrWrongMP
		}
		if proof.Block < proof.Votes.Start || proof.Block >= proof.Votes.Start+idx.Block(len(proof.Votes.Votes)) {
			return ErrWrongMP
		}
	}
	if proof := mp.EpochVoteDoublesign; proof != nil {
		count++
		if err := v.ValidateEV(proof.Pair[0]); err != nil {
			return ErrWrongMP
		}
		if err := v.ValidateEV(proof.Pair[1]); err != nil {
			return ErrWrongMP
		}
		if proof.Pair[0].Epoch != proof.Pair[1].Epoch {
			return ErrWrongMP
		}
		if proof.Pair[0].Vote == proof.Pair[1].Vote {
			return ErrWrongMP
		}
	}
	if proof := mp.WrongEpochVote; proof != nil {
		count++
		if err := v.ValidateEV(proof.Votes); err != nil {
			return ErrWrongMP
		}
	}
	if count != 1 {
		return ErrMalformedMP
	}
	return nil
}

func (v *Checker) checkTxs(e inter.EventPayloadI) error {
	for _, tx := range e.Txs() {
		if err := validateTx(tx); err != nil {
			return err
		}
	}
	return nil
}

// Validate event
func (v *Checker) Validate(e inter.EventPayloadI) error {
	if err := v.base.Validate(e); err != nil {
		return err
	}
	if e.GasPowerUsed() >= math.MaxInt64-1 || e.GasPowerLeft().Max() >= math.MaxInt64-1 {
		return base.ErrHugeValue
	}
	if e.CreationTime() <= 0 || e.MedianTime() <= 0 {
		return ErrZeroTime
	}
	if err := v.checkTxs(e); err != nil {
		return err
	}
	for _, mp := range e.MisbehaviourProofs() {
		if err := v.validateMP(mp); err != nil {
			return err
		}
	}
	if err := v.validateEV(e.Epoch(), e.EpochVote(), false); err != nil {
		return err
	}
	if err := v.validateBVs(e.Epoch(), e.BlockVotes(), false); err != nil {
		return err
	}

	return nil
}

func (v *Checker) validateBVs(eventEpoch idx.Epoch, bvs inter.LlrBlockVotes, greedy bool) error {
	if bvs.Epoch > eventEpoch {
		return FutureBVsEpoch
	}
	if bvs.Start >= math.MaxInt64/2 {
		return base.ErrHugeValue
	}
	if bvs.Epoch >= math.MaxInt32-1 {
		return base.ErrHugeValue
	}
	if len(bvs.Votes) > MaxBlockVotesPerEvent {
		return TooManyBVs
	}
	if ((bvs.Start == 0) != (len(bvs.Votes) == 0)) || ((bvs.Start == 0) != (bvs.Epoch == 0)) {
		return MalformedBVs
	}
	if ((bvs.Start == 0) != (len(bvs.Votes) == 0)) || ((bvs.Start == 0) != (bvs.Epoch == 0)) {
		return MalformedBVs
	}
	if greedy && bvs.Epoch == 0 {
		return EmptyBVs
	}
	return nil
}

func (v *Checker) validateEV(eventEpoch idx.Epoch, ev inter.LlrEpochVote, greedy bool) error {
	if ev.Epoch > eventEpoch {
		return FutureEVEpoch
	}
	if (ev.Epoch == 0) != (ev.Vote == hash.Zero) {
		return MalformedEV
	}
	if ev.Epoch >= math.MaxInt32-1 {
		return base.ErrHugeValue
	}
	if greedy && ev.Epoch == 0 {
		return EmptyEV
	}
	return nil
}

func (v *Checker) ValidateBVs(bvs inter.LlrSignedBlockVotes) error {
	return v.validateBVs(bvs.EventLocator.Epoch, bvs.LlrBlockVotes, true)
}

func (v *Checker) ValidateEV(ev inter.LlrSignedEpochVote) error {
	return v.validateEV(ev.EventLocator.Epoch, ev.LlrEpochVote, true)
}
