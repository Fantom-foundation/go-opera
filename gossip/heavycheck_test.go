package gossip

import (
	"bytes"
	"math"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/Fantom-foundation/go-opera/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-opera/eventcheck/heavycheck"
	"github.com/Fantom-foundation/go-opera/inter"
)

type LLRHeavyCheckTestSuite struct {
	suite.Suite

	env        *testEnv
	me         *inter.MutableEventPayload
	startEpoch idx.Epoch
}

func (s *LLRHeavyCheckTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	const (
		validatorsNum = 10
		startEpoch    = 1
	)

	env := newTestEnv(startEpoch, validatorsNum)

	em := env.emitters[0]
	e, err := em.EmitEvent()
	s.Require().NoError(err)
	s.Require().NotNil(e)

	s.env = env
	s.me = mutableEventPayloadFromImmutable(e)
	s.startEpoch = idx.Epoch(startEpoch)
}

func (s *LLRHeavyCheckTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.env.Close()
}

func (s *LLRHeavyCheckTestSuite) TestHeavyCheckValidateEV() {

	var ev inter.LlrSignedEpochVote

	testCases := []struct {
		name    string
		errExp  error
		pretest func()
	}{
		{
			"validateEV returns nil",
			nil,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"validateEV returns ErrUnknownEpochEV",
			heavycheck.ErrUnknownEpochEV,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"epochcheck.ErrAuth",
			epochcheck.ErrAuth,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(100)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"ErrWrongPayloadHash",
			heavycheck.ErrWrongPayloadHash,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(hash.Hash{})

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				ev = inter.AsSignedEpochVote(s.me)
			},
		},

		{
			"ErrWrongEventSig",
			heavycheck.ErrWrongEventSig,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(4)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				ev = inter.AsSignedEpochVote(s.me)
				ev.Signed.Locator.Creator = 5
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupSuite()
			tc.pretest()

			err := s.env.checkers.Heavycheck.ValidateEV(ev)

			if tc.errExp != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.errExp.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}

}

func (s *LLRHeavyCheckTestSuite) TestHeavyCheckValidateBVs() {
	var bv inter.LlrSignedBlockVotes

	testCases := []struct {
		name    string
		errExp  error
		pretest func()
	}{
		{
			"success",
			nil,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetCreator(2)

				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[1], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"ErrUnknownEpochBVs",
			heavycheck.ErrUnknownEpochBVs,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: 25,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(2)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[1], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"ErrImpossibleBVsEpoch",
			heavycheck.ErrImpossibleBVsEpoch,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 0,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(2)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[1], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"ErrUnknownEpochBVs",
			heavycheck.ErrUnknownEpochBVs,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: 0,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}
				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"epochcheck.ErrAuth",
			epochcheck.ErrAuth,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				invalidValidatorID := idx.ValidatorID(100)

				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(invalidValidatorID)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"ErrWrongPayloadHash",
			heavycheck.ErrWrongPayloadHash,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}
				emptyPayload := hash.Hash{}

				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(emptyPayload)

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"ErrWrongEventSig",
			heavycheck.ErrWrongEventSig,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetVersion(1)
				s.me.SetBlockVotes(bv.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(4)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				bv = inter.AsSignedBlockVotes(s.me)
				bv.Signed.Locator.Creator = 5
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupSuite()
			tc.pretest()

			err := s.env.checkers.Heavycheck.ValidateBVs(bv)

			if tc.errExp != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.errExp.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func mutableEventPayloadFromImmutable(e *inter.EventPayload) *inter.MutableEventPayload {
	me := &inter.MutableEventPayload{}
	me.SetVersion(e.Version())
	me.SetNetForkID(e.NetForkID())
	me.SetCreator(e.Creator())
	me.SetEpoch(e.Epoch())
	me.SetCreationTime(e.CreationTime())
	me.SetMedianTime(e.MedianTime())
	me.SetPrevEpochHash(e.PrevEpochHash())
	me.SetExtra(e.Extra())
	me.SetGasPowerLeft(e.GasPowerLeft())
	me.SetGasPowerUsed(e.GasPowerUsed())
	me.SetPayloadHash(e.PayloadHash())
	me.SetSig(e.Sig())
	me.SetTxs(e.Txs())
	me.SetMisbehaviourProofs(e.MisbehaviourProofs())
	me.SetBlockVotes(e.BlockVotes())
	me.SetEpochVote(e.EpochVote())
	return me
}

func (s *LLRHeavyCheckTestSuite) TestHeavyCheckValidateEvent() {

	testCases := []struct {
		name    string
		errExp  error
		pretest func()
	}{
		{
			"success",
			nil,
			func() {
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetCreator(3)
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"epochcheck.ErrNotRelevant",
			epochcheck.ErrNotRelevant,
			func() {
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch + 1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"epochcheck.ErrAuth",
			epochcheck.ErrAuth,
			func() {
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				invalidCreator := idx.ValidatorID(100)
				s.me.SetCreator(invalidCreator)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"ErrWrongEventSig",
			heavycheck.ErrWrongEventSig,
			func() {
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetCreator(3)
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[1], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"ErrMalformedTxSig",
			heavycheck.ErrMalformedTxSig,
			func() {
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetCreator(3)
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: h.Big(),
					Gas:      math.MaxUint64,
					To:       nil,
					Value:    h.Big(),
					Data:     []byte{},
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				s.me.SetTxs(txs)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"ErrWrongPayloadHash",
			heavycheck.ErrWrongPayloadHash,
			func() {
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)

				invalidPayloadHash := hash.Hash{}
				s.me.SetPayloadHash(invalidPayloadHash)

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"EpochVote().Epoch == 0",
			nil,
			func() {
				ev := inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: 0,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetEpochVote(ev.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"EpochVote().Epoch != 0, matchPubkey returns heavycheck.ErrUnknownEpochEV",
			heavycheck.ErrUnknownEpochEV,
			func() {
				ev := inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetEpochVote(ev.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"EpochVote().Epoch != 0, matchPubkey returns epochcheck.ErrAuth",
			epochcheck.ErrAuth,
			func() {
				ev := inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetEpochVote(ev.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				invalidCreator := idx.ValidatorID(100)
				s.me.SetCreator(invalidCreator)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"EpochVote().Epoch != 0, matchPubkey returns nil",
			nil,
			func() {
				ev := inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}

				s.me.SetEpochVote(ev.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				ev = inter.AsSignedEpochVote(s.me)
			},
		},
		{
			"BlockVote().Epoch == 0",
			nil,
			func() {
				bv := inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: 0,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetBlockVotes(bv.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"BlockVote().Epoch != 0, validateBVsEpoch returns nil",
			nil,
			func() {
				bv := inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetBlockVotes(bv.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
			},
		},
		{
			"blockvote epoch is 0, block vote epoch does not match event epoch,matchPubkey returns nil",
			nil,
			func() {
				bv := inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}

				s.me.SetBlockVotes(bv.Val)
				s.me.SetVersion(1)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetSeq(idx.Event(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))
				s.me.SetCreator(3)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[2], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)
				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupSuite()
			tc.pretest()

			err := s.env.checkers.Heavycheck.ValidateEvent(s.me)

			if tc.errExp != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.errExp.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestLLRHeavyCheckTestSuite(t *testing.T) {
	t.Skip() // skip until fixed
	suite.Run(t, new(LLRHeavyCheckTestSuite))
}
