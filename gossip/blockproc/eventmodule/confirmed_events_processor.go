package eventmodule

import (
	"github.com/Fantom-foundation/go-opera/gossip/blockproc"
	"github.com/Fantom-foundation/go-opera/inter"
)

type ValidatorEventsModule struct{}

func New() *ValidatorEventsModule {
	return &ValidatorEventsModule{}
}

func (m *ValidatorEventsModule) Start(bs blockproc.BlockState, es blockproc.EpochState) blockproc.ConfirmedEventsProcessor {
	return &ValidatorEventsProcessor{
		es:                     es,
		bs:                     bs,
		validatorHighestEvents: make(inter.EventIs, es.Validators.Len()),
	}
}

type ValidatorEventsProcessor struct {
	es                     blockproc.EpochState
	bs                     blockproc.BlockState
	validatorHighestEvents inter.EventIs
}

func (p *ValidatorEventsProcessor) ProcessConfirmedEvent(e inter.EventI) {
	creatorIdx := p.es.Validators.GetIdx(e.Creator())
	prev := p.validatorHighestEvents[creatorIdx]
	if prev == nil || e.Seq() > prev.Seq() {
		p.validatorHighestEvents[creatorIdx] = e
	}
	p.bs.EpochGas += e.GasPowerUsed()
}

func (p *ValidatorEventsProcessor) Finalize(block blockproc.BlockCtx) blockproc.BlockState {
	for _, v := range block.CBlock.Cheaters {
		creatorIdx := p.es.Validators.GetIdx(v)
		p.validatorHighestEvents[creatorIdx] = nil
	}
	for creatorIdx, e := range p.validatorHighestEvents {
		if e == nil {
			continue
		}
		info := p.bs.ValidatorStates[creatorIdx]
		if block.Idx <= info.LastBlock+p.es.Rules.Economy.BlockMissedSlack {
			if e.MedianTime() > info.LastOnlineTime {
				info.Uptime += e.MedianTime() - info.LastOnlineTime
			}
		}
		info.LastGasPowerLeft = e.GasPowerLeft()
		info.LastOnlineTime = e.MedianTime()
		info.LastBlock = p.bs.LastBlock
		info.LastEvent = e.ID()
		p.bs.ValidatorStates[creatorIdx] = info
	}
	return p.bs
}
