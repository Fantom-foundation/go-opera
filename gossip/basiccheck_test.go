package gossip

import (
	"bytes"
	"math"
	"math/big"
	"testing"

	lbasiccheck "github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	"github.com/Fantom-foundation/go-opera/inter"
)

type LLRBasicCheckTestSuite struct {
	suite.Suite

	env        *testEnv
	me         *inter.MutableEventPayload
	startEpoch idx.Epoch
}

func (s *LLRBasicCheckTestSuite) SetupSuite() {
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

func (s *LLRBasicCheckTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.env.Close()
}

func (s *LLRBasicCheckTestSuite) TestBasicCheckValidate() {

	testCases := []struct {
		name    string
		pretest func()
		errExp  error
	}{

		{
			"ErrWrongNetForkID",
			func() {
				s.me.SetNetForkID(1)
			},
			basiccheck.ErrWrongNetForkID,
		},

		{
			"Validate checkLimits ErrHugeValue",
			func() {
				s.me.SetEpoch(math.MaxInt32 - 1)
			},
			lbasiccheck.ErrHugeValue,
		},
		{
			"Validate checkInited checkInited ErrNotInited ",
			func() {
				s.me.SetSeq(0)
			},
			lbasiccheck.ErrNotInited,
		},
		{
			"Validate checkInited ErrNoParents",
			func() {
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				s.me.SetSeq(idx.Event(2))
				parents := hash.Events{}
				s.me.SetParents(parents)
			},
			lbasiccheck.ErrNoParents,
		},
		{
			"Validate ErrHugeValue-1",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				s.me.SetGasPowerUsed(math.MaxInt64 - 1)
			},
			lbasiccheck.ErrHugeValue,
		},
		{
			"Validate ErrHugeValue-2",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				s.me.SetGasPowerLeft(inter.GasPowerLeft{Gas: [2]uint64{math.MaxInt64 - 1, math.MaxInt64}})
			},
			lbasiccheck.ErrHugeValue,
		},
		{
			"Validate ErrZeroTime-1",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				s.me.SetCreationTime(0)
			},
			basiccheck.ErrZeroTime,
		},
		{
			"Validate ErrZeroTime-2",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				s.me.SetMedianTime(0)
			},
			basiccheck.ErrZeroTime,
		},
		{
			"Validate checkTxs validateTx ErrNegativeValue-1",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: h.Big(),
					Gas:      math.MaxUint64,
					To:       nil,
					Value:    big.NewInt(-1000),
					Data:     []byte{},
					V:        big.NewInt(0xff),
					R:        h.Big(),
					S:        h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				s.me.SetTxs(txs)
			},
			basiccheck.ErrNegativeValue,
		},
		{
			"Validate checkTxs validateTx ErrNegativeValue-2",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: big.NewInt(-1000),
					Gas:      math.MaxUint64,
					To:       nil,
					Value:    h.Big(),
					Data:     []byte{},
					V:        big.NewInt(0xff),
					R:        h.Big(),
					S:        h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				s.me.SetTxs(txs)
			},
			basiccheck.ErrNegativeValue,
		},
		{
			"Validate checkTxs validateTx ErrIntrinsicGas",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))
				tx1 := types.NewTx(&types.LegacyTx{
					Nonce:    math.MaxUint64,
					GasPrice: h.Big(),
					Gas:      0,
					To:       nil,
					Value:    h.Big(),
					Data:     []byte{},
					V:        big.NewInt(0xff),
					R:        h.Big(),
					S:        h.Big(),
				})
				txs := types.Transactions{}
				txs = append(txs, tx1)
				s.me.SetTxs(txs)
			},
			basiccheck.ErrIntrinsicGas,
		},

		{
			"Validate checkTxs validateTx ErrTipAboveFeeCap",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))

				tx1 := types.NewTx(&types.DynamicFeeTx{
					Nonce:     math.MaxUint64,
					To:        nil,
					Data:      []byte{},
					Gas:       math.MaxUint64,
					Value:     h.Big(),
					ChainID:   new(big.Int),
					GasTipCap: big.NewInt(1000),
					GasFeeCap: new(big.Int),
					V:         big.NewInt(0xff),
					R:         h.Big(),
					S:         h.Big(),
				})

				txs := types.Transactions{}
				txs = append(txs, tx1)
				s.me.SetTxs(txs)
			},
			basiccheck.ErrTipAboveFeeCap,
		},
		{
			"Validate returns nil",
			func() {
				s.me.SetSeq(idx.Event(1))
				s.me.SetEpoch(idx.Epoch(1))
				s.me.SetFrame(idx.Frame(1))
				s.me.SetLamport(idx.Lamport(1))

				h := hash.BytesToEvent(bytes.Repeat([]byte{math.MaxUint8}, 32))

				tx1 := types.NewTx(&types.DynamicFeeTx{
					Nonce:     math.MaxUint64,
					To:        nil,
					Data:      []byte{},
					Gas:       math.MaxUint64,
					Value:     h.Big(),
					ChainID:   new(big.Int),
					GasTipCap: new(big.Int),
					GasFeeCap: big.NewInt(1000),
					V:         big.NewInt(0xff),
					R:         h.Big(),
					S:         h.Big(),
				})

				txs := types.Transactions{}
				txs = append(txs, tx1)
				s.me.SetTxs(txs)
			},
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupSuite()
			tc.pretest()

			err := s.env.checkers.Basiccheck.Validate(s.me)

			if tc.errExp != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.errExp.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *LLRBasicCheckTestSuite) TestBasicCheckValidateEV() {

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
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
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
			"validateEventLocator ErrWrongNetForkID",
			basiccheck.ErrWrongNetForkID,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetNetForkID(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
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
			"validateEventLocator ErrHugeValue-1 e.Seq >= math.MaxInt32-1 ",
			lbasiccheck.ErrHugeValue,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetSeq(idx.Event(math.MaxInt32 - 1))
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
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
			"validateEventLocator ErrHugeValue-2 e.Epoch >= math.MaxInt32-1",
			lbasiccheck.ErrHugeValue,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(math.MaxInt32 - 1))
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
			"validateEventLocator ErrHugeValue-3 e.Lamport >= math.MaxInt32-1",
			lbasiccheck.ErrHugeValue,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetLamport(idx.Lamport(math.MaxInt32 - 1))
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(math.MaxInt32 - 1))
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
			"validateEV ErrFutureEVEpoch",
			basiccheck.FutureEVEpoch,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch + 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(s.startEpoch)
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
			"validateEV MalformedEV-1",
			basiccheck.MalformedEV,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: 0,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(s.startEpoch)
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
			"validateEV MalformedEV-2",
			basiccheck.MalformedEV,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: s.startEpoch,
						Vote:  hash.Zero,
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(s.startEpoch)
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
			"validateEV ErrHugeValue",
			lbasiccheck.ErrHugeValue,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: math.MaxInt32 - 1,
						Vote:  hash.HexToHash("0x01"),
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(math.MaxInt32 - 1))
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
			"validateEV EmptyEV",
			basiccheck.EmptyEV,
			func() {
				ev = inter.LlrSignedEpochVote{
					Val: inter.LlrEpochVote{
						Epoch: 0,
						Vote:  hash.Zero,
					},
				}
				s.me.SetVersion(1)
				s.me.SetEpochVote(ev.Val)
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
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
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupSuite()
			tc.pretest()

			err := s.env.checkers.Basiccheck.ValidateEV(ev)

			if tc.errExp != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.errExp.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}

}

func (s *LLRBasicCheckTestSuite) TestBasicCheckValidateBV() {
	var bv inter.LlrSignedBlockVotes

	testCases := []struct {
		name    string
		errExp  error
		pretest func()
	}{
		{
			"validateBV returns nil",
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
				s.me.SetBlockVotes(bv.Val)
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
			"validateBV ErrWrongNetForkID",
			basiccheck.ErrWrongNetForkID,
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
				s.me.SetBlockVotes(bv.Val)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))
				s.me.SetNetForkID(1)
				sig, err := s.env.signer.Sign(s.env.pubkeys[1], s.me.HashToSign().Bytes())
				s.Require().NoError(err)
				sSig := inter.Signature{}
				copy(sSig[:], sig)
				s.me.SetSig(sSig)

				bv = inter.AsSignedBlockVotes(s.me)
			},
		},
		{
			"validateBVs ErrHugeValue e.Seq >= math.MaxInt32-1",
			lbasiccheck.ErrHugeValue,
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

				s.me.SetSeq(idx.Event(math.MaxInt32 - 1))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs ErrHugeValue e.Epoch >= math.MaxInt32-1",
			lbasiccheck.ErrHugeValue,
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

				s.me.SetEpoch(idx.Epoch(math.MaxInt32 - 1))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs ErrHugeValue e.Lamport >= math.MaxInt32-1",
			lbasiccheck.ErrHugeValue,
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

				s.me.SetLamport(idx.Lamport(math.MaxInt32 - 1))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs FutureBVsEpoc",
			basiccheck.FutureBVsEpoch,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: s.startEpoch + 1,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs ErrHugeValue-1",
			basiccheck.FutureBVsEpoch,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: math.MaxInt64 / 2,
						Epoch: s.startEpoch + 1,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs ErrHugeValue-2",
			basiccheck.FutureBVsEpoch,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: math.MaxInt32 - 1,
						Votes: []hash.Hash{
							hash.Zero,
							hash.HexToHash("0x01"),
						},
					},
				}
				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs TooManyBVs",
			basiccheck.TooManyBVs,
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

				for j := 0; j < basiccheck.MaxBlockVotesPerEvent+1; j++ {
					bv.Val.Votes = append(bv.Val.Votes, hash.HexToHash("0x01"))
				}

				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs MalformedBVs",
			basiccheck.MalformedBVs,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 1,
						Epoch: 0,
						Votes: []hash.Hash{},
					},
				}

				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetBlockVotes(bv.Val)
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
			"validateBVs EmptyBVs",
			basiccheck.EmptyBVs,
			func() {
				bv = inter.LlrSignedBlockVotes{
					Val: inter.LlrBlockVotes{
						Start: 0,
						Epoch: 0,
						Votes: []hash.Hash{},
					},
				}

				s.me.SetEpoch(idx.Epoch(s.startEpoch))
				s.me.SetBlockVotes(bv.Val)
				s.me.SetPayloadHash(inter.CalcPayloadHash(s.me))

				sig, err := s.env.signer.Sign(s.env.pubkeys[1], s.me.HashToSign().Bytes())
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

			err := s.env.checkers.Basiccheck.ValidateBVs(bv)

			if tc.errExp != nil {
				s.Require().Error(err)
				s.Require().EqualError(err, tc.errExp.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}

}

func TestBasicCheckIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LLRBasicCheckTestSuite))
}
