package proclogger

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/ibr"
	"github.com/Fantom-foundation/go-opera/inter/ier"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
)

type dagSum struct {
	connected       idx.Event
	totalProcessing time.Duration
}

type llrSum struct {
	bvs idx.Block
	brs idx.Block
	evs idx.Epoch
	ers idx.Epoch
}

type Logger struct {
	// summary accumulators
	dagSum dagSum
	llrSum llrSum

	// latest logged data
	lastEpoch     idx.Epoch
	lastBlock     idx.Block
	lastID        hash.Event
	lastEventTime inter.Timestamp
	lastLlrTime   inter.Timestamp

	nextLogging time.Time

	emitting  bool
	noSummary bool

	logger.Instance
}

func (l *Logger) summary(now time.Time) {
	if l.noSummary {
		return
	}
	if now.After(l.nextLogging) {
		if l.llrSum != (llrSum{}) {
			age := utils.PrettyDuration(now.Sub(l.lastLlrTime.Time())).String()
			if l.lastLlrTime <= l.lastEventTime {
				age = "none"
			}
			l.Log.Info("New LLR summary", "last_epoch", l.lastEpoch, "last_block", l.lastBlock,
				"new_evs", l.llrSum.evs, "new_ers", l.llrSum.ers, "new_bvs", l.llrSum.bvs, "new_brs", l.llrSum.brs, "age", age)
		}
		if l.dagSum != (dagSum{}) {
			l.Log.Info("New DAG summary", "new", l.dagSum.connected, "last_id", l.lastID.String(),
				"age", utils.PrettyDuration(now.Sub(l.lastEventTime.Time())), "t", utils.PrettyDuration(l.dagSum.totalProcessing))
		}
		l.dagSum = dagSum{}
		l.llrSum = llrSum{}
		l.nextLogging = now.Add(8 * time.Second)
	}
}

// BlockVotesConnectionStarted starts the BVs logging
// Not safe for concurrent use
func (l *Logger) BlockVotesConnectionStarted(bvs inter.LlrSignedBlockVotes) func() {
	if bvs.Val.Epoch == 0 {
		return func() {}
	}
	l.llrSum.bvs += idx.Block(len(bvs.Val.Votes))

	start := time.Now()

	return func() {
		if l.lastBlock < bvs.Val.LastBlock() {
			l.lastBlock = bvs.Val.LastBlock()
		}
		now := time.Now()
		// logging for the individual item
		msg := "New BVs"
		logType := l.Log.Debug
		if l.emitting {
			msg = "New BVs emitted"
			logType = l.Log.Info
		}
		logType(msg, "id", bvs.Signed.Locator.ID(), "by", bvs.Signed.Locator.Creator,
			"blocks", fmt.Sprintf("%d-%d", bvs.Val.Start, bvs.Val.LastBlock()),
			"t", utils.PrettyDuration(now.Sub(start)))
		l.summary(now)
	}
}

// BlockRecordConnectionStarted starts the BR logging
// Not safe for concurrent use
func (l *Logger) BlockRecordConnectionStarted(br ibr.LlrIdxFullBlockRecord) func() {
	l.llrSum.brs++

	start := time.Now()

	return func() {
		if l.lastBlock < br.Idx {
			l.lastBlock = br.Idx
		}
		if l.lastLlrTime < br.Time {
			l.lastLlrTime = br.Time
		}
		now := time.Now()
		// logging for the individual item
		msg := "New BR"
		logType := l.Log.Debug
		logType(msg, "block", br.Idx,
			"age", utils.PrettyDuration(now.Sub(br.Time.Time())),
			"t", utils.PrettyDuration(now.Sub(start)))
		l.summary(now)
	}
}

// EpochVoteConnectionStarted starts the EV logging
// Not safe for concurrent use
func (l *Logger) EpochVoteConnectionStarted(ev inter.LlrSignedEpochVote) func() {
	if ev.Val.Epoch == 0 {
		return func() {}
	}
	l.llrSum.evs++

	start := time.Now()

	return func() {
		if l.lastEpoch < ev.Val.Epoch {
			l.lastEpoch = ev.Val.Epoch
		}
		now := time.Now()
		// logging for the individual item
		msg := "New EV"
		logType := l.Log.Debug
		if l.emitting {
			msg = "New EV emitted"
			logType = l.Log.Info
		}
		logType(msg, "id", ev.Signed.Locator.ID(), "by", ev.Signed.Locator.Creator,
			"epoch", ev.Val.Epoch,
			"t", utils.PrettyDuration(now.Sub(start)))
		l.summary(now)
	}
}

// EpochRecordConnectionStarted starts the ER logging
// Not safe for concurrent use
func (l *Logger) EpochRecordConnectionStarted(er ier.LlrIdxFullEpochRecord) func() {
	l.llrSum.ers++

	start := time.Now()

	return func() {
		if l.lastEpoch < er.Idx {
			l.lastEpoch = er.Idx
		}
		if l.lastLlrTime < er.EpochState.EpochStart {
			l.lastLlrTime = er.EpochState.EpochStart
		}
		now := time.Now()
		// logging for the individual item
		msg := "New ER"
		logType := l.Log.Debug
		logType(msg, "epoch", er.Idx,
			"age", utils.PrettyDuration(now.Sub(er.EpochState.EpochStart.Time())),
			"t", utils.PrettyDuration(now.Sub(start)))
		l.summary(now)
	}
}
