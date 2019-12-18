package ethapi

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
)

// PublicSfcAPI provides an API to access SFC related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicSfcAPI struct {
	b Backend
}

// NewPublicSfcAPI creates a new SFC protocol API.
func NewPublicSfcAPI(b Backend) *PublicSfcAPI {
	return &PublicSfcAPI{b}
}

// GetValidationScore returns staker's ValidationScore.
func (s *PublicSfcAPI) GetValidationScore(ctx context.Context, stakerID hexutil.Uint) (*hexutil.Big, error) {
	v, err := s.b.GetValidationScore(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), err
}

// GetOriginationScore returns staker's OriginationScore.
func (s *PublicSfcAPI) GetOriginationScore(ctx context.Context, stakerID hexutil.Uint) (*hexutil.Big, error) {
	v, err := s.b.GetOriginationScore(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), err
}

// GetRewardWeights returns staker's reward weights.
func (s *PublicSfcAPI) GetRewardWeights(ctx context.Context, stakerID hexutil.Uint) (map[string]interface{}, error) {
	baseRewardWeight, txRewardWeight, err := s.b.GetRewardWeights(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	if baseRewardWeight == nil || txRewardWeight == nil {
		return nil, nil
	}
	return map[string]interface{}{
		"baseRewardWeight": (*hexutil.Big)(baseRewardWeight),
		"txRewardWeight":   (*hexutil.Big)(txRewardWeight),
	}, nil
}

// GetStakerPoI returns staker's PoI.
func (s *PublicSfcAPI) GetStakerPoI(ctx context.Context, stakerID hexutil.Uint) (*hexutil.Big, error) {
	v, err := s.b.GetStakerPoI(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), err
}

// GetDowntime returns staker's Downtime.
func (s *PublicSfcAPI) GetDowntime(ctx context.Context, stakerID hexutil.Uint) (map[string]interface{}, error) {
	blocks, period, err := s.b.GetDowntime(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"missedBlocks": hexutil.Uint64(blocks),
		"downtime":     hexutil.Uint64(period),
	}, nil
}

// GetDelegatorClaimedRewards returns sum of claimed rewards in past, by this delegator
func (s *PublicSfcAPI) GetDelegatorClaimedRewards(ctx context.Context, addr common.Address) (*hexutil.Big, error) {
	v, err := s.b.GetDelegatorClaimedRewards(ctx, addr)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), err
}

// GetStakerClaimedRewards returns sum of claimed rewards in past, by this staker
func (s *PublicSfcAPI) GetStakerClaimedRewards(ctx context.Context, stakerID hexutil.Uint64) (*hexutil.Big, error) {
	v, err := s.b.GetStakerClaimedRewards(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), err
}

// GetStakerDelegatorsClaimedRewards returns sum of claimed rewards in past, by this delegators of this staker
func (s *PublicSfcAPI) GetStakerDelegatorsClaimedRewards(ctx context.Context, stakerID hexutil.Uint64) (*hexutil.Big, error) {
	v, err := s.b.GetStakerDelegatorsClaimedRewards(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(v), err
}

// RPCMarshalStaker converts the given staker to the RPC output .
func RPCMarshalStaker(it sfctype.SfcStakerAndID) map[string]interface{} {
	return map[string]interface{}{
		"id":               hexutil.Uint64(it.StakerID),
		"totalStake":       (*hexutil.Big)(it.Staker.CalcTotalStake()),
		"stake":            (*hexutil.Big)(it.Staker.StakeAmount),
		"delegatedMe":      (*hexutil.Big)(it.Staker.DelegatedMe),
		"isValidator":      it.Staker.IsValidator,
		"isActive":         it.Staker.Ok(),
		"isCheater":        it.Staker.IsCheater(),
		"isOffline":        it.Staker.Offline(),
		"address":          it.Staker.Address,
		"createdEpoch":     hexutil.Uint64(it.Staker.CreatedEpoch),
		"createdTime":      hexutil.Uint64(it.Staker.CreatedTime),
		"deactivatedEpoch": hexutil.Uint64(it.Staker.DeactivatedEpoch),
		"deactivatedTime":  hexutil.Uint64(it.Staker.DeactivatedTime),
	}
}

func (s *PublicSfcAPI) addStakerMetricFields(ctx context.Context, res map[string]interface{}, stakerID idx.StakerID) (map[string]interface{}, error) {
	blocks, period, err := s.b.GetDowntime(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	res["missedBlocks"] = hexutil.Uint64(blocks)
	res["downtime"] = hexutil.Uint64(period)

	poi, err := s.b.GetStakerPoI(ctx, stakerID)
	if err != nil {
		return nil, err
	}
	res["poi"] = (*hexutil.Big)(poi)

	baseRewardWeight, txRewardWeight, err := s.b.GetRewardWeights(ctx, stakerID)
	if err != nil {
		return nil, err
	}
	res["baseRewardWeight"] = (*hexutil.Big)(baseRewardWeight)
	res["txRewardWeight"] = (*hexutil.Big)(txRewardWeight)

	validationScore, err := s.b.GetValidationScore(ctx, stakerID)
	if err != nil {
		return nil, err
	}
	res["validationScore"] = (*hexutil.Big)(validationScore)

	originationScore, err := s.b.GetOriginationScore(ctx, stakerID)
	if err != nil {
		return nil, err
	}
	res["originationScore"] = (*hexutil.Big)(originationScore)

	claimedRewards, err := s.b.GetStakerClaimedRewards(ctx, stakerID)
	if err != nil {
		return nil, err
	}
	res["claimedRewards"] = (*hexutil.Big)(claimedRewards)

	delegatorsClaimedRewards, err := s.b.GetStakerDelegatorsClaimedRewards(ctx, stakerID)
	if err != nil {
		return nil, err
	}
	res["delegatorsClaimedRewards"] = (*hexutil.Big)(delegatorsClaimedRewards)

	return res, nil
}

func (s *PublicSfcAPI) addDelegatorMetricFields(ctx context.Context, res map[string]interface{}, address common.Address) (map[string]interface{}, error) {
	claimedRewards, err := s.b.GetDelegatorClaimedRewards(ctx, address)
	if err != nil {
		return nil, err
	}
	res["claimedRewards"] = (*hexutil.Big)(claimedRewards)

	return res, nil
}

// GetStaker returns SFC staker's info
// Verbosity. Number. If >= 1, then include base field. If >= 2, then include metrics.
func (s *PublicSfcAPI) GetStaker(ctx context.Context, stakerID hexutil.Uint, verbosity hexutil.Uint64) (map[string]interface{}, error) {
	staker, err := s.b.GetStaker(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}
	if staker == nil {
		return nil, nil
	}
	it := sfctype.SfcStakerAndID{
		StakerID: idx.StakerID(stakerID),
		Staker:   staker,
	}
	stakerRPC := RPCMarshalStaker(it)
	if verbosity <= 1 {
		return stakerRPC, nil
	}
	return s.addStakerMetricFields(ctx, stakerRPC, idx.StakerID(stakerID))
}

// GetStakerByAddress returns SFC staker's info by address
// Verbosity. Number. If 0, then include only stakerID. If >= 1, then include base field. If >= 2, then include metrics.
func (s *PublicSfcAPI) GetStakerByAddress(ctx context.Context, address common.Address, verbosity hexutil.Uint64) (map[string]interface{}, error) {
	stakerID, err := s.b.GetStakerID(ctx, address)
	if err != nil {
		return nil, err
	}
	if stakerID == 0 {
		return nil, nil
	}
	if verbosity == 0 {
		// ID only
		return map[string]interface{}{
			"id": hexutil.Uint64(stakerID),
		}, nil
	}
	return s.GetStaker(ctx, hexutil.Uint(stakerID), verbosity)
}

// GetStakers returns SFC stakers info
// Verbosity. Number. If 0, then include only stakerIDs. If >= 1, then include base field. If >= 2, then include metrics (including downtime if validator).
func (s *PublicSfcAPI) GetStakers(ctx context.Context, verbosity hexutil.Uint64) ([]interface{}, error) {
	stakers, err := s.b.GetStakers(ctx)
	if err != nil {
		return nil, err
	}

	if verbosity == 0 {
		// IDs only
		ids := make([]interface{}, len(stakers))
		for i, it := range stakers {
			ids[i] = hexutil.Uint64(it.StakerID).String()
		}
		return ids, nil
	}

	stakersRPC := make([]interface{}, len(stakers))
	for i, it := range stakers {
		stakerRPC := RPCMarshalStaker(it)
		if verbosity >= 2 {
			stakerRPC, err = s.addStakerMetricFields(ctx, stakerRPC, it.StakerID)
			if err != nil {
				return nil, err
			}
		}
		stakersRPC[i] = stakerRPC
	}

	return stakersRPC, err
}

// RPCMarshalDelegator converts the given delegator to the RPC output .
func RPCMarshalDelegator(it sfctype.SfcDelegatorAndAddr) map[string]interface{} {
	return map[string]interface{}{
		"address":          it.Addr,
		"toStakerID":       hexutil.Uint64(it.Delegator.ToStakerID),
		"amount":           (*hexutil.Big)(it.Delegator.Amount),
		"createdEpoch":     hexutil.Uint64(it.Delegator.CreatedEpoch),
		"createdTime":      hexutil.Uint64(it.Delegator.CreatedTime),
		"deactivatedEpoch": hexutil.Uint64(it.Delegator.DeactivatedEpoch),
		"deactivatedTime":  hexutil.Uint64(it.Delegator.DeactivatedTime),
	}
}

// GetDelegatorsOf returns SFC delegators who delegated to a staker
// Verbosity. Number. If 0, then include only addresses. If >= 1, then include base fields. If >= 2, then include metrics.
func (s *PublicSfcAPI) GetDelegatorsOf(ctx context.Context, stakerID hexutil.Uint64, verbosity hexutil.Uint64) ([]interface{}, error) {
	delegators, err := s.b.GetDelegatorsOf(ctx, idx.StakerID(stakerID))
	if err != nil {
		return nil, err
	}

	if verbosity == 0 {
		// Addresses only
		addresses := make([]interface{}, len(delegators))
		for i, it := range delegators {
			addresses[i] = it.Addr.String()
		}
		return addresses, nil
	}

	delegatorsRPC := make([]interface{}, len(delegators))
	for i, it := range delegators {
		delegatorRPC := RPCMarshalDelegator(it)
		if verbosity >= 2 {
			delegatorRPC, err = s.addDelegatorMetricFields(ctx, delegatorRPC, it.Addr)
			if err != nil {
				return nil, err
			}
		}
		delegatorsRPC[i] = delegatorRPC
	}

	return delegatorsRPC, err
}

// GetDelegator returns SFC delegator info
// Verbosity. Number. If >= 1, then include base fields. If >= 2, then include metrics.
func (s *PublicSfcAPI) GetDelegator(ctx context.Context, addr common.Address, verbosity hexutil.Uint64) (map[string]interface{}, error) {
	delegator, err := s.b.GetDelegator(ctx, addr)
	if err != nil {
		return nil, err
	}
	if delegator == nil {
		return nil, nil
	}
	it := sfctype.SfcDelegatorAndAddr{
		Addr:      addr,
		Delegator: delegator,
	}
	delegatorRPC := RPCMarshalDelegator(it)
	if verbosity <= 1 {
		return delegatorRPC, nil
	}
	return s.addDelegatorMetricFields(ctx, delegatorRPC, addr)
}
