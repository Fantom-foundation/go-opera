// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sfc100

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractABI is the input ABI used to generate the binding from.
const ContractABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"ChangedValidatorStatus\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupExtraReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupBaseReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unlockedReward\",\"type\":\"uint256\"}],\"name\":\"ClaimedRewards\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"}],\"name\":\"CreatedValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"DeactivatedValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Delegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"LockedUpStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupExtraReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupBaseReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unlockedReward\",\"type\":\"uint256\"}],\"name\":\"RestakedRewards\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Undelegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"penalty\",\"type\":\"uint256\"}],\"name\":\"UnlockedStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"UpdatedBaseRewardPerSec\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blocksNum\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"}],\"name\":\"UpdatedOfflinePenaltyThreshold\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"UpdatedSlashingRefundRatio\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"syncPubkey\",\"type\":\"bool\"}],\"name\":\"_syncValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"baseRewardPerSecond\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"claimRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"contractCommission\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"createValidator\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentSealedEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"deactivateValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochAccumulatedOriginatedTxsFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochAccumulatedRewardPerToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochAccumulatedUptime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochOfflineBlocks\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochOfflineTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochReceivedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalBaseRewardWeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalTxRewardWeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseRewardPerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"getEpochValidatorIDs\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"getLockedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getLockupInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getStashedLockupRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"lockupExtraReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupBaseReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"unlockedReward\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"getUnlockedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getValidator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"receivedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"getValidatorID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getValidatorPubkey\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getWithdrawalRequest\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sealedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_totalSupply\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"nodeDriver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"isLockedUp\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"isSlashed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastValidatorID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"lockStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxDelegatedRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxLockupDuration\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minLockupDuration\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minSelfStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"offlinePenaltyThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"blocksNum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"pendingRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"restakeRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"rewardsStash\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTime\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"}],\"name\":\"sealEpoch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nextValidatorIDs\",\"type\":\"uint256[]\"}],\"name\":\"sealEpochValidators\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupFromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupEndTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewards\",\"type\":\"uint256\"}],\"name\":\"setGenesisDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"setGenesisValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slashingRefundRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"stashRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"stashedRewardsUntilEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalActiveStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSlashedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"unlockStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"unlockedRewardRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"updateBaseRewardPerSecond\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blocksNum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"}],\"name\":\"updateOfflinePenaltyThreshold\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"updateSlashingRefundRatio\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"validatorCommission\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"bytes3\",\"name\":\"\",\"type\":\"bytes3\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalPeriodEpochs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"withdrawalPeriodTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minStakeIncrease\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minStakeDecrease\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minDelegation\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minDelegationIncrease\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minDelegationDecrease\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakeLockPeriodTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakeLockPeriodEpochs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegationLockPeriodTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegationLockPeriodEpochs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_toStakerID\",\"type\":\"uint256\"}],\"name\":\"delegations\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paidUntilEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_stakerID\",\"type\":\"uint256\"}],\"name\":\"stakers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"paidUntilEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"delegatedMe\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"dagAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"sfcAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_addr\",\"type\":\"address\"}],\"name\":\"getStakerID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakeTotalAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegationsTotalAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"}],\"name\":\"isDelegationLockedUp\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakerID\",\"type\":\"uint256\"}],\"name\":\"isStakeLockedUp\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakersLastID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakersNum\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegationsNum\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"}],\"name\":\"lockedDelegations\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"fromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakerID\",\"type\":\"uint256\"}],\"name\":\"lockedStakes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"fromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"createDelegation\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"calcDelegationRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakerID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"calcValidatorRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"}],\"name\":\"claimDelegationRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"}],\"name\":\"claimDelegationCompoundRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"claimValidatorRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"claimValidatorCompoundRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"prepareToWithdrawStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"prepareToWithdrawStakePartial\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdrawStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"prepareToWithdrawDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"prepareToWithdrawDelegationPartial\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"withdrawDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"partialWithdrawByRequest\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"lockDuration\",\"type\":\"uint256\"}],\"name\":\"lockUpStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"lockDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toStakerID\",\"type\":\"uint256\"}],\"name\":\"lockUpDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
var ContractBin = "0x608060405234801561001057600080fd5b5061603e80620000216000396000f3fe6080604052600436106106705760003560e01c80638b0e9f3f1161034e578063c65ee0e1116101bb578063de67f215116100f7578063ebdf104c116100a0578063f3ae5b1a1161007a578063f3ae5b1a1461190a578063f8b18d8a1461164e578063f99837e614611934578063fd5e6dd11461196457610670565b8063ebdf104c1461174f578063ec6a7f1c146118c2578063f2fde38b146118d757610670565b8063df4f49d4116100d1578063df4f49d414611678578063e08d7e66146116a2578063e261641a1461171f57610670565b8063de67f215146115e8578063df00c9221461161e578063df0e307a1461164e57610670565b8063cfd5fa0c11610164578063d9a7c1f91161013e578063d9a7c1f91461153a578063dc31e1af1461154f578063dc599bb11461157f578063dd099bb6146115af57610670565b8063cfd5fa0c14611483578063cfdbb7cd146114bc578063d845fc90146114f557610670565b8063cc8343aa11610195578063cc8343aa146113ee578063cda5826a14611420578063cfd476631461144a57610670565b8063c65ee0e1146113af578063c7be95de146113d9578063cb1c4e671461070357610670565b8063b1e643391161028a578063bb03a4bd11610233578063c3de580e1161020d578063c3de580e1461135b578063c41b640514611385578063c4b5dd7e14610703578063c5f530af1461139a57610670565b8063bb03a4bd14611310578063bed9d86114611346578063c312eb071461104e57610670565b8063b810e41111610264578063b810e41114611248578063b82b842714611281578063b88a37e21461129657610670565b8063b1e6433914611189578063b5d89627146111b3578063b6d9edd51461121e57610670565b806396c7ee46116102f7578063a4b89fab116102d1578063a4b89fab1461109b578063a5a470ad146110cb578063a77865151461113b578063a86a056f1461115057610670565b806396c7ee4614610fef5780639fa6dd351461104e578063a198d2291461106b57610670565b80638da5cb5b116103285780638da5cb5b14610f735780638f32d59b14610fa457806396060e7114610fb957610670565b80638b0e9f3f14610ef55780638b1a0d1114610f0a5780638cddb01514610f3a57610670565b806339b80c00116104ec5780636099ecb2116104285780636f498663116103d15780637cacb1d6116103ab5780637cacb1d614610e035780637f664d8714610e1857806381d9dc7a14610718578063854873e114610e5657610670565b80636f49866314610da0578063715018a614610dd95780637667180814610dee57610670565b806363321e271161040257806363321e2714610d1f578063650acd6614610d52578063670322f814610d6757610670565b80636099ecb214610cb657806360c7e37f1461070357806361e53fcc14610cef57610670565b80634f864df41161049557806354fd4d501161046f57806354fd4d5014610c2757806358f95b8014610c715780635e2308d2146109e65780635fab23a814610ca157610670565b80634f864df414610b465780634feb92f314610b7c57806354d77ed21461088e57610670565b8063441a3e70116104c6578063441a3e7014610ad15780634bd202dc14610b015780634f7c4efb14610b1657610670565b806339b80c0014610a925780633d0317fe14610abc5780633fee10a81461088e57610670565b80631d3ac42c116105bb57806326682c7111610564578063295cccba1161053e578063295cccba14610a105780632cedb09714610a3a57806330fa992914610a68578063375b3c0a14610a7d57610670565b806326682c71146109b65780632709275e146109e657806328f73148146109fb57610670565b80631f270152116105955780631f270152146108d3578063223fae09146109305780632265f284146109a157610670565b80631d3ac42c1461085e5780631d58179c1461088e5780631e702f83146108a357610670565b80630d4955e31161061d57806318160ddd116105f757806318160ddd146107e457806318f628d4146107f957806319ddb54f1461070357610670565b80630d4955e3146107815780630d7b26091461079657806312622d0e146107ab57610670565b806308728f6e1161064e57806308728f6e1461071857806308c368741461072d5780630962ef791461075757610670565b80630135b1db14610675578063019e2729146106ba5780630298599214610703575b600080fd5b34801561068157600080fd5b506106a86004803603602081101561069857600080fd5b50356001600160a01b03166119ea565b60408051918252519081900360200190f35b3480156106c657600080fd5b50610701600480360360808110156106dd57600080fd5b508035906020810135906001600160a01b03604082013581169160600135166119fc565b005b34801561070f57600080fd5b506106a8611b73565b34801561072457600080fd5b506106a8611b79565b34801561073957600080fd5b506107016004803603602081101561075057600080fd5b5035611b7f565b34801561076357600080fd5b506107016004803603602081101561077a57600080fd5b5035611c4b565b34801561078d57600080fd5b506106a8611d28565b3480156107a257600080fd5b506106a8611d30565b3480156107b757600080fd5b506106a8600480360360408110156107ce57600080fd5b506001600160a01b038135169060200135611d37565b3480156107f057600080fd5b506106a8611dc0565b34801561080557600080fd5b50610701600480360361012081101561081d57600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e0810135906101000135611dc6565b34801561086a57600080fd5b506106a86004803603604081101561088157600080fd5b5080359060200135611f26565b34801561089a57600080fd5b506106a86120db565b3480156108af57600080fd5b50610701600480360360408110156108c657600080fd5b50803590602001356120ea565b3480156108df57600080fd5b50610912600480360360608110156108f657600080fd5b506001600160a01b038135169060208101359060400135612199565b60408051938452602084019290925282820152519081900360600190f35b34801561093c57600080fd5b506109696004803603604081101561095357600080fd5b506001600160a01b0381351690602001356121cb565b604080519788526020880196909652868601949094526060860192909252608085015260a084015260c0830152519081900360e00190f35b3480156109ad57600080fd5b506106a861223f565b3480156109c257600080fd5b50610701600480360360408110156109d957600080fd5b5080359060200135612251565b3480156109f257600080fd5b506106a8612271565b348015610a0757600080fd5b506106a861228d565b348015610a1c57600080fd5b5061070160048036036020811015610a3357600080fd5b5035612293565b348015610a4657600080fd5b50610a4f6122ac565b6040805192835260208301919091528051918290030190f35b348015610a7457600080fd5b506106a86122b6565b348015610a8957600080fd5b506106a86122c9565b348015610a9e57600080fd5b5061096960048036036020811015610ab557600080fd5b50356122d3565b348015610ac857600080fd5b506106a8612315565b348015610add57600080fd5b5061070160048036036040811015610af457600080fd5b5080359060200135612326565b348015610b0d57600080fd5b506106a8612680565b348015610b2257600080fd5b5061070160048036036040811015610b3957600080fd5b5080359060200135612685565b348015610b5257600080fd5b5061070160048036036060811015610b6957600080fd5b50803590602081013590604001356127c9565b348015610b8857600080fd5b506107016004803603610100811015610ba057600080fd5b6001600160a01b0382351691602081013591810190606081016040820135640100000000811115610bd057600080fd5b820183602082011115610be257600080fd5b80359060200191846001830284011164010000000083111715610c0457600080fd5b919350915080359060208101359060408101359060608101359060800135612b57565b348015610c3357600080fd5b50610c3c612bfd565b604080517fffffff00000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b348015610c7d57600080fd5b506106a860048036036040811015610c9457600080fd5b5080359060200135612c21565b348015610cad57600080fd5b506106a8612c3e565b348015610cc257600080fd5b506106a860048036036040811015610cd957600080fd5b506001600160a01b038135169060200135612c44565b348015610cfb57600080fd5b506106a860048036036040811015610d1257600080fd5b5080359060200135612c82565b348015610d2b57600080fd5b506106a860048036036020811015610d4257600080fd5b50356001600160a01b0316612ca3565b348015610d5e57600080fd5b506106a8612cc2565b348015610d7357600080fd5b506106a860048036036040811015610d8a57600080fd5b506001600160a01b038135169060200135612cc7565b348015610dac57600080fd5b506106a860048036036040811015610dc357600080fd5b506001600160a01b038135169060200135612d08565b348015610de557600080fd5b50610701612d72565b348015610dfa57600080fd5b506106a8612e2d565b348015610e0f57600080fd5b506106a8612e36565b348015610e2457600080fd5b50610e4260048036036020811015610e3b57600080fd5b5035612e3c565b604080519115158252519081900360200190f35b348015610e6257600080fd5b50610e8060048036036020811015610e7957600080fd5b5035612e61565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610eba578181015183820152602001610ea2565b50505050905090810190601f168015610ee75780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b348015610f0157600080fd5b506106a8612f1a565b348015610f1657600080fd5b5061070160048036036040811015610f2d57600080fd5b5080359060200135612f20565b348015610f4657600080fd5b5061070160048036036040811015610f5d57600080fd5b506001600160a01b038135169060200135612fc2565b348015610f7f57600080fd5b50610f8861301d565b604080516001600160a01b039092168252519081900360200190f35b348015610fb057600080fd5b50610e4261302c565b348015610fc557600080fd5b5061091260048036036060811015610fdc57600080fd5b508035906020810135906040013561303d565b348015610ffb57600080fd5b506110286004803603604081101561101257600080fd5b506001600160a01b038135169060200135613095565b604080519485526020850193909352838301919091526060830152519081900360800190f35b6107016004803603602081101561106457600080fd5b50356130c7565b34801561107757600080fd5b506106a86004803603604081101561108e57600080fd5b50803590602001356130d5565b3480156110a757600080fd5b50610701600480360360408110156110be57600080fd5b50803590602001356130f6565b610701600480360360208110156110e157600080fd5b8101906020810181356401000000008111156110fc57600080fd5b82018360208201111561110e57600080fd5b8035906020019184600183028401116401000000008311171561113057600080fd5b50909250905061311e565b34801561114757600080fd5b506106a8613219565b34801561115c57600080fd5b506106a86004803603604081101561117357600080fd5b506001600160a01b03813516906020013561322f565b34801561119557600080fd5b50610701600480360360208110156111ac57600080fd5b503561324c565b3480156111bf57600080fd5b506111dd600480360360208110156111d657600080fd5b5035613299565b604080519788526020880196909652868601949094526060860192909252608085015260a08401526001600160a01b031660c0830152519081900360e00190f35b34801561122a57600080fd5b506107016004803603602081101561124157600080fd5b50356132df565b34801561125457600080fd5b506109126004803603604081101561126b57600080fd5b506001600160a01b0381351690602001356133d1565b34801561128d57600080fd5b506106a86133fd565b3480156112a257600080fd5b506112c0600480360360208110156112b957600080fd5b5035613404565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156112fc5781810151838201526020016112e4565b505050509050019250505060405180910390f35b34801561131c57600080fd5b506107016004803603606081101561133357600080fd5b5080359060208101359060400135613469565b34801561135257600080fd5b50610701613474565b34801561136757600080fd5b50610e426004803603602081101561137e57600080fd5b50356134c1565b34801561139157600080fd5b5061070161324c565b3480156113a657600080fd5b506106a86134d8565b3480156113bb57600080fd5b506106a8600480360360208110156113d257600080fd5b50356134e7565b3480156113e557600080fd5b506106a86134f9565b3480156113fa57600080fd5b506107016004803603604081101561141157600080fd5b508035906020013515156134ff565b34801561142c57600080fd5b506107016004803603602081101561144357600080fd5b5035613737565b34801561145657600080fd5b506106a86004803603604081101561146d57600080fd5b506001600160a01b038135169060200135613750565b34801561148f57600080fd5b50610e42600480360360408110156114a657600080fd5b506001600160a01b03813516906020013561376d565b3480156114c857600080fd5b50610e42600480360360408110156114df57600080fd5b506001600160a01b038135169060200135613775565b34801561150157600080fd5b506109126004803603608081101561151857600080fd5b506001600160a01b0381351690602081013590604081013590606001356137dd565b34801561154657600080fd5b506106a861381b565b34801561155b57600080fd5b506106a86004803603604081101561157257600080fd5b5080359060200135613821565b34801561158b57600080fd5b50610701600480360360408110156115a257600080fd5b5080359060200135613842565b3480156115bb57600080fd5b50610912600480360360408110156115d257600080fd5b506001600160a01b03813516906020013561384b565b3480156115f457600080fd5b506107016004803603606081101561160b57600080fd5b50803590602081013590604001356138b7565b34801561162a57600080fd5b506106a86004803603604081101561164157600080fd5b5080359060200135613bf5565b34801561165a57600080fd5b506107016004803603602081101561167157600080fd5b5035613474565b34801561168457600080fd5b506109126004803603602081101561169b57600080fd5b5035613c16565b3480156116ae57600080fd5b50610701600480360360208110156116c557600080fd5b8101906020810181356401000000008111156116e057600080fd5b8201836020820111156116f257600080fd5b8035906020019184602083028401116401000000008311171561171457600080fd5b509092509050613c4c565b34801561172b57600080fd5b506106a86004803603604081101561174257600080fd5b5080359060200135613d2c565b34801561175b57600080fd5b506107016004803603608081101561177257600080fd5b81019060208101813564010000000081111561178d57600080fd5b82018360208201111561179f57600080fd5b803590602001918460208302840111640100000000831117156117c157600080fd5b9193909290916020810190356401000000008111156117df57600080fd5b8201836020820111156117f157600080fd5b8035906020019184602083028401116401000000008311171561181357600080fd5b91939092909160208101903564010000000081111561183157600080fd5b82018360208201111561184357600080fd5b8035906020019184602083028401116401000000008311171561186557600080fd5b91939092909160208101903564010000000081111561188357600080fd5b82018360208201111561189557600080fd5b803590602001918460208302840111640100000000831117156118b757600080fd5b509092509050613d4d565b3480156118ce57600080fd5b506106a8613f29565b3480156118e357600080fd5b50610701600480360360208110156118fa57600080fd5b50356001600160a01b0316613f33565b34801561191657600080fd5b506107016004803603602081101561192d57600080fd5b5035613f95565b34801561194057600080fd5b506107016004803603604081101561195757600080fd5b5080359060200135613fb8565b34801561197057600080fd5b5061198e6004803603602081101561198757600080fd5b5035613fc1565b604080519a8b5260208b0199909952898901979097526060890195909552608088019390935260a087019190915260c086015260e08501526001600160a01b039081166101008501521661012083015251908190036101400190f35b60696020526000908152604090205481565b600054610100900460ff1680611a155750611a156140e4565b80611a23575060005460ff16155b611a5e5760405162461bcd60e51b815260040180806020018281038252602e815260200180615f3e602e913960400191505060405180910390fd5b600054610100900460ff16158015611ac457600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b611acd826140ea565b6067859055606680547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b03851617905560768490556755cfe697852e904c6075556103e86078556203f480607955611b2b61424b565b6000868152607760205260409020600701558015611b6c57600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555b5050505050565b60015b90565b606b5490565b33611b88615d35565b611b92828461424f565b60208101518151919250600091611bae9163ffffffff61435416565b9050611bd18385611bcc85604001518561435490919063ffffffff16565b6143ae565b6001600160a01b0383166000818152607360209081526040808320888452825291829020805485019055845185820151868401518451928352928201528083019190915290518692917f4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26919081900360600190a350505050565b33611c54615d35565b611c5e828461424f565b9050816001600160a01b03166108fc611c9c8360400151611c908560200151866000015161435490919063ffffffff16565b9063ffffffff61435416565b6040518115909202916000818181858888f19350505050158015611cc4573d6000803e3d6000fd5b5082826001600160a01b03167fc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf4583600001518460200151856040015160405180848152602001838152602001828152602001935050505060405180910390a3505050565b6301e1338090565b6212750090565b6000611d438383613775565b611d7157506001600160a01b0382166000908152607260209081526040808320848452909152902054611dba565b6001600160a01b038316600081815260736020908152604080832086845282528083205493835260728252808320868452909152902054611db79163ffffffff6144b816565b90505b92915050565b60765481565b611dcf336144fa565b611e0a5760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b611e1589898961450e565b6001600160a01b0389166000908152606f602090815260408083208b84529091529020600201819055611e4787614685565b8515611f1b5786861115611e8c5760405162461bcd60e51b815260040180806020018281038252602c815260200180615fde602c913960400191505060405180910390fd5b6001600160a01b03891660008181526073602090815260408083208c845282528083208a8155600181018a90556002810189905560038101889055848452607483528184208d855283529281902086905580518781529182018a9052805192938c9390927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e7892908290030190a3505b505050505050505050565b336000818152607360209081526040808320868452909152812090919083611f95576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b611f9f8286613775565b611ff0576040805162461bcd60e51b815260206004820152600d60248201527f6e6f74206c6f636b656420757000000000000000000000000000000000000000604482015290519081900360640190fd5b8054841115612046576040805162461bcd60e51b815260206004820152601760248201527f6e6f7420656e6f756768206c6f636b6564207374616b65000000000000000000604482015290519081900360640190fd5b6120508286614705565b506000612063838787856000015461484d565b825486900383556001600160a01b03841660008181526072602090815260408083208b8452825291829020805485900390558151898152908101849052815193945089937fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9929181900390910190a395945050505050565b60006120e5612cc2565b905090565b6120f3336144fa565b61212e5760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b80612180576040805162461bcd60e51b815260206004820152600c60248201527f77726f6e67207374617475730000000000000000000000000000000000000000604482015290519081900360640190fd5b61218a8282614996565b6121958260006134ff565b5050565b607160209081526000938452604080852082529284528284209052825290208054600182015460029092015490919083565b6001600160a01b03821660009081526072602090815260408083208484529091528120548190819081908190819081908061221c575060009650869550859450849350839250829150819050612233565b600197508796506000955085945092508591508790505b92959891949750929550565b6000612249614ac0565b601002905090565b3360009081526069602052604090205461226c8184846127c9565b505050565b6000606461227d614ac0565b601e028161228757fe5b04905090565b606d5481565b3360009081526069602052604090205461219581611c4b565b6078546079549091565b60006122c0612315565b606c5403905090565b60006120e56134d8565b607760205280600052604060002060009150905080600701549080600801549080600901549080600a01549080600b01549080600c01549080600d0154905087565b60006064606c546018028161228757fe5b3361232f615d35565b506001600160a01b0381166000908152607160209081526040808320868452825280832085845282529182902082516060810184528154808252600183015493820193909352600290910154928101929092526123d3576040805162461bcd60e51b815260206004820152601560248201527f7265717565737420646f65736e27742065786973740000000000000000000000604482015290519081900360640190fd5b6020808201518251600087815260689093526040909220600101549091901580159061240f575060008681526068602052604090206001015482115b15612430575050600084815260686020526040902060018101546002909101545b6124386133fd565b820161244261424b565b1015612495576040805162461bcd60e51b815260206004820152601660248201527f6e6f7420656e6f7567682074696d652070617373656400000000000000000000604482015290519081900360640190fd5b61249d612cc2565b81016124a7612e2d565b10156124fa576040805162461bcd60e51b815260206004820152601860248201527f6e6f7420656e6f7567682065706f636873207061737365640000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152607160209081526040808320898452825280832088845290915281206002015490612533886134c1565b905060006125558383607a60008d815260200190815260200160002054614acc565b6001600160a01b03881660009081526071602090815260408083208d845282528083208c845290915281208181556001810182905560020155606e80548201905590508083116125ec576040805162461bcd60e51b815260206004820152601660248201527f7374616b652069732066756c6c7920736c617368656400000000000000000000604482015290519081900360640190fd5b6001600160a01b0387166108fc612609858463ffffffff6144b816565b6040518115909202916000818181858888f19350505050158015612631573d6000803e3d6000fd5b508789886001600160a01b03167f75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21866040518082815260200191505060405180910390a4505050505050505050565b600090565b61268d61302c565b6126de576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6126e7826134c1565b612738576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f722069736e277420736c6173686564000000000000000000604482015290519081900360640190fd5b612740614ac0565b81111561277e5760405162461bcd60e51b8152600401808060200182810382526021815260200180615f6c6021913960400191505060405180910390fd5b6000828152607a60209081526040918290208390558151838152915184927f047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb91792908290030190a25050565b336127d48185614705565b506000821161282a576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6128348185611d37565b821115612888576040805162461bcd60e51b815260206004820152601960248201527f6e6f7420656e6f75676820756e6c6f636b6564207374616b6500000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607160209081526040808320878452825280832086845290915290206002015415612909576040805162461bcd60e51b815260206004820152601360248201527f7772494420616c72656164792065786973747300000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607260209081526040808320878452825280832080548690039055606890915290206003015461294e908363ffffffff6144b816565b600085815260686020526040902060030155606c54612973908363ffffffff6144b816565b606c556000848152606860205260409020546129a057606d5461299c908363ffffffff6144b816565b606d555b60006129ab85614b2e565b90508015612a58576129bb6134d8565b811015612a0f576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b612a1885614b60565b612a535760405162461bcd60e51b8152600401808060200182810382526029815260200180615fb56029913960400191505060405180910390fd5b612a63565b612a63856001614996565b6001600160a01b038216600090815260716020908152604080832088845282528083208784529091529020600201839055612a9c612e2d565b6001600160a01b03831660009081526071602090815260408083208984528252808320888452909152902055612ad061424b565b6001600160a01b03831660009081526071602090815260408083208984528252808320888452909152812060010191909155612b0d9086906134ff565b8385836001600160a01b03167fd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57866040518082815260200191505060405180910390a45050505050565b612b60336144fa565b612b9b5760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b612be3898989898080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b92508a91508990508888614ba8565b606b54881115611f1b57606b889055505050505050505050565b7f333030000000000000000000000000000000000000000000000000000000000090565b600091825260776020908152604080842092845291905290205490565b606e5481565b6000612c4e615d35565b612c588484614d6f565b805160208201516040830151929350612c7a92611c909163ffffffff61435416565b949350505050565b60009182526077602090815260408084209284526001909201905290205490565b6001600160a01b0381166000908152606960205260409020545b919050565b600390565b6000612cd38383613775565b612cdf57506000611dba565b506001600160a01b03919091166000908152607360209081526040808320938352929052205490565b6000612d12615d35565b506001600160a01b0383166000908152606f6020908152604080832085845282529182902082516060810184528154808252600183015493820184905260029092015493810184905292612c7a929091611c90919063ffffffff61435416565b612d7a61302c565b612dcb576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055565b60675460010190565b60675481565b600081815260686020526040812060060154611dba906001600160a01b031683613775565b606a6020908152600091825260409182902080548351601f60027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff61010060018616150201909316929092049182018490048402810184019094528084529091830182828015612f125780601f10612ee757610100808354040283529160200191612f12565b820191906000526020600020905b815481529060010190602001808311612ef557829003601f168201915b505050505081565b606c5481565b612f2861302c565b612f79576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b60798190556078829055604080518381526020810183905281517f702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34929181900390910190a15050565b612fcc8282614705565b612195576040805162461bcd60e51b815260206004820152601060248201527f6e6f7468696e6720746f20737461736800000000000000000000000000000000604482015290519081900360640190fd5b6033546001600160a01b031690565b6033546001600160a01b0316331490565b600083815260686020526040812060060154819081908190613068906001600160a01b031688612c44565b9050806130805750600092506001915082905061308c565b60675490935091508190505b93509350939050565b607360209081526000928352604080842090915290825290208054600182015460028301546003909301549192909184565b6130d23382346143ae565b50565b60009182526077602090815260408084209284526005909201905290205490565b33600090815260726020908152604080832084845290915290205461219590829084906138b7565b6131266134d8565b34101561317a576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b806131cc576040805162461bcd60e51b815260206004820152600c60248201527f656d707479207075626b65790000000000000000000000000000000000000000604482015290519081900360640190fd5b61320c3383838080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250614ddd92505050565b61219533606b54346143ae565b60006064613225614ac0565b600f028161228757fe5b607060209081526000928352604080842090915290825290205481565b6040805162461bcd60e51b815260206004820152601f60248201527f75736520534643763320756e64656c656761746528292066756e6374696f6e00604482015290519081900360640190fd5b606860205260009081526040902080546001820154600283015460038401546004850154600586015460069096015494959394929391929091906001600160a01b031687565b6132e761302c565b613338576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6801c985c8903591eb20811115613396576040805162461bcd60e51b815260206004820152601b60248201527f746f6f206c617267652072657761726420706572207365636f6e640000000000604482015290519081900360640190fd5b60758190556040805182815290517f8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae23969181900360200190a150565b607460209081526000928352604080842090915290825290208054600182015460029092015490919083565b62093a8090565b60008181526077602090815260409182902060060180548351818402810184019094528084526060939283018282801561345d57602002820191906000526020600020905b815481526020019060010190808311613449575b50505050509050919050565b61226c8284836127c9565b6040805162461bcd60e51b815260206004820152601d60248201527f75736520534643763320776974686472617728292066756e6374696f6e000000604482015290519081900360640190fd5b600090815260686020526040902054608016151590565b6a02a055184a310c1260000090565b607a6020526000908152604090205481565b606b5481565b61350882614e08565b613559576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090206003810154905415613577575060005b606654604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018690526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b1580156135e457600080fd5b505af11580156135f8573d6000803e3d6000fd5b5050505081801561360857508015155b1561226c576066546000848152606a60205260409081902081517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810187815260248201938452825460027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6001831615610100020190911604604483018190526001600160a01b039095169463242a6e3f948994939091606490910190849080156136f95780601f106136ce576101008083540402835291602001916136f9565b820191906000526020600020905b8154815290600101906020018083116136dc57829003601f168201915b50509350505050600060405180830381600087803b15801561371a57600080fd5b505af115801561372e573d6000803e3d6000fd5b50505050505050565b3360009081526069602052604090205461219581611b7f565b607260209081526000928352604080842090915290825290205481565b6000611db783835b6001600160a01b038216600090815260736020908152604080832084845290915281206002015415801590611db757506001600160a01b03831660009081526073602090815260408083208584529091529020600201546137d461424b565b11159392505050565b6000806000806137ed8888612c44565b90508061380557506000925060019150829050613811565b60675490935091508190505b9450945094915050565b60755481565b60009182526077602090815260408084209284526003909201905290205490565b61219581611b7f565b6000806000613858615d56565b505050506001600160a01b03919091166000908152607360209081526040808320938352928152908290208251608081018452815481526001820154928101839052600282015493810184905260039091015460609091018190529092565b338161390a576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6139148185613775565b15613966576040805162461bcd60e51b815260206004820152601160248201527f616c7265616479206c6f636b6564207570000000000000000000000000000000604482015290519081900360640190fd5b6139708185611d37565b8211156139c4576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b60008481526068602052604090205415613a25576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b613a2d611d30565b8310158015613a435750613a3f611d28565b8311155b613a94576040805162461bcd60e51b815260206004820152601260248201527f696e636f7272656374206475726174696f6e0000000000000000000000000000604482015290519081900360640190fd5b6000613aa284611c9061424b565b6000868152606860205260409020600601549091506001600160a01b039081169083168114613b30576001600160a01b0381166000908152607360209081526040808320898452909152902060020154821115613b305760405162461bcd60e51b8152600401808060200182810382526028815260200180615f8d6028913960400191505060405180910390fd5b613b3a8387614705565b506001600160a01b03831660009081526073602090815260408083208984529091529020848155613b69612e2d565b6001808301919091556002808301859055600383018890556001600160a01b03861660008181526074602090815260408083208d845282528083208381559586018390559490930155825189815291820188905282518a9391927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e7892908290030190a350505050505050565b60009182526077602090815260408084209284526002909201905290205490565b60008181526068602052604081206006015481908190613c3f906001600160a01b03168561384b565b9250925092509193909250565b613c55336144fa565b613c905760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b600060776000613c9e612e2d565b8152602001908152602001600020905060008090505b82811015613d17576000848483818110613cca57fe5b60209081029290920135600081815260688452604080822060030154948890529020839055600c860154909350613d0891508263ffffffff61435416565b600c8501555050600101613cb4565b50613d26600682018484615d7e565b50505050565b60009182526077602090815260408084209284526004909201905290205490565b613d56336144fa565b613d915760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b600060776000613d9f612e2d565b81526020019081526020016000209050606081600601805480602002602001604051908101604052809291908181526020018280548015613dff57602002820191906000526020600020905b815481526020019060010190808311613deb575b50505050509050613e8682828c8c80806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050508b8b80806020026020016040519081016040528093929190818152602001838360200280828437600092019190915250614e1f92505050565b613ef5828288888080602002602001604051908101604052809392919081815260200183836020028082843760009201919091525050604080516020808c0282810182019093528b82529093508b92508a918291850190849080828437600092019190915250614f2e92505050565b613efd612e2d565b606755613f0861424b565b600783015550607554600b820155607654600d909101555050505050505050565b60006120e56133fd565b613f3b61302c565b613f8c576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6130d28161554e565b336000908152606960205260409020546121958183613fb382614b2e565b6138b7565b61219581611c4b565b600080600080600080600080600080613fd8615dc9565b5060008b815260686020908152604091829020825160e08101845281548082526001830154938201939093526002820154938101939093526003810154606084015260048101546080840152600581015460a0840152600601546001600160a01b031660c083015260081415614052576101008152614074565b8051608014156140655760018152614074565b80516001141561407457600081525b600061407f8d614b2e565b9050816000015182608001518360a00151846040015185602001518560016140b4888a606001516144b890919063ffffffff16565b8960c001518a60c001518393509b509b509b509b509b509b509b509b509b509b5050509193959799509193959799565b303b1590565b600054610100900460ff168061410357506141036140e4565b80614111575060005460ff16155b61414c5760405162461bcd60e51b815260040180806020018281038252602e815260200180615f3e602e913960400191505060405180910390fd5b600054610100900460ff161580156141b257600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0384811691909117918290556040519116906000907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a3801561219557600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555050565b4290565b614257615d35565b6142618383614705565b50506001600160a01b0382166000908152606f60209081526040808320848452825280832081516060810183528154808252600183015494820185905260029092015492810183905293926142bf92611c909163ffffffff61435416565b905080614313576040805162461bcd60e51b815260206004820152600c60248201527f7a65726f20726577617264730000000000000000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152606f602090815260408083208684529091528120818155600181018290556002015561434d81614685565b5092915050565b600082820183811015611db7576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b6143b782614e08565b614408576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090205415614469576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b61447483838361450e565b61447d82614b60565b61226c5760405162461bcd60e51b8152600401808060200182810382526029815260200180615fb56029913960400191505060405180910390fd5b6000611db783836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f770000815250615607565b6066546001600160a01b0390811691161490565b60008111614563576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b61456d8383614705565b506001600160a01b03831660009081526072602090815260408083208584529091529020546145a2908263ffffffff61435416565b6001600160a01b03841660009081526072602090815260408083208684528252808320939093556068905220600301546145e2818363ffffffff61435416565b600084815260686020526040902060030155606c54614607908363ffffffff61435416565b606c5560008381526068602052604090205461463457606d54614630908363ffffffff61435416565b606d555b61463f8382156134ff565b60408051838152905184916001600160a01b038716917f9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb9181900360200190a350505050565b606654604080517f66e7ea0f0000000000000000000000000000000000000000000000000000000081523060048201526024810184905290516001600160a01b03909216916366e7ea0f9160448082019260009290919082900301818387803b1580156146f157600080fd5b505af1158015611b6c573d6000803e3d6000fd5b600061470f615d35565b614719848461569e565b9050614724836157d6565b6001600160a01b0385166000818152607060209081526040808320888452825280832094909455918152606f82528281208682528252829020825160608101845281548152600182015492810192909252600201549181019190915261478a9082615831565b6001600160a01b0385166000818152606f60209081526040808320888452825280832085518155858301516001808301919091559582015160029182015593835260748252808320888452825291829020825160608101845281548152948101549185019190915290910154908201526148049082615831565b6001600160a01b03851660009081526074602090815260408083208784528252918290208351815590830151600180830191909155929091015160029091015591505092915050565b6001600160a01b03841660009081526074602090815260408083208684529091528120548190614895908490614889908763ffffffff6158a316565b9063ffffffff6158fc16565b6001600160a01b0387166000908152607460209081526040808320898452909152812060010154919250906148d6908590614889908863ffffffff6158a316565b905060028104820160006148ee86614889848a6158a3565b6001600160a01b038a1660009081526074602090815260408083208c8452909152902054909150614925908563ffffffff6144b816565b6001600160a01b038a1660009081526074602090815260408083208c845290915290209081556001015461495990846144b8565b6001600160a01b038a1660009081526074602090815260408083208c845290915290206001015586811061498a5750855b98975050505050505050565b6000828152606860205260409020541580156149b157508015155b156149de57600082815260686020526040902060030154606d546149da9163ffffffff6144b816565b606d555b60008281526068602052604090205481111561219557600082815260686020526040902081815560020154614a8657614a15612e2d565b600083815260686020526040902060020155614a2f61424b565b6000838152606860209081526040918290206001810184905560020154825190815290810192909252805184927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e4792908290030190a25b60408051828152905183917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25050565b670de0b6b3a764000090565b6000821580614ae25750614ade614ac0565b8210155b15614aef57506000614b27565b614b1a6001611c90614aff614ac0565b61488986614b0b614ac0565b8a91900363ffffffff6158a316565b905083811115614b275750825b9392505050565b6000818152606860209081526040808320600601546001600160a01b0316835260728252808320938352929052205490565b6000614b8d614b6d614ac0565b614889614b7861223f565b614b8186614b2e565b9063ffffffff6158a316565b60008381526068602052604090206003015411159050919050565b6001600160a01b03881660009081526069602052604090205415614c13576040805162461bcd60e51b815260206004820152601860248201527f76616c696461746f7220616c7265616479206578697374730000000000000000604482015290519081900360640190fd5b6001600160a01b03881660008181526069602090815260408083208b90558a8352606882528083208981556004810189905560058101889055600181018690556002810187905560060180547fffffffffffffffffffffffff000000000000000000000000000000000000000016909417909355606a81529190208751614c9c92890190615e0f565b50876001600160a01b0316877f49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf8686604051808381526020018281526020019250505060405180910390a38115614d28576040805183815260208101839052815189927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47928290030190a25b8415614d655760408051868152905188917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25b5050505050505050565b614d77615d35565b614d7f615d35565b614d89848461569e565b6001600160a01b0385166000908152606f602090815260408083208784528252918290208251606081018452815481526001820154928101929092526002015491810191909152909150612c7a9082615831565b606b80546001019081905561226c8382846000614df8612e2d565b614e0061424b565b600080614ba8565b600090815260686020526040902060050154151590565b60005b8351811015611b6c57607854828281518110614e3a57fe5b6020026020010151118015614e645750607954838281518110614e5957fe5b602002602001015110155b15614ea557614e87848281518110614e7857fe5b60200260200101516008614996565b614ea5848281518110614e9657fe5b602002602001015160006134ff565b828181518110614eb157fe5b6020026020010151856004016000868481518110614ecb57fe5b6020026020010151815260200190815260200160002081905550818181518110614ef157fe5b6020026020010151856005016000868481518110614f0b57fe5b602090810291909101810151825281019190915260400160002055600101614e22565b614f36615e7d565b6040518060c001604052808551604051908082528060200260200182016040528015614f6c578160200160208202803883390190505b508152602001600081526020018551604051908082528060200260200182016040528015614fa4578160200160208202803883390190505b508152602001600081526020016000815260200160008152509050600060776000614fde6001614fd2612e2d565b9063ffffffff6144b816565b8152602081019190915260400160002060016080840152600781015490915061500561424b565b111561501f57806007015461501861424b565b0360808301525b60005b85518110156150ea57826080015185828151811061503c57fe5b602002602001015185838151811061505057fe5b6020026020010151028161506057fe5b048360400151828151811061507157fe5b6020026020010181815250506150ab8360400151828151811061509057fe5b6020026020010151846060015161435490919063ffffffff16565b606084015283516150dd908590839081106150c257fe5b60200260200101518460a0015161435490919063ffffffff16565b60a0840152600101615022565b5060005b85518110156151bb57826080015185828151811061510857fe5b6020026020010151846080015187848151811061512157fe5b60200260200101518a60000160008b878151811061513b57fe5b6020026020010151815260200190815260200160002054028161515a57fe5b04028161516357fe5b048360000151828151811061517457fe5b6020026020010181815250506151ae8360000151828151811061519357fe5b6020026020010151846020015161435490919063ffffffff16565b60208401526001016150ee565b5060005b85518110156155265760006151f78460800151607554866000015185815181106151e557fe5b6020026020010151876020015161593e565b90506152336152268560a001518660400151858151811061521457fe5b6020026020010151876060015161597f565b829063ffffffff61435416565b9050600087838151811061524357fe5b6020908102919091018101516000818152606890925260408220600601549092506001600160a01b03169061527f8461527a613219565b6159dc565b6001600160a01b03831660009081526072602090815260408083208784529091529020549091508015615426576000816152b98587612cc7565b8402816152c257fe5b0490508083036152d0615d35565b6001600160a01b03861660009081526073602090815260408083208a84529091529020600301546153029084906159f9565b905061530c615d35565b6153178360006159f9565b6001600160a01b0388166000908152606f602090815260408083208c8452825291829020825160608101845281548152600182015492810192909252600201549181019190915290915061536c908383615aea565b6001600160a01b0388166000818152606f602090815260408083208d84528252808320855181558583015160018083019190915595820151600291820155938352607482528083208d8452825291829020825160608101845281548152948101549185019190915290910154908201526153e7908383615aea565b6001600160a01b03881660009081526074602090815260408083208c845282529182902083518155908301516001820155910151600290910155505050505b600084815260686020526040812060030154838703918115615458578161544b614ac0565b84028161545457fe5b0490505b808a600101600089815260200190815260200160002054018f6001016000898152602001908152602001600020819055508b898151811061549557fe5b60200260200101518a600301600089815260200190815260200160002054018f6003016000898152602001908152602001600020819055508c89815181106154d957fe5b60200260200101518a600201600089815260200190815260200160002054018f600201600089815260200190815260200160002081905550505050505050505080806001019150506151bf565b505060a081015160088601556020810151600986015560600151600a90940193909355505050565b6001600160a01b0381166155935760405162461bcd60e51b8152600401808060200182810382526026815260200180615ece6026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b600081848411156156965760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b8381101561565b578181015183820152602001615643565b50505050905090810190601f1680156156885780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b6156a6615d35565b6001600160a01b0383166000908152607060209081526040808320858452909152812054906156d4846157d6565b905060006156e28686615b05565b9050818111156156ef5750805b828110156156fa5750815b6001600160a01b0386166000818152607360209081526040808320898452825280832093835260728252808320898452909152812054825490919061574690839063ffffffff6144b816565b9050600061575a84600001548a8988615be2565b9050615764615d35565b6157728286600301546159f9565b9050615780838b8a89615be2565b915061578a615d35565b6157958360006159f9565b90506157a3858c898b615be2565b92506157ad615d35565b6157b88460006159f9565b90506157c5838383615aea565b9d9c50505050505050505050505050565b600081815260686020526040812060020154156158295760008281526068602052604090206002015460675410156158115750606754612cbd565b50600081815260686020526040902060020154612cbd565b505060675490565b615839615d35565b604080516060810190915282518451829161585a919063ffffffff61435416565b815260200161587a8460200151866020015161435490919063ffffffff16565b815260200161589a8460400151866040015161435490919063ffffffff16565b90529392505050565b6000826158b257506000611dba565b828202828482816158bf57fe5b0414611db75760405162461bcd60e51b8152600401808060200182810382526021815260200180615f1d6021913960400191505060405180910390fd5b6000611db783836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f000000000000815250615c50565b60008261594d57506000612c7a565b600061595f868663ffffffff6158a316565b905061597583614889838763ffffffff6158a316565b9695505050505050565b60008261598e57506000614b27565b60006159a483614889878763ffffffff6158a316565b90506159d36159b1614ac0565b6148896159bc612271565b6159c4614ac0565b8591900363ffffffff6158a316565b95945050505050565b6000611db76159e9614ac0565b614889858563ffffffff6158a316565b615a01615d35565b604051806060016040528060008152602001600081526020016000815250905081600014615abc576000615a33612271565b615a3b614ac0565b0390506000615a5b615a4b611d28565b614889848763ffffffff6158a316565b90506000615a84615a6a614ac0565b61488984615a76612271565b8a910163ffffffff6158a316565b9050615aa9615a91614ac0565b614889615a9c612271565b899063ffffffff6158a316565b602085018190529003835250611dba9050565b615adf615ac7614ac0565b614889615ad2612271565b869063ffffffff6158a316565b604082015292915050565b615af2615d35565b612c7a615aff8585615831565b83615831565b6001600160a01b0382166000908152607360209081526040808320848452909152812060010154606754615b3a858583615cb5565b15615b48579150611dba9050565b615b53858584615cb5565b615b6257600092505050611dba565b80821115615b7557600092505050611dba565b80821015615ba857600281830104615b8e868683615cb5565b15615b9e57806001019250615ba2565b8091505b50615b75565b80615bb857600092505050611dba565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01949350505050565b6000818310615bf357506000612c7a565b600083815260776020818152604080842088855260019081018352818520548786529383528185208986520190915290912054615c45615c31614ac0565b61488989614b81858763ffffffff6144b816565b979650505050505050565b60008183615c9f5760405162461bcd60e51b815260206004820181815283516024840152835190928392604490910191908501908083836000831561565b578181015183820152602001615643565b506000838581615cab57fe5b0495945050505050565b6001600160a01b03831660009081526073602090815260408083208584529091528120600101548210801590612c7a57506001600160a01b0384166000908152607360209081526040808320868452909152902060020154615d1683615d20565b1115949350505050565b60009081526077602052604090206007015490565b60405180606001604052806000815260200160008152602001600081525090565b6040518060800160405280600081526020016000815260200160008152602001600081525090565b828054828255906000526020600020908101928215615db9579160200282015b82811115615db9578235825591602001919060010190615d9e565b50615dc5929150615eb3565b5090565b6040518060e0016040528060008152602001600081526020016000815260200160008152602001600081526020016000815260200160006001600160a01b031681525090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10615e5057805160ff1916838001178555615db9565b82800160010185558215615db9579182015b82811115615db9578251825591602001919060010190615e62565b6040518060c001604052806060815260200160008152602001606081526020016000815260200160008152602001600081525090565b611b7691905b80821115615dc55760008155600101615eb956fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f206164647265737363616c6c6572206973206e6f7420746865204e6f64654472697665724175746820636f6e7472616374536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f77436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a65646d757374206265206c657373207468616e206f7220657175616c20746f20312e3076616c696461746f72206c6f636b757020706572696f642077696c6c20656e64206561726c69657276616c696461746f7227732064656c65676174696f6e73206c696d69742069732065786365656465646c6f636b6564207374616b652069732067726561746572207468616e207468652077686f6c65207374616b65a265627a7a723158204bc860db30ff399bf4cc12003067479c85b0a58310e666fbfa89ae82ed3dfd0a64736f6c634300050c0032"

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(ContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// BaseRewardPerSecond is a free data retrieval call binding the contract method 0xd9a7c1f9.
//
// Solidity: function baseRewardPerSecond() view returns(uint256)
func (_Contract *ContractCaller) BaseRewardPerSecond(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "baseRewardPerSecond")
	return *ret0, err
}

// BaseRewardPerSecond is a free data retrieval call binding the contract method 0xd9a7c1f9.
//
// Solidity: function baseRewardPerSecond() view returns(uint256)
func (_Contract *ContractSession) BaseRewardPerSecond() (*big.Int, error) {
	return _Contract.Contract.BaseRewardPerSecond(&_Contract.CallOpts)
}

// BaseRewardPerSecond is a free data retrieval call binding the contract method 0xd9a7c1f9.
//
// Solidity: function baseRewardPerSecond() view returns(uint256)
func (_Contract *ContractCallerSession) BaseRewardPerSecond() (*big.Int, error) {
	return _Contract.Contract.BaseRewardPerSecond(&_Contract.CallOpts)
}

// CalcDelegationRewards is a free data retrieval call binding the contract method 0xd845fc90.
//
// Solidity: function calcDelegationRewards(address delegator, uint256 toStakerID, uint256 , uint256 ) view returns(uint256, uint256, uint256)
func (_Contract *ContractCaller) CalcDelegationRewards(opts *bind.CallOpts, delegator common.Address, toStakerID *big.Int, arg2 *big.Int, arg3 *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
	}
	err := _Contract.contract.Call(opts, out, "calcDelegationRewards", delegator, toStakerID, arg2, arg3)
	return *ret0, *ret1, *ret2, err
}

// CalcDelegationRewards is a free data retrieval call binding the contract method 0xd845fc90.
//
// Solidity: function calcDelegationRewards(address delegator, uint256 toStakerID, uint256 , uint256 ) view returns(uint256, uint256, uint256)
func (_Contract *ContractSession) CalcDelegationRewards(delegator common.Address, toStakerID *big.Int, arg2 *big.Int, arg3 *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	return _Contract.Contract.CalcDelegationRewards(&_Contract.CallOpts, delegator, toStakerID, arg2, arg3)
}

// CalcDelegationRewards is a free data retrieval call binding the contract method 0xd845fc90.
//
// Solidity: function calcDelegationRewards(address delegator, uint256 toStakerID, uint256 , uint256 ) view returns(uint256, uint256, uint256)
func (_Contract *ContractCallerSession) CalcDelegationRewards(delegator common.Address, toStakerID *big.Int, arg2 *big.Int, arg3 *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	return _Contract.Contract.CalcDelegationRewards(&_Contract.CallOpts, delegator, toStakerID, arg2, arg3)
}

// CalcValidatorRewards is a free data retrieval call binding the contract method 0x96060e71.
//
// Solidity: function calcValidatorRewards(uint256 stakerID, uint256 , uint256 ) view returns(uint256, uint256, uint256)
func (_Contract *ContractCaller) CalcValidatorRewards(opts *bind.CallOpts, stakerID *big.Int, arg1 *big.Int, arg2 *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	var (
		ret0 = new(*big.Int)
		ret1 = new(*big.Int)
		ret2 = new(*big.Int)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
	}
	err := _Contract.contract.Call(opts, out, "calcValidatorRewards", stakerID, arg1, arg2)
	return *ret0, *ret1, *ret2, err
}

// CalcValidatorRewards is a free data retrieval call binding the contract method 0x96060e71.
//
// Solidity: function calcValidatorRewards(uint256 stakerID, uint256 , uint256 ) view returns(uint256, uint256, uint256)
func (_Contract *ContractSession) CalcValidatorRewards(stakerID *big.Int, arg1 *big.Int, arg2 *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	return _Contract.Contract.CalcValidatorRewards(&_Contract.CallOpts, stakerID, arg1, arg2)
}

// CalcValidatorRewards is a free data retrieval call binding the contract method 0x96060e71.
//
// Solidity: function calcValidatorRewards(uint256 stakerID, uint256 , uint256 ) view returns(uint256, uint256, uint256)
func (_Contract *ContractCallerSession) CalcValidatorRewards(stakerID *big.Int, arg1 *big.Int, arg2 *big.Int) (*big.Int, *big.Int, *big.Int, error) {
	return _Contract.Contract.CalcValidatorRewards(&_Contract.CallOpts, stakerID, arg1, arg2)
}

// ContractCommission is a free data retrieval call binding the contract method 0x2709275e.
//
// Solidity: function contractCommission() pure returns(uint256)
func (_Contract *ContractCaller) ContractCommission(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "contractCommission")
	return *ret0, err
}

// ContractCommission is a free data retrieval call binding the contract method 0x2709275e.
//
// Solidity: function contractCommission() pure returns(uint256)
func (_Contract *ContractSession) ContractCommission() (*big.Int, error) {
	return _Contract.Contract.ContractCommission(&_Contract.CallOpts)
}

// ContractCommission is a free data retrieval call binding the contract method 0x2709275e.
//
// Solidity: function contractCommission() pure returns(uint256)
func (_Contract *ContractCallerSession) ContractCommission() (*big.Int, error) {
	return _Contract.Contract.ContractCommission(&_Contract.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_Contract *ContractCaller) CurrentEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "currentEpoch")
	return *ret0, err
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_Contract *ContractSession) CurrentEpoch() (*big.Int, error) {
	return _Contract.Contract.CurrentEpoch(&_Contract.CallOpts)
}

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_Contract *ContractCallerSession) CurrentEpoch() (*big.Int, error) {
	return _Contract.Contract.CurrentEpoch(&_Contract.CallOpts)
}

// CurrentSealedEpoch is a free data retrieval call binding the contract method 0x7cacb1d6.
//
// Solidity: function currentSealedEpoch() view returns(uint256)
func (_Contract *ContractCaller) CurrentSealedEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "currentSealedEpoch")
	return *ret0, err
}

// CurrentSealedEpoch is a free data retrieval call binding the contract method 0x7cacb1d6.
//
// Solidity: function currentSealedEpoch() view returns(uint256)
func (_Contract *ContractSession) CurrentSealedEpoch() (*big.Int, error) {
	return _Contract.Contract.CurrentSealedEpoch(&_Contract.CallOpts)
}

// CurrentSealedEpoch is a free data retrieval call binding the contract method 0x7cacb1d6.
//
// Solidity: function currentSealedEpoch() view returns(uint256)
func (_Contract *ContractCallerSession) CurrentSealedEpoch() (*big.Int, error) {
	return _Contract.Contract.CurrentSealedEpoch(&_Contract.CallOpts)
}

// DelegationLockPeriodEpochs is a free data retrieval call binding the contract method 0x1d58179c.
//
// Solidity: function delegationLockPeriodEpochs() pure returns(uint256)
func (_Contract *ContractCaller) DelegationLockPeriodEpochs(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "delegationLockPeriodEpochs")
	return *ret0, err
}

// DelegationLockPeriodEpochs is a free data retrieval call binding the contract method 0x1d58179c.
//
// Solidity: function delegationLockPeriodEpochs() pure returns(uint256)
func (_Contract *ContractSession) DelegationLockPeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.DelegationLockPeriodEpochs(&_Contract.CallOpts)
}

// DelegationLockPeriodEpochs is a free data retrieval call binding the contract method 0x1d58179c.
//
// Solidity: function delegationLockPeriodEpochs() pure returns(uint256)
func (_Contract *ContractCallerSession) DelegationLockPeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.DelegationLockPeriodEpochs(&_Contract.CallOpts)
}

// DelegationLockPeriodTime is a free data retrieval call binding the contract method 0xec6a7f1c.
//
// Solidity: function delegationLockPeriodTime() pure returns(uint256)
func (_Contract *ContractCaller) DelegationLockPeriodTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "delegationLockPeriodTime")
	return *ret0, err
}

// DelegationLockPeriodTime is a free data retrieval call binding the contract method 0xec6a7f1c.
//
// Solidity: function delegationLockPeriodTime() pure returns(uint256)
func (_Contract *ContractSession) DelegationLockPeriodTime() (*big.Int, error) {
	return _Contract.Contract.DelegationLockPeriodTime(&_Contract.CallOpts)
}

// DelegationLockPeriodTime is a free data retrieval call binding the contract method 0xec6a7f1c.
//
// Solidity: function delegationLockPeriodTime() pure returns(uint256)
func (_Contract *ContractCallerSession) DelegationLockPeriodTime() (*big.Int, error) {
	return _Contract.Contract.DelegationLockPeriodTime(&_Contract.CallOpts)
}

// Delegations is a free data retrieval call binding the contract method 0x223fae09.
//
// Solidity: function delegations(address _from, uint256 _toStakerID) view returns(uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime, uint256 amount, uint256 paidUntilEpoch, uint256 toStakerID)
func (_Contract *ContractCaller) Delegations(opts *bind.CallOpts, _from common.Address, _toStakerID *big.Int) (struct {
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	Amount           *big.Int
	PaidUntilEpoch   *big.Int
	ToStakerID       *big.Int
}, error) {
	ret := new(struct {
		CreatedEpoch     *big.Int
		CreatedTime      *big.Int
		DeactivatedEpoch *big.Int
		DeactivatedTime  *big.Int
		Amount           *big.Int
		PaidUntilEpoch   *big.Int
		ToStakerID       *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "delegations", _from, _toStakerID)
	return *ret, err
}

// Delegations is a free data retrieval call binding the contract method 0x223fae09.
//
// Solidity: function delegations(address _from, uint256 _toStakerID) view returns(uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime, uint256 amount, uint256 paidUntilEpoch, uint256 toStakerID)
func (_Contract *ContractSession) Delegations(_from common.Address, _toStakerID *big.Int) (struct {
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	Amount           *big.Int
	PaidUntilEpoch   *big.Int
	ToStakerID       *big.Int
}, error) {
	return _Contract.Contract.Delegations(&_Contract.CallOpts, _from, _toStakerID)
}

// Delegations is a free data retrieval call binding the contract method 0x223fae09.
//
// Solidity: function delegations(address _from, uint256 _toStakerID) view returns(uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime, uint256 amount, uint256 paidUntilEpoch, uint256 toStakerID)
func (_Contract *ContractCallerSession) Delegations(_from common.Address, _toStakerID *big.Int) (struct {
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	Amount           *big.Int
	PaidUntilEpoch   *big.Int
	ToStakerID       *big.Int
}, error) {
	return _Contract.Contract.Delegations(&_Contract.CallOpts, _from, _toStakerID)
}

// DelegationsNum is a free data retrieval call binding the contract method 0x4bd202dc.
//
// Solidity: function delegationsNum() pure returns(uint256)
func (_Contract *ContractCaller) DelegationsNum(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "delegationsNum")
	return *ret0, err
}

// DelegationsNum is a free data retrieval call binding the contract method 0x4bd202dc.
//
// Solidity: function delegationsNum() pure returns(uint256)
func (_Contract *ContractSession) DelegationsNum() (*big.Int, error) {
	return _Contract.Contract.DelegationsNum(&_Contract.CallOpts)
}

// DelegationsNum is a free data retrieval call binding the contract method 0x4bd202dc.
//
// Solidity: function delegationsNum() pure returns(uint256)
func (_Contract *ContractCallerSession) DelegationsNum() (*big.Int, error) {
	return _Contract.Contract.DelegationsNum(&_Contract.CallOpts)
}

// DelegationsTotalAmount is a free data retrieval call binding the contract method 0x30fa9929.
//
// Solidity: function delegationsTotalAmount() view returns(uint256)
func (_Contract *ContractCaller) DelegationsTotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "delegationsTotalAmount")
	return *ret0, err
}

// DelegationsTotalAmount is a free data retrieval call binding the contract method 0x30fa9929.
//
// Solidity: function delegationsTotalAmount() view returns(uint256)
func (_Contract *ContractSession) DelegationsTotalAmount() (*big.Int, error) {
	return _Contract.Contract.DelegationsTotalAmount(&_Contract.CallOpts)
}

// DelegationsTotalAmount is a free data retrieval call binding the contract method 0x30fa9929.
//
// Solidity: function delegationsTotalAmount() view returns(uint256)
func (_Contract *ContractCallerSession) DelegationsTotalAmount() (*big.Int, error) {
	return _Contract.Contract.DelegationsTotalAmount(&_Contract.CallOpts)
}

// GetEpochAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0xdc31e1af.
//
// Solidity: function getEpochAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochAccumulatedOriginatedTxsFee(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochAccumulatedOriginatedTxsFee", epoch, validatorID)
	return *ret0, err
}

// GetEpochAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0xdc31e1af.
//
// Solidity: function getEpochAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochAccumulatedOriginatedTxsFee(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochAccumulatedOriginatedTxsFee(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0xdc31e1af.
//
// Solidity: function getEpochAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochAccumulatedOriginatedTxsFee(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochAccumulatedOriginatedTxsFee(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochAccumulatedRewardPerToken is a free data retrieval call binding the contract method 0x61e53fcc.
//
// Solidity: function getEpochAccumulatedRewardPerToken(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochAccumulatedRewardPerToken(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochAccumulatedRewardPerToken", epoch, validatorID)
	return *ret0, err
}

// GetEpochAccumulatedRewardPerToken is a free data retrieval call binding the contract method 0x61e53fcc.
//
// Solidity: function getEpochAccumulatedRewardPerToken(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochAccumulatedRewardPerToken(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochAccumulatedRewardPerToken(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochAccumulatedRewardPerToken is a free data retrieval call binding the contract method 0x61e53fcc.
//
// Solidity: function getEpochAccumulatedRewardPerToken(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochAccumulatedRewardPerToken(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochAccumulatedRewardPerToken(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochAccumulatedUptime is a free data retrieval call binding the contract method 0xdf00c922.
//
// Solidity: function getEpochAccumulatedUptime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochAccumulatedUptime(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochAccumulatedUptime", epoch, validatorID)
	return *ret0, err
}

// GetEpochAccumulatedUptime is a free data retrieval call binding the contract method 0xdf00c922.
//
// Solidity: function getEpochAccumulatedUptime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochAccumulatedUptime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochAccumulatedUptime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochAccumulatedUptime is a free data retrieval call binding the contract method 0xdf00c922.
//
// Solidity: function getEpochAccumulatedUptime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochAccumulatedUptime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochAccumulatedUptime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochOfflineBlocks is a free data retrieval call binding the contract method 0xa198d229.
//
// Solidity: function getEpochOfflineBlocks(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochOfflineBlocks(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochOfflineBlocks", epoch, validatorID)
	return *ret0, err
}

// GetEpochOfflineBlocks is a free data retrieval call binding the contract method 0xa198d229.
//
// Solidity: function getEpochOfflineBlocks(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochOfflineBlocks(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochOfflineBlocks(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochOfflineBlocks is a free data retrieval call binding the contract method 0xa198d229.
//
// Solidity: function getEpochOfflineBlocks(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochOfflineBlocks(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochOfflineBlocks(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochOfflineTime is a free data retrieval call binding the contract method 0xe261641a.
//
// Solidity: function getEpochOfflineTime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochOfflineTime(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochOfflineTime", epoch, validatorID)
	return *ret0, err
}

// GetEpochOfflineTime is a free data retrieval call binding the contract method 0xe261641a.
//
// Solidity: function getEpochOfflineTime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochOfflineTime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochOfflineTime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochOfflineTime is a free data retrieval call binding the contract method 0xe261641a.
//
// Solidity: function getEpochOfflineTime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochOfflineTime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochOfflineTime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochReceivedStake is a free data retrieval call binding the contract method 0x58f95b80.
//
// Solidity: function getEpochReceivedStake(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochReceivedStake(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochReceivedStake", epoch, validatorID)
	return *ret0, err
}

// GetEpochReceivedStake is a free data retrieval call binding the contract method 0x58f95b80.
//
// Solidity: function getEpochReceivedStake(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochReceivedStake(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochReceivedStake(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochReceivedStake is a free data retrieval call binding the contract method 0x58f95b80.
//
// Solidity: function getEpochReceivedStake(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochReceivedStake(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochReceivedStake(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshot is a free data retrieval call binding the contract method 0x39b80c00.
//
// Solidity: function getEpochSnapshot(uint256 ) view returns(uint256 endTime, uint256 epochFee, uint256 totalBaseRewardWeight, uint256 totalTxRewardWeight, uint256 baseRewardPerSecond, uint256 totalStake, uint256 totalSupply)
func (_Contract *ContractCaller) GetEpochSnapshot(opts *bind.CallOpts, arg0 *big.Int) (struct {
	EndTime               *big.Int
	EpochFee              *big.Int
	TotalBaseRewardWeight *big.Int
	TotalTxRewardWeight   *big.Int
	BaseRewardPerSecond   *big.Int
	TotalStake            *big.Int
	TotalSupply           *big.Int
}, error) {
	ret := new(struct {
		EndTime               *big.Int
		EpochFee              *big.Int
		TotalBaseRewardWeight *big.Int
		TotalTxRewardWeight   *big.Int
		BaseRewardPerSecond   *big.Int
		TotalStake            *big.Int
		TotalSupply           *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getEpochSnapshot", arg0)
	return *ret, err
}

// GetEpochSnapshot is a free data retrieval call binding the contract method 0x39b80c00.
//
// Solidity: function getEpochSnapshot(uint256 ) view returns(uint256 endTime, uint256 epochFee, uint256 totalBaseRewardWeight, uint256 totalTxRewardWeight, uint256 baseRewardPerSecond, uint256 totalStake, uint256 totalSupply)
func (_Contract *ContractSession) GetEpochSnapshot(arg0 *big.Int) (struct {
	EndTime               *big.Int
	EpochFee              *big.Int
	TotalBaseRewardWeight *big.Int
	TotalTxRewardWeight   *big.Int
	BaseRewardPerSecond   *big.Int
	TotalStake            *big.Int
	TotalSupply           *big.Int
}, error) {
	return _Contract.Contract.GetEpochSnapshot(&_Contract.CallOpts, arg0)
}

// GetEpochSnapshot is a free data retrieval call binding the contract method 0x39b80c00.
//
// Solidity: function getEpochSnapshot(uint256 ) view returns(uint256 endTime, uint256 epochFee, uint256 totalBaseRewardWeight, uint256 totalTxRewardWeight, uint256 baseRewardPerSecond, uint256 totalStake, uint256 totalSupply)
func (_Contract *ContractCallerSession) GetEpochSnapshot(arg0 *big.Int) (struct {
	EndTime               *big.Int
	EpochFee              *big.Int
	TotalBaseRewardWeight *big.Int
	TotalTxRewardWeight   *big.Int
	BaseRewardPerSecond   *big.Int
	TotalStake            *big.Int
	TotalSupply           *big.Int
}, error) {
	return _Contract.Contract.GetEpochSnapshot(&_Contract.CallOpts, arg0)
}

// GetEpochValidatorIDs is a free data retrieval call binding the contract method 0xb88a37e2.
//
// Solidity: function getEpochValidatorIDs(uint256 epoch) view returns(uint256[])
func (_Contract *ContractCaller) GetEpochValidatorIDs(opts *bind.CallOpts, epoch *big.Int) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochValidatorIDs", epoch)
	return *ret0, err
}

// GetEpochValidatorIDs is a free data retrieval call binding the contract method 0xb88a37e2.
//
// Solidity: function getEpochValidatorIDs(uint256 epoch) view returns(uint256[])
func (_Contract *ContractSession) GetEpochValidatorIDs(epoch *big.Int) ([]*big.Int, error) {
	return _Contract.Contract.GetEpochValidatorIDs(&_Contract.CallOpts, epoch)
}

// GetEpochValidatorIDs is a free data retrieval call binding the contract method 0xb88a37e2.
//
// Solidity: function getEpochValidatorIDs(uint256 epoch) view returns(uint256[])
func (_Contract *ContractCallerSession) GetEpochValidatorIDs(epoch *big.Int) ([]*big.Int, error) {
	return _Contract.Contract.GetEpochValidatorIDs(&_Contract.CallOpts, epoch)
}

// GetLockedStake is a free data retrieval call binding the contract method 0x670322f8.
//
// Solidity: function getLockedStake(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractCaller) GetLockedStake(opts *bind.CallOpts, delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getLockedStake", delegator, toValidatorID)
	return *ret0, err
}

// GetLockedStake is a free data retrieval call binding the contract method 0x670322f8.
//
// Solidity: function getLockedStake(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractSession) GetLockedStake(delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetLockedStake(&_Contract.CallOpts, delegator, toValidatorID)
}

// GetLockedStake is a free data retrieval call binding the contract method 0x670322f8.
//
// Solidity: function getLockedStake(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetLockedStake(delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetLockedStake(&_Contract.CallOpts, delegator, toValidatorID)
}

// GetLockupInfo is a free data retrieval call binding the contract method 0x96c7ee46.
//
// Solidity: function getLockupInfo(address , uint256 ) view returns(uint256 lockedStake, uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractCaller) GetLockupInfo(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	LockedStake *big.Int
	FromEpoch   *big.Int
	EndTime     *big.Int
	Duration    *big.Int
}, error) {
	ret := new(struct {
		LockedStake *big.Int
		FromEpoch   *big.Int
		EndTime     *big.Int
		Duration    *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getLockupInfo", arg0, arg1)
	return *ret, err
}

// GetLockupInfo is a free data retrieval call binding the contract method 0x96c7ee46.
//
// Solidity: function getLockupInfo(address , uint256 ) view returns(uint256 lockedStake, uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractSession) GetLockupInfo(arg0 common.Address, arg1 *big.Int) (struct {
	LockedStake *big.Int
	FromEpoch   *big.Int
	EndTime     *big.Int
	Duration    *big.Int
}, error) {
	return _Contract.Contract.GetLockupInfo(&_Contract.CallOpts, arg0, arg1)
}

// GetLockupInfo is a free data retrieval call binding the contract method 0x96c7ee46.
//
// Solidity: function getLockupInfo(address , uint256 ) view returns(uint256 lockedStake, uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractCallerSession) GetLockupInfo(arg0 common.Address, arg1 *big.Int) (struct {
	LockedStake *big.Int
	FromEpoch   *big.Int
	EndTime     *big.Int
	Duration    *big.Int
}, error) {
	return _Contract.Contract.GetLockupInfo(&_Contract.CallOpts, arg0, arg1)
}

// GetStake is a free data retrieval call binding the contract method 0xcfd47663.
//
// Solidity: function getStake(address , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) GetStake(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getStake", arg0, arg1)
	return *ret0, err
}

// GetStake is a free data retrieval call binding the contract method 0xcfd47663.
//
// Solidity: function getStake(address , uint256 ) view returns(uint256)
func (_Contract *ContractSession) GetStake(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetStake(&_Contract.CallOpts, arg0, arg1)
}

// GetStake is a free data retrieval call binding the contract method 0xcfd47663.
//
// Solidity: function getStake(address , uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) GetStake(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetStake(&_Contract.CallOpts, arg0, arg1)
}

// GetStakerID is a free data retrieval call binding the contract method 0x63321e27.
//
// Solidity: function getStakerID(address _addr) view returns(uint256)
func (_Contract *ContractCaller) GetStakerID(opts *bind.CallOpts, _addr common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getStakerID", _addr)
	return *ret0, err
}

// GetStakerID is a free data retrieval call binding the contract method 0x63321e27.
//
// Solidity: function getStakerID(address _addr) view returns(uint256)
func (_Contract *ContractSession) GetStakerID(_addr common.Address) (*big.Int, error) {
	return _Contract.Contract.GetStakerID(&_Contract.CallOpts, _addr)
}

// GetStakerID is a free data retrieval call binding the contract method 0x63321e27.
//
// Solidity: function getStakerID(address _addr) view returns(uint256)
func (_Contract *ContractCallerSession) GetStakerID(_addr common.Address) (*big.Int, error) {
	return _Contract.Contract.GetStakerID(&_Contract.CallOpts, _addr)
}

// GetStashedLockupRewards is a free data retrieval call binding the contract method 0xb810e411.
//
// Solidity: function getStashedLockupRewards(address , uint256 ) view returns(uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractCaller) GetStashedLockupRewards(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
}, error) {
	ret := new(struct {
		LockupExtraReward *big.Int
		LockupBaseReward  *big.Int
		UnlockedReward    *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getStashedLockupRewards", arg0, arg1)
	return *ret, err
}

// GetStashedLockupRewards is a free data retrieval call binding the contract method 0xb810e411.
//
// Solidity: function getStashedLockupRewards(address , uint256 ) view returns(uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractSession) GetStashedLockupRewards(arg0 common.Address, arg1 *big.Int) (struct {
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
}, error) {
	return _Contract.Contract.GetStashedLockupRewards(&_Contract.CallOpts, arg0, arg1)
}

// GetStashedLockupRewards is a free data retrieval call binding the contract method 0xb810e411.
//
// Solidity: function getStashedLockupRewards(address , uint256 ) view returns(uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractCallerSession) GetStashedLockupRewards(arg0 common.Address, arg1 *big.Int) (struct {
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
}, error) {
	return _Contract.Contract.GetStashedLockupRewards(&_Contract.CallOpts, arg0, arg1)
}

// GetUnlockedStake is a free data retrieval call binding the contract method 0x12622d0e.
//
// Solidity: function getUnlockedStake(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractCaller) GetUnlockedStake(opts *bind.CallOpts, delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getUnlockedStake", delegator, toValidatorID)
	return *ret0, err
}

// GetUnlockedStake is a free data retrieval call binding the contract method 0x12622d0e.
//
// Solidity: function getUnlockedStake(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractSession) GetUnlockedStake(delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetUnlockedStake(&_Contract.CallOpts, delegator, toValidatorID)
}

// GetUnlockedStake is a free data retrieval call binding the contract method 0x12622d0e.
//
// Solidity: function getUnlockedStake(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetUnlockedStake(delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetUnlockedStake(&_Contract.CallOpts, delegator, toValidatorID)
}

// GetValidator is a free data retrieval call binding the contract method 0xb5d89627.
//
// Solidity: function getValidator(uint256 ) view returns(uint256 status, uint256 deactivatedTime, uint256 deactivatedEpoch, uint256 receivedStake, uint256 createdEpoch, uint256 createdTime, address auth)
func (_Contract *ContractCaller) GetValidator(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Status           *big.Int
	DeactivatedTime  *big.Int
	DeactivatedEpoch *big.Int
	ReceivedStake    *big.Int
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	Auth             common.Address
}, error) {
	ret := new(struct {
		Status           *big.Int
		DeactivatedTime  *big.Int
		DeactivatedEpoch *big.Int
		ReceivedStake    *big.Int
		CreatedEpoch     *big.Int
		CreatedTime      *big.Int
		Auth             common.Address
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getValidator", arg0)
	return *ret, err
}

// GetValidator is a free data retrieval call binding the contract method 0xb5d89627.
//
// Solidity: function getValidator(uint256 ) view returns(uint256 status, uint256 deactivatedTime, uint256 deactivatedEpoch, uint256 receivedStake, uint256 createdEpoch, uint256 createdTime, address auth)
func (_Contract *ContractSession) GetValidator(arg0 *big.Int) (struct {
	Status           *big.Int
	DeactivatedTime  *big.Int
	DeactivatedEpoch *big.Int
	ReceivedStake    *big.Int
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	Auth             common.Address
}, error) {
	return _Contract.Contract.GetValidator(&_Contract.CallOpts, arg0)
}

// GetValidator is a free data retrieval call binding the contract method 0xb5d89627.
//
// Solidity: function getValidator(uint256 ) view returns(uint256 status, uint256 deactivatedTime, uint256 deactivatedEpoch, uint256 receivedStake, uint256 createdEpoch, uint256 createdTime, address auth)
func (_Contract *ContractCallerSession) GetValidator(arg0 *big.Int) (struct {
	Status           *big.Int
	DeactivatedTime  *big.Int
	DeactivatedEpoch *big.Int
	ReceivedStake    *big.Int
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	Auth             common.Address
}, error) {
	return _Contract.Contract.GetValidator(&_Contract.CallOpts, arg0)
}

// GetValidatorID is a free data retrieval call binding the contract method 0x0135b1db.
//
// Solidity: function getValidatorID(address ) view returns(uint256)
func (_Contract *ContractCaller) GetValidatorID(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getValidatorID", arg0)
	return *ret0, err
}

// GetValidatorID is a free data retrieval call binding the contract method 0x0135b1db.
//
// Solidity: function getValidatorID(address ) view returns(uint256)
func (_Contract *ContractSession) GetValidatorID(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.GetValidatorID(&_Contract.CallOpts, arg0)
}

// GetValidatorID is a free data retrieval call binding the contract method 0x0135b1db.
//
// Solidity: function getValidatorID(address ) view returns(uint256)
func (_Contract *ContractCallerSession) GetValidatorID(arg0 common.Address) (*big.Int, error) {
	return _Contract.Contract.GetValidatorID(&_Contract.CallOpts, arg0)
}

// GetValidatorPubkey is a free data retrieval call binding the contract method 0x854873e1.
//
// Solidity: function getValidatorPubkey(uint256 ) view returns(bytes)
func (_Contract *ContractCaller) GetValidatorPubkey(opts *bind.CallOpts, arg0 *big.Int) ([]byte, error) {
	var (
		ret0 = new([]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getValidatorPubkey", arg0)
	return *ret0, err
}

// GetValidatorPubkey is a free data retrieval call binding the contract method 0x854873e1.
//
// Solidity: function getValidatorPubkey(uint256 ) view returns(bytes)
func (_Contract *ContractSession) GetValidatorPubkey(arg0 *big.Int) ([]byte, error) {
	return _Contract.Contract.GetValidatorPubkey(&_Contract.CallOpts, arg0)
}

// GetValidatorPubkey is a free data retrieval call binding the contract method 0x854873e1.
//
// Solidity: function getValidatorPubkey(uint256 ) view returns(bytes)
func (_Contract *ContractCallerSession) GetValidatorPubkey(arg0 *big.Int) ([]byte, error) {
	return _Contract.Contract.GetValidatorPubkey(&_Contract.CallOpts, arg0)
}

// GetWithdrawalRequest is a free data retrieval call binding the contract method 0x1f270152.
//
// Solidity: function getWithdrawalRequest(address , uint256 , uint256 ) view returns(uint256 epoch, uint256 time, uint256 amount)
func (_Contract *ContractCaller) GetWithdrawalRequest(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int, arg2 *big.Int) (struct {
	Epoch  *big.Int
	Time   *big.Int
	Amount *big.Int
}, error) {
	ret := new(struct {
		Epoch  *big.Int
		Time   *big.Int
		Amount *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getWithdrawalRequest", arg0, arg1, arg2)
	return *ret, err
}

// GetWithdrawalRequest is a free data retrieval call binding the contract method 0x1f270152.
//
// Solidity: function getWithdrawalRequest(address , uint256 , uint256 ) view returns(uint256 epoch, uint256 time, uint256 amount)
func (_Contract *ContractSession) GetWithdrawalRequest(arg0 common.Address, arg1 *big.Int, arg2 *big.Int) (struct {
	Epoch  *big.Int
	Time   *big.Int
	Amount *big.Int
}, error) {
	return _Contract.Contract.GetWithdrawalRequest(&_Contract.CallOpts, arg0, arg1, arg2)
}

// GetWithdrawalRequest is a free data retrieval call binding the contract method 0x1f270152.
//
// Solidity: function getWithdrawalRequest(address , uint256 , uint256 ) view returns(uint256 epoch, uint256 time, uint256 amount)
func (_Contract *ContractCallerSession) GetWithdrawalRequest(arg0 common.Address, arg1 *big.Int, arg2 *big.Int) (struct {
	Epoch  *big.Int
	Time   *big.Int
	Amount *big.Int
}, error) {
	return _Contract.Contract.GetWithdrawalRequest(&_Contract.CallOpts, arg0, arg1, arg2)
}

// IsDelegationLockedUp is a free data retrieval call binding the contract method 0xcfd5fa0c.
//
// Solidity: function isDelegationLockedUp(address delegator, uint256 toStakerID) view returns(bool)
func (_Contract *ContractCaller) IsDelegationLockedUp(opts *bind.CallOpts, delegator common.Address, toStakerID *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "isDelegationLockedUp", delegator, toStakerID)
	return *ret0, err
}

// IsDelegationLockedUp is a free data retrieval call binding the contract method 0xcfd5fa0c.
//
// Solidity: function isDelegationLockedUp(address delegator, uint256 toStakerID) view returns(bool)
func (_Contract *ContractSession) IsDelegationLockedUp(delegator common.Address, toStakerID *big.Int) (bool, error) {
	return _Contract.Contract.IsDelegationLockedUp(&_Contract.CallOpts, delegator, toStakerID)
}

// IsDelegationLockedUp is a free data retrieval call binding the contract method 0xcfd5fa0c.
//
// Solidity: function isDelegationLockedUp(address delegator, uint256 toStakerID) view returns(bool)
func (_Contract *ContractCallerSession) IsDelegationLockedUp(delegator common.Address, toStakerID *big.Int) (bool, error) {
	return _Contract.Contract.IsDelegationLockedUp(&_Contract.CallOpts, delegator, toStakerID)
}

// IsLockedUp is a free data retrieval call binding the contract method 0xcfdbb7cd.
//
// Solidity: function isLockedUp(address delegator, uint256 toValidatorID) view returns(bool)
func (_Contract *ContractCaller) IsLockedUp(opts *bind.CallOpts, delegator common.Address, toValidatorID *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "isLockedUp", delegator, toValidatorID)
	return *ret0, err
}

// IsLockedUp is a free data retrieval call binding the contract method 0xcfdbb7cd.
//
// Solidity: function isLockedUp(address delegator, uint256 toValidatorID) view returns(bool)
func (_Contract *ContractSession) IsLockedUp(delegator common.Address, toValidatorID *big.Int) (bool, error) {
	return _Contract.Contract.IsLockedUp(&_Contract.CallOpts, delegator, toValidatorID)
}

// IsLockedUp is a free data retrieval call binding the contract method 0xcfdbb7cd.
//
// Solidity: function isLockedUp(address delegator, uint256 toValidatorID) view returns(bool)
func (_Contract *ContractCallerSession) IsLockedUp(delegator common.Address, toValidatorID *big.Int) (bool, error) {
	return _Contract.Contract.IsLockedUp(&_Contract.CallOpts, delegator, toValidatorID)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_Contract *ContractCaller) IsOwner(opts *bind.CallOpts) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "isOwner")
	return *ret0, err
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_Contract *ContractSession) IsOwner() (bool, error) {
	return _Contract.Contract.IsOwner(&_Contract.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_Contract *ContractCallerSession) IsOwner() (bool, error) {
	return _Contract.Contract.IsOwner(&_Contract.CallOpts)
}

// IsSlashed is a free data retrieval call binding the contract method 0xc3de580e.
//
// Solidity: function isSlashed(uint256 validatorID) view returns(bool)
func (_Contract *ContractCaller) IsSlashed(opts *bind.CallOpts, validatorID *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "isSlashed", validatorID)
	return *ret0, err
}

// IsSlashed is a free data retrieval call binding the contract method 0xc3de580e.
//
// Solidity: function isSlashed(uint256 validatorID) view returns(bool)
func (_Contract *ContractSession) IsSlashed(validatorID *big.Int) (bool, error) {
	return _Contract.Contract.IsSlashed(&_Contract.CallOpts, validatorID)
}

// IsSlashed is a free data retrieval call binding the contract method 0xc3de580e.
//
// Solidity: function isSlashed(uint256 validatorID) view returns(bool)
func (_Contract *ContractCallerSession) IsSlashed(validatorID *big.Int) (bool, error) {
	return _Contract.Contract.IsSlashed(&_Contract.CallOpts, validatorID)
}

// IsStakeLockedUp is a free data retrieval call binding the contract method 0x7f664d87.
//
// Solidity: function isStakeLockedUp(uint256 stakerID) view returns(bool)
func (_Contract *ContractCaller) IsStakeLockedUp(opts *bind.CallOpts, stakerID *big.Int) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "isStakeLockedUp", stakerID)
	return *ret0, err
}

// IsStakeLockedUp is a free data retrieval call binding the contract method 0x7f664d87.
//
// Solidity: function isStakeLockedUp(uint256 stakerID) view returns(bool)
func (_Contract *ContractSession) IsStakeLockedUp(stakerID *big.Int) (bool, error) {
	return _Contract.Contract.IsStakeLockedUp(&_Contract.CallOpts, stakerID)
}

// IsStakeLockedUp is a free data retrieval call binding the contract method 0x7f664d87.
//
// Solidity: function isStakeLockedUp(uint256 stakerID) view returns(bool)
func (_Contract *ContractCallerSession) IsStakeLockedUp(stakerID *big.Int) (bool, error) {
	return _Contract.Contract.IsStakeLockedUp(&_Contract.CallOpts, stakerID)
}

// LastValidatorID is a free data retrieval call binding the contract method 0xc7be95de.
//
// Solidity: function lastValidatorID() view returns(uint256)
func (_Contract *ContractCaller) LastValidatorID(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "lastValidatorID")
	return *ret0, err
}

// LastValidatorID is a free data retrieval call binding the contract method 0xc7be95de.
//
// Solidity: function lastValidatorID() view returns(uint256)
func (_Contract *ContractSession) LastValidatorID() (*big.Int, error) {
	return _Contract.Contract.LastValidatorID(&_Contract.CallOpts)
}

// LastValidatorID is a free data retrieval call binding the contract method 0xc7be95de.
//
// Solidity: function lastValidatorID() view returns(uint256)
func (_Contract *ContractCallerSession) LastValidatorID() (*big.Int, error) {
	return _Contract.Contract.LastValidatorID(&_Contract.CallOpts)
}

// LockedDelegations is a free data retrieval call binding the contract method 0xdd099bb6.
//
// Solidity: function lockedDelegations(address delegator, uint256 toStakerID) view returns(uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractCaller) LockedDelegations(opts *bind.CallOpts, delegator common.Address, toStakerID *big.Int) (struct {
	FromEpoch *big.Int
	EndTime   *big.Int
	Duration  *big.Int
}, error) {
	ret := new(struct {
		FromEpoch *big.Int
		EndTime   *big.Int
		Duration  *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "lockedDelegations", delegator, toStakerID)
	return *ret, err
}

// LockedDelegations is a free data retrieval call binding the contract method 0xdd099bb6.
//
// Solidity: function lockedDelegations(address delegator, uint256 toStakerID) view returns(uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractSession) LockedDelegations(delegator common.Address, toStakerID *big.Int) (struct {
	FromEpoch *big.Int
	EndTime   *big.Int
	Duration  *big.Int
}, error) {
	return _Contract.Contract.LockedDelegations(&_Contract.CallOpts, delegator, toStakerID)
}

// LockedDelegations is a free data retrieval call binding the contract method 0xdd099bb6.
//
// Solidity: function lockedDelegations(address delegator, uint256 toStakerID) view returns(uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractCallerSession) LockedDelegations(delegator common.Address, toStakerID *big.Int) (struct {
	FromEpoch *big.Int
	EndTime   *big.Int
	Duration  *big.Int
}, error) {
	return _Contract.Contract.LockedDelegations(&_Contract.CallOpts, delegator, toStakerID)
}

// LockedStakes is a free data retrieval call binding the contract method 0xdf4f49d4.
//
// Solidity: function lockedStakes(uint256 stakerID) view returns(uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractCaller) LockedStakes(opts *bind.CallOpts, stakerID *big.Int) (struct {
	FromEpoch *big.Int
	EndTime   *big.Int
	Duration  *big.Int
}, error) {
	ret := new(struct {
		FromEpoch *big.Int
		EndTime   *big.Int
		Duration  *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "lockedStakes", stakerID)
	return *ret, err
}

// LockedStakes is a free data retrieval call binding the contract method 0xdf4f49d4.
//
// Solidity: function lockedStakes(uint256 stakerID) view returns(uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractSession) LockedStakes(stakerID *big.Int) (struct {
	FromEpoch *big.Int
	EndTime   *big.Int
	Duration  *big.Int
}, error) {
	return _Contract.Contract.LockedStakes(&_Contract.CallOpts, stakerID)
}

// LockedStakes is a free data retrieval call binding the contract method 0xdf4f49d4.
//
// Solidity: function lockedStakes(uint256 stakerID) view returns(uint256 fromEpoch, uint256 endTime, uint256 duration)
func (_Contract *ContractCallerSession) LockedStakes(stakerID *big.Int) (struct {
	FromEpoch *big.Int
	EndTime   *big.Int
	Duration  *big.Int
}, error) {
	return _Contract.Contract.LockedStakes(&_Contract.CallOpts, stakerID)
}

// MaxDelegatedRatio is a free data retrieval call binding the contract method 0x2265f284.
//
// Solidity: function maxDelegatedRatio() pure returns(uint256)
func (_Contract *ContractCaller) MaxDelegatedRatio(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "maxDelegatedRatio")
	return *ret0, err
}

// MaxDelegatedRatio is a free data retrieval call binding the contract method 0x2265f284.
//
// Solidity: function maxDelegatedRatio() pure returns(uint256)
func (_Contract *ContractSession) MaxDelegatedRatio() (*big.Int, error) {
	return _Contract.Contract.MaxDelegatedRatio(&_Contract.CallOpts)
}

// MaxDelegatedRatio is a free data retrieval call binding the contract method 0x2265f284.
//
// Solidity: function maxDelegatedRatio() pure returns(uint256)
func (_Contract *ContractCallerSession) MaxDelegatedRatio() (*big.Int, error) {
	return _Contract.Contract.MaxDelegatedRatio(&_Contract.CallOpts)
}

// MaxLockupDuration is a free data retrieval call binding the contract method 0x0d4955e3.
//
// Solidity: function maxLockupDuration() pure returns(uint256)
func (_Contract *ContractCaller) MaxLockupDuration(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "maxLockupDuration")
	return *ret0, err
}

// MaxLockupDuration is a free data retrieval call binding the contract method 0x0d4955e3.
//
// Solidity: function maxLockupDuration() pure returns(uint256)
func (_Contract *ContractSession) MaxLockupDuration() (*big.Int, error) {
	return _Contract.Contract.MaxLockupDuration(&_Contract.CallOpts)
}

// MaxLockupDuration is a free data retrieval call binding the contract method 0x0d4955e3.
//
// Solidity: function maxLockupDuration() pure returns(uint256)
func (_Contract *ContractCallerSession) MaxLockupDuration() (*big.Int, error) {
	return _Contract.Contract.MaxLockupDuration(&_Contract.CallOpts)
}

// MinDelegation is a free data retrieval call binding the contract method 0x02985992.
//
// Solidity: function minDelegation() pure returns(uint256)
func (_Contract *ContractCaller) MinDelegation(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minDelegation")
	return *ret0, err
}

// MinDelegation is a free data retrieval call binding the contract method 0x02985992.
//
// Solidity: function minDelegation() pure returns(uint256)
func (_Contract *ContractSession) MinDelegation() (*big.Int, error) {
	return _Contract.Contract.MinDelegation(&_Contract.CallOpts)
}

// MinDelegation is a free data retrieval call binding the contract method 0x02985992.
//
// Solidity: function minDelegation() pure returns(uint256)
func (_Contract *ContractCallerSession) MinDelegation() (*big.Int, error) {
	return _Contract.Contract.MinDelegation(&_Contract.CallOpts)
}

// MinDelegationDecrease is a free data retrieval call binding the contract method 0xcb1c4e67.
//
// Solidity: function minDelegationDecrease() pure returns(uint256)
func (_Contract *ContractCaller) MinDelegationDecrease(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minDelegationDecrease")
	return *ret0, err
}

// MinDelegationDecrease is a free data retrieval call binding the contract method 0xcb1c4e67.
//
// Solidity: function minDelegationDecrease() pure returns(uint256)
func (_Contract *ContractSession) MinDelegationDecrease() (*big.Int, error) {
	return _Contract.Contract.MinDelegationDecrease(&_Contract.CallOpts)
}

// MinDelegationDecrease is a free data retrieval call binding the contract method 0xcb1c4e67.
//
// Solidity: function minDelegationDecrease() pure returns(uint256)
func (_Contract *ContractCallerSession) MinDelegationDecrease() (*big.Int, error) {
	return _Contract.Contract.MinDelegationDecrease(&_Contract.CallOpts)
}

// MinDelegationIncrease is a free data retrieval call binding the contract method 0x60c7e37f.
//
// Solidity: function minDelegationIncrease() pure returns(uint256)
func (_Contract *ContractCaller) MinDelegationIncrease(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minDelegationIncrease")
	return *ret0, err
}

// MinDelegationIncrease is a free data retrieval call binding the contract method 0x60c7e37f.
//
// Solidity: function minDelegationIncrease() pure returns(uint256)
func (_Contract *ContractSession) MinDelegationIncrease() (*big.Int, error) {
	return _Contract.Contract.MinDelegationIncrease(&_Contract.CallOpts)
}

// MinDelegationIncrease is a free data retrieval call binding the contract method 0x60c7e37f.
//
// Solidity: function minDelegationIncrease() pure returns(uint256)
func (_Contract *ContractCallerSession) MinDelegationIncrease() (*big.Int, error) {
	return _Contract.Contract.MinDelegationIncrease(&_Contract.CallOpts)
}

// MinLockupDuration is a free data retrieval call binding the contract method 0x0d7b2609.
//
// Solidity: function minLockupDuration() pure returns(uint256)
func (_Contract *ContractCaller) MinLockupDuration(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minLockupDuration")
	return *ret0, err
}

// MinLockupDuration is a free data retrieval call binding the contract method 0x0d7b2609.
//
// Solidity: function minLockupDuration() pure returns(uint256)
func (_Contract *ContractSession) MinLockupDuration() (*big.Int, error) {
	return _Contract.Contract.MinLockupDuration(&_Contract.CallOpts)
}

// MinLockupDuration is a free data retrieval call binding the contract method 0x0d7b2609.
//
// Solidity: function minLockupDuration() pure returns(uint256)
func (_Contract *ContractCallerSession) MinLockupDuration() (*big.Int, error) {
	return _Contract.Contract.MinLockupDuration(&_Contract.CallOpts)
}

// MinSelfStake is a free data retrieval call binding the contract method 0xc5f530af.
//
// Solidity: function minSelfStake() pure returns(uint256)
func (_Contract *ContractCaller) MinSelfStake(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minSelfStake")
	return *ret0, err
}

// MinSelfStake is a free data retrieval call binding the contract method 0xc5f530af.
//
// Solidity: function minSelfStake() pure returns(uint256)
func (_Contract *ContractSession) MinSelfStake() (*big.Int, error) {
	return _Contract.Contract.MinSelfStake(&_Contract.CallOpts)
}

// MinSelfStake is a free data retrieval call binding the contract method 0xc5f530af.
//
// Solidity: function minSelfStake() pure returns(uint256)
func (_Contract *ContractCallerSession) MinSelfStake() (*big.Int, error) {
	return _Contract.Contract.MinSelfStake(&_Contract.CallOpts)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() pure returns(uint256)
func (_Contract *ContractCaller) MinStake(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minStake")
	return *ret0, err
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() pure returns(uint256)
func (_Contract *ContractSession) MinStake() (*big.Int, error) {
	return _Contract.Contract.MinStake(&_Contract.CallOpts)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() pure returns(uint256)
func (_Contract *ContractCallerSession) MinStake() (*big.Int, error) {
	return _Contract.Contract.MinStake(&_Contract.CallOpts)
}

// MinStakeDecrease is a free data retrieval call binding the contract method 0x19ddb54f.
//
// Solidity: function minStakeDecrease() pure returns(uint256)
func (_Contract *ContractCaller) MinStakeDecrease(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minStakeDecrease")
	return *ret0, err
}

// MinStakeDecrease is a free data retrieval call binding the contract method 0x19ddb54f.
//
// Solidity: function minStakeDecrease() pure returns(uint256)
func (_Contract *ContractSession) MinStakeDecrease() (*big.Int, error) {
	return _Contract.Contract.MinStakeDecrease(&_Contract.CallOpts)
}

// MinStakeDecrease is a free data retrieval call binding the contract method 0x19ddb54f.
//
// Solidity: function minStakeDecrease() pure returns(uint256)
func (_Contract *ContractCallerSession) MinStakeDecrease() (*big.Int, error) {
	return _Contract.Contract.MinStakeDecrease(&_Contract.CallOpts)
}

// MinStakeIncrease is a free data retrieval call binding the contract method 0xc4b5dd7e.
//
// Solidity: function minStakeIncrease() pure returns(uint256)
func (_Contract *ContractCaller) MinStakeIncrease(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "minStakeIncrease")
	return *ret0, err
}

// MinStakeIncrease is a free data retrieval call binding the contract method 0xc4b5dd7e.
//
// Solidity: function minStakeIncrease() pure returns(uint256)
func (_Contract *ContractSession) MinStakeIncrease() (*big.Int, error) {
	return _Contract.Contract.MinStakeIncrease(&_Contract.CallOpts)
}

// MinStakeIncrease is a free data retrieval call binding the contract method 0xc4b5dd7e.
//
// Solidity: function minStakeIncrease() pure returns(uint256)
func (_Contract *ContractCallerSession) MinStakeIncrease() (*big.Int, error) {
	return _Contract.Contract.MinStakeIncrease(&_Contract.CallOpts)
}

// OfflinePenaltyThreshold is a free data retrieval call binding the contract method 0x2cedb097.
//
// Solidity: function offlinePenaltyThreshold() view returns(uint256 blocksNum, uint256 time)
func (_Contract *ContractCaller) OfflinePenaltyThreshold(opts *bind.CallOpts) (struct {
	BlocksNum *big.Int
	Time      *big.Int
}, error) {
	ret := new(struct {
		BlocksNum *big.Int
		Time      *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "offlinePenaltyThreshold")
	return *ret, err
}

// OfflinePenaltyThreshold is a free data retrieval call binding the contract method 0x2cedb097.
//
// Solidity: function offlinePenaltyThreshold() view returns(uint256 blocksNum, uint256 time)
func (_Contract *ContractSession) OfflinePenaltyThreshold() (struct {
	BlocksNum *big.Int
	Time      *big.Int
}, error) {
	return _Contract.Contract.OfflinePenaltyThreshold(&_Contract.CallOpts)
}

// OfflinePenaltyThreshold is a free data retrieval call binding the contract method 0x2cedb097.
//
// Solidity: function offlinePenaltyThreshold() view returns(uint256 blocksNum, uint256 time)
func (_Contract *ContractCallerSession) OfflinePenaltyThreshold() (struct {
	BlocksNum *big.Int
	Time      *big.Int
}, error) {
	return _Contract.Contract.OfflinePenaltyThreshold(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "owner")
	return *ret0, err
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCallerSession) Owner() (common.Address, error) {
	return _Contract.Contract.Owner(&_Contract.CallOpts)
}

// PendingRewards is a free data retrieval call binding the contract method 0x6099ecb2.
//
// Solidity: function pendingRewards(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractCaller) PendingRewards(opts *bind.CallOpts, delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "pendingRewards", delegator, toValidatorID)
	return *ret0, err
}

// PendingRewards is a free data retrieval call binding the contract method 0x6099ecb2.
//
// Solidity: function pendingRewards(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractSession) PendingRewards(delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.PendingRewards(&_Contract.CallOpts, delegator, toValidatorID)
}

// PendingRewards is a free data retrieval call binding the contract method 0x6099ecb2.
//
// Solidity: function pendingRewards(address delegator, uint256 toValidatorID) view returns(uint256)
func (_Contract *ContractCallerSession) PendingRewards(delegator common.Address, toValidatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.PendingRewards(&_Contract.CallOpts, delegator, toValidatorID)
}

// RewardsStash is a free data retrieval call binding the contract method 0x6f498663.
//
// Solidity: function rewardsStash(address delegator, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) RewardsStash(opts *bind.CallOpts, delegator common.Address, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "rewardsStash", delegator, validatorID)
	return *ret0, err
}

// RewardsStash is a free data retrieval call binding the contract method 0x6f498663.
//
// Solidity: function rewardsStash(address delegator, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) RewardsStash(delegator common.Address, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.RewardsStash(&_Contract.CallOpts, delegator, validatorID)
}

// RewardsStash is a free data retrieval call binding the contract method 0x6f498663.
//
// Solidity: function rewardsStash(address delegator, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) RewardsStash(delegator common.Address, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.RewardsStash(&_Contract.CallOpts, delegator, validatorID)
}

// SlashingRefundRatio is a free data retrieval call binding the contract method 0xc65ee0e1.
//
// Solidity: function slashingRefundRatio(uint256 ) view returns(uint256)
func (_Contract *ContractCaller) SlashingRefundRatio(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "slashingRefundRatio", arg0)
	return *ret0, err
}

// SlashingRefundRatio is a free data retrieval call binding the contract method 0xc65ee0e1.
//
// Solidity: function slashingRefundRatio(uint256 ) view returns(uint256)
func (_Contract *ContractSession) SlashingRefundRatio(arg0 *big.Int) (*big.Int, error) {
	return _Contract.Contract.SlashingRefundRatio(&_Contract.CallOpts, arg0)
}

// SlashingRefundRatio is a free data retrieval call binding the contract method 0xc65ee0e1.
//
// Solidity: function slashingRefundRatio(uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) SlashingRefundRatio(arg0 *big.Int) (*big.Int, error) {
	return _Contract.Contract.SlashingRefundRatio(&_Contract.CallOpts, arg0)
}

// StakeLockPeriodEpochs is a free data retrieval call binding the contract method 0x54d77ed2.
//
// Solidity: function stakeLockPeriodEpochs() pure returns(uint256)
func (_Contract *ContractCaller) StakeLockPeriodEpochs(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "stakeLockPeriodEpochs")
	return *ret0, err
}

// StakeLockPeriodEpochs is a free data retrieval call binding the contract method 0x54d77ed2.
//
// Solidity: function stakeLockPeriodEpochs() pure returns(uint256)
func (_Contract *ContractSession) StakeLockPeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.StakeLockPeriodEpochs(&_Contract.CallOpts)
}

// StakeLockPeriodEpochs is a free data retrieval call binding the contract method 0x54d77ed2.
//
// Solidity: function stakeLockPeriodEpochs() pure returns(uint256)
func (_Contract *ContractCallerSession) StakeLockPeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.StakeLockPeriodEpochs(&_Contract.CallOpts)
}

// StakeLockPeriodTime is a free data retrieval call binding the contract method 0x3fee10a8.
//
// Solidity: function stakeLockPeriodTime() pure returns(uint256)
func (_Contract *ContractCaller) StakeLockPeriodTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "stakeLockPeriodTime")
	return *ret0, err
}

// StakeLockPeriodTime is a free data retrieval call binding the contract method 0x3fee10a8.
//
// Solidity: function stakeLockPeriodTime() pure returns(uint256)
func (_Contract *ContractSession) StakeLockPeriodTime() (*big.Int, error) {
	return _Contract.Contract.StakeLockPeriodTime(&_Contract.CallOpts)
}

// StakeLockPeriodTime is a free data retrieval call binding the contract method 0x3fee10a8.
//
// Solidity: function stakeLockPeriodTime() pure returns(uint256)
func (_Contract *ContractCallerSession) StakeLockPeriodTime() (*big.Int, error) {
	return _Contract.Contract.StakeLockPeriodTime(&_Contract.CallOpts)
}

// StakeTotalAmount is a free data retrieval call binding the contract method 0x3d0317fe.
//
// Solidity: function stakeTotalAmount() view returns(uint256)
func (_Contract *ContractCaller) StakeTotalAmount(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "stakeTotalAmount")
	return *ret0, err
}

// StakeTotalAmount is a free data retrieval call binding the contract method 0x3d0317fe.
//
// Solidity: function stakeTotalAmount() view returns(uint256)
func (_Contract *ContractSession) StakeTotalAmount() (*big.Int, error) {
	return _Contract.Contract.StakeTotalAmount(&_Contract.CallOpts)
}

// StakeTotalAmount is a free data retrieval call binding the contract method 0x3d0317fe.
//
// Solidity: function stakeTotalAmount() view returns(uint256)
func (_Contract *ContractCallerSession) StakeTotalAmount() (*big.Int, error) {
	return _Contract.Contract.StakeTotalAmount(&_Contract.CallOpts)
}

// Stakers is a free data retrieval call binding the contract method 0xfd5e6dd1.
//
// Solidity: function stakers(uint256 _stakerID) view returns(uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime, uint256 stakeAmount, uint256 paidUntilEpoch, uint256 delegatedMe, address dagAddress, address sfcAddress)
func (_Contract *ContractCaller) Stakers(opts *bind.CallOpts, _stakerID *big.Int) (struct {
	Status           *big.Int
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	StakeAmount      *big.Int
	PaidUntilEpoch   *big.Int
	DelegatedMe      *big.Int
	DagAddress       common.Address
	SfcAddress       common.Address
}, error) {
	ret := new(struct {
		Status           *big.Int
		CreatedEpoch     *big.Int
		CreatedTime      *big.Int
		DeactivatedEpoch *big.Int
		DeactivatedTime  *big.Int
		StakeAmount      *big.Int
		PaidUntilEpoch   *big.Int
		DelegatedMe      *big.Int
		DagAddress       common.Address
		SfcAddress       common.Address
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "stakers", _stakerID)
	return *ret, err
}

// Stakers is a free data retrieval call binding the contract method 0xfd5e6dd1.
//
// Solidity: function stakers(uint256 _stakerID) view returns(uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime, uint256 stakeAmount, uint256 paidUntilEpoch, uint256 delegatedMe, address dagAddress, address sfcAddress)
func (_Contract *ContractSession) Stakers(_stakerID *big.Int) (struct {
	Status           *big.Int
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	StakeAmount      *big.Int
	PaidUntilEpoch   *big.Int
	DelegatedMe      *big.Int
	DagAddress       common.Address
	SfcAddress       common.Address
}, error) {
	return _Contract.Contract.Stakers(&_Contract.CallOpts, _stakerID)
}

// Stakers is a free data retrieval call binding the contract method 0xfd5e6dd1.
//
// Solidity: function stakers(uint256 _stakerID) view returns(uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime, uint256 stakeAmount, uint256 paidUntilEpoch, uint256 delegatedMe, address dagAddress, address sfcAddress)
func (_Contract *ContractCallerSession) Stakers(_stakerID *big.Int) (struct {
	Status           *big.Int
	CreatedEpoch     *big.Int
	CreatedTime      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	StakeAmount      *big.Int
	PaidUntilEpoch   *big.Int
	DelegatedMe      *big.Int
	DagAddress       common.Address
	SfcAddress       common.Address
}, error) {
	return _Contract.Contract.Stakers(&_Contract.CallOpts, _stakerID)
}

// StakersLastID is a free data retrieval call binding the contract method 0x81d9dc7a.
//
// Solidity: function stakersLastID() view returns(uint256)
func (_Contract *ContractCaller) StakersLastID(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "stakersLastID")
	return *ret0, err
}

// StakersLastID is a free data retrieval call binding the contract method 0x81d9dc7a.
//
// Solidity: function stakersLastID() view returns(uint256)
func (_Contract *ContractSession) StakersLastID() (*big.Int, error) {
	return _Contract.Contract.StakersLastID(&_Contract.CallOpts)
}

// StakersLastID is a free data retrieval call binding the contract method 0x81d9dc7a.
//
// Solidity: function stakersLastID() view returns(uint256)
func (_Contract *ContractCallerSession) StakersLastID() (*big.Int, error) {
	return _Contract.Contract.StakersLastID(&_Contract.CallOpts)
}

// StakersNum is a free data retrieval call binding the contract method 0x08728f6e.
//
// Solidity: function stakersNum() view returns(uint256)
func (_Contract *ContractCaller) StakersNum(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "stakersNum")
	return *ret0, err
}

// StakersNum is a free data retrieval call binding the contract method 0x08728f6e.
//
// Solidity: function stakersNum() view returns(uint256)
func (_Contract *ContractSession) StakersNum() (*big.Int, error) {
	return _Contract.Contract.StakersNum(&_Contract.CallOpts)
}

// StakersNum is a free data retrieval call binding the contract method 0x08728f6e.
//
// Solidity: function stakersNum() view returns(uint256)
func (_Contract *ContractCallerSession) StakersNum() (*big.Int, error) {
	return _Contract.Contract.StakersNum(&_Contract.CallOpts)
}

// StashedRewardsUntilEpoch is a free data retrieval call binding the contract method 0xa86a056f.
//
// Solidity: function stashedRewardsUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) StashedRewardsUntilEpoch(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "stashedRewardsUntilEpoch", arg0, arg1)
	return *ret0, err
}

// StashedRewardsUntilEpoch is a free data retrieval call binding the contract method 0xa86a056f.
//
// Solidity: function stashedRewardsUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractSession) StashedRewardsUntilEpoch(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.StashedRewardsUntilEpoch(&_Contract.CallOpts, arg0, arg1)
}

// StashedRewardsUntilEpoch is a free data retrieval call binding the contract method 0xa86a056f.
//
// Solidity: function stashedRewardsUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) StashedRewardsUntilEpoch(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.StashedRewardsUntilEpoch(&_Contract.CallOpts, arg0, arg1)
}

// TotalActiveStake is a free data retrieval call binding the contract method 0x28f73148.
//
// Solidity: function totalActiveStake() view returns(uint256)
func (_Contract *ContractCaller) TotalActiveStake(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "totalActiveStake")
	return *ret0, err
}

// TotalActiveStake is a free data retrieval call binding the contract method 0x28f73148.
//
// Solidity: function totalActiveStake() view returns(uint256)
func (_Contract *ContractSession) TotalActiveStake() (*big.Int, error) {
	return _Contract.Contract.TotalActiveStake(&_Contract.CallOpts)
}

// TotalActiveStake is a free data retrieval call binding the contract method 0x28f73148.
//
// Solidity: function totalActiveStake() view returns(uint256)
func (_Contract *ContractCallerSession) TotalActiveStake() (*big.Int, error) {
	return _Contract.Contract.TotalActiveStake(&_Contract.CallOpts)
}

// TotalSlashedStake is a free data retrieval call binding the contract method 0x5fab23a8.
//
// Solidity: function totalSlashedStake() view returns(uint256)
func (_Contract *ContractCaller) TotalSlashedStake(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "totalSlashedStake")
	return *ret0, err
}

// TotalSlashedStake is a free data retrieval call binding the contract method 0x5fab23a8.
//
// Solidity: function totalSlashedStake() view returns(uint256)
func (_Contract *ContractSession) TotalSlashedStake() (*big.Int, error) {
	return _Contract.Contract.TotalSlashedStake(&_Contract.CallOpts)
}

// TotalSlashedStake is a free data retrieval call binding the contract method 0x5fab23a8.
//
// Solidity: function totalSlashedStake() view returns(uint256)
func (_Contract *ContractCallerSession) TotalSlashedStake() (*big.Int, error) {
	return _Contract.Contract.TotalSlashedStake(&_Contract.CallOpts)
}

// TotalStake is a free data retrieval call binding the contract method 0x8b0e9f3f.
//
// Solidity: function totalStake() view returns(uint256)
func (_Contract *ContractCaller) TotalStake(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "totalStake")
	return *ret0, err
}

// TotalStake is a free data retrieval call binding the contract method 0x8b0e9f3f.
//
// Solidity: function totalStake() view returns(uint256)
func (_Contract *ContractSession) TotalStake() (*big.Int, error) {
	return _Contract.Contract.TotalStake(&_Contract.CallOpts)
}

// TotalStake is a free data retrieval call binding the contract method 0x8b0e9f3f.
//
// Solidity: function totalStake() view returns(uint256)
func (_Contract *ContractCallerSession) TotalStake() (*big.Int, error) {
	return _Contract.Contract.TotalStake(&_Contract.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Contract *ContractCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Contract *ContractSession) TotalSupply() (*big.Int, error) {
	return _Contract.Contract.TotalSupply(&_Contract.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Contract *ContractCallerSession) TotalSupply() (*big.Int, error) {
	return _Contract.Contract.TotalSupply(&_Contract.CallOpts)
}

// UnlockedRewardRatio is a free data retrieval call binding the contract method 0x5e2308d2.
//
// Solidity: function unlockedRewardRatio() pure returns(uint256)
func (_Contract *ContractCaller) UnlockedRewardRatio(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "unlockedRewardRatio")
	return *ret0, err
}

// UnlockedRewardRatio is a free data retrieval call binding the contract method 0x5e2308d2.
//
// Solidity: function unlockedRewardRatio() pure returns(uint256)
func (_Contract *ContractSession) UnlockedRewardRatio() (*big.Int, error) {
	return _Contract.Contract.UnlockedRewardRatio(&_Contract.CallOpts)
}

// UnlockedRewardRatio is a free data retrieval call binding the contract method 0x5e2308d2.
//
// Solidity: function unlockedRewardRatio() pure returns(uint256)
func (_Contract *ContractCallerSession) UnlockedRewardRatio() (*big.Int, error) {
	return _Contract.Contract.UnlockedRewardRatio(&_Contract.CallOpts)
}

// ValidatorCommission is a free data retrieval call binding the contract method 0xa7786515.
//
// Solidity: function validatorCommission() pure returns(uint256)
func (_Contract *ContractCaller) ValidatorCommission(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "validatorCommission")
	return *ret0, err
}

// ValidatorCommission is a free data retrieval call binding the contract method 0xa7786515.
//
// Solidity: function validatorCommission() pure returns(uint256)
func (_Contract *ContractSession) ValidatorCommission() (*big.Int, error) {
	return _Contract.Contract.ValidatorCommission(&_Contract.CallOpts)
}

// ValidatorCommission is a free data retrieval call binding the contract method 0xa7786515.
//
// Solidity: function validatorCommission() pure returns(uint256)
func (_Contract *ContractCallerSession) ValidatorCommission() (*big.Int, error) {
	return _Contract.Contract.ValidatorCommission(&_Contract.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(bytes3)
func (_Contract *ContractCaller) Version(opts *bind.CallOpts) ([3]byte, error) {
	var (
		ret0 = new([3]byte)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "version")
	return *ret0, err
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(bytes3)
func (_Contract *ContractSession) Version() ([3]byte, error) {
	return _Contract.Contract.Version(&_Contract.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() pure returns(bytes3)
func (_Contract *ContractCallerSession) Version() ([3]byte, error) {
	return _Contract.Contract.Version(&_Contract.CallOpts)
}

// WithdrawalPeriodEpochs is a free data retrieval call binding the contract method 0x650acd66.
//
// Solidity: function withdrawalPeriodEpochs() pure returns(uint256)
func (_Contract *ContractCaller) WithdrawalPeriodEpochs(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "withdrawalPeriodEpochs")
	return *ret0, err
}

// WithdrawalPeriodEpochs is a free data retrieval call binding the contract method 0x650acd66.
//
// Solidity: function withdrawalPeriodEpochs() pure returns(uint256)
func (_Contract *ContractSession) WithdrawalPeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.WithdrawalPeriodEpochs(&_Contract.CallOpts)
}

// WithdrawalPeriodEpochs is a free data retrieval call binding the contract method 0x650acd66.
//
// Solidity: function withdrawalPeriodEpochs() pure returns(uint256)
func (_Contract *ContractCallerSession) WithdrawalPeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.WithdrawalPeriodEpochs(&_Contract.CallOpts)
}

// WithdrawalPeriodTime is a free data retrieval call binding the contract method 0xb82b8427.
//
// Solidity: function withdrawalPeriodTime() pure returns(uint256)
func (_Contract *ContractCaller) WithdrawalPeriodTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "withdrawalPeriodTime")
	return *ret0, err
}

// WithdrawalPeriodTime is a free data retrieval call binding the contract method 0xb82b8427.
//
// Solidity: function withdrawalPeriodTime() pure returns(uint256)
func (_Contract *ContractSession) WithdrawalPeriodTime() (*big.Int, error) {
	return _Contract.Contract.WithdrawalPeriodTime(&_Contract.CallOpts)
}

// WithdrawalPeriodTime is a free data retrieval call binding the contract method 0xb82b8427.
//
// Solidity: function withdrawalPeriodTime() pure returns(uint256)
func (_Contract *ContractCallerSession) WithdrawalPeriodTime() (*big.Int, error) {
	return _Contract.Contract.WithdrawalPeriodTime(&_Contract.CallOpts)
}

// SyncValidator is a paid mutator transaction binding the contract method 0xcc8343aa.
//
// Solidity: function _syncValidator(uint256 validatorID, bool syncPubkey) returns()
func (_Contract *ContractTransactor) SyncValidator(opts *bind.TransactOpts, validatorID *big.Int, syncPubkey bool) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "_syncValidator", validatorID, syncPubkey)
}

// SyncValidator is a paid mutator transaction binding the contract method 0xcc8343aa.
//
// Solidity: function _syncValidator(uint256 validatorID, bool syncPubkey) returns()
func (_Contract *ContractSession) SyncValidator(validatorID *big.Int, syncPubkey bool) (*types.Transaction, error) {
	return _Contract.Contract.SyncValidator(&_Contract.TransactOpts, validatorID, syncPubkey)
}

// SyncValidator is a paid mutator transaction binding the contract method 0xcc8343aa.
//
// Solidity: function _syncValidator(uint256 validatorID, bool syncPubkey) returns()
func (_Contract *ContractTransactorSession) SyncValidator(validatorID *big.Int, syncPubkey bool) (*types.Transaction, error) {
	return _Contract.Contract.SyncValidator(&_Contract.TransactOpts, validatorID, syncPubkey)
}

// ClaimDelegationCompoundRewards is a paid mutator transaction binding the contract method 0xdc599bb1.
//
// Solidity: function claimDelegationCompoundRewards(uint256 , uint256 toStakerID) returns()
func (_Contract *ContractTransactor) ClaimDelegationCompoundRewards(opts *bind.TransactOpts, arg0 *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "claimDelegationCompoundRewards", arg0, toStakerID)
}

// ClaimDelegationCompoundRewards is a paid mutator transaction binding the contract method 0xdc599bb1.
//
// Solidity: function claimDelegationCompoundRewards(uint256 , uint256 toStakerID) returns()
func (_Contract *ContractSession) ClaimDelegationCompoundRewards(arg0 *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimDelegationCompoundRewards(&_Contract.TransactOpts, arg0, toStakerID)
}

// ClaimDelegationCompoundRewards is a paid mutator transaction binding the contract method 0xdc599bb1.
//
// Solidity: function claimDelegationCompoundRewards(uint256 , uint256 toStakerID) returns()
func (_Contract *ContractTransactorSession) ClaimDelegationCompoundRewards(arg0 *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimDelegationCompoundRewards(&_Contract.TransactOpts, arg0, toStakerID)
}

// ClaimDelegationRewards is a paid mutator transaction binding the contract method 0xf99837e6.
//
// Solidity: function claimDelegationRewards(uint256 , uint256 toStakerID) returns()
func (_Contract *ContractTransactor) ClaimDelegationRewards(opts *bind.TransactOpts, arg0 *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "claimDelegationRewards", arg0, toStakerID)
}

// ClaimDelegationRewards is a paid mutator transaction binding the contract method 0xf99837e6.
//
// Solidity: function claimDelegationRewards(uint256 , uint256 toStakerID) returns()
func (_Contract *ContractSession) ClaimDelegationRewards(arg0 *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimDelegationRewards(&_Contract.TransactOpts, arg0, toStakerID)
}

// ClaimDelegationRewards is a paid mutator transaction binding the contract method 0xf99837e6.
//
// Solidity: function claimDelegationRewards(uint256 , uint256 toStakerID) returns()
func (_Contract *ContractTransactorSession) ClaimDelegationRewards(arg0 *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimDelegationRewards(&_Contract.TransactOpts, arg0, toStakerID)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x0962ef79.
//
// Solidity: function claimRewards(uint256 toValidatorID) returns()
func (_Contract *ContractTransactor) ClaimRewards(opts *bind.TransactOpts, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "claimRewards", toValidatorID)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x0962ef79.
//
// Solidity: function claimRewards(uint256 toValidatorID) returns()
func (_Contract *ContractSession) ClaimRewards(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimRewards(&_Contract.TransactOpts, toValidatorID)
}

// ClaimRewards is a paid mutator transaction binding the contract method 0x0962ef79.
//
// Solidity: function claimRewards(uint256 toValidatorID) returns()
func (_Contract *ContractTransactorSession) ClaimRewards(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimRewards(&_Contract.TransactOpts, toValidatorID)
}

// ClaimValidatorCompoundRewards is a paid mutator transaction binding the contract method 0xcda5826a.
//
// Solidity: function claimValidatorCompoundRewards(uint256 ) returns()
func (_Contract *ContractTransactor) ClaimValidatorCompoundRewards(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "claimValidatorCompoundRewards", arg0)
}

// ClaimValidatorCompoundRewards is a paid mutator transaction binding the contract method 0xcda5826a.
//
// Solidity: function claimValidatorCompoundRewards(uint256 ) returns()
func (_Contract *ContractSession) ClaimValidatorCompoundRewards(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimValidatorCompoundRewards(&_Contract.TransactOpts, arg0)
}

// ClaimValidatorCompoundRewards is a paid mutator transaction binding the contract method 0xcda5826a.
//
// Solidity: function claimValidatorCompoundRewards(uint256 ) returns()
func (_Contract *ContractTransactorSession) ClaimValidatorCompoundRewards(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimValidatorCompoundRewards(&_Contract.TransactOpts, arg0)
}

// ClaimValidatorRewards is a paid mutator transaction binding the contract method 0x295cccba.
//
// Solidity: function claimValidatorRewards(uint256 ) returns()
func (_Contract *ContractTransactor) ClaimValidatorRewards(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "claimValidatorRewards", arg0)
}

// ClaimValidatorRewards is a paid mutator transaction binding the contract method 0x295cccba.
//
// Solidity: function claimValidatorRewards(uint256 ) returns()
func (_Contract *ContractSession) ClaimValidatorRewards(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimValidatorRewards(&_Contract.TransactOpts, arg0)
}

// ClaimValidatorRewards is a paid mutator transaction binding the contract method 0x295cccba.
//
// Solidity: function claimValidatorRewards(uint256 ) returns()
func (_Contract *ContractTransactorSession) ClaimValidatorRewards(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.ClaimValidatorRewards(&_Contract.TransactOpts, arg0)
}

// CreateDelegation is a paid mutator transaction binding the contract method 0xc312eb07.
//
// Solidity: function createDelegation(uint256 toValidatorID) payable returns()
func (_Contract *ContractTransactor) CreateDelegation(opts *bind.TransactOpts, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "createDelegation", toValidatorID)
}

// CreateDelegation is a paid mutator transaction binding the contract method 0xc312eb07.
//
// Solidity: function createDelegation(uint256 toValidatorID) payable returns()
func (_Contract *ContractSession) CreateDelegation(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.CreateDelegation(&_Contract.TransactOpts, toValidatorID)
}

// CreateDelegation is a paid mutator transaction binding the contract method 0xc312eb07.
//
// Solidity: function createDelegation(uint256 toValidatorID) payable returns()
func (_Contract *ContractTransactorSession) CreateDelegation(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.CreateDelegation(&_Contract.TransactOpts, toValidatorID)
}

// CreateValidator is a paid mutator transaction binding the contract method 0xa5a470ad.
//
// Solidity: function createValidator(bytes pubkey) payable returns()
func (_Contract *ContractTransactor) CreateValidator(opts *bind.TransactOpts, pubkey []byte) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "createValidator", pubkey)
}

// CreateValidator is a paid mutator transaction binding the contract method 0xa5a470ad.
//
// Solidity: function createValidator(bytes pubkey) payable returns()
func (_Contract *ContractSession) CreateValidator(pubkey []byte) (*types.Transaction, error) {
	return _Contract.Contract.CreateValidator(&_Contract.TransactOpts, pubkey)
}

// CreateValidator is a paid mutator transaction binding the contract method 0xa5a470ad.
//
// Solidity: function createValidator(bytes pubkey) payable returns()
func (_Contract *ContractTransactorSession) CreateValidator(pubkey []byte) (*types.Transaction, error) {
	return _Contract.Contract.CreateValidator(&_Contract.TransactOpts, pubkey)
}

// DeactivateValidator is a paid mutator transaction binding the contract method 0x1e702f83.
//
// Solidity: function deactivateValidator(uint256 validatorID, uint256 status) returns()
func (_Contract *ContractTransactor) DeactivateValidator(opts *bind.TransactOpts, validatorID *big.Int, status *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "deactivateValidator", validatorID, status)
}

// DeactivateValidator is a paid mutator transaction binding the contract method 0x1e702f83.
//
// Solidity: function deactivateValidator(uint256 validatorID, uint256 status) returns()
func (_Contract *ContractSession) DeactivateValidator(validatorID *big.Int, status *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.DeactivateValidator(&_Contract.TransactOpts, validatorID, status)
}

// DeactivateValidator is a paid mutator transaction binding the contract method 0x1e702f83.
//
// Solidity: function deactivateValidator(uint256 validatorID, uint256 status) returns()
func (_Contract *ContractTransactorSession) DeactivateValidator(validatorID *big.Int, status *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.DeactivateValidator(&_Contract.TransactOpts, validatorID, status)
}

// Delegate is a paid mutator transaction binding the contract method 0x9fa6dd35.
//
// Solidity: function delegate(uint256 toValidatorID) payable returns()
func (_Contract *ContractTransactor) Delegate(opts *bind.TransactOpts, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "delegate", toValidatorID)
}

// Delegate is a paid mutator transaction binding the contract method 0x9fa6dd35.
//
// Solidity: function delegate(uint256 toValidatorID) payable returns()
func (_Contract *ContractSession) Delegate(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Delegate(&_Contract.TransactOpts, toValidatorID)
}

// Delegate is a paid mutator transaction binding the contract method 0x9fa6dd35.
//
// Solidity: function delegate(uint256 toValidatorID) payable returns()
func (_Contract *ContractTransactorSession) Delegate(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Delegate(&_Contract.TransactOpts, toValidatorID)
}

// Initialize is a paid mutator transaction binding the contract method 0x019e2729.
//
// Solidity: function initialize(uint256 sealedEpoch, uint256 _totalSupply, address nodeDriver, address owner) returns()
func (_Contract *ContractTransactor) Initialize(opts *bind.TransactOpts, sealedEpoch *big.Int, _totalSupply *big.Int, nodeDriver common.Address, owner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "initialize", sealedEpoch, _totalSupply, nodeDriver, owner)
}

// Initialize is a paid mutator transaction binding the contract method 0x019e2729.
//
// Solidity: function initialize(uint256 sealedEpoch, uint256 _totalSupply, address nodeDriver, address owner) returns()
func (_Contract *ContractSession) Initialize(sealedEpoch *big.Int, _totalSupply *big.Int, nodeDriver common.Address, owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, sealedEpoch, _totalSupply, nodeDriver, owner)
}

// Initialize is a paid mutator transaction binding the contract method 0x019e2729.
//
// Solidity: function initialize(uint256 sealedEpoch, uint256 _totalSupply, address nodeDriver, address owner) returns()
func (_Contract *ContractTransactorSession) Initialize(sealedEpoch *big.Int, _totalSupply *big.Int, nodeDriver common.Address, owner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.Initialize(&_Contract.TransactOpts, sealedEpoch, _totalSupply, nodeDriver, owner)
}

// LockStake is a paid mutator transaction binding the contract method 0xde67f215.
//
// Solidity: function lockStake(uint256 toValidatorID, uint256 lockupDuration, uint256 amount) returns()
func (_Contract *ContractTransactor) LockStake(opts *bind.TransactOpts, toValidatorID *big.Int, lockupDuration *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "lockStake", toValidatorID, lockupDuration, amount)
}

// LockStake is a paid mutator transaction binding the contract method 0xde67f215.
//
// Solidity: function lockStake(uint256 toValidatorID, uint256 lockupDuration, uint256 amount) returns()
func (_Contract *ContractSession) LockStake(toValidatorID *big.Int, lockupDuration *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LockStake(&_Contract.TransactOpts, toValidatorID, lockupDuration, amount)
}

// LockStake is a paid mutator transaction binding the contract method 0xde67f215.
//
// Solidity: function lockStake(uint256 toValidatorID, uint256 lockupDuration, uint256 amount) returns()
func (_Contract *ContractTransactorSession) LockStake(toValidatorID *big.Int, lockupDuration *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LockStake(&_Contract.TransactOpts, toValidatorID, lockupDuration, amount)
}

// LockUpDelegation is a paid mutator transaction binding the contract method 0xa4b89fab.
//
// Solidity: function lockUpDelegation(uint256 lockDuration, uint256 toStakerID) returns()
func (_Contract *ContractTransactor) LockUpDelegation(opts *bind.TransactOpts, lockDuration *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "lockUpDelegation", lockDuration, toStakerID)
}

// LockUpDelegation is a paid mutator transaction binding the contract method 0xa4b89fab.
//
// Solidity: function lockUpDelegation(uint256 lockDuration, uint256 toStakerID) returns()
func (_Contract *ContractSession) LockUpDelegation(lockDuration *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LockUpDelegation(&_Contract.TransactOpts, lockDuration, toStakerID)
}

// LockUpDelegation is a paid mutator transaction binding the contract method 0xa4b89fab.
//
// Solidity: function lockUpDelegation(uint256 lockDuration, uint256 toStakerID) returns()
func (_Contract *ContractTransactorSession) LockUpDelegation(lockDuration *big.Int, toStakerID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LockUpDelegation(&_Contract.TransactOpts, lockDuration, toStakerID)
}

// LockUpStake is a paid mutator transaction binding the contract method 0xf3ae5b1a.
//
// Solidity: function lockUpStake(uint256 lockDuration) returns()
func (_Contract *ContractTransactor) LockUpStake(opts *bind.TransactOpts, lockDuration *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "lockUpStake", lockDuration)
}

// LockUpStake is a paid mutator transaction binding the contract method 0xf3ae5b1a.
//
// Solidity: function lockUpStake(uint256 lockDuration) returns()
func (_Contract *ContractSession) LockUpStake(lockDuration *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LockUpStake(&_Contract.TransactOpts, lockDuration)
}

// LockUpStake is a paid mutator transaction binding the contract method 0xf3ae5b1a.
//
// Solidity: function lockUpStake(uint256 lockDuration) returns()
func (_Contract *ContractTransactorSession) LockUpStake(lockDuration *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LockUpStake(&_Contract.TransactOpts, lockDuration)
}

// PartialWithdrawByRequest is a paid mutator transaction binding the contract method 0xf8b18d8a.
//
// Solidity: function partialWithdrawByRequest(uint256 ) returns()
func (_Contract *ContractTransactor) PartialWithdrawByRequest(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "partialWithdrawByRequest", arg0)
}

// PartialWithdrawByRequest is a paid mutator transaction binding the contract method 0xf8b18d8a.
//
// Solidity: function partialWithdrawByRequest(uint256 ) returns()
func (_Contract *ContractSession) PartialWithdrawByRequest(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PartialWithdrawByRequest(&_Contract.TransactOpts, arg0)
}

// PartialWithdrawByRequest is a paid mutator transaction binding the contract method 0xf8b18d8a.
//
// Solidity: function partialWithdrawByRequest(uint256 ) returns()
func (_Contract *ContractTransactorSession) PartialWithdrawByRequest(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PartialWithdrawByRequest(&_Contract.TransactOpts, arg0)
}

// PrepareToWithdrawDelegation is a paid mutator transaction binding the contract method 0xb1e64339.
//
// Solidity: function prepareToWithdrawDelegation(uint256 ) returns()
func (_Contract *ContractTransactor) PrepareToWithdrawDelegation(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "prepareToWithdrawDelegation", arg0)
}

// PrepareToWithdrawDelegation is a paid mutator transaction binding the contract method 0xb1e64339.
//
// Solidity: function prepareToWithdrawDelegation(uint256 ) returns()
func (_Contract *ContractSession) PrepareToWithdrawDelegation(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawDelegation(&_Contract.TransactOpts, arg0)
}

// PrepareToWithdrawDelegation is a paid mutator transaction binding the contract method 0xb1e64339.
//
// Solidity: function prepareToWithdrawDelegation(uint256 ) returns()
func (_Contract *ContractTransactorSession) PrepareToWithdrawDelegation(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawDelegation(&_Contract.TransactOpts, arg0)
}

// PrepareToWithdrawDelegationPartial is a paid mutator transaction binding the contract method 0xbb03a4bd.
//
// Solidity: function prepareToWithdrawDelegationPartial(uint256 wrID, uint256 toStakerID, uint256 amount) returns()
func (_Contract *ContractTransactor) PrepareToWithdrawDelegationPartial(opts *bind.TransactOpts, wrID *big.Int, toStakerID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "prepareToWithdrawDelegationPartial", wrID, toStakerID, amount)
}

// PrepareToWithdrawDelegationPartial is a paid mutator transaction binding the contract method 0xbb03a4bd.
//
// Solidity: function prepareToWithdrawDelegationPartial(uint256 wrID, uint256 toStakerID, uint256 amount) returns()
func (_Contract *ContractSession) PrepareToWithdrawDelegationPartial(wrID *big.Int, toStakerID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawDelegationPartial(&_Contract.TransactOpts, wrID, toStakerID, amount)
}

// PrepareToWithdrawDelegationPartial is a paid mutator transaction binding the contract method 0xbb03a4bd.
//
// Solidity: function prepareToWithdrawDelegationPartial(uint256 wrID, uint256 toStakerID, uint256 amount) returns()
func (_Contract *ContractTransactorSession) PrepareToWithdrawDelegationPartial(wrID *big.Int, toStakerID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawDelegationPartial(&_Contract.TransactOpts, wrID, toStakerID, amount)
}

// PrepareToWithdrawStake is a paid mutator transaction binding the contract method 0xc41b6405.
//
// Solidity: function prepareToWithdrawStake() returns()
func (_Contract *ContractTransactor) PrepareToWithdrawStake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "prepareToWithdrawStake")
}

// PrepareToWithdrawStake is a paid mutator transaction binding the contract method 0xc41b6405.
//
// Solidity: function prepareToWithdrawStake() returns()
func (_Contract *ContractSession) PrepareToWithdrawStake() (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawStake(&_Contract.TransactOpts)
}

// PrepareToWithdrawStake is a paid mutator transaction binding the contract method 0xc41b6405.
//
// Solidity: function prepareToWithdrawStake() returns()
func (_Contract *ContractTransactorSession) PrepareToWithdrawStake() (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawStake(&_Contract.TransactOpts)
}

// PrepareToWithdrawStakePartial is a paid mutator transaction binding the contract method 0x26682c71.
//
// Solidity: function prepareToWithdrawStakePartial(uint256 wrID, uint256 amount) returns()
func (_Contract *ContractTransactor) PrepareToWithdrawStakePartial(opts *bind.TransactOpts, wrID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "prepareToWithdrawStakePartial", wrID, amount)
}

// PrepareToWithdrawStakePartial is a paid mutator transaction binding the contract method 0x26682c71.
//
// Solidity: function prepareToWithdrawStakePartial(uint256 wrID, uint256 amount) returns()
func (_Contract *ContractSession) PrepareToWithdrawStakePartial(wrID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawStakePartial(&_Contract.TransactOpts, wrID, amount)
}

// PrepareToWithdrawStakePartial is a paid mutator transaction binding the contract method 0x26682c71.
//
// Solidity: function prepareToWithdrawStakePartial(uint256 wrID, uint256 amount) returns()
func (_Contract *ContractTransactorSession) PrepareToWithdrawStakePartial(wrID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.PrepareToWithdrawStakePartial(&_Contract.TransactOpts, wrID, amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _Contract.Contract.RenounceOwnership(&_Contract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Contract *ContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Contract.Contract.RenounceOwnership(&_Contract.TransactOpts)
}

// RestakeRewards is a paid mutator transaction binding the contract method 0x08c36874.
//
// Solidity: function restakeRewards(uint256 toValidatorID) returns()
func (_Contract *ContractTransactor) RestakeRewards(opts *bind.TransactOpts, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "restakeRewards", toValidatorID)
}

// RestakeRewards is a paid mutator transaction binding the contract method 0x08c36874.
//
// Solidity: function restakeRewards(uint256 toValidatorID) returns()
func (_Contract *ContractSession) RestakeRewards(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RestakeRewards(&_Contract.TransactOpts, toValidatorID)
}

// RestakeRewards is a paid mutator transaction binding the contract method 0x08c36874.
//
// Solidity: function restakeRewards(uint256 toValidatorID) returns()
func (_Contract *ContractTransactorSession) RestakeRewards(toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RestakeRewards(&_Contract.TransactOpts, toValidatorID)
}

// SealEpoch is a paid mutator transaction binding the contract method 0xebdf104c.
//
// Solidity: function sealEpoch(uint256[] offlineTime, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee) returns()
func (_Contract *ContractTransactor) SealEpoch(opts *bind.TransactOpts, offlineTime []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "sealEpoch", offlineTime, offlineBlocks, uptimes, originatedTxsFee)
}

// SealEpoch is a paid mutator transaction binding the contract method 0xebdf104c.
//
// Solidity: function sealEpoch(uint256[] offlineTime, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee) returns()
func (_Contract *ContractSession) SealEpoch(offlineTime []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpoch(&_Contract.TransactOpts, offlineTime, offlineBlocks, uptimes, originatedTxsFee)
}

// SealEpoch is a paid mutator transaction binding the contract method 0xebdf104c.
//
// Solidity: function sealEpoch(uint256[] offlineTime, uint256[] offlineBlocks, uint256[] uptimes, uint256[] originatedTxsFee) returns()
func (_Contract *ContractTransactorSession) SealEpoch(offlineTime []*big.Int, offlineBlocks []*big.Int, uptimes []*big.Int, originatedTxsFee []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpoch(&_Contract.TransactOpts, offlineTime, offlineBlocks, uptimes, originatedTxsFee)
}

// SealEpochValidators is a paid mutator transaction binding the contract method 0xe08d7e66.
//
// Solidity: function sealEpochValidators(uint256[] nextValidatorIDs) returns()
func (_Contract *ContractTransactor) SealEpochValidators(opts *bind.TransactOpts, nextValidatorIDs []*big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "sealEpochValidators", nextValidatorIDs)
}

// SealEpochValidators is a paid mutator transaction binding the contract method 0xe08d7e66.
//
// Solidity: function sealEpochValidators(uint256[] nextValidatorIDs) returns()
func (_Contract *ContractSession) SealEpochValidators(nextValidatorIDs []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpochValidators(&_Contract.TransactOpts, nextValidatorIDs)
}

// SealEpochValidators is a paid mutator transaction binding the contract method 0xe08d7e66.
//
// Solidity: function sealEpochValidators(uint256[] nextValidatorIDs) returns()
func (_Contract *ContractTransactorSession) SealEpochValidators(nextValidatorIDs []*big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SealEpochValidators(&_Contract.TransactOpts, nextValidatorIDs)
}

// SetGenesisDelegation is a paid mutator transaction binding the contract method 0x18f628d4.
//
// Solidity: function setGenesisDelegation(address delegator, uint256 toValidatorID, uint256 stake, uint256 lockedStake, uint256 lockupFromEpoch, uint256 lockupEndTime, uint256 lockupDuration, uint256 earlyUnlockPenalty, uint256 rewards) returns()
func (_Contract *ContractTransactor) SetGenesisDelegation(opts *bind.TransactOpts, delegator common.Address, toValidatorID *big.Int, stake *big.Int, lockedStake *big.Int, lockupFromEpoch *big.Int, lockupEndTime *big.Int, lockupDuration *big.Int, earlyUnlockPenalty *big.Int, rewards *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setGenesisDelegation", delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
}

// SetGenesisDelegation is a paid mutator transaction binding the contract method 0x18f628d4.
//
// Solidity: function setGenesisDelegation(address delegator, uint256 toValidatorID, uint256 stake, uint256 lockedStake, uint256 lockupFromEpoch, uint256 lockupEndTime, uint256 lockupDuration, uint256 earlyUnlockPenalty, uint256 rewards) returns()
func (_Contract *ContractSession) SetGenesisDelegation(delegator common.Address, toValidatorID *big.Int, stake *big.Int, lockedStake *big.Int, lockupFromEpoch *big.Int, lockupEndTime *big.Int, lockupDuration *big.Int, earlyUnlockPenalty *big.Int, rewards *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisDelegation(&_Contract.TransactOpts, delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
}

// SetGenesisDelegation is a paid mutator transaction binding the contract method 0x18f628d4.
//
// Solidity: function setGenesisDelegation(address delegator, uint256 toValidatorID, uint256 stake, uint256 lockedStake, uint256 lockupFromEpoch, uint256 lockupEndTime, uint256 lockupDuration, uint256 earlyUnlockPenalty, uint256 rewards) returns()
func (_Contract *ContractTransactorSession) SetGenesisDelegation(delegator common.Address, toValidatorID *big.Int, stake *big.Int, lockedStake *big.Int, lockupFromEpoch *big.Int, lockupEndTime *big.Int, lockupDuration *big.Int, earlyUnlockPenalty *big.Int, rewards *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisDelegation(&_Contract.TransactOpts, delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
}

// SetGenesisValidator is a paid mutator transaction binding the contract method 0x4feb92f3.
//
// Solidity: function setGenesisValidator(address auth, uint256 validatorID, bytes pubkey, uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime) returns()
func (_Contract *ContractTransactor) SetGenesisValidator(opts *bind.TransactOpts, auth common.Address, validatorID *big.Int, pubkey []byte, status *big.Int, createdEpoch *big.Int, createdTime *big.Int, deactivatedEpoch *big.Int, deactivatedTime *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "setGenesisValidator", auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
}

// SetGenesisValidator is a paid mutator transaction binding the contract method 0x4feb92f3.
//
// Solidity: function setGenesisValidator(address auth, uint256 validatorID, bytes pubkey, uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime) returns()
func (_Contract *ContractSession) SetGenesisValidator(auth common.Address, validatorID *big.Int, pubkey []byte, status *big.Int, createdEpoch *big.Int, createdTime *big.Int, deactivatedEpoch *big.Int, deactivatedTime *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisValidator(&_Contract.TransactOpts, auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
}

// SetGenesisValidator is a paid mutator transaction binding the contract method 0x4feb92f3.
//
// Solidity: function setGenesisValidator(address auth, uint256 validatorID, bytes pubkey, uint256 status, uint256 createdEpoch, uint256 createdTime, uint256 deactivatedEpoch, uint256 deactivatedTime) returns()
func (_Contract *ContractTransactorSession) SetGenesisValidator(auth common.Address, validatorID *big.Int, pubkey []byte, status *big.Int, createdEpoch *big.Int, createdTime *big.Int, deactivatedEpoch *big.Int, deactivatedTime *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.SetGenesisValidator(&_Contract.TransactOpts, auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
}

// StashRewards is a paid mutator transaction binding the contract method 0x8cddb015.
//
// Solidity: function stashRewards(address delegator, uint256 toValidatorID) returns()
func (_Contract *ContractTransactor) StashRewards(opts *bind.TransactOpts, delegator common.Address, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "stashRewards", delegator, toValidatorID)
}

// StashRewards is a paid mutator transaction binding the contract method 0x8cddb015.
//
// Solidity: function stashRewards(address delegator, uint256 toValidatorID) returns()
func (_Contract *ContractSession) StashRewards(delegator common.Address, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.StashRewards(&_Contract.TransactOpts, delegator, toValidatorID)
}

// StashRewards is a paid mutator transaction binding the contract method 0x8cddb015.
//
// Solidity: function stashRewards(address delegator, uint256 toValidatorID) returns()
func (_Contract *ContractTransactorSession) StashRewards(delegator common.Address, toValidatorID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.StashRewards(&_Contract.TransactOpts, delegator, toValidatorID)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferOwnership(&_Contract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Contract *ContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Contract.Contract.TransferOwnership(&_Contract.TransactOpts, newOwner)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4f864df4.
//
// Solidity: function undelegate(uint256 toValidatorID, uint256 wrID, uint256 amount) returns()
func (_Contract *ContractTransactor) Undelegate(opts *bind.TransactOpts, toValidatorID *big.Int, wrID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "undelegate", toValidatorID, wrID, amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4f864df4.
//
// Solidity: function undelegate(uint256 toValidatorID, uint256 wrID, uint256 amount) returns()
func (_Contract *ContractSession) Undelegate(toValidatorID *big.Int, wrID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Undelegate(&_Contract.TransactOpts, toValidatorID, wrID, amount)
}

// Undelegate is a paid mutator transaction binding the contract method 0x4f864df4.
//
// Solidity: function undelegate(uint256 toValidatorID, uint256 wrID, uint256 amount) returns()
func (_Contract *ContractTransactorSession) Undelegate(toValidatorID *big.Int, wrID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Undelegate(&_Contract.TransactOpts, toValidatorID, wrID, amount)
}

// UnlockStake is a paid mutator transaction binding the contract method 0x1d3ac42c.
//
// Solidity: function unlockStake(uint256 toValidatorID, uint256 amount) returns(uint256)
func (_Contract *ContractTransactor) UnlockStake(opts *bind.TransactOpts, toValidatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "unlockStake", toValidatorID, amount)
}

// UnlockStake is a paid mutator transaction binding the contract method 0x1d3ac42c.
//
// Solidity: function unlockStake(uint256 toValidatorID, uint256 amount) returns(uint256)
func (_Contract *ContractSession) UnlockStake(toValidatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UnlockStake(&_Contract.TransactOpts, toValidatorID, amount)
}

// UnlockStake is a paid mutator transaction binding the contract method 0x1d3ac42c.
//
// Solidity: function unlockStake(uint256 toValidatorID, uint256 amount) returns(uint256)
func (_Contract *ContractTransactorSession) UnlockStake(toValidatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UnlockStake(&_Contract.TransactOpts, toValidatorID, amount)
}

// UpdateBaseRewardPerSecond is a paid mutator transaction binding the contract method 0xb6d9edd5.
//
// Solidity: function updateBaseRewardPerSecond(uint256 value) returns()
func (_Contract *ContractTransactor) UpdateBaseRewardPerSecond(opts *bind.TransactOpts, value *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateBaseRewardPerSecond", value)
}

// UpdateBaseRewardPerSecond is a paid mutator transaction binding the contract method 0xb6d9edd5.
//
// Solidity: function updateBaseRewardPerSecond(uint256 value) returns()
func (_Contract *ContractSession) UpdateBaseRewardPerSecond(value *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateBaseRewardPerSecond(&_Contract.TransactOpts, value)
}

// UpdateBaseRewardPerSecond is a paid mutator transaction binding the contract method 0xb6d9edd5.
//
// Solidity: function updateBaseRewardPerSecond(uint256 value) returns()
func (_Contract *ContractTransactorSession) UpdateBaseRewardPerSecond(value *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateBaseRewardPerSecond(&_Contract.TransactOpts, value)
}

// UpdateOfflinePenaltyThreshold is a paid mutator transaction binding the contract method 0x8b1a0d11.
//
// Solidity: function updateOfflinePenaltyThreshold(uint256 blocksNum, uint256 time) returns()
func (_Contract *ContractTransactor) UpdateOfflinePenaltyThreshold(opts *bind.TransactOpts, blocksNum *big.Int, time *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateOfflinePenaltyThreshold", blocksNum, time)
}

// UpdateOfflinePenaltyThreshold is a paid mutator transaction binding the contract method 0x8b1a0d11.
//
// Solidity: function updateOfflinePenaltyThreshold(uint256 blocksNum, uint256 time) returns()
func (_Contract *ContractSession) UpdateOfflinePenaltyThreshold(blocksNum *big.Int, time *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateOfflinePenaltyThreshold(&_Contract.TransactOpts, blocksNum, time)
}

// UpdateOfflinePenaltyThreshold is a paid mutator transaction binding the contract method 0x8b1a0d11.
//
// Solidity: function updateOfflinePenaltyThreshold(uint256 blocksNum, uint256 time) returns()
func (_Contract *ContractTransactorSession) UpdateOfflinePenaltyThreshold(blocksNum *big.Int, time *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateOfflinePenaltyThreshold(&_Contract.TransactOpts, blocksNum, time)
}

// UpdateSlashingRefundRatio is a paid mutator transaction binding the contract method 0x4f7c4efb.
//
// Solidity: function updateSlashingRefundRatio(uint256 validatorID, uint256 refundRatio) returns()
func (_Contract *ContractTransactor) UpdateSlashingRefundRatio(opts *bind.TransactOpts, validatorID *big.Int, refundRatio *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateSlashingRefundRatio", validatorID, refundRatio)
}

// UpdateSlashingRefundRatio is a paid mutator transaction binding the contract method 0x4f7c4efb.
//
// Solidity: function updateSlashingRefundRatio(uint256 validatorID, uint256 refundRatio) returns()
func (_Contract *ContractSession) UpdateSlashingRefundRatio(validatorID *big.Int, refundRatio *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateSlashingRefundRatio(&_Contract.TransactOpts, validatorID, refundRatio)
}

// UpdateSlashingRefundRatio is a paid mutator transaction binding the contract method 0x4f7c4efb.
//
// Solidity: function updateSlashingRefundRatio(uint256 validatorID, uint256 refundRatio) returns()
func (_Contract *ContractTransactorSession) UpdateSlashingRefundRatio(validatorID *big.Int, refundRatio *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.UpdateSlashingRefundRatio(&_Contract.TransactOpts, validatorID, refundRatio)
}

// Withdraw is a paid mutator transaction binding the contract method 0x441a3e70.
//
// Solidity: function withdraw(uint256 toValidatorID, uint256 wrID) returns()
func (_Contract *ContractTransactor) Withdraw(opts *bind.TransactOpts, toValidatorID *big.Int, wrID *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "withdraw", toValidatorID, wrID)
}

// Withdraw is a paid mutator transaction binding the contract method 0x441a3e70.
//
// Solidity: function withdraw(uint256 toValidatorID, uint256 wrID) returns()
func (_Contract *ContractSession) Withdraw(toValidatorID *big.Int, wrID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Withdraw(&_Contract.TransactOpts, toValidatorID, wrID)
}

// Withdraw is a paid mutator transaction binding the contract method 0x441a3e70.
//
// Solidity: function withdraw(uint256 toValidatorID, uint256 wrID) returns()
func (_Contract *ContractTransactorSession) Withdraw(toValidatorID *big.Int, wrID *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.Withdraw(&_Contract.TransactOpts, toValidatorID, wrID)
}

// WithdrawDelegation is a paid mutator transaction binding the contract method 0xdf0e307a.
//
// Solidity: function withdrawDelegation(uint256 ) returns()
func (_Contract *ContractTransactor) WithdrawDelegation(opts *bind.TransactOpts, arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "withdrawDelegation", arg0)
}

// WithdrawDelegation is a paid mutator transaction binding the contract method 0xdf0e307a.
//
// Solidity: function withdrawDelegation(uint256 ) returns()
func (_Contract *ContractSession) WithdrawDelegation(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.WithdrawDelegation(&_Contract.TransactOpts, arg0)
}

// WithdrawDelegation is a paid mutator transaction binding the contract method 0xdf0e307a.
//
// Solidity: function withdrawDelegation(uint256 ) returns()
func (_Contract *ContractTransactorSession) WithdrawDelegation(arg0 *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.WithdrawDelegation(&_Contract.TransactOpts, arg0)
}

// WithdrawStake is a paid mutator transaction binding the contract method 0xbed9d861.
//
// Solidity: function withdrawStake() returns()
func (_Contract *ContractTransactor) WithdrawStake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "withdrawStake")
}

// WithdrawStake is a paid mutator transaction binding the contract method 0xbed9d861.
//
// Solidity: function withdrawStake() returns()
func (_Contract *ContractSession) WithdrawStake() (*types.Transaction, error) {
	return _Contract.Contract.WithdrawStake(&_Contract.TransactOpts)
}

// WithdrawStake is a paid mutator transaction binding the contract method 0xbed9d861.
//
// Solidity: function withdrawStake() returns()
func (_Contract *ContractTransactorSession) WithdrawStake() (*types.Transaction, error) {
	return _Contract.Contract.WithdrawStake(&_Contract.TransactOpts)
}

// ContractChangedValidatorStatusIterator is returned from FilterChangedValidatorStatus and is used to iterate over the raw logs and unpacked data for ChangedValidatorStatus events raised by the Contract contract.
type ContractChangedValidatorStatusIterator struct {
	Event *ContractChangedValidatorStatus // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractChangedValidatorStatusIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractChangedValidatorStatus)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractChangedValidatorStatus)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractChangedValidatorStatusIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractChangedValidatorStatusIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractChangedValidatorStatus represents a ChangedValidatorStatus event raised by the Contract contract.
type ContractChangedValidatorStatus struct {
	ValidatorID *big.Int
	Status      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterChangedValidatorStatus is a free log retrieval operation binding the contract event 0xcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e.
//
// Solidity: event ChangedValidatorStatus(uint256 indexed validatorID, uint256 status)
func (_Contract *ContractFilterer) FilterChangedValidatorStatus(opts *bind.FilterOpts, validatorID []*big.Int) (*ContractChangedValidatorStatusIterator, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ChangedValidatorStatus", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractChangedValidatorStatusIterator{contract: _Contract.contract, event: "ChangedValidatorStatus", logs: logs, sub: sub}, nil
}

// WatchChangedValidatorStatus is a free log subscription operation binding the contract event 0xcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e.
//
// Solidity: event ChangedValidatorStatus(uint256 indexed validatorID, uint256 status)
func (_Contract *ContractFilterer) WatchChangedValidatorStatus(opts *bind.WatchOpts, sink chan<- *ContractChangedValidatorStatus, validatorID []*big.Int) (event.Subscription, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ChangedValidatorStatus", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractChangedValidatorStatus)
				if err := _Contract.contract.UnpackLog(event, "ChangedValidatorStatus", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChangedValidatorStatus is a log parse operation binding the contract event 0xcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e.
//
// Solidity: event ChangedValidatorStatus(uint256 indexed validatorID, uint256 status)
func (_Contract *ContractFilterer) ParseChangedValidatorStatus(log types.Log) (*ContractChangedValidatorStatus, error) {
	event := new(ContractChangedValidatorStatus)
	if err := _Contract.contract.UnpackLog(event, "ChangedValidatorStatus", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractClaimedRewardsIterator is returned from FilterClaimedRewards and is used to iterate over the raw logs and unpacked data for ClaimedRewards events raised by the Contract contract.
type ContractClaimedRewardsIterator struct {
	Event *ContractClaimedRewards // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractClaimedRewardsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractClaimedRewards)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractClaimedRewards)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractClaimedRewardsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractClaimedRewardsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractClaimedRewards represents a ClaimedRewards event raised by the Contract contract.
type ContractClaimedRewards struct {
	Delegator         common.Address
	ToValidatorID     *big.Int
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterClaimedRewards is a free log retrieval operation binding the contract event 0xc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf45.
//
// Solidity: event ClaimedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractFilterer) FilterClaimedRewards(opts *bind.FilterOpts, delegator []common.Address, toValidatorID []*big.Int) (*ContractClaimedRewardsIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "ClaimedRewards", delegatorRule, toValidatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractClaimedRewardsIterator{contract: _Contract.contract, event: "ClaimedRewards", logs: logs, sub: sub}, nil
}

// WatchClaimedRewards is a free log subscription operation binding the contract event 0xc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf45.
//
// Solidity: event ClaimedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractFilterer) WatchClaimedRewards(opts *bind.WatchOpts, sink chan<- *ContractClaimedRewards, delegator []common.Address, toValidatorID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "ClaimedRewards", delegatorRule, toValidatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractClaimedRewards)
				if err := _Contract.contract.UnpackLog(event, "ClaimedRewards", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseClaimedRewards is a log parse operation binding the contract event 0xc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf45.
//
// Solidity: event ClaimedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractFilterer) ParseClaimedRewards(log types.Log) (*ContractClaimedRewards, error) {
	event := new(ContractClaimedRewards)
	if err := _Contract.contract.UnpackLog(event, "ClaimedRewards", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractCreatedValidatorIterator is returned from FilterCreatedValidator and is used to iterate over the raw logs and unpacked data for CreatedValidator events raised by the Contract contract.
type ContractCreatedValidatorIterator struct {
	Event *ContractCreatedValidator // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractCreatedValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractCreatedValidator)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractCreatedValidator)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractCreatedValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractCreatedValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractCreatedValidator represents a CreatedValidator event raised by the Contract contract.
type ContractCreatedValidator struct {
	ValidatorID  *big.Int
	Auth         common.Address
	CreatedEpoch *big.Int
	CreatedTime  *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCreatedValidator is a free log retrieval operation binding the contract event 0x49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf.
//
// Solidity: event CreatedValidator(uint256 indexed validatorID, address indexed auth, uint256 createdEpoch, uint256 createdTime)
func (_Contract *ContractFilterer) FilterCreatedValidator(opts *bind.FilterOpts, validatorID []*big.Int, auth []common.Address) (*ContractCreatedValidatorIterator, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}
	var authRule []interface{}
	for _, authItem := range auth {
		authRule = append(authRule, authItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "CreatedValidator", validatorIDRule, authRule)
	if err != nil {
		return nil, err
	}
	return &ContractCreatedValidatorIterator{contract: _Contract.contract, event: "CreatedValidator", logs: logs, sub: sub}, nil
}

// WatchCreatedValidator is a free log subscription operation binding the contract event 0x49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf.
//
// Solidity: event CreatedValidator(uint256 indexed validatorID, address indexed auth, uint256 createdEpoch, uint256 createdTime)
func (_Contract *ContractFilterer) WatchCreatedValidator(opts *bind.WatchOpts, sink chan<- *ContractCreatedValidator, validatorID []*big.Int, auth []common.Address) (event.Subscription, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}
	var authRule []interface{}
	for _, authItem := range auth {
		authRule = append(authRule, authItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "CreatedValidator", validatorIDRule, authRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractCreatedValidator)
				if err := _Contract.contract.UnpackLog(event, "CreatedValidator", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCreatedValidator is a log parse operation binding the contract event 0x49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf.
//
// Solidity: event CreatedValidator(uint256 indexed validatorID, address indexed auth, uint256 createdEpoch, uint256 createdTime)
func (_Contract *ContractFilterer) ParseCreatedValidator(log types.Log) (*ContractCreatedValidator, error) {
	event := new(ContractCreatedValidator)
	if err := _Contract.contract.UnpackLog(event, "CreatedValidator", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractDeactivatedValidatorIterator is returned from FilterDeactivatedValidator and is used to iterate over the raw logs and unpacked data for DeactivatedValidator events raised by the Contract contract.
type ContractDeactivatedValidatorIterator struct {
	Event *ContractDeactivatedValidator // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDeactivatedValidatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDeactivatedValidator)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDeactivatedValidator)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDeactivatedValidatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDeactivatedValidatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDeactivatedValidator represents a DeactivatedValidator event raised by the Contract contract.
type ContractDeactivatedValidator struct {
	ValidatorID      *big.Int
	DeactivatedEpoch *big.Int
	DeactivatedTime  *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterDeactivatedValidator is a free log retrieval operation binding the contract event 0xac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47.
//
// Solidity: event DeactivatedValidator(uint256 indexed validatorID, uint256 deactivatedEpoch, uint256 deactivatedTime)
func (_Contract *ContractFilterer) FilterDeactivatedValidator(opts *bind.FilterOpts, validatorID []*big.Int) (*ContractDeactivatedValidatorIterator, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "DeactivatedValidator", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractDeactivatedValidatorIterator{contract: _Contract.contract, event: "DeactivatedValidator", logs: logs, sub: sub}, nil
}

// WatchDeactivatedValidator is a free log subscription operation binding the contract event 0xac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47.
//
// Solidity: event DeactivatedValidator(uint256 indexed validatorID, uint256 deactivatedEpoch, uint256 deactivatedTime)
func (_Contract *ContractFilterer) WatchDeactivatedValidator(opts *bind.WatchOpts, sink chan<- *ContractDeactivatedValidator, validatorID []*big.Int) (event.Subscription, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "DeactivatedValidator", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDeactivatedValidator)
				if err := _Contract.contract.UnpackLog(event, "DeactivatedValidator", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeactivatedValidator is a log parse operation binding the contract event 0xac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47.
//
// Solidity: event DeactivatedValidator(uint256 indexed validatorID, uint256 deactivatedEpoch, uint256 deactivatedTime)
func (_Contract *ContractFilterer) ParseDeactivatedValidator(log types.Log) (*ContractDeactivatedValidator, error) {
	event := new(ContractDeactivatedValidator)
	if err := _Contract.contract.UnpackLog(event, "DeactivatedValidator", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractDelegatedIterator is returned from FilterDelegated and is used to iterate over the raw logs and unpacked data for Delegated events raised by the Contract contract.
type ContractDelegatedIterator struct {
	Event *ContractDelegated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDelegated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDelegated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDelegated represents a Delegated event raised by the Contract contract.
type ContractDelegated struct {
	Delegator     common.Address
	ToValidatorID *big.Int
	Amount        *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDelegated is a free log retrieval operation binding the contract event 0x9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb.
//
// Solidity: event Delegated(address indexed delegator, uint256 indexed toValidatorID, uint256 amount)
func (_Contract *ContractFilterer) FilterDelegated(opts *bind.FilterOpts, delegator []common.Address, toValidatorID []*big.Int) (*ContractDelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Delegated", delegatorRule, toValidatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractDelegatedIterator{contract: _Contract.contract, event: "Delegated", logs: logs, sub: sub}, nil
}

// WatchDelegated is a free log subscription operation binding the contract event 0x9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb.
//
// Solidity: event Delegated(address indexed delegator, uint256 indexed toValidatorID, uint256 amount)
func (_Contract *ContractFilterer) WatchDelegated(opts *bind.WatchOpts, sink chan<- *ContractDelegated, delegator []common.Address, toValidatorID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Delegated", delegatorRule, toValidatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDelegated)
				if err := _Contract.contract.UnpackLog(event, "Delegated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegated is a log parse operation binding the contract event 0x9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb.
//
// Solidity: event Delegated(address indexed delegator, uint256 indexed toValidatorID, uint256 amount)
func (_Contract *ContractFilterer) ParseDelegated(log types.Log) (*ContractDelegated, error) {
	event := new(ContractDelegated)
	if err := _Contract.contract.UnpackLog(event, "Delegated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractLockedUpStakeIterator is returned from FilterLockedUpStake and is used to iterate over the raw logs and unpacked data for LockedUpStake events raised by the Contract contract.
type ContractLockedUpStakeIterator struct {
	Event *ContractLockedUpStake // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractLockedUpStakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractLockedUpStake)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractLockedUpStake)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractLockedUpStakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractLockedUpStakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractLockedUpStake represents a LockedUpStake event raised by the Contract contract.
type ContractLockedUpStake struct {
	Delegator   common.Address
	ValidatorID *big.Int
	Duration    *big.Int
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterLockedUpStake is a free log retrieval operation binding the contract event 0x138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78.
//
// Solidity: event LockedUpStake(address indexed delegator, uint256 indexed validatorID, uint256 duration, uint256 amount)
func (_Contract *ContractFilterer) FilterLockedUpStake(opts *bind.FilterOpts, delegator []common.Address, validatorID []*big.Int) (*ContractLockedUpStakeIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "LockedUpStake", delegatorRule, validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractLockedUpStakeIterator{contract: _Contract.contract, event: "LockedUpStake", logs: logs, sub: sub}, nil
}

// WatchLockedUpStake is a free log subscription operation binding the contract event 0x138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78.
//
// Solidity: event LockedUpStake(address indexed delegator, uint256 indexed validatorID, uint256 duration, uint256 amount)
func (_Contract *ContractFilterer) WatchLockedUpStake(opts *bind.WatchOpts, sink chan<- *ContractLockedUpStake, delegator []common.Address, validatorID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "LockedUpStake", delegatorRule, validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractLockedUpStake)
				if err := _Contract.contract.UnpackLog(event, "LockedUpStake", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseLockedUpStake is a log parse operation binding the contract event 0x138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78.
//
// Solidity: event LockedUpStake(address indexed delegator, uint256 indexed validatorID, uint256 duration, uint256 amount)
func (_Contract *ContractFilterer) ParseLockedUpStake(log types.Log) (*ContractLockedUpStake, error) {
	event := new(ContractLockedUpStake)
	if err := _Contract.contract.UnpackLog(event, "LockedUpStake", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Contract contract.
type ContractOwnershipTransferredIterator struct {
	Event *ContractOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractOwnershipTransferred represents a OwnershipTransferred event raised by the Contract contract.
type ContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractOwnershipTransferredIterator{contract: _Contract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractOwnershipTransferred)
				if err := _Contract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Contract *ContractFilterer) ParseOwnershipTransferred(log types.Log) (*ContractOwnershipTransferred, error) {
	event := new(ContractOwnershipTransferred)
	if err := _Contract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractRestakedRewardsIterator is returned from FilterRestakedRewards and is used to iterate over the raw logs and unpacked data for RestakedRewards events raised by the Contract contract.
type ContractRestakedRewardsIterator struct {
	Event *ContractRestakedRewards // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractRestakedRewardsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRestakedRewards)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractRestakedRewards)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractRestakedRewardsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRestakedRewardsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRestakedRewards represents a RestakedRewards event raised by the Contract contract.
type ContractRestakedRewards struct {
	Delegator         common.Address
	ToValidatorID     *big.Int
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRestakedRewards is a free log retrieval operation binding the contract event 0x4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26.
//
// Solidity: event RestakedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractFilterer) FilterRestakedRewards(opts *bind.FilterOpts, delegator []common.Address, toValidatorID []*big.Int) (*ContractRestakedRewardsIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RestakedRewards", delegatorRule, toValidatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractRestakedRewardsIterator{contract: _Contract.contract, event: "RestakedRewards", logs: logs, sub: sub}, nil
}

// WatchRestakedRewards is a free log subscription operation binding the contract event 0x4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26.
//
// Solidity: event RestakedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractFilterer) WatchRestakedRewards(opts *bind.WatchOpts, sink chan<- *ContractRestakedRewards, delegator []common.Address, toValidatorID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RestakedRewards", delegatorRule, toValidatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRestakedRewards)
				if err := _Contract.contract.UnpackLog(event, "RestakedRewards", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRestakedRewards is a log parse operation binding the contract event 0x4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26.
//
// Solidity: event RestakedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractFilterer) ParseRestakedRewards(log types.Log) (*ContractRestakedRewards, error) {
	event := new(ContractRestakedRewards)
	if err := _Contract.contract.UnpackLog(event, "RestakedRewards", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractUndelegatedIterator is returned from FilterUndelegated and is used to iterate over the raw logs and unpacked data for Undelegated events raised by the Contract contract.
type ContractUndelegatedIterator struct {
	Event *ContractUndelegated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUndelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUndelegated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUndelegated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUndelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUndelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUndelegated represents a Undelegated event raised by the Contract contract.
type ContractUndelegated struct {
	Delegator     common.Address
	ToValidatorID *big.Int
	WrID          *big.Int
	Amount        *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterUndelegated is a free log retrieval operation binding the contract event 0xd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57.
//
// Solidity: event Undelegated(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount)
func (_Contract *ContractFilterer) FilterUndelegated(opts *bind.FilterOpts, delegator []common.Address, toValidatorID []*big.Int, wrID []*big.Int) (*ContractUndelegatedIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}
	var wrIDRule []interface{}
	for _, wrIDItem := range wrID {
		wrIDRule = append(wrIDRule, wrIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Undelegated", delegatorRule, toValidatorIDRule, wrIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractUndelegatedIterator{contract: _Contract.contract, event: "Undelegated", logs: logs, sub: sub}, nil
}

// WatchUndelegated is a free log subscription operation binding the contract event 0xd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57.
//
// Solidity: event Undelegated(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount)
func (_Contract *ContractFilterer) WatchUndelegated(opts *bind.WatchOpts, sink chan<- *ContractUndelegated, delegator []common.Address, toValidatorID []*big.Int, wrID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}
	var wrIDRule []interface{}
	for _, wrIDItem := range wrID {
		wrIDRule = append(wrIDRule, wrIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Undelegated", delegatorRule, toValidatorIDRule, wrIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUndelegated)
				if err := _Contract.contract.UnpackLog(event, "Undelegated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUndelegated is a log parse operation binding the contract event 0xd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57.
//
// Solidity: event Undelegated(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount)
func (_Contract *ContractFilterer) ParseUndelegated(log types.Log) (*ContractUndelegated, error) {
	event := new(ContractUndelegated)
	if err := _Contract.contract.UnpackLog(event, "Undelegated", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractUnlockedStakeIterator is returned from FilterUnlockedStake and is used to iterate over the raw logs and unpacked data for UnlockedStake events raised by the Contract contract.
type ContractUnlockedStakeIterator struct {
	Event *ContractUnlockedStake // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUnlockedStakeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUnlockedStake)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUnlockedStake)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUnlockedStakeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUnlockedStakeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUnlockedStake represents a UnlockedStake event raised by the Contract contract.
type ContractUnlockedStake struct {
	Delegator   common.Address
	ValidatorID *big.Int
	Amount      *big.Int
	Penalty     *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnlockedStake is a free log retrieval operation binding the contract event 0xef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9.
//
// Solidity: event UnlockedStake(address indexed delegator, uint256 indexed validatorID, uint256 amount, uint256 penalty)
func (_Contract *ContractFilterer) FilterUnlockedStake(opts *bind.FilterOpts, delegator []common.Address, validatorID []*big.Int) (*ContractUnlockedStakeIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UnlockedStake", delegatorRule, validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractUnlockedStakeIterator{contract: _Contract.contract, event: "UnlockedStake", logs: logs, sub: sub}, nil
}

// WatchUnlockedStake is a free log subscription operation binding the contract event 0xef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9.
//
// Solidity: event UnlockedStake(address indexed delegator, uint256 indexed validatorID, uint256 amount, uint256 penalty)
func (_Contract *ContractFilterer) WatchUnlockedStake(opts *bind.WatchOpts, sink chan<- *ContractUnlockedStake, delegator []common.Address, validatorID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UnlockedStake", delegatorRule, validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUnlockedStake)
				if err := _Contract.contract.UnpackLog(event, "UnlockedStake", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnlockedStake is a log parse operation binding the contract event 0xef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9.
//
// Solidity: event UnlockedStake(address indexed delegator, uint256 indexed validatorID, uint256 amount, uint256 penalty)
func (_Contract *ContractFilterer) ParseUnlockedStake(log types.Log) (*ContractUnlockedStake, error) {
	event := new(ContractUnlockedStake)
	if err := _Contract.contract.UnpackLog(event, "UnlockedStake", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractUpdatedBaseRewardPerSecIterator is returned from FilterUpdatedBaseRewardPerSec and is used to iterate over the raw logs and unpacked data for UpdatedBaseRewardPerSec events raised by the Contract contract.
type ContractUpdatedBaseRewardPerSecIterator struct {
	Event *ContractUpdatedBaseRewardPerSec // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdatedBaseRewardPerSecIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdatedBaseRewardPerSec)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdatedBaseRewardPerSec)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdatedBaseRewardPerSecIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdatedBaseRewardPerSecIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdatedBaseRewardPerSec represents a UpdatedBaseRewardPerSec event raised by the Contract contract.
type ContractUpdatedBaseRewardPerSec struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterUpdatedBaseRewardPerSec is a free log retrieval operation binding the contract event 0x8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae2396.
//
// Solidity: event UpdatedBaseRewardPerSec(uint256 value)
func (_Contract *ContractFilterer) FilterUpdatedBaseRewardPerSec(opts *bind.FilterOpts) (*ContractUpdatedBaseRewardPerSecIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdatedBaseRewardPerSec")
	if err != nil {
		return nil, err
	}
	return &ContractUpdatedBaseRewardPerSecIterator{contract: _Contract.contract, event: "UpdatedBaseRewardPerSec", logs: logs, sub: sub}, nil
}

// WatchUpdatedBaseRewardPerSec is a free log subscription operation binding the contract event 0x8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae2396.
//
// Solidity: event UpdatedBaseRewardPerSec(uint256 value)
func (_Contract *ContractFilterer) WatchUpdatedBaseRewardPerSec(opts *bind.WatchOpts, sink chan<- *ContractUpdatedBaseRewardPerSec) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdatedBaseRewardPerSec")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdatedBaseRewardPerSec)
				if err := _Contract.contract.UnpackLog(event, "UpdatedBaseRewardPerSec", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdatedBaseRewardPerSec is a log parse operation binding the contract event 0x8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae2396.
//
// Solidity: event UpdatedBaseRewardPerSec(uint256 value)
func (_Contract *ContractFilterer) ParseUpdatedBaseRewardPerSec(log types.Log) (*ContractUpdatedBaseRewardPerSec, error) {
	event := new(ContractUpdatedBaseRewardPerSec)
	if err := _Contract.contract.UnpackLog(event, "UpdatedBaseRewardPerSec", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractUpdatedOfflinePenaltyThresholdIterator is returned from FilterUpdatedOfflinePenaltyThreshold and is used to iterate over the raw logs and unpacked data for UpdatedOfflinePenaltyThreshold events raised by the Contract contract.
type ContractUpdatedOfflinePenaltyThresholdIterator struct {
	Event *ContractUpdatedOfflinePenaltyThreshold // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdatedOfflinePenaltyThresholdIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdatedOfflinePenaltyThreshold)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdatedOfflinePenaltyThreshold)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdatedOfflinePenaltyThresholdIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdatedOfflinePenaltyThresholdIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdatedOfflinePenaltyThreshold represents a UpdatedOfflinePenaltyThreshold event raised by the Contract contract.
type ContractUpdatedOfflinePenaltyThreshold struct {
	BlocksNum *big.Int
	Period    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUpdatedOfflinePenaltyThreshold is a free log retrieval operation binding the contract event 0x702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34.
//
// Solidity: event UpdatedOfflinePenaltyThreshold(uint256 blocksNum, uint256 period)
func (_Contract *ContractFilterer) FilterUpdatedOfflinePenaltyThreshold(opts *bind.FilterOpts) (*ContractUpdatedOfflinePenaltyThresholdIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdatedOfflinePenaltyThreshold")
	if err != nil {
		return nil, err
	}
	return &ContractUpdatedOfflinePenaltyThresholdIterator{contract: _Contract.contract, event: "UpdatedOfflinePenaltyThreshold", logs: logs, sub: sub}, nil
}

// WatchUpdatedOfflinePenaltyThreshold is a free log subscription operation binding the contract event 0x702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34.
//
// Solidity: event UpdatedOfflinePenaltyThreshold(uint256 blocksNum, uint256 period)
func (_Contract *ContractFilterer) WatchUpdatedOfflinePenaltyThreshold(opts *bind.WatchOpts, sink chan<- *ContractUpdatedOfflinePenaltyThreshold) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdatedOfflinePenaltyThreshold")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdatedOfflinePenaltyThreshold)
				if err := _Contract.contract.UnpackLog(event, "UpdatedOfflinePenaltyThreshold", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdatedOfflinePenaltyThreshold is a log parse operation binding the contract event 0x702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34.
//
// Solidity: event UpdatedOfflinePenaltyThreshold(uint256 blocksNum, uint256 period)
func (_Contract *ContractFilterer) ParseUpdatedOfflinePenaltyThreshold(log types.Log) (*ContractUpdatedOfflinePenaltyThreshold, error) {
	event := new(ContractUpdatedOfflinePenaltyThreshold)
	if err := _Contract.contract.UnpackLog(event, "UpdatedOfflinePenaltyThreshold", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractUpdatedSlashingRefundRatioIterator is returned from FilterUpdatedSlashingRefundRatio and is used to iterate over the raw logs and unpacked data for UpdatedSlashingRefundRatio events raised by the Contract contract.
type ContractUpdatedSlashingRefundRatioIterator struct {
	Event *ContractUpdatedSlashingRefundRatio // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractUpdatedSlashingRefundRatioIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractUpdatedSlashingRefundRatio)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractUpdatedSlashingRefundRatio)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractUpdatedSlashingRefundRatioIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractUpdatedSlashingRefundRatioIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractUpdatedSlashingRefundRatio represents a UpdatedSlashingRefundRatio event raised by the Contract contract.
type ContractUpdatedSlashingRefundRatio struct {
	ValidatorID *big.Int
	RefundRatio *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUpdatedSlashingRefundRatio is a free log retrieval operation binding the contract event 0x047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb917.
//
// Solidity: event UpdatedSlashingRefundRatio(uint256 indexed validatorID, uint256 refundRatio)
func (_Contract *ContractFilterer) FilterUpdatedSlashingRefundRatio(opts *bind.FilterOpts, validatorID []*big.Int) (*ContractUpdatedSlashingRefundRatioIterator, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "UpdatedSlashingRefundRatio", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractUpdatedSlashingRefundRatioIterator{contract: _Contract.contract, event: "UpdatedSlashingRefundRatio", logs: logs, sub: sub}, nil
}

// WatchUpdatedSlashingRefundRatio is a free log subscription operation binding the contract event 0x047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb917.
//
// Solidity: event UpdatedSlashingRefundRatio(uint256 indexed validatorID, uint256 refundRatio)
func (_Contract *ContractFilterer) WatchUpdatedSlashingRefundRatio(opts *bind.WatchOpts, sink chan<- *ContractUpdatedSlashingRefundRatio, validatorID []*big.Int) (event.Subscription, error) {

	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "UpdatedSlashingRefundRatio", validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractUpdatedSlashingRefundRatio)
				if err := _Contract.contract.UnpackLog(event, "UpdatedSlashingRefundRatio", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpdatedSlashingRefundRatio is a log parse operation binding the contract event 0x047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb917.
//
// Solidity: event UpdatedSlashingRefundRatio(uint256 indexed validatorID, uint256 refundRatio)
func (_Contract *ContractFilterer) ParseUpdatedSlashingRefundRatio(log types.Log) (*ContractUpdatedSlashingRefundRatio, error) {
	event := new(ContractUpdatedSlashingRefundRatio)
	if err := _Contract.contract.UnpackLog(event, "UpdatedSlashingRefundRatio", log); err != nil {
		return nil, err
	}
	return event, nil
}

// ContractWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Contract contract.
type ContractWithdrawnIterator struct {
	Event *ContractWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractWithdrawn represents a Withdrawn event raised by the Contract contract.
type ContractWithdrawn struct {
	Delegator     common.Address
	ToValidatorID *big.Int
	WrID          *big.Int
	Amount        *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21.
//
// Solidity: event Withdrawn(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount)
func (_Contract *ContractFilterer) FilterWithdrawn(opts *bind.FilterOpts, delegator []common.Address, toValidatorID []*big.Int, wrID []*big.Int) (*ContractWithdrawnIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}
	var wrIDRule []interface{}
	for _, wrIDItem := range wrID {
		wrIDRule = append(wrIDRule, wrIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "Withdrawn", delegatorRule, toValidatorIDRule, wrIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractWithdrawnIterator{contract: _Contract.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21.
//
// Solidity: event Withdrawn(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount)
func (_Contract *ContractFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *ContractWithdrawn, delegator []common.Address, toValidatorID []*big.Int, wrID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var toValidatorIDRule []interface{}
	for _, toValidatorIDItem := range toValidatorID {
		toValidatorIDRule = append(toValidatorIDRule, toValidatorIDItem)
	}
	var wrIDRule []interface{}
	for _, wrIDItem := range wrID {
		wrIDRule = append(wrIDRule, wrIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "Withdrawn", delegatorRule, toValidatorIDRule, wrIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractWithdrawn)
				if err := _Contract.contract.UnpackLog(event, "Withdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawn is a log parse operation binding the contract event 0x75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21.
//
// Solidity: event Withdrawn(address indexed delegator, uint256 indexed toValidatorID, uint256 indexed wrID, uint256 amount)
func (_Contract *ContractFilterer) ParseWithdrawn(log types.Log) (*ContractWithdrawn, error) {
	event := new(ContractWithdrawn)
	if err := _Contract.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	return event, nil
}

var ContractBinRuntime = "0x6080604052600436106106705760003560e01c80638b0e9f3f1161034e578063c65ee0e1116101bb578063de67f215116100f7578063ebdf104c116100a0578063f3ae5b1a1161007a578063f3ae5b1a1461190a578063f8b18d8a1461164e578063f99837e614611934578063fd5e6dd11461196457610670565b8063ebdf104c1461174f578063ec6a7f1c146118c2578063f2fde38b146118d757610670565b8063df4f49d4116100d1578063df4f49d414611678578063e08d7e66146116a2578063e261641a1461171f57610670565b8063de67f215146115e8578063df00c9221461161e578063df0e307a1461164e57610670565b8063cfd5fa0c11610164578063d9a7c1f91161013e578063d9a7c1f91461153a578063dc31e1af1461154f578063dc599bb11461157f578063dd099bb6146115af57610670565b8063cfd5fa0c14611483578063cfdbb7cd146114bc578063d845fc90146114f557610670565b8063cc8343aa11610195578063cc8343aa146113ee578063cda5826a14611420578063cfd476631461144a57610670565b8063c65ee0e1146113af578063c7be95de146113d9578063cb1c4e671461070357610670565b8063b1e643391161028a578063bb03a4bd11610233578063c3de580e1161020d578063c3de580e1461135b578063c41b640514611385578063c4b5dd7e14610703578063c5f530af1461139a57610670565b8063bb03a4bd14611310578063bed9d86114611346578063c312eb071461104e57610670565b8063b810e41111610264578063b810e41114611248578063b82b842714611281578063b88a37e21461129657610670565b8063b1e6433914611189578063b5d89627146111b3578063b6d9edd51461121e57610670565b806396c7ee46116102f7578063a4b89fab116102d1578063a4b89fab1461109b578063a5a470ad146110cb578063a77865151461113b578063a86a056f1461115057610670565b806396c7ee4614610fef5780639fa6dd351461104e578063a198d2291461106b57610670565b80638da5cb5b116103285780638da5cb5b14610f735780638f32d59b14610fa457806396060e7114610fb957610670565b80638b0e9f3f14610ef55780638b1a0d1114610f0a5780638cddb01514610f3a57610670565b806339b80c00116104ec5780636099ecb2116104285780636f498663116103d15780637cacb1d6116103ab5780637cacb1d614610e035780637f664d8714610e1857806381d9dc7a14610718578063854873e114610e5657610670565b80636f49866314610da0578063715018a614610dd95780637667180814610dee57610670565b806363321e271161040257806363321e2714610d1f578063650acd6614610d52578063670322f814610d6757610670565b80636099ecb214610cb657806360c7e37f1461070357806361e53fcc14610cef57610670565b80634f864df41161049557806354fd4d501161046f57806354fd4d5014610c2757806358f95b8014610c715780635e2308d2146109e65780635fab23a814610ca157610670565b80634f864df414610b465780634feb92f314610b7c57806354d77ed21461088e57610670565b8063441a3e70116104c6578063441a3e7014610ad15780634bd202dc14610b015780634f7c4efb14610b1657610670565b806339b80c0014610a925780633d0317fe14610abc5780633fee10a81461088e57610670565b80631d3ac42c116105bb57806326682c7111610564578063295cccba1161053e578063295cccba14610a105780632cedb09714610a3a57806330fa992914610a68578063375b3c0a14610a7d57610670565b806326682c71146109b65780632709275e146109e657806328f73148146109fb57610670565b80631f270152116105955780631f270152146108d3578063223fae09146109305780632265f284146109a157610670565b80631d3ac42c1461085e5780631d58179c1461088e5780631e702f83146108a357610670565b80630d4955e31161061d57806318160ddd116105f757806318160ddd146107e457806318f628d4146107f957806319ddb54f1461070357610670565b80630d4955e3146107815780630d7b26091461079657806312622d0e146107ab57610670565b806308728f6e1161064e57806308728f6e1461071857806308c368741461072d5780630962ef791461075757610670565b80630135b1db14610675578063019e2729146106ba5780630298599214610703575b600080fd5b34801561068157600080fd5b506106a86004803603602081101561069857600080fd5b50356001600160a01b03166119ea565b60408051918252519081900360200190f35b3480156106c657600080fd5b50610701600480360360808110156106dd57600080fd5b508035906020810135906001600160a01b03604082013581169160600135166119fc565b005b34801561070f57600080fd5b506106a8611b73565b34801561072457600080fd5b506106a8611b79565b34801561073957600080fd5b506107016004803603602081101561075057600080fd5b5035611b7f565b34801561076357600080fd5b506107016004803603602081101561077a57600080fd5b5035611c4b565b34801561078d57600080fd5b506106a8611d28565b3480156107a257600080fd5b506106a8611d30565b3480156107b757600080fd5b506106a8600480360360408110156107ce57600080fd5b506001600160a01b038135169060200135611d37565b3480156107f057600080fd5b506106a8611dc0565b34801561080557600080fd5b50610701600480360361012081101561081d57600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e0810135906101000135611dc6565b34801561086a57600080fd5b506106a86004803603604081101561088157600080fd5b5080359060200135611f26565b34801561089a57600080fd5b506106a86120db565b3480156108af57600080fd5b50610701600480360360408110156108c657600080fd5b50803590602001356120ea565b3480156108df57600080fd5b50610912600480360360608110156108f657600080fd5b506001600160a01b038135169060208101359060400135612199565b60408051938452602084019290925282820152519081900360600190f35b34801561093c57600080fd5b506109696004803603604081101561095357600080fd5b506001600160a01b0381351690602001356121cb565b604080519788526020880196909652868601949094526060860192909252608085015260a084015260c0830152519081900360e00190f35b3480156109ad57600080fd5b506106a861223f565b3480156109c257600080fd5b50610701600480360360408110156109d957600080fd5b5080359060200135612251565b3480156109f257600080fd5b506106a8612271565b348015610a0757600080fd5b506106a861228d565b348015610a1c57600080fd5b5061070160048036036020811015610a3357600080fd5b5035612293565b348015610a4657600080fd5b50610a4f6122ac565b6040805192835260208301919091528051918290030190f35b348015610a7457600080fd5b506106a86122b6565b348015610a8957600080fd5b506106a86122c9565b348015610a9e57600080fd5b5061096960048036036020811015610ab557600080fd5b50356122d3565b348015610ac857600080fd5b506106a8612315565b348015610add57600080fd5b5061070160048036036040811015610af457600080fd5b5080359060200135612326565b348015610b0d57600080fd5b506106a8612680565b348015610b2257600080fd5b5061070160048036036040811015610b3957600080fd5b5080359060200135612685565b348015610b5257600080fd5b5061070160048036036060811015610b6957600080fd5b50803590602081013590604001356127c9565b348015610b8857600080fd5b506107016004803603610100811015610ba057600080fd5b6001600160a01b0382351691602081013591810190606081016040820135640100000000811115610bd057600080fd5b820183602082011115610be257600080fd5b80359060200191846001830284011164010000000083111715610c0457600080fd5b919350915080359060208101359060408101359060608101359060800135612b57565b348015610c3357600080fd5b50610c3c612bfd565b604080517fffffff00000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b348015610c7d57600080fd5b506106a860048036036040811015610c9457600080fd5b5080359060200135612c21565b348015610cad57600080fd5b506106a8612c3e565b348015610cc257600080fd5b506106a860048036036040811015610cd957600080fd5b506001600160a01b038135169060200135612c44565b348015610cfb57600080fd5b506106a860048036036040811015610d1257600080fd5b5080359060200135612c82565b348015610d2b57600080fd5b506106a860048036036020811015610d4257600080fd5b50356001600160a01b0316612ca3565b348015610d5e57600080fd5b506106a8612cc2565b348015610d7357600080fd5b506106a860048036036040811015610d8a57600080fd5b506001600160a01b038135169060200135612cc7565b348015610dac57600080fd5b506106a860048036036040811015610dc357600080fd5b506001600160a01b038135169060200135612d08565b348015610de557600080fd5b50610701612d72565b348015610dfa57600080fd5b506106a8612e2d565b348015610e0f57600080fd5b506106a8612e36565b348015610e2457600080fd5b50610e4260048036036020811015610e3b57600080fd5b5035612e3c565b604080519115158252519081900360200190f35b348015610e6257600080fd5b50610e8060048036036020811015610e7957600080fd5b5035612e61565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610eba578181015183820152602001610ea2565b50505050905090810190601f168015610ee75780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b348015610f0157600080fd5b506106a8612f1a565b348015610f1657600080fd5b5061070160048036036040811015610f2d57600080fd5b5080359060200135612f20565b348015610f4657600080fd5b5061070160048036036040811015610f5d57600080fd5b506001600160a01b038135169060200135612fc2565b348015610f7f57600080fd5b50610f8861301d565b604080516001600160a01b039092168252519081900360200190f35b348015610fb057600080fd5b50610e4261302c565b348015610fc557600080fd5b5061091260048036036060811015610fdc57600080fd5b508035906020810135906040013561303d565b348015610ffb57600080fd5b506110286004803603604081101561101257600080fd5b506001600160a01b038135169060200135613095565b604080519485526020850193909352838301919091526060830152519081900360800190f35b6107016004803603602081101561106457600080fd5b50356130c7565b34801561107757600080fd5b506106a86004803603604081101561108e57600080fd5b50803590602001356130d5565b3480156110a757600080fd5b50610701600480360360408110156110be57600080fd5b50803590602001356130f6565b610701600480360360208110156110e157600080fd5b8101906020810181356401000000008111156110fc57600080fd5b82018360208201111561110e57600080fd5b8035906020019184600183028401116401000000008311171561113057600080fd5b50909250905061311e565b34801561114757600080fd5b506106a8613219565b34801561115c57600080fd5b506106a86004803603604081101561117357600080fd5b506001600160a01b03813516906020013561322f565b34801561119557600080fd5b50610701600480360360208110156111ac57600080fd5b503561324c565b3480156111bf57600080fd5b506111dd600480360360208110156111d657600080fd5b5035613299565b604080519788526020880196909652868601949094526060860192909252608085015260a08401526001600160a01b031660c0830152519081900360e00190f35b34801561122a57600080fd5b506107016004803603602081101561124157600080fd5b50356132df565b34801561125457600080fd5b506109126004803603604081101561126b57600080fd5b506001600160a01b0381351690602001356133d1565b34801561128d57600080fd5b506106a86133fd565b3480156112a257600080fd5b506112c0600480360360208110156112b957600080fd5b5035613404565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156112fc5781810151838201526020016112e4565b505050509050019250505060405180910390f35b34801561131c57600080fd5b506107016004803603606081101561133357600080fd5b5080359060208101359060400135613469565b34801561135257600080fd5b50610701613474565b34801561136757600080fd5b50610e426004803603602081101561137e57600080fd5b50356134c1565b34801561139157600080fd5b5061070161324c565b3480156113a657600080fd5b506106a86134d8565b3480156113bb57600080fd5b506106a8600480360360208110156113d257600080fd5b50356134e7565b3480156113e557600080fd5b506106a86134f9565b3480156113fa57600080fd5b506107016004803603604081101561141157600080fd5b508035906020013515156134ff565b34801561142c57600080fd5b506107016004803603602081101561144357600080fd5b5035613737565b34801561145657600080fd5b506106a86004803603604081101561146d57600080fd5b506001600160a01b038135169060200135613750565b34801561148f57600080fd5b50610e42600480360360408110156114a657600080fd5b506001600160a01b03813516906020013561376d565b3480156114c857600080fd5b50610e42600480360360408110156114df57600080fd5b506001600160a01b038135169060200135613775565b34801561150157600080fd5b506109126004803603608081101561151857600080fd5b506001600160a01b0381351690602081013590604081013590606001356137dd565b34801561154657600080fd5b506106a861381b565b34801561155b57600080fd5b506106a86004803603604081101561157257600080fd5b5080359060200135613821565b34801561158b57600080fd5b50610701600480360360408110156115a257600080fd5b5080359060200135613842565b3480156115bb57600080fd5b50610912600480360360408110156115d257600080fd5b506001600160a01b03813516906020013561384b565b3480156115f457600080fd5b506107016004803603606081101561160b57600080fd5b50803590602081013590604001356138b7565b34801561162a57600080fd5b506106a86004803603604081101561164157600080fd5b5080359060200135613bf5565b34801561165a57600080fd5b506107016004803603602081101561167157600080fd5b5035613474565b34801561168457600080fd5b506109126004803603602081101561169b57600080fd5b5035613c16565b3480156116ae57600080fd5b50610701600480360360208110156116c557600080fd5b8101906020810181356401000000008111156116e057600080fd5b8201836020820111156116f257600080fd5b8035906020019184602083028401116401000000008311171561171457600080fd5b509092509050613c4c565b34801561172b57600080fd5b506106a86004803603604081101561174257600080fd5b5080359060200135613d2c565b34801561175b57600080fd5b506107016004803603608081101561177257600080fd5b81019060208101813564010000000081111561178d57600080fd5b82018360208201111561179f57600080fd5b803590602001918460208302840111640100000000831117156117c157600080fd5b9193909290916020810190356401000000008111156117df57600080fd5b8201836020820111156117f157600080fd5b8035906020019184602083028401116401000000008311171561181357600080fd5b91939092909160208101903564010000000081111561183157600080fd5b82018360208201111561184357600080fd5b8035906020019184602083028401116401000000008311171561186557600080fd5b91939092909160208101903564010000000081111561188357600080fd5b82018360208201111561189557600080fd5b803590602001918460208302840111640100000000831117156118b757600080fd5b509092509050613d4d565b3480156118ce57600080fd5b506106a8613f29565b3480156118e357600080fd5b50610701600480360360208110156118fa57600080fd5b50356001600160a01b0316613f33565b34801561191657600080fd5b506107016004803603602081101561192d57600080fd5b5035613f95565b34801561194057600080fd5b506107016004803603604081101561195757600080fd5b5080359060200135613fb8565b34801561197057600080fd5b5061198e6004803603602081101561198757600080fd5b5035613fc1565b604080519a8b5260208b0199909952898901979097526060890195909552608088019390935260a087019190915260c086015260e08501526001600160a01b039081166101008501521661012083015251908190036101400190f35b60696020526000908152604090205481565b600054610100900460ff1680611a155750611a156140e4565b80611a23575060005460ff16155b611a5e5760405162461bcd60e51b815260040180806020018281038252602e815260200180615f3e602e913960400191505060405180910390fd5b600054610100900460ff16158015611ac457600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b611acd826140ea565b6067859055606680547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b03851617905560768490556755cfe697852e904c6075556103e86078556203f480607955611b2b61424b565b6000868152607760205260409020600701558015611b6c57600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555b5050505050565b60015b90565b606b5490565b33611b88615d35565b611b92828461424f565b60208101518151919250600091611bae9163ffffffff61435416565b9050611bd18385611bcc85604001518561435490919063ffffffff16565b6143ae565b6001600160a01b0383166000818152607360209081526040808320888452825291829020805485019055845185820151868401518451928352928201528083019190915290518692917f4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26919081900360600190a350505050565b33611c54615d35565b611c5e828461424f565b9050816001600160a01b03166108fc611c9c8360400151611c908560200151866000015161435490919063ffffffff16565b9063ffffffff61435416565b6040518115909202916000818181858888f19350505050158015611cc4573d6000803e3d6000fd5b5082826001600160a01b03167fc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf4583600001518460200151856040015160405180848152602001838152602001828152602001935050505060405180910390a3505050565b6301e1338090565b6212750090565b6000611d438383613775565b611d7157506001600160a01b0382166000908152607260209081526040808320848452909152902054611dba565b6001600160a01b038316600081815260736020908152604080832086845282528083205493835260728252808320868452909152902054611db79163ffffffff6144b816565b90505b92915050565b60765481565b611dcf336144fa565b611e0a5760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b611e1589898961450e565b6001600160a01b0389166000908152606f602090815260408083208b84529091529020600201819055611e4787614685565b8515611f1b5786861115611e8c5760405162461bcd60e51b815260040180806020018281038252602c815260200180615fde602c913960400191505060405180910390fd5b6001600160a01b03891660008181526073602090815260408083208c845282528083208a8155600181018a90556002810189905560038101889055848452607483528184208d855283529281902086905580518781529182018a9052805192938c9390927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e7892908290030190a3505b505050505050505050565b336000818152607360209081526040808320868452909152812090919083611f95576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b611f9f8286613775565b611ff0576040805162461bcd60e51b815260206004820152600d60248201527f6e6f74206c6f636b656420757000000000000000000000000000000000000000604482015290519081900360640190fd5b8054841115612046576040805162461bcd60e51b815260206004820152601760248201527f6e6f7420656e6f756768206c6f636b6564207374616b65000000000000000000604482015290519081900360640190fd5b6120508286614705565b506000612063838787856000015461484d565b825486900383556001600160a01b03841660008181526072602090815260408083208b8452825291829020805485900390558151898152908101849052815193945089937fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9929181900390910190a395945050505050565b60006120e5612cc2565b905090565b6120f3336144fa565b61212e5760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b80612180576040805162461bcd60e51b815260206004820152600c60248201527f77726f6e67207374617475730000000000000000000000000000000000000000604482015290519081900360640190fd5b61218a8282614996565b6121958260006134ff565b5050565b607160209081526000938452604080852082529284528284209052825290208054600182015460029092015490919083565b6001600160a01b03821660009081526072602090815260408083208484529091528120548190819081908190819081908061221c575060009650869550859450849350839250829150819050612233565b600197508796506000955085945092508591508790505b92959891949750929550565b6000612249614ac0565b601002905090565b3360009081526069602052604090205461226c8184846127c9565b505050565b6000606461227d614ac0565b601e028161228757fe5b04905090565b606d5481565b3360009081526069602052604090205461219581611c4b565b6078546079549091565b60006122c0612315565b606c5403905090565b60006120e56134d8565b607760205280600052604060002060009150905080600701549080600801549080600901549080600a01549080600b01549080600c01549080600d0154905087565b60006064606c546018028161228757fe5b3361232f615d35565b506001600160a01b0381166000908152607160209081526040808320868452825280832085845282529182902082516060810184528154808252600183015493820193909352600290910154928101929092526123d3576040805162461bcd60e51b815260206004820152601560248201527f7265717565737420646f65736e27742065786973740000000000000000000000604482015290519081900360640190fd5b6020808201518251600087815260689093526040909220600101549091901580159061240f575060008681526068602052604090206001015482115b15612430575050600084815260686020526040902060018101546002909101545b6124386133fd565b820161244261424b565b1015612495576040805162461bcd60e51b815260206004820152601660248201527f6e6f7420656e6f7567682074696d652070617373656400000000000000000000604482015290519081900360640190fd5b61249d612cc2565b81016124a7612e2d565b10156124fa576040805162461bcd60e51b815260206004820152601860248201527f6e6f7420656e6f7567682065706f636873207061737365640000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152607160209081526040808320898452825280832088845290915281206002015490612533886134c1565b905060006125558383607a60008d815260200190815260200160002054614acc565b6001600160a01b03881660009081526071602090815260408083208d845282528083208c845290915281208181556001810182905560020155606e80548201905590508083116125ec576040805162461bcd60e51b815260206004820152601660248201527f7374616b652069732066756c6c7920736c617368656400000000000000000000604482015290519081900360640190fd5b6001600160a01b0387166108fc612609858463ffffffff6144b816565b6040518115909202916000818181858888f19350505050158015612631573d6000803e3d6000fd5b508789886001600160a01b03167f75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21866040518082815260200191505060405180910390a4505050505050505050565b600090565b61268d61302c565b6126de576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6126e7826134c1565b612738576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f722069736e277420736c6173686564000000000000000000604482015290519081900360640190fd5b612740614ac0565b81111561277e5760405162461bcd60e51b8152600401808060200182810382526021815260200180615f6c6021913960400191505060405180910390fd5b6000828152607a60209081526040918290208390558151838152915184927f047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb91792908290030190a25050565b336127d48185614705565b506000821161282a576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6128348185611d37565b821115612888576040805162461bcd60e51b815260206004820152601960248201527f6e6f7420656e6f75676820756e6c6f636b6564207374616b6500000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607160209081526040808320878452825280832086845290915290206002015415612909576040805162461bcd60e51b815260206004820152601360248201527f7772494420616c72656164792065786973747300000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607260209081526040808320878452825280832080548690039055606890915290206003015461294e908363ffffffff6144b816565b600085815260686020526040902060030155606c54612973908363ffffffff6144b816565b606c556000848152606860205260409020546129a057606d5461299c908363ffffffff6144b816565b606d555b60006129ab85614b2e565b90508015612a58576129bb6134d8565b811015612a0f576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b612a1885614b60565b612a535760405162461bcd60e51b8152600401808060200182810382526029815260200180615fb56029913960400191505060405180910390fd5b612a63565b612a63856001614996565b6001600160a01b038216600090815260716020908152604080832088845282528083208784529091529020600201839055612a9c612e2d565b6001600160a01b03831660009081526071602090815260408083208984528252808320888452909152902055612ad061424b565b6001600160a01b03831660009081526071602090815260408083208984528252808320888452909152812060010191909155612b0d9086906134ff565b8385836001600160a01b03167fd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57866040518082815260200191505060405180910390a45050505050565b612b60336144fa565b612b9b5760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b612be3898989898080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b92508a91508990508888614ba8565b606b54881115611f1b57606b889055505050505050505050565b7f333030000000000000000000000000000000000000000000000000000000000090565b600091825260776020908152604080842092845291905290205490565b606e5481565b6000612c4e615d35565b612c588484614d6f565b805160208201516040830151929350612c7a92611c909163ffffffff61435416565b949350505050565b60009182526077602090815260408084209284526001909201905290205490565b6001600160a01b0381166000908152606960205260409020545b919050565b600390565b6000612cd38383613775565b612cdf57506000611dba565b506001600160a01b03919091166000908152607360209081526040808320938352929052205490565b6000612d12615d35565b506001600160a01b0383166000908152606f6020908152604080832085845282529182902082516060810184528154808252600183015493820184905260029092015493810184905292612c7a929091611c90919063ffffffff61435416565b612d7a61302c565b612dcb576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055565b60675460010190565b60675481565b600081815260686020526040812060060154611dba906001600160a01b031683613775565b606a6020908152600091825260409182902080548351601f60027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff61010060018616150201909316929092049182018490048402810184019094528084529091830182828015612f125780601f10612ee757610100808354040283529160200191612f12565b820191906000526020600020905b815481529060010190602001808311612ef557829003601f168201915b505050505081565b606c5481565b612f2861302c565b612f79576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b60798190556078829055604080518381526020810183905281517f702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34929181900390910190a15050565b612fcc8282614705565b612195576040805162461bcd60e51b815260206004820152601060248201527f6e6f7468696e6720746f20737461736800000000000000000000000000000000604482015290519081900360640190fd5b6033546001600160a01b031690565b6033546001600160a01b0316331490565b600083815260686020526040812060060154819081908190613068906001600160a01b031688612c44565b9050806130805750600092506001915082905061308c565b60675490935091508190505b93509350939050565b607360209081526000928352604080842090915290825290208054600182015460028301546003909301549192909184565b6130d23382346143ae565b50565b60009182526077602090815260408084209284526005909201905290205490565b33600090815260726020908152604080832084845290915290205461219590829084906138b7565b6131266134d8565b34101561317a576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b806131cc576040805162461bcd60e51b815260206004820152600c60248201527f656d707479207075626b65790000000000000000000000000000000000000000604482015290519081900360640190fd5b61320c3383838080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250614ddd92505050565b61219533606b54346143ae565b60006064613225614ac0565b600f028161228757fe5b607060209081526000928352604080842090915290825290205481565b6040805162461bcd60e51b815260206004820152601f60248201527f75736520534643763320756e64656c656761746528292066756e6374696f6e00604482015290519081900360640190fd5b606860205260009081526040902080546001820154600283015460038401546004850154600586015460069096015494959394929391929091906001600160a01b031687565b6132e761302c565b613338576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6801c985c8903591eb20811115613396576040805162461bcd60e51b815260206004820152601b60248201527f746f6f206c617267652072657761726420706572207365636f6e640000000000604482015290519081900360640190fd5b60758190556040805182815290517f8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae23969181900360200190a150565b607460209081526000928352604080842090915290825290208054600182015460029092015490919083565b62093a8090565b60008181526077602090815260409182902060060180548351818402810184019094528084526060939283018282801561345d57602002820191906000526020600020905b815481526020019060010190808311613449575b50505050509050919050565b61226c8284836127c9565b6040805162461bcd60e51b815260206004820152601d60248201527f75736520534643763320776974686472617728292066756e6374696f6e000000604482015290519081900360640190fd5b600090815260686020526040902054608016151590565b6a02a055184a310c1260000090565b607a6020526000908152604090205481565b606b5481565b61350882614e08565b613559576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090206003810154905415613577575060005b606654604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018690526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b1580156135e457600080fd5b505af11580156135f8573d6000803e3d6000fd5b5050505081801561360857508015155b1561226c576066546000848152606a60205260409081902081517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810187815260248201938452825460027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6001831615610100020190911604604483018190526001600160a01b039095169463242a6e3f948994939091606490910190849080156136f95780601f106136ce576101008083540402835291602001916136f9565b820191906000526020600020905b8154815290600101906020018083116136dc57829003601f168201915b50509350505050600060405180830381600087803b15801561371a57600080fd5b505af115801561372e573d6000803e3d6000fd5b50505050505050565b3360009081526069602052604090205461219581611b7f565b607260209081526000928352604080842090915290825290205481565b6000611db783835b6001600160a01b038216600090815260736020908152604080832084845290915281206002015415801590611db757506001600160a01b03831660009081526073602090815260408083208584529091529020600201546137d461424b565b11159392505050565b6000806000806137ed8888612c44565b90508061380557506000925060019150829050613811565b60675490935091508190505b9450945094915050565b60755481565b60009182526077602090815260408084209284526003909201905290205490565b61219581611b7f565b6000806000613858615d56565b505050506001600160a01b03919091166000908152607360209081526040808320938352928152908290208251608081018452815481526001820154928101839052600282015493810184905260039091015460609091018190529092565b338161390a576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6139148185613775565b15613966576040805162461bcd60e51b815260206004820152601160248201527f616c7265616479206c6f636b6564207570000000000000000000000000000000604482015290519081900360640190fd5b6139708185611d37565b8211156139c4576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b60008481526068602052604090205415613a25576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b613a2d611d30565b8310158015613a435750613a3f611d28565b8311155b613a94576040805162461bcd60e51b815260206004820152601260248201527f696e636f7272656374206475726174696f6e0000000000000000000000000000604482015290519081900360640190fd5b6000613aa284611c9061424b565b6000868152606860205260409020600601549091506001600160a01b039081169083168114613b30576001600160a01b0381166000908152607360209081526040808320898452909152902060020154821115613b305760405162461bcd60e51b8152600401808060200182810382526028815260200180615f8d6028913960400191505060405180910390fd5b613b3a8387614705565b506001600160a01b03831660009081526073602090815260408083208984529091529020848155613b69612e2d565b6001808301919091556002808301859055600383018890556001600160a01b03861660008181526074602090815260408083208d845282528083208381559586018390559490930155825189815291820188905282518a9391927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e7892908290030190a350505050505050565b60009182526077602090815260408084209284526002909201905290205490565b60008181526068602052604081206006015481908190613c3f906001600160a01b03168561384b565b9250925092509193909250565b613c55336144fa565b613c905760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b600060776000613c9e612e2d565b8152602001908152602001600020905060008090505b82811015613d17576000848483818110613cca57fe5b60209081029290920135600081815260688452604080822060030154948890529020839055600c860154909350613d0891508263ffffffff61435416565b600c8501555050600101613cb4565b50613d26600682018484615d7e565b50505050565b60009182526077602090815260408084209284526004909201905290205490565b613d56336144fa565b613d915760405162461bcd60e51b8152600401808060200182810382526029815260200180615ef46029913960400191505060405180910390fd5b600060776000613d9f612e2d565b81526020019081526020016000209050606081600601805480602002602001604051908101604052809291908181526020018280548015613dff57602002820191906000526020600020905b815481526020019060010190808311613deb575b50505050509050613e8682828c8c80806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050508b8b80806020026020016040519081016040528093929190818152602001838360200280828437600092019190915250614e1f92505050565b613ef5828288888080602002602001604051908101604052809392919081815260200183836020028082843760009201919091525050604080516020808c0282810182019093528b82529093508b92508a918291850190849080828437600092019190915250614f2e92505050565b613efd612e2d565b606755613f0861424b565b600783015550607554600b820155607654600d909101555050505050505050565b60006120e56133fd565b613f3b61302c565b613f8c576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6130d28161554e565b336000908152606960205260409020546121958183613fb382614b2e565b6138b7565b61219581611c4b565b600080600080600080600080600080613fd8615dc9565b5060008b815260686020908152604091829020825160e08101845281548082526001830154938201939093526002820154938101939093526003810154606084015260048101546080840152600581015460a0840152600601546001600160a01b031660c083015260081415614052576101008152614074565b8051608014156140655760018152614074565b80516001141561407457600081525b600061407f8d614b2e565b9050816000015182608001518360a00151846040015185602001518560016140b4888a606001516144b890919063ffffffff16565b8960c001518a60c001518393509b509b509b509b509b509b509b509b509b509b5050509193959799509193959799565b303b1590565b600054610100900460ff168061410357506141036140e4565b80614111575060005460ff16155b61414c5760405162461bcd60e51b815260040180806020018281038252602e815260200180615f3e602e913960400191505060405180910390fd5b600054610100900460ff161580156141b257600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff909116610100171660011790555b603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0384811691909117918290556040519116906000907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a3801561219557600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff1690555050565b4290565b614257615d35565b6142618383614705565b50506001600160a01b0382166000908152606f60209081526040808320848452825280832081516060810183528154808252600183015494820185905260029092015492810183905293926142bf92611c909163ffffffff61435416565b905080614313576040805162461bcd60e51b815260206004820152600c60248201527f7a65726f20726577617264730000000000000000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152606f602090815260408083208684529091528120818155600181018290556002015561434d81614685565b5092915050565b600082820183811015611db7576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b6143b782614e08565b614408576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090205415614469576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b61447483838361450e565b61447d82614b60565b61226c5760405162461bcd60e51b8152600401808060200182810382526029815260200180615fb56029913960400191505060405180910390fd5b6000611db783836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f770000815250615607565b6066546001600160a01b0390811691161490565b60008111614563576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b61456d8383614705565b506001600160a01b03831660009081526072602090815260408083208584529091529020546145a2908263ffffffff61435416565b6001600160a01b03841660009081526072602090815260408083208684528252808320939093556068905220600301546145e2818363ffffffff61435416565b600084815260686020526040902060030155606c54614607908363ffffffff61435416565b606c5560008381526068602052604090205461463457606d54614630908363ffffffff61435416565b606d555b61463f8382156134ff565b60408051838152905184916001600160a01b038716917f9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb9181900360200190a350505050565b606654604080517f66e7ea0f0000000000000000000000000000000000000000000000000000000081523060048201526024810184905290516001600160a01b03909216916366e7ea0f9160448082019260009290919082900301818387803b1580156146f157600080fd5b505af1158015611b6c573d6000803e3d6000fd5b600061470f615d35565b614719848461569e565b9050614724836157d6565b6001600160a01b0385166000818152607060209081526040808320888452825280832094909455918152606f82528281208682528252829020825160608101845281548152600182015492810192909252600201549181019190915261478a9082615831565b6001600160a01b0385166000818152606f60209081526040808320888452825280832085518155858301516001808301919091559582015160029182015593835260748252808320888452825291829020825160608101845281548152948101549185019190915290910154908201526148049082615831565b6001600160a01b03851660009081526074602090815260408083208784528252918290208351815590830151600180830191909155929091015160029091015591505092915050565b6001600160a01b03841660009081526074602090815260408083208684529091528120548190614895908490614889908763ffffffff6158a316565b9063ffffffff6158fc16565b6001600160a01b0387166000908152607460209081526040808320898452909152812060010154919250906148d6908590614889908863ffffffff6158a316565b905060028104820160006148ee86614889848a6158a3565b6001600160a01b038a1660009081526074602090815260408083208c8452909152902054909150614925908563ffffffff6144b816565b6001600160a01b038a1660009081526074602090815260408083208c845290915290209081556001015461495990846144b8565b6001600160a01b038a1660009081526074602090815260408083208c845290915290206001015586811061498a5750855b98975050505050505050565b6000828152606860205260409020541580156149b157508015155b156149de57600082815260686020526040902060030154606d546149da9163ffffffff6144b816565b606d555b60008281526068602052604090205481111561219557600082815260686020526040902081815560020154614a8657614a15612e2d565b600083815260686020526040902060020155614a2f61424b565b6000838152606860209081526040918290206001810184905560020154825190815290810192909252805184927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e4792908290030190a25b60408051828152905183917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25050565b670de0b6b3a764000090565b6000821580614ae25750614ade614ac0565b8210155b15614aef57506000614b27565b614b1a6001611c90614aff614ac0565b61488986614b0b614ac0565b8a91900363ffffffff6158a316565b905083811115614b275750825b9392505050565b6000818152606860209081526040808320600601546001600160a01b0316835260728252808320938352929052205490565b6000614b8d614b6d614ac0565b614889614b7861223f565b614b8186614b2e565b9063ffffffff6158a316565b60008381526068602052604090206003015411159050919050565b6001600160a01b03881660009081526069602052604090205415614c13576040805162461bcd60e51b815260206004820152601860248201527f76616c696461746f7220616c7265616479206578697374730000000000000000604482015290519081900360640190fd5b6001600160a01b03881660008181526069602090815260408083208b90558a8352606882528083208981556004810189905560058101889055600181018690556002810187905560060180547fffffffffffffffffffffffff000000000000000000000000000000000000000016909417909355606a81529190208751614c9c92890190615e0f565b50876001600160a01b0316877f49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf8686604051808381526020018281526020019250505060405180910390a38115614d28576040805183815260208101839052815189927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47928290030190a25b8415614d655760408051868152905188917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25b5050505050505050565b614d77615d35565b614d7f615d35565b614d89848461569e565b6001600160a01b0385166000908152606f602090815260408083208784528252918290208251606081018452815481526001820154928101929092526002015491810191909152909150612c7a9082615831565b606b80546001019081905561226c8382846000614df8612e2d565b614e0061424b565b600080614ba8565b600090815260686020526040902060050154151590565b60005b8351811015611b6c57607854828281518110614e3a57fe5b6020026020010151118015614e645750607954838281518110614e5957fe5b602002602001015110155b15614ea557614e87848281518110614e7857fe5b60200260200101516008614996565b614ea5848281518110614e9657fe5b602002602001015160006134ff565b828181518110614eb157fe5b6020026020010151856004016000868481518110614ecb57fe5b6020026020010151815260200190815260200160002081905550818181518110614ef157fe5b6020026020010151856005016000868481518110614f0b57fe5b602090810291909101810151825281019190915260400160002055600101614e22565b614f36615e7d565b6040518060c001604052808551604051908082528060200260200182016040528015614f6c578160200160208202803883390190505b508152602001600081526020018551604051908082528060200260200182016040528015614fa4578160200160208202803883390190505b508152602001600081526020016000815260200160008152509050600060776000614fde6001614fd2612e2d565b9063ffffffff6144b816565b8152602081019190915260400160002060016080840152600781015490915061500561424b565b111561501f57806007015461501861424b565b0360808301525b60005b85518110156150ea57826080015185828151811061503c57fe5b602002602001015185838151811061505057fe5b6020026020010151028161506057fe5b048360400151828151811061507157fe5b6020026020010181815250506150ab8360400151828151811061509057fe5b6020026020010151846060015161435490919063ffffffff16565b606084015283516150dd908590839081106150c257fe5b60200260200101518460a0015161435490919063ffffffff16565b60a0840152600101615022565b5060005b85518110156151bb57826080015185828151811061510857fe5b6020026020010151846080015187848151811061512157fe5b60200260200101518a60000160008b878151811061513b57fe5b6020026020010151815260200190815260200160002054028161515a57fe5b04028161516357fe5b048360000151828151811061517457fe5b6020026020010181815250506151ae8360000151828151811061519357fe5b6020026020010151846020015161435490919063ffffffff16565b60208401526001016150ee565b5060005b85518110156155265760006151f78460800151607554866000015185815181106151e557fe5b6020026020010151876020015161593e565b90506152336152268560a001518660400151858151811061521457fe5b6020026020010151876060015161597f565b829063ffffffff61435416565b9050600087838151811061524357fe5b6020908102919091018101516000818152606890925260408220600601549092506001600160a01b03169061527f8461527a613219565b6159dc565b6001600160a01b03831660009081526072602090815260408083208784529091529020549091508015615426576000816152b98587612cc7565b8402816152c257fe5b0490508083036152d0615d35565b6001600160a01b03861660009081526073602090815260408083208a84529091529020600301546153029084906159f9565b905061530c615d35565b6153178360006159f9565b6001600160a01b0388166000908152606f602090815260408083208c8452825291829020825160608101845281548152600182015492810192909252600201549181019190915290915061536c908383615aea565b6001600160a01b0388166000818152606f602090815260408083208d84528252808320855181558583015160018083019190915595820151600291820155938352607482528083208d8452825291829020825160608101845281548152948101549185019190915290910154908201526153e7908383615aea565b6001600160a01b03881660009081526074602090815260408083208c845282529182902083518155908301516001820155910151600290910155505050505b600084815260686020526040812060030154838703918115615458578161544b614ac0565b84028161545457fe5b0490505b808a600101600089815260200190815260200160002054018f6001016000898152602001908152602001600020819055508b898151811061549557fe5b60200260200101518a600301600089815260200190815260200160002054018f6003016000898152602001908152602001600020819055508c89815181106154d957fe5b60200260200101518a600201600089815260200190815260200160002054018f600201600089815260200190815260200160002081905550505050505050505080806001019150506151bf565b505060a081015160088601556020810151600986015560600151600a90940193909355505050565b6001600160a01b0381166155935760405162461bcd60e51b8152600401808060200182810382526026815260200180615ece6026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b600081848411156156965760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b8381101561565b578181015183820152602001615643565b50505050905090810190601f1680156156885780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b6156a6615d35565b6001600160a01b0383166000908152607060209081526040808320858452909152812054906156d4846157d6565b905060006156e28686615b05565b9050818111156156ef5750805b828110156156fa5750815b6001600160a01b0386166000818152607360209081526040808320898452825280832093835260728252808320898452909152812054825490919061574690839063ffffffff6144b816565b9050600061575a84600001548a8988615be2565b9050615764615d35565b6157728286600301546159f9565b9050615780838b8a89615be2565b915061578a615d35565b6157958360006159f9565b90506157a3858c898b615be2565b92506157ad615d35565b6157b88460006159f9565b90506157c5838383615aea565b9d9c50505050505050505050505050565b600081815260686020526040812060020154156158295760008281526068602052604090206002015460675410156158115750606754612cbd565b50600081815260686020526040902060020154612cbd565b505060675490565b615839615d35565b604080516060810190915282518451829161585a919063ffffffff61435416565b815260200161587a8460200151866020015161435490919063ffffffff16565b815260200161589a8460400151866040015161435490919063ffffffff16565b90529392505050565b6000826158b257506000611dba565b828202828482816158bf57fe5b0414611db75760405162461bcd60e51b8152600401808060200182810382526021815260200180615f1d6021913960400191505060405180910390fd5b6000611db783836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f000000000000815250615c50565b60008261594d57506000612c7a565b600061595f868663ffffffff6158a316565b905061597583614889838763ffffffff6158a316565b9695505050505050565b60008261598e57506000614b27565b60006159a483614889878763ffffffff6158a316565b90506159d36159b1614ac0565b6148896159bc612271565b6159c4614ac0565b8591900363ffffffff6158a316565b95945050505050565b6000611db76159e9614ac0565b614889858563ffffffff6158a316565b615a01615d35565b604051806060016040528060008152602001600081526020016000815250905081600014615abc576000615a33612271565b615a3b614ac0565b0390506000615a5b615a4b611d28565b614889848763ffffffff6158a316565b90506000615a84615a6a614ac0565b61488984615a76612271565b8a910163ffffffff6158a316565b9050615aa9615a91614ac0565b614889615a9c612271565b899063ffffffff6158a316565b602085018190529003835250611dba9050565b615adf615ac7614ac0565b614889615ad2612271565b869063ffffffff6158a316565b604082015292915050565b615af2615d35565b612c7a615aff8585615831565b83615831565b6001600160a01b0382166000908152607360209081526040808320848452909152812060010154606754615b3a858583615cb5565b15615b48579150611dba9050565b615b53858584615cb5565b615b6257600092505050611dba565b80821115615b7557600092505050611dba565b80821015615ba857600281830104615b8e868683615cb5565b15615b9e57806001019250615ba2565b8091505b50615b75565b80615bb857600092505050611dba565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01949350505050565b6000818310615bf357506000612c7a565b600083815260776020818152604080842088855260019081018352818520548786529383528185208986520190915290912054615c45615c31614ac0565b61488989614b81858763ffffffff6144b816565b979650505050505050565b60008183615c9f5760405162461bcd60e51b815260206004820181815283516024840152835190928392604490910191908501908083836000831561565b578181015183820152602001615643565b506000838581615cab57fe5b0495945050505050565b6001600160a01b03831660009081526073602090815260408083208584529091528120600101548210801590612c7a57506001600160a01b0384166000908152607360209081526040808320868452909152902060020154615d1683615d20565b1115949350505050565b60009081526077602052604090206007015490565b60405180606001604052806000815260200160008152602001600081525090565b6040518060800160405280600081526020016000815260200160008152602001600081525090565b828054828255906000526020600020908101928215615db9579160200282015b82811115615db9578235825591602001919060010190615d9e565b50615dc5929150615eb3565b5090565b6040518060e0016040528060008152602001600081526020016000815260200160008152602001600081526020016000815260200160006001600160a01b031681525090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f10615e5057805160ff1916838001178555615db9565b82800160010185558215615db9579182015b82811115615db9578251825591602001919060010190615e62565b6040518060c001604052806060815260200160008152602001606081526020016000815260200160008152602001600081525090565b611b7691905b80821115615dc55760008155600101615eb956fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f206164647265737363616c6c6572206973206e6f7420746865204e6f64654472697665724175746820636f6e7472616374536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f77436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a65646d757374206265206c657373207468616e206f7220657175616c20746f20312e3076616c696461746f72206c6f636b757020706572696f642077696c6c20656e64206561726c69657276616c696461746f7227732064656c65676174696f6e73206c696d69742069732065786365656465646c6f636b6564207374616b652069732067726561746572207468616e207468652077686f6c65207374616b65a265627a7a723158204bc860db30ff399bf4cc12003067479c85b0a58310e666fbfa89ae82ed3dfd0a64736f6c634300050c0032"
