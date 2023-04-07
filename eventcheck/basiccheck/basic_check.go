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
	ErrWrongNetForkID = errors.New("wrong network fork ID")
	ErrZeroTime       = errors.New("event has zero timestamp")
	ErrNegativeValue  = errors.New("negative value")
	ErrIntrinsicGas   = errors.New("intrinsic gas too low")
	// ErrTipAboveFeeCap is a sanity error to ensure no one is able to specify a
	// transaction with a tip higher than the total fee cap.
	ErrTipAboveFeeCap = errors.New("max priority fee per gas higher than max fee per gas")
	ErrWrongMP        = errors.New("inconsistent misbehaviour proof")
	ErrNoCrimeInMP    = errors.New("action in misbehaviour proof isn't a criminal offence")
	ErrWrongCreatorMP = errors.New("wrong creator in misbehaviour proof")
	ErrMPTooLate      = errors.New("too old misbehaviour proof")
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
	MaxLiableEpochs       = 32768
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
	intrGas, err := evmcore.IntrinsicGas(tx.Data(), tx.AccessList(), tx.To() == nil, tx.Type() == types.AccountAbstractionTxType)
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

func (v *Checker) validateMP(msgEpoch idx.Epoch, mp inter.MisbehaviourProof) error {
	count := 0
	if proof := mp.EventsDoublesign; proof != nil {
		count++
		if err := v.validateEventLocator(proof.Pair[0].Locator); err != nil {
			return err
		}
		if err := v.validateEventLocator(proof.Pair[1].Locator); err != nil {
			return err
		}
		if proof.Pair[0].Locator.Creator != proof.Pair[1].Locator.Creator {
			return ErrWrongCreatorMP
		}
		if proof.Pair[0].Locator.Epoch != proof.Pair[1].Locator.Epoch {
			return ErrNoCrimeInMP
		}
		if proof.Pair[0].Locator.Seq != proof.Pair[1].Locator.Seq {
			return ErrNoCrimeInMP
		}
		if proof.Pair[0].Locator == proof.Pair[1].Locator {
			return ErrNoCrimeInMP
		}
		if msgEpoch > proof.Pair[0].Locator.Epoch+MaxLiableEpochs {
			return ErrMPTooLate
		}
	}
	if proof := mp.BlockVoteDoublesign; proof != nil {
		count++
		if err := v.ValidateBVs(proof.Pair[0]); err != nil {
			return err
		}
		if err := v.ValidateBVs(proof.Pair[1]); err != nil {
			return err
		}
		if proof.Pair[0].Signed.Locator.Creator != proof.Pair[1].Signed.Locator.Creator {
			return ErrWrongCreatorMP
		}
		if proof.Block < proof.Pair[0].Val.Start || proof.Block >= proof.Pair[0].Val.Start+idx.Block(len(proof.Pair[0].Val.Votes)) {
			return ErrWrongMP
		}
		if proof.Block < proof.Pair[1].Val.Start || proof.Block >= proof.Pair[1].Val.Start+idx.Block(len(proof.Pair[1].Val.Votes)) {
			return ErrWrongMP
		}
		if proof.GetVote(0) == proof.GetVote(1) && proof.Pair[0].Val.Epoch == proof.Pair[1].Val.Epoch {
			return ErrNoCrimeInMP
		}
		if msgEpoch > proof.Pair[0].Val.Epoch+MaxLiableEpochs || msgEpoch > proof.Pair[1].Val.Epoch+MaxLiableEpochs {
			return ErrMPTooLate
		}
	}
	if proof := mp.WrongBlockVote; proof != nil {
		count++
		for i, pal := range proof.Pals {
			if err := v.ValidateBVs(pal); err != nil {
				return err
			}
			if proof.Block < pal.Val.Start || proof.Block >= pal.Val.Start+idx.Block(len(pal.Val.Votes)) {
				return ErrWrongMP
			}
			if msgEpoch > pal.Val.Epoch+MaxLiableEpochs {
				return ErrMPTooLate
			}
			// see MinAccomplicesForProof
			if proof.WrongEpoch {
				if i > 0 && pal.Val.Epoch != proof.Pals[i-1].Val.Epoch {
					return ErrNoCrimeInMP
				}
			} else {
				if i > 0 && proof.GetVote(i-1) != proof.GetVote(i) {
					return ErrNoCrimeInMP
				}
			}
			for _, prev := range proof.Pals[:i] {
				if prev.Signed.Locator.Creator == pal.Signed.Locator.Creator {
					return ErrWrongCreatorMP
				}
			}
		}
	}
	if proof := mp.EpochVoteDoublesign; proof != nil {
		count++
		if err := v.ValidateEV(proof.Pair[0]); err != nil {
			return err
		}
		if err := v.ValidateEV(proof.Pair[1]); err != nil {
			return err
		}
		if proof.Pair[0].Signed.Locator.Creator != proof.Pair[1].Signed.Locator.Creator {
			return ErrWrongCreatorMP
		}
		if proof.Pair[0].Val.Epoch != proof.Pair[1].Val.Epoch {
			return ErrNoCrimeInMP
		}
		if proof.Pair[0].Val.Vote == proof.Pair[1].Val.Vote {
			return ErrNoCrimeInMP
		}
		if msgEpoch > proof.Pair[0].Val.Epoch+MaxLiableEpochs {
			return ErrMPTooLate
		}
	}
	if proof := mp.WrongEpochVote; proof != nil {
		count++
		for i, pal := range proof.Pals {
			if err := v.ValidateEV(pal); err != nil {
				return err
			}
			if msgEpoch > pal.Val.Epoch+MaxLiableEpochs {
				return ErrMPTooLate
			}
			// see MinAccomplicesForProof
			if i > 0 && proof.Pals[i-1].Val != proof.Pals[i].Val {
				return ErrNoCrimeInMP
			}
			for _, prev := range proof.Pals[:i] {
				if prev.Signed.Locator.Creator == pal.Signed.Locator.Creator {
					return ErrWrongCreatorMP
				}
			}
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
	if e.NetForkID() != 0 {
		return ErrWrongNetForkID
	}
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
		if err := v.validateMP(e.Epoch(), mp); err != nil {
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

func (v *Checker) validateEventLocator(e inter.EventLocator) error {
	if e.NetForkID != 0 {
		return ErrWrongNetForkID
	}
	if e.Seq >= math.MaxInt32-1 || e.Epoch >= math.MaxInt32-1 ||
		e.Lamport >= math.MaxInt32-1 {
		return base.ErrHugeValue
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
	if err := v.validateEventLocator(bvs.Signed.Locator); err != nil {
		return err
	}
	return v.validateBVs(bvs.Signed.Locator.Epoch, bvs.Val, true)
}

func (v *Checker) ValidateEV(ev inter.LlrSignedEpochVote) error {
	if err := v.validateEventLocator(ev.Signed.Locator); err != nil {
		return err
	}
	return v.validateEV(ev.Signed.Locator.Epoch, ev.Val, true)
}
