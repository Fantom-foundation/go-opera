package dagstreamseeder

import (
	"errors"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream"
	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamseeder"
	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/gossip/protocols/dag/dagstream"
)

var (
	ErrWrongType        = errors.New("wrong request type")
	ErrWrongSelectorLen = errors.New("wrong event selector length")
)

type Seeder struct {
	*basestreamseeder.BaseSeeder
}

type Callbacks struct {
	ForEachEvent func(start []byte, onEvent func(key hash.Event, eventB rlp.RawValue) bool)
}

type Peer struct {
	ID           string
	SendChunk    func(dagstream.Response, hash.Events) error
	Misbehaviour func(error)
}

func New(cfg Config, callbacks Callbacks) *Seeder {
	return &Seeder{
		BaseSeeder: basestreamseeder.New(basestreamseeder.Config(cfg), basestreamseeder.Callbacks{
			ForEachItem: func(start basestream.Locator, rType basestream.RequestType, onKey func(basestream.Locator) bool, onAppended func(basestream.Payload) bool) basestream.Payload {
				res := &dagstream.Payload{
					IDs:    hash.Events{},
					Events: []rlp.RawValue{},
					Size:   0,
				}
				callbacks.ForEachEvent(start.(dagstream.Locator), func(key hash.Event, eventB rlp.RawValue) bool {
					if !onKey(dagstream.Locator(key.Bytes())) {
						return false
					}
					if rType == dagstream.RequestIDs {
						res.AddID(key, len(eventB))
					} else {
						res.AddEvent(key, eventB)
					}
					return onAppended(res)
				})
				return res
			},
		}),
	}
}

func (s *Seeder) NotifyRequestReceived(peer Peer, r dagstream.Request) (err error, peerErr error) {
	if len(r.Session.Start) > len(hash.ZeroEvent) || len(r.Session.Stop) > len(hash.ZeroEvent) {
		return nil, ErrWrongSelectorLen
	}
	if r.Type != dagstream.RequestIDs && r.Type != dagstream.RequestEvents {
		return nil, ErrWrongType
	}
	rType := r.Type
	return s.BaseSeeder.NotifyRequestReceived(basestreamseeder.Peer{
		ID: peer.ID,
		SendChunk: func(response basestream.Response) error {
			payload := response.Payload.(*dagstream.Payload)
			payloadIDs := payload.IDs
			if rType == dagstream.RequestEvents {
				payloadIDs = payloadIDs[:0]
			}
			return peer.SendChunk(dagstream.Response{
				SessionID: response.SessionID,
				Done:      response.Done,
				IDs:       payloadIDs,
				Events:    payload.Events,
			}, payload.IDs)
		},
		Misbehaviour: peer.Misbehaviour,
	}, basestream.Request{
		Session: basestream.Session{
			ID:    r.Session.ID,
			Start: r.Session.Start,
			Stop:  r.Session.Stop,
		},
		Type:           r.Type,
		MaxPayloadNum:  uint32(r.Limit.Num),
		MaxPayloadSize: r.Limit.Size,
		MaxChunks:      r.MaxChunks,
	})
}
