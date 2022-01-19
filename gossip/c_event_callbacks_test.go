package gossip

import (
	"math"
	"testing"

	"github.com/Fantom-foundation/go-opera/inter"
	//	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	lbasiccheck "github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	//      "github.com/ethereum/go-ethereum/core/types"
	//      "github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

/*
Plan:
basiccheck - table driven tess to evaluate every test case
	Validate(e inter.EventPayloadI)
	ValidateBVs(bvs inter.LlrSignedBlockVotes)
	ValidateEV(ev inter.LlrSignedEpochVote)

bvallcheck
  test c.HeavyCheck.Enqueue(bvs, checked) in HeavyCheck
epochcheck
   CalcGasPowerUsed  test for few test cases
   (v *Checker).checkGas check for few errors
   CheckTxs test driven tests generate a bunch of transactions and check for all error case
   func (v *Checker) Validate(e inter.EventPayloadI) error { calls Validate(e) from base check
evalcheck
   Enqueue calls .HeavyCheck.Enqueue(evs, checked) to test HeavyCheck
gaspowercheck
   -func (v *Checker) CalcGasPower(e inter.EventI, selfParent inter.EventI) (inter.GasPowerLeft, error) {
    generate some events , test for epochcheck.ErrNotRelevant
    test for res.Gas[i] and test calcGasPower under the hood
   - CalcValidatorGasPower test for calculations
      calcValidatorGasPowerPerSec test calculations
   - test for errors func (v *Checker) Validate(e inter.EventI, selfParent inter.EventI) error {
heavycheck
    find out how to test EnqueueEvent, EnqueueBVs, EnqueueEV
	test ValidateEventLocator against many error cases : test driven tests
    test matchPubkey for various error cases
    test ValidateEventLocator for many error cases
	test ValidateBVs
	test ValidateEV
	test ValidateEvent for many error cases : table driven tests . put some MisbehaviourProofs() on it
parentscheck
	Validateevent func (v *Checker) Validate(e inter.EventI, parents inter.EventIs) error {
all.go
// Validate runs all the checks except Poset-related
it is better to test a single check
func (v *Checkers) Validate(e inter.EventPayloadI, parents inter.EventIs) error {
	runs all checks

*/

type LLREventCallbacksTestSuite struct {
	// we can inherit another suite here and insert it here
	suite.Suite

	env *testEnv
	me  *inter.MutableEventPayload
}

func (s *LLREventCallbacksTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	const validatorsNum = 10

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
	s.me = mutableEventPayloadFromImmutable(e)

}

func (s *LLREventCallbacksTestSuite) TearDownSuite() {
	s.T().Log("tearing down test suite")
	s.env.Close()
}

func (s *LLREventCallbacksTestSuite) TestBasicCheckValidate() {

	/*
		   const (
				rounds        = 60
				validatorsNum = 10
				startEpoch    = 1
			)

			require := require.New(t)

			//creating generator
			generator := newTestEnv(startEpoch, validatorsNum)
			defer generator.Close()

		   em := generator.emitters[0]
	*/

	testCases := []struct {
		name    string
		pretest func()
		errExp  error
	}{

		{"ErrWrongNetForkID",
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
       "Validate checkInited checkInited ErrNoParents",
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
	}

	for _, tc := range testCases {
		tc := tc
		s.SetupSuite()
		tc.pretest()

		err := s.env.checkers.Basiccheck.Validate(s.me)

		if tc.errExp != nil {
			s.Require().Error(err)
			s.Require().EqualError(err, tc.errExp.Error())
		} else {
			s.Require().NoError(err)
		}

	}
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LLREventCallbacksTestSuite))
}
