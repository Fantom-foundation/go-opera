package poset

import (
	"github.com/golang/protobuf/proto"
)

type pendingRound struct {
	Index   int64
	Decided bool
}

type RoundInfo struct {
	Message RoundInfoMessage
}

func NewRoundInfo() *RoundInfo {
	return &RoundInfo{
		Message: RoundInfoMessage{
			Events: make(map[string]*RoundEvent),
		},
	}
}

func (r *RoundInfo) AddEvent(x string, clotho bool) {
	_, ok := r.Message.Events[x]
	if !ok {
		r.Message.Events[x] = &RoundEvent{
			Clotho: clotho,
		}
	}
}

func (r *RoundInfo) SetConsensusEvent(x string) {
	e, ok := r.Message.Events[x]
	if !ok {
		e = &RoundEvent{}
	}
	e.Consensus = true
	r.Message.Events[x] = e
}

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

//return true if no clothos' fame is left undefined
func (r *RoundInfo) ClothoDecided() bool {
	for _, e := range r.Message.Events {
		if e.Clotho && e.Atropos == Trilean_UNDEFINED {
			return false
		}
	}
	return true
}

//return clothos
func (r *RoundInfo) Clotho() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Clotho {
			res = append(res, x)
		}
	}
	return res
}

func (r *RoundInfo) RoundEvents() []string {
	var res []string
	for x, e := range r.Message.Events {
		if !e.Consensus {
			res = append(res, x)
		}
	}
	return res
}

//return consensus events
func (r *RoundInfo) ConsensusEvents() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Consensus {
			res = append(res, x)
		}
	}
	return res
}

//return Atropos
func (r *RoundInfo) Atropos() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Clotho && e.Atropos == Trilean_TRUE {
			res = append(res, x)
		}
	}
	return res
}

func (r *RoundInfo) IsDecided(clotho string) bool {
	w, ok := r.Message.Events[clotho]
	return ok && w.Clotho && w.Atropos != Trilean_UNDEFINED
}

func (r *RoundInfo) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(&r.Message); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (r *RoundInfo) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, &r.Message)
}

func (r *RoundInfo) IsQueued() bool {
	return r.Message.Queued
}

func (this *RoundEvent) Equals(that *RoundEvent) bool {
	return this.Consensus == that.Consensus &&
		this.Clotho == that.Clotho &&
		this.Atropos == that.Atropos
}

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

func (this *RoundInfo) Equals(that *RoundInfo) bool {
	return this.Message.Queued == that.Message.Queued &&
		EqualsMapStringRoundEvent(this.Message.Events, that.Message.Events)
}
