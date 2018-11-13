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
	queued bool
}

func NewRoundInfo() *RoundInfo {
	return &RoundInfo{
		Message: RoundInfoMessage{
			Events: make(map[string]*RoundEvent),
		},
	}
}

func (r *RoundInfo) AddEvent(x string, witness bool) {
	_, ok := r.Message.Events[x]
	if !ok {
		r.Message.Events[x] = &RoundEvent{
			Witness: witness,
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

func (r *RoundInfo) SetFame(x string, f bool) {
	e, ok := r.Message.Events[x]
	if !ok {
		e = &RoundEvent{
			Witness: true,
		}
	}
	if f {
		e.Famous = Trilean_TRUE
	} else {
		e.Famous = Trilean_FALSE
	}
	r.Message.Events[x] = e
}

//return true if no witnesses' fame is left undefined
func (r *RoundInfo) WitnessesDecided() bool {
	for _, e := range r.Message.Events {
		if e.Witness && e.Famous == Trilean_UNDEFINED {
			return false
		}
	}
	return true
}

//return witnesses
func (r *RoundInfo) Witnesses() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Witness {
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

//return famous witnesses
func (r *RoundInfo) FamousWitnesses() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Witness && e.Famous == Trilean_TRUE {
			res = append(res, x)
		}
	}
	return res
}

func (r *RoundInfo) IsDecided(witness string) bool {
	w, ok := r.Message.Events[witness]
	return ok && w.Witness && w.Famous != Trilean_UNDEFINED
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
	return r.queued
}

func (this *RoundEvent) Equals(that *RoundEvent) bool {
	return this.Consensus == that.Consensus &&
		this.Witness == that.Witness &&
		this.Famous == that.Famous
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
	return this.queued == that.queued &&
		EqualsMapStringRoundEvent(this.Message.Events, that.Message.Events)
}
