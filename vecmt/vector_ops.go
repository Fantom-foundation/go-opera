package vecmt

import (
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/vecengine"
	"github.com/Fantom-foundation/lachesis-base/vecfc"

	"github.com/Fantom-foundation/go-opera/inter"
)

type CreationTimer interface {
	CreationTime() inter.Timestamp
}

func (b *HighestBefore) InitWithEvent(i idx.Validator, e dag.Event) {
	b.VSeq.InitWithEvent(i, e)
	b.VTime.Set(i, e.(CreationTimer).CreationTime())
}

func (b *HighestBefore) IsEmpty(i idx.Validator) bool {
	return b.VSeq.IsEmpty(i)
}

func (b *HighestBefore) IsForkDetected(i idx.Validator) bool {
	return b.VSeq.IsForkDetected(i)
}

func (b *HighestBefore) Seq(i idx.Validator) idx.Event {
	return b.VSeq.Seq(i)
}

func (b *HighestBefore) MinSeq(i idx.Validator) idx.Event {
	return b.VSeq.MinSeq(i)
}

func (b *HighestBefore) SetForkDetected(i idx.Validator) {
	b.VSeq.SetForkDetected(i)
}

func (self *HighestBefore) CollectFrom(_other vecengine.HighestBeforeI, num idx.Validator) {
	other := _other.(*HighestBefore)
	for branchID := idx.Validator(0); branchID < num; branchID++ {
		hisSeq := other.VSeq.Get(branchID)
		if hisSeq.Seq == 0 && !hisSeq.IsForkDetected() {
			// hisSeq doesn't observe anything about this branchID
			continue
		}
		mySeq := self.VSeq.Get(branchID)

		if mySeq.IsForkDetected() {
			// mySeq observes the maximum already
			continue
		}
		if hisSeq.IsForkDetected() {
			// set fork detected
			self.SetForkDetected(branchID)
		} else {
			if mySeq.Seq == 0 || mySeq.MinSeq > hisSeq.MinSeq {
				// take hisSeq.MinSeq
				mySeq.MinSeq = hisSeq.MinSeq
				self.VSeq.Set(branchID, mySeq)
			}
			if mySeq.Seq < hisSeq.Seq {
				// take hisSeq.Seq
				mySeq.Seq = hisSeq.Seq
				self.VSeq.Set(branchID, mySeq)
				self.VTime.Set(branchID, other.VTime.Get(branchID))
			}
		}
	}
}

func (self *HighestBefore) GatherFrom(to idx.Validator, _other vecengine.HighestBeforeI, from []idx.Validator) {
	other := _other.(*HighestBefore)
	// read all branches to find highest event
	highestBranchSeq := vecfc.BranchSeq{}
	highestBranchTime := inter.Timestamp(0)
	for _, branchID := range from {
		vseq := other.VSeq.Get(branchID)
		vtime := other.VTime.Get(branchID)
		if vseq.IsForkDetected() {
			highestBranchSeq = vseq
			break
		}
		if vseq.Seq > highestBranchSeq.Seq {
			highestBranchSeq = vseq
			highestBranchTime = vtime
		}
	}
	self.VSeq.Set(to, highestBranchSeq)
	self.VTime.Set(to, highestBranchTime)
}
