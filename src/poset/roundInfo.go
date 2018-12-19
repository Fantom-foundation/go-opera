package poset

import (
	"github.com/golang/protobuf/proto"
)

type pendingRound struct {
	Index   int64
	Decided bool
}

// protobuf message wrapper for created event messages
type RoundCreated struct {
	Message RoundCreatedMessage
}

// RoundCreated constructor
func NewRoundCreated() *RoundCreated {
	return &RoundCreated{
		Message: RoundCreatedMessage{
			Events: make(map[string]*RoundEvent),
		},
	}
}

// RoundReceived constructor
func NewRoundReceived() *RoundReceived {
	return &RoundReceived{
		Rounds: []string{},
	}
}

// add event to created events message
func (r *RoundCreated) AddEvent(x string, clotho bool) {
	_, ok := r.Message.Events[x]
	if !ok {
		r.Message.Events[x] = &RoundEvent{
			Clotho: clotho,
		}
	}
}

// mark the given event as a consensus event
func (r *RoundCreated) SetConsensusEvent(x string) {
	e, ok := r.Message.Events[x]
	if !ok {
		e = &RoundEvent{}
	}
	e.Consensus = true
	r.Message.Events[x] = e
}

// set the received round for the given event
func (r *RoundCreated) SetRoundReceived(x string, round int64) {
	e, ok := r.Message.Events[x]

	if !ok {
		return
	}

	e.RoundReceived = round

	r.Message.Events[x] = e
}

// set whether the given event is Atropos, otherwise it is Clotho when not found
func (r *RoundCreated) SetAtropos(x string, f bool) {
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

// return true if no clothos' fame is left undefined
func (r *RoundCreated) ClothoDecided() bool {
	for _, e := range r.Message.Events {
		if e.Clotho && e.Atropos == Trilean_UNDEFINED {
			return false
		}
	}
	return true
}

// return clothos
func (r *RoundCreated) Clotho() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Clotho {
			res = append(res, x)
		}
	}
	return res
}

// return all non-consensus events
func (r *RoundCreated) RoundEvents() []string {
	var res []string
	for x, e := range r.Message.Events {
		if !e.Consensus {
			res = append(res, x)
		}
	}
	return res
}

// return consensus events
func (r *RoundCreated) ConsensusEvents() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Consensus {
			res = append(res, x)
		}
	}
	return res
}

// return Atropos
func (r *RoundCreated) Atropos() []string {
	var res []string
	for x, e := range r.Message.Events {
		if e.Clotho && e.Atropos == Trilean_TRUE {
			res = append(res, x)
		}
	}
	return res
}

// return whether the given even is Atropos or Clotho
func (r *RoundCreated) IsDecided(clotho string) bool {
	w, ok := r.Message.Events[clotho]
	return ok && w.Clotho && w.Atropos != Trilean_UNDEFINED
}

// serialise RoundCreated using protobuf
func (r *RoundCreated) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(&r.Message); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// serialise RoundReceived using protobuf
func (r *RoundReceived) ProtoMarshal() ([]byte, error) {
	var bf proto.Buffer
	bf.SetDeterministic(true)
	if err := bf.Marshal(r); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

// de-serialise RoundCreated using protobuf
func (r *RoundCreated) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, &r.Message)
}

// de-serialise RoundReceived using protobuf
func (r *RoundReceived) ProtoUnmarshal(data []byte) error {
	return proto.Unmarshal(data, r)
}

// is the RoundCreated queued in PendingRounds
func (r *RoundCreated) IsQueued() bool {
	return r.Message.Queued
}

// are the round events the same
func (this *RoundEvent) Equals(that *RoundEvent) bool {
	return this.Consensus == that.Consensus &&
		this.Clotho == that.Clotho &&
		this.Atropos == that.Atropos &&
		this.RoundReceived == that.RoundReceived
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

func (this *RoundCreated) Equals(that *RoundCreated) bool {
	return this.Message.Queued == that.Message.Queued &&
		EqualsMapStringRoundEvent(this.Message.Events, that.Message.Events)
}
