package sfcapi

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Events

//event CreatedValidator(uint256 indexed validatorID, address indexed auth, uint256 createdEpoch, uint256 createdTime);
//event DeactivatedValidator(uint256 indexed validatorID, uint256 deactivatedEpoch, uint256 deactivatedTime);
//event ChangedValidatorStatus(uint256 indexed validatorID, uint256 status);
//event Delegated(address indexed delegator, uint256 indexed toValidatorID, uint256 amount);
//event Undelegated(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount);
//event ClaimedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 rewards);

var (
	// Topics of SFC contract logs
	Topics = struct {
		ClaimedRewards          common.Hash
		RestakedRewards         common.Hash
		ClaimedDelegationReward common.Hash
		ClaimedValidatorReward  common.Hash
		CreatedValidator        common.Hash
		DeactivatedValidator    common.Hash
		ChangedValidatorStatus  common.Hash
		Delegated               common.Hash
		Undelegated             common.Hash
	}{
		ClaimedRewards:          crypto.Keccak256Hash([]byte("ClaimedRewards(address,uint256,uint256,uint256,uint256)")),
		RestakedRewards:         crypto.Keccak256Hash([]byte("RestakedRewards(address,uint256,uint256,uint256,uint256)")),
		ClaimedDelegationReward: crypto.Keccak256Hash([]byte("ClaimedDelegationReward(address,uint256,uint256,uint256,uint256)")),
		ClaimedValidatorReward:  crypto.Keccak256Hash([]byte("ClaimedValidatorReward(uint256,uint256,uint256,uint256)")),
		CreatedValidator:        crypto.Keccak256Hash([]byte("CreatedValidator(uint256,address,uint256,uint256)")),
		DeactivatedValidator:    crypto.Keccak256Hash([]byte("DeactivatedValidator(uint256,uint256,uint256)")),
		ChangedValidatorStatus:  crypto.Keccak256Hash([]byte("ChangedValidatorStatus(uint256,uint256)")),
		Delegated:               crypto.Keccak256Hash([]byte("Delegated(address,uint256,uint256)")),
		Undelegated:             crypto.Keccak256Hash([]byte("Undelegated(address,uint256,uint256,uint256)")),
	}
)
