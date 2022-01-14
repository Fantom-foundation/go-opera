package gossip

import (
	"testing"

	"github.com/stretchr/testify/suite"
	//	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/inter"
	//	"github.com/Fantom-foundation/go-opera/utils"

	"github.com/Fantom-foundation/lachesis-base/hash"
	//	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type LLRCallbacksTestSuite struct {
	// we can inherit another suite here and insert it here
	suite.Suite

	env *testEnv
	e   *inter.EventPayload
	me  *inter.MutableEventPayload
	bvs inter.LlrSignedBlockVotes
	ev  inter.LlrSignedEpochVote
}

func (s *LLRCallbacksTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	const (
		validatorsNum = 10
		////rounds        = 50
	)

	env := newTestEnv(1, validatorsNum)

	// generate txs and multiple blocks
	// TODO consider declare a standalone function
	/*
		for n := uint64(0); n < rounds; n++ {
			// transfers
			txs := make([]*types.Transaction, validatorsNum)
			for i := idx.Validator(0); i < validatorsNum; i++ {
				from := i % validatorsNum
				to := 0
				txs[i] = env.Transfer(idx.ValidatorID(from+1), idx.ValidatorID(to+1), utils.ToFtm(100))
			}
			tm := sameEpoch
			if n%10 == 0 {
				tm = nextEpoch
			}
			_, err := env.ApplyTxs(tm, txs...)
			s.Require().NoError(err)
		}
	*/

	em := env.emitters[0]
	e, err := em.EmitEvent()
	s.Require().NoError(err)
	s.Require().NotNil(e)

	s.env = env
	s.e = e
	s.me = mutableEventPayloadFromImmutable(e)

	s.bvs = inter.AsSignedBlockVotes(s.me)
	s.ev = inter.AsSignedEpochVote(s.me)

}

func (s *LLRCallbacksTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.env.Close()
}

func (s *LLRCallbacksTestSuite) Test_processBlockVotesWithZeroVotes() {
	s.bvs.Val.Votes = []hash.Hash{}
	s.Require().Nil(s.env.ProcessBlockVotes(s.bvs))
}

func (s *LLRCallbacksTestSuite) Test_processBlockVotesErrAlreadyProcessedBVs() {
	s.env.store.SetBlockVotes(s.bvs)
	s.Require().EqualError(s.env.ProcessBlockVotes(s.bvs), eventcheck.ErrAlreadyProcessedBVs.Error())
}

func (s *LLRCallbacksTestSuite) Test_processBlockVotesErrUnknownEpochBVs() {
	bv := inter.LlrSignedBlockVotes{
		Val: inter.LlrBlockVotes{
			Start: s.env.store.GetLatestBlockIndex() - 1,
			Epoch: 1,
			Votes: []hash.Hash{
				hash.Zero,
				hash.HexToHash("0x01"),
			},
		},
	}

	s.Require().EqualError(s.env.ProcessBlockVotes(bv), eventcheck.ErrUnknownEpochBVs.Error())
}

func (s *LLRCallbacksTestSuite) Test_processBlockVoteWonBrIsNil() {
	s.bvs.Val.Votes = []hash.Hash{hash.HexToHash("0x01"), hash.Zero}
	//s.env.store.SetBlockVotes(s.bvs)
	s.Require().Nil(s.env.ProcessBlockVotes(s.bvs))
}

func TestLLRCallbacksTestSuite(t *testing.T) {
	suite.Run(t, new(LLRCallbacksTestSuite))
}

func (s *LLRCallbacksTestSuite) Test_processEpochVoteWithZeroEpoch() {
	s.ev.Val.Epoch = 0
	s.Require().Nil(s.env.ProcessEpochVote(s.ev))
}

func (s *LLRCallbacksTestSuite) Test_processEpochVoteErrAlreadyProcessedEV() {
	s.ev.Val.Epoch = 10
	s.env.store.SetEpochVote(s.ev)
	s.Require().EqualError(s.env.ProcessEpochVote(s.ev), eventcheck.ErrAlreadyProcessedEV.Error())
}

func (s *LLRCallbacksTestSuite) Test_processEpochVoteErrUnknownEpochEV() {
	// print outr everything in GetHistoryEpochState
	ev := inter.LlrSignedEpochVote{
		Val: inter.LlrEpochVote{
			Epoch: 1,
			Vote:  hash.HexToHash("0x01"),
		},
	}

	s.Require().EqualError(s.env.ProcessEpochVote(ev), eventcheck.ErrUnknownEpochEV.Error())
}

// TODO
func Test_actualizeLowestIndex(t *testing.T) {

}

func mutableEventPayloadFromImmutable(e *inter.EventPayload) *inter.MutableEventPayload {
	// we migrate immutable payload to mutable payload
	// we set
	// we test against errors in processEvent

	me := &inter.MutableEventPayload{}
	me.SetVersion(e.Version())
	me.SetNetForkID(e.NetForkID())
	me.SetCreator(e.Creator()) //check in Validate
	me.SetEpoch(e.Epoch())     // check in Validate
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

// TODO make standalone test that is called by a set of validators
func Test_processBlockVoteLLRVotingDoubleSignIsMet(t *testing.T) {

}
