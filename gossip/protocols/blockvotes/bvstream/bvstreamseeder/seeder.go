package bvstreamseeder

import (
	"errors"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream"
	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamseeder"
	"github.com/Fantom-foundation/lachesis-base/hash"

	"github.com/Fantom-foundation/go-opera/gossip/protocols/blockvotes/bvstream"
)

var (
	ErrWrongType        = errors.New("wrong request type")
	ErrWrongSelectorLen = errors.New("wrong event selector length")
)

type Seeder struct {
	*basestreamseeder.BaseSeeder
}

type Callbacks struct {
	Iterate func(locator []byte, f func(key []byte, bvs rlp.RawValue) bool)
}

type Peer struct {
	ID           string
	SendChunk    func(bvstream.Response) error
	Misbehaviour func(error)
}

func New(cfg Config, callbacks Callbacks) *Seeder {
	return &Seeder{
		BaseSeeder: basestreamseeder.New(basestreamseeder.Config(cfg), basestreamseeder.Callbacks{
			ForEachItem: func(start basestream.Locator, _ basestream.RequestType, onKey func(key basestream.Locator) bool, onAppended func(items basestream.Payload) bool) basestream.Payload {
				res := &bvstream.Payload{
					Items: []rlp.RawValue{},
					Keys:  []bvstream.Locator{},
					Size:  0,
				}
				st := start.(bvstream.Locator)
				callbacks.Iterate(st, func(bkey []byte, bvs rlp.RawValue) bool {
					key := bvstream.Locator(bkey)
					if !onKey(key) {
						return false
					}
					res.AddSignedBlockVotes(key, bvs)
					return onAppended(res)
				})
				return res
			},
		}),
	}
}

func (s *Seeder) NotifyRequestReceived(peer Peer, r bvstream.Request) (err error, peerErr error) {
	if len(r.Session.Start) > len(hash.ZeroEvent)+4+8 || len(r.Session.Stop) > len(hash.ZeroEvent)+4+8 {
		return nil, ErrWrongSelectorLen
	}
	if r.Type != 0 {
		return nil, ErrWrongType
	}
	return s.BaseSeeder.NotifyRequestReceived(basestreamseeder.Peer{
		ID: peer.ID,
		SendChunk: func(response basestream.Response) error {
			return peer.SendChunk(bvstream.Response{
				SessionID: response.SessionID,
				Done:      response.Done,
				Payload:   response.Payload.(*bvstream.Payload).Items,
			})
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
