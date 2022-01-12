package gossip

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Fantom-foundation/go-opera/eventcheck"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/lachesis-base/hash"
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

	const validatorsNum = 1

	env := newTestEnv(1, validatorsNum)

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

// TODO fix that
func (s *LLRCallbacksTestSuite) Test_processBlockVotesErrUnknownEpochBVs() {
	s.bvs.Val.Epoch = 100
	s.Require().EqualError(s.env.ProcessBlockVotes(s.bvs), eventcheck.ErrUnknownEpochBVs.Error())
}

// test passes
func (s *LLRCallbacksTestSuite) Test_processBlockVoteWonBrIsNil() {
	/*

		ProcessBlockVotes b 0
		range for loop
		processBlockVote block, epoch, bv, val, vals 0 0 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1] 0 [1:1776356839]
		processBlockVote &llrs 0xc0003aa700
		processBlockVote newWeight 1776356839
		processBlockVote vals.TotalWeight()/3+1 592118947
		processBlockVote, wonBr == nil
		range for loop
		processBlockVote block, epoch, bv, val, vals 1 0 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0] 0 [1:1776356839]
		processBlockVote &llrs 0xc0003aa708
		processBlockVote newWeight 1776356839
		processBlockVote vals.TotalWeight()/3+1 592118947
		processBlockVote, wonBr == nil
		PASS
	*/
	// TODO should i require more test cases here?
	s.bvs.Val.Votes = []hash.Hash{hash.HexToHash("0x01"), hash.Zero}
	//s.env.store.SetBlockVotes(s.bvs)
	s.Require().Nil(s.env.ProcessBlockVotes(s.bvs))
}

func (s *LLRCallbacksTestSuite) Test_processBlockVoteLLRVotingDoubleSignIsMet() {
	// TODO figure out how to invoke processBlockVote to trigger LLRVotingDoubleSignIsMet
	// Consider to use FakeEvent to resolve this issue
	/*
		    func processBlockVote(){
				...

				} else if *wonBr != bv {
						fmt.Println("processBlockVote, wonBr != nil")
						s.Log.Error("LLR voting doublesign is met", "block", block)
					}
				}

			}

		    this is TestMisbehaviourProofsBlockVoteDoublesign test in mps_test.go

			Pair: [2]inter.LlrSignedBlockVotes{
						{
							Val: inter.LlrBlockVotes{
								Start: env.store.GetLatestBlockIndex() - 1,
								Epoch: startEpoch,
								Votes: []hash.Hash{
									hash.Zero,
									hash.HexToHash("0x01"),
								},
							},
						},
						{
							Val: inter.LlrBlockVotes{
								Start: env.store.GetLatestBlockIndex(),
								Epoch: startEpoch,
								Votes: []hash.Hash{
									hash.HexToHash("0x02"),
								},
							},
						},
					},
			just wonder how to apply it in my test case

	*/

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
	s.ev.Val.Epoch = 10
	s.Require().EqualError(s.env.ProcessEpochVote(s.ev), eventcheck.ErrUnknownEpochEV.Error())
}

/*

func (v *Checker) ValidateEV(ev inter.LlrSignedEpochVote) error {
	return v.ValidateEventLocator(ev.Signed, ev.Val.Epoch-1, ErrUnknownEpochEV, func() bool {
		return ev.CalcPayloadHash() == ev.Signed.Locator.PayloadHash
	})
}

func (v *Checker) ValidateEventLocator(e inter.SignedEventLocator, authEpoch idx.Epoch, authErr error, checkPayload func() bool) error {
	pubkeys := v.reader.GetEpochPubKeysOf(authEpoch)
	if len(pubkeys) == 0 {
		return authErr
	}
	pubkey, ok := pubkeys[e.Locator.Creator]
	if !ok {
		return epochcheck.ErrAuth
	}
	if checkPayload != nil && !checkPayload() {
		return ErrWrongPayloadHash
	}
	if !verifySignature(e.Locator.HashToSign(), e.Sig, pubkey) {
		return ErrWrongEventSig
	}
	return nil
}


*/

// WIP
func (s *LLRCallbacksTestSuite) Test_ValidateEV() {

	// TODO determine testcases
	// TODO check signature as well
	testCases := []struct {
		name string
		val  inter.LlrEpochVote
		//errExp  error
	}{
		{"zero", inter.LlrEpochVote{Epoch: s.ev.Val.Epoch - 1, Vote: hash.HexToHash("0x01")}},
		{"first", inter.LlrEpochVote{Epoch: s.ev.Val.Epoch, Vote: hash.HexToHash("0x01")}},
		{"second", inter.LlrEpochVote{Epoch: s.ev.Val.Epoch + 1, Vote: hash.HexToHash("0x02")}},
		{"third", inter.LlrEpochVote{Epoch: s.ev.Val.Epoch + 2, Vote: hash.HexToHash("0x03")}},
	}

	for _, tc := range testCases {
		tc := tc

		s.SetupSuite()

		s.ev.Val = tc.val
		// find out how to get checker
		s.Require().NoError(s.env.checkers.Heavycheck.ValidateEV(s.ev))
	}

}

func (s *LLRCallbacksTestSuite) Test_ValidateBVs() {
	s.ev.Val.Epoch = 10
	s.Require().Nil(s.env.ProcessEpochVote(s.ev))
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
 

/*
func ExampleProcessBlockVotes() {
	//votes := []hash.Hash{hash.Zero,hash.HexToHash("0x01")}
    // take event e.g. empty event
	// make mutable event from it?
	//

}
*/
