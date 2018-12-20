package poset

import (
	"github.com/golang/protobuf/proto"
)

type pendingRound struct {
	Index   int64
	Decided bool
}

// RoundInfo info for the round
type RoundInfo struct {
	Message RoundInfoMessage
}

// NewRoundInfo creates a new round info struct
func NewRoundInfo() *RoundInfo {
	return &RoundInfo{
		Message: RoundInfoMessage{
			Events: make(map[string]*RoundEvent),
		},
	}
}

// AddEvent add event to round info (optionally set clotho)
func (r *RoundInfo) AddEvent(x string, clotho bool) {
	_, ok := r.Message.Events[x]
	if !ok {
		r.Message.Events[x] = &RoundEvent{
			Clotho: clotho,
		}
	}
}

// SetConsensusEvent set an event as an consensus event
func (r *RoundInfo) SetConsensusEvent(x string) {
	e, ok := r.Message.Events[x]
	if !ok {
		e = &RoundEvent{}
	}
	e.Consensus = true
	r.Message.Events[x] = e
}

// SetAtropos sets an event as san atropos
func (r *RoundInfo) SetAtropos(x string, f bool) {
	e, ok := r.Message.Events[x]
	if !ok {
		e = &RoundEvent{
			Clotho: true,
		}
	}
	if f {
		e.Atropos = Trilean_TRUE
	} else {
		e.Atropos = Trilean_FALSE
	}
	r.Message.Events[x] = e
}

// ClothoDecided return true if no clothos' fame is left undefined
func (r *RoundInfo) ClothoDecided() bool {
	for _, e := range r.Message.Events {
		if e.Clotho && e.Atropos == Trilean_UNDEFINED {
			return false
		}
	}
	return true
}

// Clotho return clothos
func (r *RoundInfo) Clotho() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Clotho {
			res = append(res, x)
		}
	}
	return res
}

// RoundEvents returns the round events
func (r *RoundInfo) RoundEvents() []string {
	var res []string
	for x, e := range r.Message.Events {
		if !e.Consensus {
			res = append(res, x)
		}
	}
	return res
}

// ConsensusEvents return consensus events
func (r *RoundInfo) ConsensusEvents() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Consensus {
			res = append(res, x)
		}
	}
	return res
}

// Atropos return Atropos
func (r *RoundInfo) Atropos() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Clotho && e.Atropos == Trilean_TRUE {
			res = append(res, x)
		}
	}
	return res
}

// IsDecided checks if the event is a decided clotho
func (r *RoundInfo) IsDecided(clotho string) bool {
	w, ok := r.Message.Events[clotho]
	return ok && w.Clotho && w.Atropos != Trilean_UNDEFINED
}

// ProtoMarshal marshals the round info to protobuff
func (r *RoundInfo) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(&r.Message); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// ProtoUnmarshal unmarshals protobuff to a round info
func (r *RoundInfo) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, &r.Message)
}

// IsQueued checks if the round is queued for processing
func (r *RoundInfo) IsQueued() bool {
	return r.Message.Queued
}

// Equals compares round events for equality
func (re *RoundEvent) Equals(that *RoundEvent) bool {
	return re.Consensus == that.Consensus &&
		re.Clotho == that.Clotho &&
		re.Atropos == that.Atropos
}

// EqualsMapStringRoundEvent compares a map string of round eventss for equality
func EqualsMapStringRoundEvent(this map[string]*RoundEvent, that map[string]*RoundEvent) bool {
	if len(this) != len(that) {
		return false
	}
	for k, v := range this {
		v2, ok := that[k]
		if !ok || !v2.Equals(v) {
			return false
		}
	}
	return true
}

// Equals compares two round info structs for equality
func (r *RoundInfo) Equals(that *RoundInfo) bool {
	return r.Message.Queued == that.Message.Queued &&
		EqualsMapStringRoundEvent(r.Message.Events, that.Message.Events)
}
