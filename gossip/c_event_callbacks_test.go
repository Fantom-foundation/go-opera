package gossip


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