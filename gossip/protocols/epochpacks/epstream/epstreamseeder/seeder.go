package epstreamseeder

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/gossip/basestream"
	"github.com/Fantom-foundation/lachesis-base/gossip/basestream/basestreamseeder"

	"github.com/cyberbono3/go-opera/gossip/protocols/epochpacks/epstream"
)

var (
	ErrWrongType = errors.New("wrong request type")
)

type Seeder struct {
	*basestreamseeder.BaseSeeder
}

type Callbacks struct {
	Iterate func(start idx.Epoch, f func(epoch idx.Epoch, eps rlp.RawValue) bool)
}

type Peer struct {
	ID           string
	SendChunk    func(epstream.Response) error
	Misbehaviour func(error)
}

func New(cfg Config, callbacks Callbacks) *Seeder {
	return &Seeder{
		BaseSeeder: basestreamseeder.New(basestreamseeder.Config(cfg), basestreamseeder.Callbacks{
			ForEachItem: func(start basestream.Locator, _ basestream.RequestType, onKey func(key basestream.Locator) bool, onAppended func(items basestream.Payload) bool) basestream.Payload {
				res := &epstream.Payload{
					Items: []rlp.RawValue{},
					Keys:  []epstream.Locator{},
					Size:  0,
				}
				st := start.(epstream.Locator)
				callbacks.Iterate(idx.Epoch(st), func(epoch idx.Epoch, eps rlp.RawValue) bool {
					key := epstream.Locator(epoch)
					if !onKey(key) {
						return false
					}
					res.AddEpochPacks(key, eps)
					return onAppended(res)
				})
				return res
			},
		}),
	}
}

func (s *Seeder) NotifyRequestReceived(peer Peer, r epstream.Request) (err error, peerErr error) {
	if r.Type != 0 {
		return nil, ErrWrongType
	}
	return s.BaseSeeder.NotifyRequestReceived(basestreamseeder.Peer{
		ID: peer.ID,
		SendChunk: func(response basestream.Response) error {
			return peer.SendChunk(epstream.Response{
				SessionID: response.SessionID,
				Done:      response.Done,
				Payload:   response.Payload.(*epstream.Payload).Items,
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
