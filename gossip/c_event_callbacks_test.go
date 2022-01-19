package gossip

import (
	"math"
	"testing"
   "bytes"
   "math/big"

	"github.com/Fantom-foundation/go-opera/inter"
	//	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Fantom-foundation/go-opera/eventcheck/basiccheck"
	lbasiccheck "github.com/Fantom-foundation/lachesis-base/eventcheck/basiccheck"
	"github.com/Fantom-foundation/lachesis-base/hash"
	//      "github.com/ethereum/go-ethereum/core/types"
	//      "github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
   "github.com/ethereum/go-ethereum/core/types"
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
   startEpoch idx.Epoch
}

func (s *LLREventCallbacksTestSuite) SetupSuite() {
	s.T().Log("setting up test suite")

	const (
      validatorsNum = 10
      startEpoch = 1
   )

	env := newTestEnv(startEpoch, validatorsNum)

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
   s.startEpoch = idx.Epoch(startEpoch)

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

            s.me.SetGasPowerUsed(math.MaxInt64-1)
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

            s.me.SetGasPowerLeft(inter.GasPowerLeft{Gas: [2]uint64{math.MaxInt64-1, math.MaxInt64}})
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
    
      /*
      {
         "Validate checkTxs validateTx ErrTipAboveFeeCap",
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
           basiccheck.ErrTipAboveFeeCap,
      },
      */
      {
         "Validate validateMP validatorvalidateEventLocator ErrWrongNetForkID",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            wrongNetforkIDMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           NetForkID: uint16(1),
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = wrongNetforkIDMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
            /*
            for i, p := range wrongNetforkIDMp.EventsDoublesign.Pair {
               sig, err := s.env.signer.Sign(s.env.pubkeys[0], p.Locator.HashToSign().Bytes())
               s.Require().NoError(err)
               copy(wrongNetforkIDMp.EventsDoublesign.Pair[i].Sig[:], sig)
            }
            */
          // s.Require().EqualError(s.env.ApplyMPs(nextEpoch, wrongNetforkIDMp), basiccheck.ErrWrongNetForkID.Error())
        //  s.Require().EqualError(s.env.ApplyMPs(time.Duration(s.startEpoch), wrongNetforkIDMp), basiccheck.ErrWrongNetForkID.Error())
           },
           basiccheck.ErrWrongNetForkID,
      },
      {
         "Validate validateMP validatorvalidateEventLocator base.ErrHugeValue-1",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            invalidSeqMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     math.MaxInt32-1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = invalidSeqMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           lbasiccheck.ErrHugeValue,
      },
      {
         "Validate validateMP validatorvalidateEventLocator base.ErrHugeValue-2",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            invalidEpochMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   math.MaxInt32-1,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = invalidEpochMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           lbasiccheck.ErrHugeValue,
      },
      {
         "Validate validateMP validatorvalidateEventLocator base.ErrHugeValue-3",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            invalidLamportMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: math.MaxInt32-1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = invalidLamportMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           lbasiccheck.ErrHugeValue,
      },
      {
         "Validate validateMP ErrWrongCreatorMP",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            wrongCreatorMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 2,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = wrongCreatorMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           basiccheck.ErrWrongCreatorMP,
      },
      {
         "Validate validateMP ErrNoCrimeInMP",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            wrongEpochMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch+1,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = wrongEpochMp 
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           basiccheck.ErrNoCrimeInMP,
      },
      {
         "Validate validateMP ErrNoCrimeInMP-2",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            wrongSeqMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     2,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = wrongSeqMp 
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           basiccheck.ErrNoCrimeInMP,
      },
      {
         "Validate validateMP ErrNoCrimeInMP-3",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(1))
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            wrongLocatorMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = wrongLocatorMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           basiccheck.ErrNoCrimeInMP,
      },
      {
         "Validate validateMP ErrMPTooLate",
          func() {
            s.me.SetSeq(idx.Event(1))
            s.me.SetEpoch(idx.Epoch(20) + basiccheck.MaxLiableEpochs)
            s.me.SetFrame(idx.Frame(1))
            s.me.SetLamport(idx.Lamport(1))

            tooLateMp := inter.MisbehaviourProof{
               EventsDoublesign: &inter.EventsDoublesign{
                  Pair: [2]inter.SignedEventLocator{
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 1,
                           Creator: 1,
                        },
                     },
                     {
                        Locator: inter.EventLocator{
                           Epoch:   s.startEpoch,
                           Seq:     1,
                           Lamport: 2,
                           Creator: 1,
                        },
                     },
                  },
               },
            }
            
            mps := make([]inter.MisbehaviourProof, 2)
            for i:=0; i < 2; i++ {
               mps[i] = tooLateMp
            }
   
            s.me.SetMisbehaviourProofs(mps)
           },
           basiccheck.ErrMPTooLate,
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
