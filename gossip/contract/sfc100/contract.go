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
const ContractABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"LockedUpStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"penalty\",\"type\":\"uint256\"}],\"name\":\"UnlockedStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"UpdatedBaseRewardPerSec\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blocksNum\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"period\",\"type\":\"uint256\"}],\"name\":\"UpdatedOfflinePenaltyThreshold\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"UpdatedSlashingRefundRatio\",\"type\":\"event\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"toEpoch\",\"type\":\"uint256\"}],\"name\":\"_nonStashedRewardsOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"syncPubkey\",\"type\":\"bool\"}],\"name\":\"_syncValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"baseRewardPerSecond\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"claimRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"claimedRewardUntilEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"contractCommission\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"createValidator\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentSealedEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"deactivateValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"delegationLockPeriodEpochs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalBaseRewardWeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalTxRewardWeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseRewardPerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotAccumulatedOriginatedTxsFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotAccumulatedRewardPerToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotAccumulatedUptime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotOfflineBlocks\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotOfflineTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotReceivedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshotValidatorIDs\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"getLockedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getLockupInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"getUnlockedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getValidator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"receivedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"getValidatorID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getValidatorPubkey\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getWithdrawalRequest\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sealedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_totalSupply\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"nodeDriver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"isLockedUp\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"isSlashed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastValidatorID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"lockStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxDelegatedRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"maxLockupDuration\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minLockupDuration\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minSelfStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"offlinePenaltyThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"blocksNum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"pendingRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"rewardsStash\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"offlineTime\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"offlineBlocks\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"uptimes\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"originatedTxsFee\",\"type\":\"uint256[]\"}],\"name\":\"sealEpoch\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nextValidatorIDs\",\"type\":\"uint256[]\"}],\"name\":\"sealEpochValidators\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupFromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupEndTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewards\",\"type\":\"uint256\"}],\"name\":\"setGenesisDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"setGenesisValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slashingRefundRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakeLockPeriodTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"stashRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalActiveStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSlashedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"unlockStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"unlockedRewardRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"unstakePeriodEpochs\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"unstakePeriodTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"updateBaseRewardPerSecond\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blocksNum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"}],\"name\":\"updateOfflinePenaltyThreshold\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"updateSlashingRefundRatio\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"validatorCommission\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"bytes3\",\"name\":\"\",\"type\":\"bytes3\"}],\"payable\":false,\"stateMutability\":\"pure\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// ContractBin is the compiled bytecode used for deploying new contracts.
var ContractBin = "0x608060405234801561001057600080fd5b50614a06806100206000396000f3fe6080604052600436106103f95760003560e01c80636f4986631161020d578063b5d8962711610128578063cfd47663116100bb578063deb6fb0d1161008a578063e59488661161006f578063e5948866146110a2578063ebdf104c146110de578063f2fde38b14611251576103f9565b8063deb6fb0d14610ff5578063e08d7e6614611025576103f9565b8063cfd4766314610f38578063cfdbb7cd14610f71578063d9a7c1f914610faa578063de67f21514610fbf576103f9565b8063c641ea28116100f7578063c641ea2814610e4d578063c65ee0e114610ec7578063c7be95de14610ef1578063cc8343aa14610f06576103f9565b8063b5d8962714610d79578063b6d9edd514610de4578063c3de580e14610e0e578063c5f530af14610e38576103f9565b80638cddb015116101a05780639fa6dd351161016f5780639fa6dd3514610ca7578063a1f0174f14610cc4578063a5a470ad14610cf4578063a778651514610d64576103f9565b80638cddb01514610bb05780638da5cb5b14610be95780638f32d59b14610c1a57806396c7ee4614610c43576103f9565b8063854873e1116101dc578063854873e114610acc5780638914d4c0146105f35780638b0e9f3f14610b6b5780638b1a0d1114610b80576103f9565b80636f49866314610a54578063715018a614610a8d5780637667180814610aa25780637cacb1d614610ab7576103f9565b80632cedb097116103185780634f864df4116102ab5780635ccfe1e81161027a5780635fab23a81161025f5780635fab23a8146109cd5780636099ecb2146109e2578063670322f814610a1b576103f9565b80635ccfe1e8146109945780635e2308d2146106aa576103f9565b80634f864df4146108395780634feb92f31461086f57806352b60bf31461091a57806354fd4d501461094a576103f9565b806341a19472116102e757806341a1947214610779578063441a3e70146107a95780634aa5ff6d146107d95780634f7c4efb14610809576103f9565b80632cedb097146106d45780632d296a9b1461070257806339b80c00146107175780633fee10a814610702576103f9565b806318f628d4116103905780631f2701521161035f5780631f270152146106385780632265f284146106955780632709275e146106aa57806328f73148146106bf576103f9565b806318f628d41461055e5780631d3ac42c146105c35780631d58179c146105f35780631e702f8314610608576103f9565b80630d7b2609116103cc5780630d7b2609146104cb57806312622d0e146104e0578063173a2c3c1461051957806318160ddd14610549576103f9565b80630135b1db146103fe578063019e2729146104435780630962ef791461048c5780630d4955e3146104b6575b600080fd5b34801561040a57600080fd5b506104316004803603602081101561042157600080fd5b50356001600160a01b0316611284565b60408051918252519081900360200190f35b34801561044f57600080fd5b5061048a6004803603608081101561046657600080fd5b508035906020810135906001600160a01b0360408201358116916060013516611296565b005b34801561049857600080fd5b5061048a600480360360208110156104af57600080fd5b5035611390565b3480156104c257600080fd5b5061043161147b565b3480156104d757600080fd5b50610431611484565b3480156104ec57600080fd5b506104316004803603604081101561050357600080fd5b506001600160a01b03813516906020013561148b565b34801561052557600080fd5b506104316004803603604081101561053c57600080fd5b5080359060200135611514565b34801561055557600080fd5b50610431611535565b34801561056a57600080fd5b5061048a600480360361012081101561058257600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e081013590610100013561153b565b3480156105cf57600080fd5b50610431600480360360408110156105e657600080fd5b508035906020013561168c565b3480156105ff57600080fd5b5061043161183d565b34801561061457600080fd5b5061048a6004803603604081101561062b57600080fd5b5080359060200135611842565b34801561064457600080fd5b506106776004803603606081101561065b57600080fd5b506001600160a01b0381351690602081013590604001356118f1565b60408051938452602084019290925282820152519081900360600190f35b3480156106a157600080fd5b50610431611923565b3480156106b657600080fd5b50610431611935565b3480156106cb57600080fd5b50610431611951565b3480156106e057600080fd5b506106e9611957565b6040805192835260208301919091528051918290030190f35b34801561070e57600080fd5b50610431611961565b34801561072357600080fd5b506107416004803603602081101561073a57600080fd5b5035611968565b604080519788526020880196909652868601949094526060860192909252608085015260a084015260c0830152519081900360e00190f35b34801561078557600080fd5b506104316004803603604081101561079c57600080fd5b50803590602001356119aa565b3480156107b557600080fd5b5061048a600480360360408110156107cc57600080fd5b50803590602001356119cb565b3480156107e557600080fd5b50610431600480360360408110156107fc57600080fd5b5080359060200135611ce2565b34801561081557600080fd5b5061048a6004803603604081101561082c57600080fd5b5080359060200135611d03565b34801561084557600080fd5b5061048a6004803603606081101561085c57600080fd5b5080359060208101359060400135611e47565b34801561087b57600080fd5b5061048a600480360361010081101561089357600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156108c357600080fd5b8201836020820111156108d557600080fd5b803590602001918460018302840111640100000000831117156108f757600080fd5b919350915080359060208101359060408101359060608101359060800135612123565b34801561092657600080fd5b506104316004803603604081101561093d57600080fd5b50803590602001356121c9565b34801561095657600080fd5b5061095f6121ea565b604080517fffffff00000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b3480156109a057600080fd5b50610431600480360360408110156109b757600080fd5b506001600160a01b03813516906020013561220e565b3480156109d957600080fd5b5061043161222b565b3480156109ee57600080fd5b5061043160048036036040811015610a0557600080fd5b506001600160a01b038135169060200135612231565b348015610a2757600080fd5b5061043160048036036040811015610a3e57600080fd5b506001600160a01b03813516906020013561227e565b348015610a6057600080fd5b5061043160048036036040811015610a7757600080fd5b506001600160a01b0381351690602001356122bf565b348015610a9957600080fd5b5061048a6122dc565b348015610aae57600080fd5b5061043161238c565b348015610ac357600080fd5b50610431612395565b348015610ad857600080fd5b50610af660048036036020811015610aef57600080fd5b503561239b565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610b30578181015183820152602001610b18565b50505050905090810190601f168015610b5d5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b348015610b7757600080fd5b50610431612436565b348015610b8c57600080fd5b5061048a60048036036040811015610ba357600080fd5b508035906020013561243c565b348015610bbc57600080fd5b5061048a60048036036040811015610bd357600080fd5b506001600160a01b0381351690602001356124de565b348015610bf557600080fd5b50610bfe612539565b604080516001600160a01b039092168252519081900360200190f35b348015610c2657600080fd5b50610c2f612548565b604080519115158252519081900360200190f35b348015610c4f57600080fd5b50610c7c60048036036040811015610c6657600080fd5b506001600160a01b038135169060200135612559565b6040805195865260208601949094528484019290925260608401526080830152519081900360a00190f35b61048a60048036036020811015610cbd57600080fd5b5035612593565b348015610cd057600080fd5b5061043160048036036040811015610ce757600080fd5b50803590602001356125a1565b61048a60048036036020811015610d0a57600080fd5b810190602081018135640100000000811115610d2557600080fd5b820183602082011115610d3757600080fd5b80359060200191846001830284011164010000000083111715610d5957600080fd5b5090925090506125be565b348015610d7057600080fd5b50610431612667565b348015610d8557600080fd5b50610da360048036036020811015610d9c57600080fd5b503561267d565b604080519788526020880196909652868601949094526060860192909252608085015260a08401526001600160a01b031660c0830152519081900360e00190f35b348015610df057600080fd5b5061048a60048036036020811015610e0757600080fd5b50356126c3565b348015610e1a57600080fd5b50610c2f60048036036020811015610e3157600080fd5b50356127b5565b348015610e4457600080fd5b506104316127cf565b348015610e5957600080fd5b50610e7760048036036020811015610e7057600080fd5b50356127de565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610eb3578181015183820152602001610e9b565b505050509050019250505060405180910390f35b348015610ed357600080fd5b5061043160048036036020811015610eea57600080fd5b5035612843565b348015610efd57600080fd5b50610431612855565b348015610f1257600080fd5b5061048a60048036036040811015610f2957600080fd5b5080359060200135151561285b565b348015610f4457600080fd5b5061043160048036036040811015610f5b57600080fd5b506001600160a01b038135169060200135612a76565b348015610f7d57600080fd5b50610c2f60048036036040811015610f9457600080fd5b506001600160a01b038135169060200135612a93565b348015610fb657600080fd5b50610431612afb565b348015610fcb57600080fd5b5061048a60048036036060811015610fe257600080fd5b5080359060208101359060400135612b01565b34801561100157600080fd5b506104316004803603604081101561101857600080fd5b5080359060200135612e18565b34801561103157600080fd5b5061048a6004803603602081101561104857600080fd5b81019060208101813564010000000081111561106357600080fd5b82018360208201111561107557600080fd5b8035906020019184602083028401116401000000008311171561109757600080fd5b509092509050612e39565b3480156110ae57600080fd5b50610431600480360360808110156110c557600080fd5b5080359060208101359060408101359060600135612f36565b3480156110ea57600080fd5b5061048a6004803603608081101561110157600080fd5b81019060208101813564010000000081111561111c57600080fd5b82018360208201111561112e57600080fd5b8035906020019184602083028401116401000000008311171561115057600080fd5b91939092909160208101903564010000000081111561116e57600080fd5b82018360208201111561118057600080fd5b803590602001918460208302840111640100000000831117156111a257600080fd5b9193909290916020810190356401000000008111156111c057600080fd5b8201836020820111156111d257600080fd5b803590602001918460208302840111640100000000831117156111f457600080fd5b91939092909160208101903564010000000081111561121257600080fd5b82018360208201111561122457600080fd5b8035906020019184602083028401116401000000008311171561124657600080fd5b509092509050612fbc565b34801561125d57600080fd5b5061048a6004803603602081101561127457600080fd5b50356001600160a01b0316613198565b60696020526000908152604090205481565b600054610100900460ff16806112af57506112af6131fa565b806112bd575060005460ff16155b6112f85760405162461bcd60e51b815260040180806020018281038252602e815260200180614906602e913960400191505060405180910390fd5b600054610100900460ff16158015611323576000805460ff1961ff0019909116610100171660011790555b61132c82613200565b60678590556066805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b03851617905560748490556755cfe697852e904c6073556103e86076556203f4806077558015611389576000805461ff00191690555b5050505050565b3361139b81836132fe565b506001600160a01b0381166000908152606f6020908152604080832085845290915290205480611412576040805162461bcd60e51b815260206004820152600c60248201527f7a65726f20726577617264730000000000000000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0382166000908152606f6020908152604080832086845290915281205561143f816133c2565b6040516001600160a01b0383169082156108fc029083906000818181858888f19350505050158015611475573d6000803e3d6000fd5b50505050565b6301e133805b90565b6212750090565b60006114978383612a93565b6114c557506001600160a01b038216600090815260716020908152604080832084845290915290205461150e565b6001600160a01b03831660008181526072602090815260408083208684528252808320549383526071825280832086845290915290205461150b9163ffffffff61344216565b90505b92915050565b60009182526075602090815260408084209284526002909201905290205490565b60745481565b61154433613484565b61157f5760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b61158a898989613498565b6001600160a01b0389166000908152606f602090815260408083208b845290915290208190556115b9876133c2565b851561168157868611156115fe5760405162461bcd60e51b815260040180806020018281038252602c8152602001806149a6602c913960400191505060405180910390fd5b6001600160a01b03891660008181526072602090815260408083208c84528252918290208981556001810189905560028101889055600381018790556004810186905582518781529182018a9052825190938c9390927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78929081900390910190a3505b505050505050505050565b3360008181526072602090815260408083208684529091528120909190836116e9576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b6116f38286612a93565b611744576040805162461bcd60e51b815260206004820152600d60248201527f6e6f74206c6f636b656420757000000000000000000000000000000000000000604482015290519081900360640190fd5b805484111561179a576040805162461bcd60e51b815260206004820152601760248201527f6e6f7420656e6f756768206c6f636b6564207374616b65000000000000000000604482015290519081900360640190fd5b6117a482866132fe565b5060006117ba82600401548684600001546135b7565b6004830180548290039055825486900383556001600160a01b03841660008181526071602090815260408083208b8452825291829020805485900390558151898152908101849052815193945089937fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9929181900390910190a395945050505050565b600390565b61184b33613484565b6118865760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b806118d8576040805162461bcd60e51b815260206004820152600c60248201527f77726f6e67207374617475730000000000000000000000000000000000000000604482015290519081900360640190fd5b6118e282826135e4565b6118ed82600061285b565b5050565b607060209081526000938452604080852082529284528284209052825290208054600182015460029092015490919083565b600061192d613693565b601002905090565b60006064611941613693565b601e028161194b57fe5b04905090565b606d5481565b6076546077549091565b62093a8090565b607560205280600052604060002060009150905080600701549080600801549080600901549080600a01549080600b01549080600c01549080600d0154905087565b60009182526075602090815260408084209284526003909201905290205490565b336119d461476b565b506001600160a01b038116600090815260706020908152604080832086845282528083208584528252918290208251606081018452815480825260018301549382019390935260029091015492810192909252611a78576040805162461bcd60e51b815260206004820152601560248201527f7265717565737420646f65736e27742065786973740000000000000000000000604482015290519081900360640190fd5b60208082015182516000878152606890935260409092206001015490919015801590611ab4575060008681526068602052604090206001015482115b15611ad5575050600084815260686020526040902060018101546002909101545b611add611961565b8201611ae761369f565b1015611b3a576040805162461bcd60e51b815260206004820152601660248201527f6e6f7420656e6f7567682074696d652070617373656400000000000000000000604482015290519081900360640190fd5b611b4261183d565b8101611b4c61238c565b1015611b9f576040805162461bcd60e51b815260206004820152601860248201527f6e6f7420656e6f7567682065706f636873207061737365640000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152607060209081526040808320898452825280832088845290915281206002015490611bd8886127b5565b90506000611bfa8383607860008d8152602001908152602001600020546136a3565b6001600160a01b03881660009081526070602090815260408083208d845282528083208c845290915281208181556001810182905560020155606e8054820190559050808311611c91576040805162461bcd60e51b815260206004820152601660248201527f7374616b652069732066756c6c7920736c617368656400000000000000000000604482015290519081900360640190fd5b6001600160a01b0387166108fc611cae858463ffffffff61344216565b6040518115909202916000818181858888f19350505050158015611cd6573d6000803e3d6000fd5b50505050505050505050565b60009182526075602090815260408084209284526001909201905290205490565b611d0b612548565b611d5c576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b611d65826127b5565b611db6576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f722069736e277420736c6173686564000000000000000000604482015290519081900360640190fd5b611dbe613693565b811115611dfc5760405162461bcd60e51b81526004018080602001828103825260218152602001806149346021913960400191505060405180910390fd5b6000828152607860209081526040918290208390558151838152915184927f047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb91792908290030190a25050565b33611e5281856132fe565b5060008211611e96576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b611ea0818561148b565b821115611ef4576040805162461bcd60e51b815260206004820152601960248201527f6e6f7420656e6f75676820756e6c6f636b6564207374616b6500000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607060209081526040808320878452825280832086845290915290206002015415611f75576040805162461bcd60e51b815260206004820152601360248201527f7772494420616c72656164792065786973747300000000000000000000000000604482015290519081900360640190fd5b6001600160a01b03811660009081526071602090815260408083208784528252808320805486900390556068909152902060030154611fba908363ffffffff61344216565b600085815260686020526040902060030155606c54611fdf908363ffffffff61344216565b606c5560008481526068602052604090205461200c57606d54612008908363ffffffff61344216565b606d555b61201584613702565b8061202657506120248461373e565b155b6120615760405162461bcd60e51b815260040180806020018281038252602981526020018061497d6029913960400191505060405180910390fd5b61206a8461373e565b612079576120798460016135e4565b6001600160a01b0381166000908152607060209081526040808320878452825280832086845290915290206002018290556120b261238c565b6001600160a01b038216600090815260706020908152604080832088845282528083208784529091529020556120e661369f565b6001600160a01b0382166000908152607060209081526040808320888452825280832087845290915281206001019190915561147590859061285b565b61212c33613484565b6121675760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b6121af898989898080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b92508a91508990508888613770565b606b5488111561168157606b889055505050505050505050565b60009182526075602090815260408084209284526005909201905290205490565b7f323032000000000000000000000000000000000000000000000000000000000090565b607960209081526000928352604080842090915290825290205481565b606e5481565b60008061223e8484613859565b506001600160a01b0385166000908152606f60209081526040808320878452909152902054909150612276908263ffffffff6139a716565b949350505050565b600061228a8383612a93565b6122965750600061150e565b506001600160a01b03919091166000908152607260209081526040808320938352929052205490565b606f60209081526000928352604080842090915290825290205481565b6122e4612548565b612335576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a36033805473ffffffffffffffffffffffffffffffffffffffff19169055565b60675460010190565b60675481565b606a6020908152600091825260409182902080548351601f60026000196101006001861615020190931692909204918201849004840281018401909452808452909183018282801561242e5780601f106124035761010080835404028352916020019161242e565b820191906000526020600020905b81548152906001019060200180831161241157829003601f168201915b505050505081565b606c5481565b612444612548565b612495576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b60778190556076829055604080518381526020810183905281517f702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34929181900390910190a15050565b6124e882826132fe565b6118ed576040805162461bcd60e51b815260206004820152601060248201527f6e6f7468696e6720746f20737461736800000000000000000000000000000000604482015290519081900360640190fd5b6033546001600160a01b031690565b6033546001600160a01b0316331490565b6072602090815260009283526040808420909152908252902080546001820154600283015460038401546004909401549293919290919085565b61259e338234613a01565b50565b600091825260756020908152604080842092845291905290205490565b6125c66127cf565b34101561261a576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b61265a3383838080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250613b0b92505050565b6118ed33606b5434613a01565b60006064612673613693565b600f028161194b57fe5b606860205260009081526040902080546001820154600283015460038401546004850154600586015460069096015494959394929391929091906001600160a01b031687565b6126cb612548565b61271c576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6801c985c8903591eb2081111561277a576040805162461bcd60e51b815260206004820152601b60248201527f746f6f206c617267652072657761726420706572207365636f6e640000000000604482015290519081900360640190fd5b60738190556040805182815290517f8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae23969181900360200190a150565b60008181526068602052604090205460801615155b919050565b6a02a055184a310c1260000090565b60008181526075602090815260409182902060060180548351818402810184019094528084526060939283018282801561283757602002820191906000526020600020905b815481526020019060010190808311612823575b50505050509050919050565b60786020526000908152604090205481565b606b5481565b61286482613b36565b6128b5576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b600082815260686020526040902060038101549054156128d3575060005b606654604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018690526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b15801561294057600080fd5b505af1158015612954573d6000803e3d6000fd5b5050505081801561296457508015155b15612a71576066546000848152606a60205260409081902081517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810187815260248201938452825460026000196001831615610100020190911604604483018190526001600160a01b039095169463242a6e3f94899493909160649091019084908015612a375780601f10612a0c57610100808354040283529160200191612a37565b820191906000526020600020905b815481529060010190602001808311612a1a57829003601f168201915b50509350505050600060405180830381600087803b158015612a5857600080fd5b505af1158015612a6c573d6000803e3d6000fd5b505050505b505050565b607160209081526000928352604080842090915290825290205481565b6001600160a01b03821660009081526072602090815260408083208484529091528120600201541580159061150b57506001600160a01b0383166000908152607260209081526040808320858452909152902060020154612af261369f565b11159392505050565b60735481565b3381612b42576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b612b4c8185612a93565b15612b9e576040805162461bcd60e51b815260206004820152601160248201527f616c7265616479206c6f636b6564207570000000000000000000000000000000604482015290519081900360640190fd5b612ba8818561148b565b821115612bfc576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b60008481526068602052604090205415612c5d576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b612c65611484565b8310158015612c7b5750612c7761147b565b8311155b612ccc576040805162461bcd60e51b815260206004820152601260248201527f696e636f7272656374206475726174696f6e0000000000000000000000000000604482015290519081900360640190fd5b6000612ce684612cda61369f565b9063ffffffff6139a716565b6000868152606860205260409020600601549091506001600160a01b039081169083168114612d74576001600160a01b0381166000908152607260209081526040808320898452909152902060020154821115612d745760405162461bcd60e51b81526004018080602001828103825260288152602001806149556028913960400191505060405180910390fd5b612d7e83876132fe565b506001600160a01b03831660009081526072602090815260408083208984529091529020848155612dad61238c565b60018201556002810183905560038101869055600060048201556040805187815260208101879052815189926001600160a01b038816927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78929081900390910190a350505050505050565b60009182526075602090815260408084209284526004909201905290205490565b612e4233613484565b612e7d5760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b600060756000612e8b61238c565b8152602001908152602001600020905060008090505b82811015612f2757600060686000868685818110612ebb57fe5b90506020020135815260200190815260200160002060030154905080836000016000878786818110612ee957fe5b90506020020135815260200190815260200160002081905550612f198184600c01546139a790919063ffffffff16565b600c84015550600101612ea1565b5061147560068201848461478c565b6000818310612f4757506000612276565b600083815260756020818152604080842088855260019081018352818520548786529383528185208986520190915290912054612fb1612f85613693565b612fa589612f99858763ffffffff61344216565b9063ffffffff613b4d16565b9063ffffffff613ba616565b979650505050505050565b612fc533613484565b6130005760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b60006075600061300e61238c565b8152602001908152602001600020905060608160060180548060200260200160405190810160405280929190818152602001828054801561306e57602002820191906000526020600020905b81548152602001906001019080831161305a575b505050505090506130f582828c8c80806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050508b8b80806020026020016040519081016040528093929190818152602001838360200280828437600092019190915250613be892505050565b613164828288888080602002602001604051908101604052809392919081815260200183836020028082843760009201919091525050604080516020808c0282810182019093528b82529093508b92508a918291850190849080828437600092019190915250613cf792505050565b61316c61238c565b60675561317761369f565b600783015550607354600b820155607454600d909101555050505050505050565b6131a0612548565b6131f1576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b61259e8161429d565b303b1590565b600054610100900460ff168061321957506132196131fa565b80613227575060005460ff16155b6132625760405162461bcd60e51b815260040180806020018281038252602e815260200180614906602e913960400191505060405180910390fd5b600054610100900460ff1615801561328d576000805460ff1961ff0019909116610100171660011790555b6033805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0384811691909117918290556040519116906000907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a380156118ed576000805461ff00191690555050565b600080600061330d8585613859565b9150915061331a8461434b565b6001600160a01b0386166000818152607960209081526040808320898452825280832094909455918152606f82528281208782529091522054613363908363ffffffff6139a716565b6001600160a01b0386166000818152606f60209081526040808320898452825280832094909455918152607282528281208782529091522060048101546133b0908363ffffffff6139a716565b60049091015550600191505092915050565b606654604080517f66e7ea0f0000000000000000000000000000000000000000000000000000000081523060048201526024810184905290516001600160a01b03909216916366e7ea0f9160448082019260009290919082900301818387803b15801561342e57600080fd5b505af1158015611389573d6000803e3d6000fd5b600061150b83836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f7700008152506143a6565b6066546001600160a01b0390811691161490565b600081116134db576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b6134e583836132fe565b506001600160a01b038316600090815260716020908152604080832085845290915290205461351a908263ffffffff6139a716565b6001600160a01b038416600090815260716020908152604080832086845282528083209390935560689052206003015461355a818363ffffffff6139a716565b600084815260686020526040902060030155606c5461357f908363ffffffff6139a716565b606c556000838152606860205260409020546135ac57606d546135a8908363ffffffff6139a716565b606d555b61147583821561285b565b6000806135ce83612fa5878763ffffffff613b4d16565b90508381106135da5750825b90505b9392505050565b6000828152606860205260409020541580156135ff57508015155b1561362c57600082815260686020526040902060030154606d546136289163ffffffff61344216565b606d555b6000828152606860205260409020548111156118ed576000828152606860205260409020818155600201546118ed5761366361369f565b60008381526068602052604090206001015561367d61238c565b6000838152606860205260409020600201555050565b670de0b6b3a764000090565b4290565b60008215806136b957506136b5613693565b8210155b156136c6575060006135dd565b6136f16001612cda6136d6613693565b612fa5866136e2613693565b8a91900363ffffffff613b4d16565b9050838111156135dd5750826135dd565b600061372361370f613693565b612fa561371a611923565b612f998661373e565b60008381526068602052604090206003015411159050919050565b6000818152606860209081526040808320600601546001600160a01b0316835260718252808320938352929052205490565b6001600160a01b038816600090815260696020526040902054156137db576040805162461bcd60e51b815260206004820152601860248201527f76616c696461746f7220616c7265616479206578697374730000000000000000604482015290519081900360640190fd5b6001600160a01b03881660008181526069602090815260408083208b90558a83526068825280832089815560048101899055600581018890556001810186905560028101879055600601805473ffffffffffffffffffffffffffffffffffffffff1916909417909355606a81529190208751611681928901906147d7565b6001600160a01b03821660009081526079602090815260408083208484529091528120548190816138898561434b565b90506000613897878761443d565b9050818111156138a45750805b828110156138af5750815b6001600160a01b03871660008181526072602090815260408083208a84528252808320938352607182528083208a845290915281205482549091906138fb90839063ffffffff61344216565b9050600061390f84600001548b8988612f36565b90506000806139228387600301546144fc565b91509150613932848d8b8a612f36565b92506000806139428560006144fc565b91509150613952878f8b8d612f36565b94506000806139628760006144fc565b909250905061397b82612cda888763ffffffff6139a716565b61398f82612cda888763ffffffff6139a716565b9e509e50505050505050505050505050509250929050565b60008282018381101561150b576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b613a0a82613b36565b613a5b576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090205415613abc576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b613ac7838383613498565b613ad082613702565b612a715760405162461bcd60e51b815260040180806020018281038252602981526020018061497d6029913960400191505060405180910390fd5b606b805460010190819055612a718382846000613b2661238c565b613b2e61369f565b600080613770565b600090815260686020526040902060050154151590565b600082613b5c5750600061150e565b82820282848281613b6957fe5b041461150b5760405162461bcd60e51b81526004018080602001828103825260218152602001806148e56021913960400191505060405180910390fd5b600061150b83836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f0000000000008152506145cb565b60005b835181101561138957607654828281518110613c0357fe5b6020026020010151118015613c2d5750607754838281518110613c2257fe5b602002602001015110155b15613c6e57613c50848281518110613c4157fe5b602002602001015160086135e4565b613c6e848281518110613c5f57fe5b6020026020010151600061285b565b828181518110613c7a57fe5b6020026020010151856004016000868481518110613c9457fe5b6020026020010151815260200190815260200160002081905550818181518110613cba57fe5b6020026020010151856005016000868481518110613cd457fe5b602090810291909101810151825281019190915260400160002055600101613beb565b613cff614845565b6040518060c001604052808551604051908082528060200260200182016040528015613d35578160200160208202803883390190505b508152602001600081526020018551604051908082528060200260200182016040528015613d6d578160200160208202803883390190505b508152602001600081526020016000815260200160008152509050600060756000613da76001613d9b61238c565b9063ffffffff61344216565b81526020810191909152604001600020600160808401526007810154909150613dce61369f565b1115613de8578060070154613de161369f565b0360808301525b60005b8551811015613ebe57613e358360800151612fa5878481518110613e0b57fe5b6020026020010151878581518110613e1f57fe5b6020026020010151613b4d90919063ffffffff16565b83604001518281518110613e4557fe5b602002602001018181525050613e7f83604001518281518110613e6457fe5b602002602001015184606001516139a790919063ffffffff16565b60608401528351613eb190859083908110613e9657fe5b60200260200101518460a001516139a790919063ffffffff16565b60a0840152600101613deb565b5060005b8551811015613f9557613f408360800151612fa5878481518110613ee257fe5b6020026020010151612f998760800151612fa58b8881518110613f0157fe5b60200260200101518e60000160008f8b81518110613f1b57fe5b6020026020010151815260200190815260200160002054613b4d90919063ffffffff16565b8351805183908110613f4e57fe5b602002602001018181525050613f8883600001518281518110613f6d57fe5b602002602001015184602001516139a790919063ffffffff16565b6020840152600101613ec2565b5060005b8551811015614275576000613fd1846080015160735486600001518581518110613fbf57fe5b60200260200101518760200151614630565b905061400d6140008560a0015186604001518581518110613fee57fe5b60200260200101518760600151614671565b829063ffffffff6139a716565b9050600087838151811061401d57fe5b6020908102919091018101516000818152606883526040808220600601546001600160a01b031680835260728552818320848452909452812091935061406a85614065612667565b6146ce565b6001600160a01b03841660009081526071602090815260408083208884529091528120549192509061409c858761227e565b8302816140a557fe5b049050600081830390506000806140c08487600301546144fc565b915091506000806140d28560006144fc565b6001600160a01b038b166000908152606f602090815260408083208f84529091529020549193509150614111908390612cda908763ffffffff6139a716565b6001600160a01b038a166000908152606f602090815260408083208e84529091529020556004880154614150908290612cda908663ffffffff6139a716565b60048901555050606c54858a0394506000935061418292509050612fa5614175613693565b859063ffffffff613b4d16565b600087815260018b0160205260409020549091506141a6908263ffffffff6139a716565b8e6001016000888152602001908152602001600020819055506141fa8b89815181106141ce57fe5b60200260200101518a6003016000898152602001908152602001600020546139a790919063ffffffff16565b8e60030160008881526020019081526020016000208190555061424e8c898151811061422257fe5b60200260200101518a6002016000898152602001908152602001600020546139a790919063ffffffff16565b600096875260028f01602052604090962095909555505060019094019350613f9992505050565b505060a081015160088601556020810151600986015560600151600a90940193909355505050565b6001600160a01b0381166142e25760405162461bcd60e51b81526004018080602001828103825260268152602001806148966026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a36033805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0392909216919091179055565b6000818152606860205260408120600201541561439e57600082815260686020526040902060020154606754101561438657506067546127ca565b506000818152606860205260409020600201546127ca565b505060675490565b600081848411156144355760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b838110156143fa5781810151838201526020016143e2565b50505050905090810190601f1680156144275780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b6001600160a01b03821660009081526072602090815260408083208484529091528120600101546067546144728585836146eb565b1561448057915061150e9050565b61448b8585846146eb565b61449a5760009250505061150e565b808211156144ad5760009250505061150e565b808210156144e0576002818301046144c68686836146eb565b156144d6578060010192506144da565b8091505b506144ad565b806144f05760009250505061150e565b60001901949350505050565b600080821561459a57600061450f611935565b614517613693565b039050600061453761452761147b565b612fa5848863ffffffff613b4d16565b905061455e614544613693565b612fa583614550611935565b8a910163ffffffff613b4d16565b935061459161456b613693565b612fa5836002614579611935565b8161458057fe5b040189613b4d90919063ffffffff16565b925050506145c4565b6145bd6145a5613693565b612fa56145b0611935565b879063ffffffff613b4d16565b9150600090505b9250929050565b6000818361461a5760405162461bcd60e51b81526020600482018181528351602484015283519092839260449091019190850190808383600083156143fa5781810151838201526020016143e2565b50600083858161462657fe5b0495945050505050565b60008261463f57506000612276565b6000614651868663ffffffff613b4d16565b905061466783612fa5838763ffffffff613b4d16565b9695505050505050565b600082614680575060006135dd565b600061469683612fa5878763ffffffff613b4d16565b90506146c56146a3613693565b612fa56146ae611935565b6146b6613693565b8591900363ffffffff613b4d16565b95945050505050565b600061150b6146db613693565b612fa5858563ffffffff613b4d16565b6001600160a01b038316600090815260726020908152604080832085845290915281206001015482108015906135da57506001600160a01b038416600090815260726020908152604080832086845290915290206002015461474c83614756565b1115949350505050565b60009081526075602052604090206007015490565b60405180606001604052806000815260200160008152602001600081525090565b8280548282559060005260206000209081019282156147c7579160200282015b828111156147c75782358255916020019190600101906147ac565b506147d392915061487b565b5090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061481857805160ff19168380011785556147c7565b828001600101855582156147c7579182015b828111156147c757825182559160200191906001019061482a565b6040518060c001604052806060815260200160008152602001606081526020016000815260200160008152602001600081525090565b61148191905b808211156147d3576000815560010161488156fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f206164647265737363616c6c6572206973206e6f7420746865204e6f64654472697665724175746820636f6e7472616374536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f77436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a65646d757374206265206c657373207468616e206f7220657175616c20746f20312e3076616c696461746f72206c6f636b757020706572696f642077696c6c20656e64206561726c69657276616c696461746f7227732064656c65676174696f6e73206c696d69742069732065786365656465646c6f636b6564207374616b652069732067726561746572207468616e207468652077686f6c65207374616b65a265627a7a7231582015ceb1c92ead96d3fabbf6f3824838f149d5774ba2d8e2395e382bade4268c1b64736f6c634300050c0032"

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

// NonStashedRewardsOf is a free data retrieval call binding the contract method 0xe5948866.
//
// Solidity: function _nonStashedRewardsOf(uint256 stakeAmount, uint256 toValidatorID, uint256 fromEpoch, uint256 toEpoch) view returns(uint256)
func (_Contract *ContractCaller) NonStashedRewardsOf(opts *bind.CallOpts, stakeAmount *big.Int, toValidatorID *big.Int, fromEpoch *big.Int, toEpoch *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "_nonStashedRewardsOf", stakeAmount, toValidatorID, fromEpoch, toEpoch)
	return *ret0, err
}

// NonStashedRewardsOf is a free data retrieval call binding the contract method 0xe5948866.
//
// Solidity: function _nonStashedRewardsOf(uint256 stakeAmount, uint256 toValidatorID, uint256 fromEpoch, uint256 toEpoch) view returns(uint256)
func (_Contract *ContractSession) NonStashedRewardsOf(stakeAmount *big.Int, toValidatorID *big.Int, fromEpoch *big.Int, toEpoch *big.Int) (*big.Int, error) {
	return _Contract.Contract.NonStashedRewardsOf(&_Contract.CallOpts, stakeAmount, toValidatorID, fromEpoch, toEpoch)
}

// NonStashedRewardsOf is a free data retrieval call binding the contract method 0xe5948866.
//
// Solidity: function _nonStashedRewardsOf(uint256 stakeAmount, uint256 toValidatorID, uint256 fromEpoch, uint256 toEpoch) view returns(uint256)
func (_Contract *ContractCallerSession) NonStashedRewardsOf(stakeAmount *big.Int, toValidatorID *big.Int, fromEpoch *big.Int, toEpoch *big.Int) (*big.Int, error) {
	return _Contract.Contract.NonStashedRewardsOf(&_Contract.CallOpts, stakeAmount, toValidatorID, fromEpoch, toEpoch)
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

// ClaimedRewardUntilEpoch is a free data retrieval call binding the contract method 0x5ccfe1e8.
//
// Solidity: function claimedRewardUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) ClaimedRewardUntilEpoch(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "claimedRewardUntilEpoch", arg0, arg1)
	return *ret0, err
}

// ClaimedRewardUntilEpoch is a free data retrieval call binding the contract method 0x5ccfe1e8.
//
// Solidity: function claimedRewardUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractSession) ClaimedRewardUntilEpoch(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.ClaimedRewardUntilEpoch(&_Contract.CallOpts, arg0, arg1)
}

// ClaimedRewardUntilEpoch is a free data retrieval call binding the contract method 0x5ccfe1e8.
//
// Solidity: function claimedRewardUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) ClaimedRewardUntilEpoch(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.ClaimedRewardUntilEpoch(&_Contract.CallOpts, arg0, arg1)
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

// GetEpochSnapshotAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0x41a19472.
//
// Solidity: function getEpochSnapshotAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochSnapshotAccumulatedOriginatedTxsFee(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotAccumulatedOriginatedTxsFee", epoch, validatorID)
	return *ret0, err
}

// GetEpochSnapshotAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0x41a19472.
//
// Solidity: function getEpochSnapshotAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochSnapshotAccumulatedOriginatedTxsFee(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotAccumulatedOriginatedTxsFee(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0x41a19472.
//
// Solidity: function getEpochSnapshotAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochSnapshotAccumulatedOriginatedTxsFee(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotAccumulatedOriginatedTxsFee(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotAccumulatedRewardPerToken is a free data retrieval call binding the contract method 0x4aa5ff6d.
//
// Solidity: function getEpochSnapshotAccumulatedRewardPerToken(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochSnapshotAccumulatedRewardPerToken(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotAccumulatedRewardPerToken", epoch, validatorID)
	return *ret0, err
}

// GetEpochSnapshotAccumulatedRewardPerToken is a free data retrieval call binding the contract method 0x4aa5ff6d.
//
// Solidity: function getEpochSnapshotAccumulatedRewardPerToken(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochSnapshotAccumulatedRewardPerToken(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotAccumulatedRewardPerToken(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotAccumulatedRewardPerToken is a free data retrieval call binding the contract method 0x4aa5ff6d.
//
// Solidity: function getEpochSnapshotAccumulatedRewardPerToken(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochSnapshotAccumulatedRewardPerToken(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotAccumulatedRewardPerToken(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotAccumulatedUptime is a free data retrieval call binding the contract method 0x173a2c3c.
//
// Solidity: function getEpochSnapshotAccumulatedUptime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochSnapshotAccumulatedUptime(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotAccumulatedUptime", epoch, validatorID)
	return *ret0, err
}

// GetEpochSnapshotAccumulatedUptime is a free data retrieval call binding the contract method 0x173a2c3c.
//
// Solidity: function getEpochSnapshotAccumulatedUptime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochSnapshotAccumulatedUptime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotAccumulatedUptime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotAccumulatedUptime is a free data retrieval call binding the contract method 0x173a2c3c.
//
// Solidity: function getEpochSnapshotAccumulatedUptime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochSnapshotAccumulatedUptime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotAccumulatedUptime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotOfflineBlocks is a free data retrieval call binding the contract method 0x52b60bf3.
//
// Solidity: function getEpochSnapshotOfflineBlocks(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochSnapshotOfflineBlocks(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotOfflineBlocks", epoch, validatorID)
	return *ret0, err
}

// GetEpochSnapshotOfflineBlocks is a free data retrieval call binding the contract method 0x52b60bf3.
//
// Solidity: function getEpochSnapshotOfflineBlocks(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochSnapshotOfflineBlocks(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotOfflineBlocks(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotOfflineBlocks is a free data retrieval call binding the contract method 0x52b60bf3.
//
// Solidity: function getEpochSnapshotOfflineBlocks(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochSnapshotOfflineBlocks(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotOfflineBlocks(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotOfflineTime is a free data retrieval call binding the contract method 0xdeb6fb0d.
//
// Solidity: function getEpochSnapshotOfflineTime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochSnapshotOfflineTime(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotOfflineTime", epoch, validatorID)
	return *ret0, err
}

// GetEpochSnapshotOfflineTime is a free data retrieval call binding the contract method 0xdeb6fb0d.
//
// Solidity: function getEpochSnapshotOfflineTime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochSnapshotOfflineTime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotOfflineTime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotOfflineTime is a free data retrieval call binding the contract method 0xdeb6fb0d.
//
// Solidity: function getEpochSnapshotOfflineTime(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochSnapshotOfflineTime(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotOfflineTime(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotReceivedStake is a free data retrieval call binding the contract method 0xa1f0174f.
//
// Solidity: function getEpochSnapshotReceivedStake(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochSnapshotReceivedStake(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotReceivedStake", epoch, validatorID)
	return *ret0, err
}

// GetEpochSnapshotReceivedStake is a free data retrieval call binding the contract method 0xa1f0174f.
//
// Solidity: function getEpochSnapshotReceivedStake(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetEpochSnapshotReceivedStake(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotReceivedStake(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotReceivedStake is a free data retrieval call binding the contract method 0xa1f0174f.
//
// Solidity: function getEpochSnapshotReceivedStake(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetEpochSnapshotReceivedStake(epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotReceivedStake(&_Contract.CallOpts, epoch, validatorID)
}

// GetEpochSnapshotValidatorIDs is a free data retrieval call binding the contract method 0xc641ea28.
//
// Solidity: function getEpochSnapshotValidatorIDs(uint256 epoch) view returns(uint256[])
func (_Contract *ContractCaller) GetEpochSnapshotValidatorIDs(opts *bind.CallOpts, epoch *big.Int) ([]*big.Int, error) {
	var (
		ret0 = new([]*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "getEpochSnapshotValidatorIDs", epoch)
	return *ret0, err
}

// GetEpochSnapshotValidatorIDs is a free data retrieval call binding the contract method 0xc641ea28.
//
// Solidity: function getEpochSnapshotValidatorIDs(uint256 epoch) view returns(uint256[])
func (_Contract *ContractSession) GetEpochSnapshotValidatorIDs(epoch *big.Int) ([]*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotValidatorIDs(&_Contract.CallOpts, epoch)
}

// GetEpochSnapshotValidatorIDs is a free data retrieval call binding the contract method 0xc641ea28.
//
// Solidity: function getEpochSnapshotValidatorIDs(uint256 epoch) view returns(uint256[])
func (_Contract *ContractCallerSession) GetEpochSnapshotValidatorIDs(epoch *big.Int) ([]*big.Int, error) {
	return _Contract.Contract.GetEpochSnapshotValidatorIDs(&_Contract.CallOpts, epoch)
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
// Solidity: function getLockupInfo(address , uint256 ) view returns(uint256 lockedStake, uint256 fromEpoch, uint256 endTime, uint256 duration, uint256 earlyUnlockPenalty)
func (_Contract *ContractCaller) GetLockupInfo(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	LockedStake        *big.Int
	FromEpoch          *big.Int
	EndTime            *big.Int
	Duration           *big.Int
	EarlyUnlockPenalty *big.Int
}, error) {
	ret := new(struct {
		LockedStake        *big.Int
		FromEpoch          *big.Int
		EndTime            *big.Int
		Duration           *big.Int
		EarlyUnlockPenalty *big.Int
	})
	out := ret
	err := _Contract.contract.Call(opts, out, "getLockupInfo", arg0, arg1)
	return *ret, err
}

// GetLockupInfo is a free data retrieval call binding the contract method 0x96c7ee46.
//
// Solidity: function getLockupInfo(address , uint256 ) view returns(uint256 lockedStake, uint256 fromEpoch, uint256 endTime, uint256 duration, uint256 earlyUnlockPenalty)
func (_Contract *ContractSession) GetLockupInfo(arg0 common.Address, arg1 *big.Int) (struct {
	LockedStake        *big.Int
	FromEpoch          *big.Int
	EndTime            *big.Int
	Duration           *big.Int
	EarlyUnlockPenalty *big.Int
}, error) {
	return _Contract.Contract.GetLockupInfo(&_Contract.CallOpts, arg0, arg1)
}

// GetLockupInfo is a free data retrieval call binding the contract method 0x96c7ee46.
//
// Solidity: function getLockupInfo(address , uint256 ) view returns(uint256 lockedStake, uint256 fromEpoch, uint256 endTime, uint256 duration, uint256 earlyUnlockPenalty)
func (_Contract *ContractCallerSession) GetLockupInfo(arg0 common.Address, arg1 *big.Int) (struct {
	LockedStake        *big.Int
	FromEpoch          *big.Int
	EndTime            *big.Int
	Duration           *big.Int
	EarlyUnlockPenalty *big.Int
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
// Solidity: function rewardsStash(address , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) RewardsStash(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "rewardsStash", arg0, arg1)
	return *ret0, err
}

// RewardsStash is a free data retrieval call binding the contract method 0x6f498663.
//
// Solidity: function rewardsStash(address , uint256 ) view returns(uint256)
func (_Contract *ContractSession) RewardsStash(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.RewardsStash(&_Contract.CallOpts, arg0, arg1)
}

// RewardsStash is a free data retrieval call binding the contract method 0x6f498663.
//
// Solidity: function rewardsStash(address , uint256 ) view returns(uint256)
func (_Contract *ContractCallerSession) RewardsStash(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contract.Contract.RewardsStash(&_Contract.CallOpts, arg0, arg1)
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

// UnstakePeriodEpochs is a free data retrieval call binding the contract method 0x8914d4c0.
//
// Solidity: function unstakePeriodEpochs() pure returns(uint256)
func (_Contract *ContractCaller) UnstakePeriodEpochs(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "unstakePeriodEpochs")
	return *ret0, err
}

// UnstakePeriodEpochs is a free data retrieval call binding the contract method 0x8914d4c0.
//
// Solidity: function unstakePeriodEpochs() pure returns(uint256)
func (_Contract *ContractSession) UnstakePeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.UnstakePeriodEpochs(&_Contract.CallOpts)
}

// UnstakePeriodEpochs is a free data retrieval call binding the contract method 0x8914d4c0.
//
// Solidity: function unstakePeriodEpochs() pure returns(uint256)
func (_Contract *ContractCallerSession) UnstakePeriodEpochs() (*big.Int, error) {
	return _Contract.Contract.UnstakePeriodEpochs(&_Contract.CallOpts)
}

// UnstakePeriodTime is a free data retrieval call binding the contract method 0x2d296a9b.
//
// Solidity: function unstakePeriodTime() pure returns(uint256)
func (_Contract *ContractCaller) UnstakePeriodTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Contract.contract.Call(opts, out, "unstakePeriodTime")
	return *ret0, err
}

// UnstakePeriodTime is a free data retrieval call binding the contract method 0x2d296a9b.
//
// Solidity: function unstakePeriodTime() pure returns(uint256)
func (_Contract *ContractSession) UnstakePeriodTime() (*big.Int, error) {
	return _Contract.Contract.UnstakePeriodTime(&_Contract.CallOpts)
}

// UnstakePeriodTime is a free data retrieval call binding the contract method 0x2d296a9b.
//
// Solidity: function unstakePeriodTime() pure returns(uint256)
func (_Contract *ContractCallerSession) UnstakePeriodTime() (*big.Int, error) {
	return _Contract.Contract.UnstakePeriodTime(&_Contract.CallOpts)
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

var ContractBinRuntime = "0x6080604052600436106103f95760003560e01c80636f4986631161020d578063b5d8962711610128578063cfd47663116100bb578063deb6fb0d1161008a578063e59488661161006f578063e5948866146110a2578063ebdf104c146110de578063f2fde38b14611251576103f9565b8063deb6fb0d14610ff5578063e08d7e6614611025576103f9565b8063cfd4766314610f38578063cfdbb7cd14610f71578063d9a7c1f914610faa578063de67f21514610fbf576103f9565b8063c641ea28116100f7578063c641ea2814610e4d578063c65ee0e114610ec7578063c7be95de14610ef1578063cc8343aa14610f06576103f9565b8063b5d8962714610d79578063b6d9edd514610de4578063c3de580e14610e0e578063c5f530af14610e38576103f9565b80638cddb015116101a05780639fa6dd351161016f5780639fa6dd3514610ca7578063a1f0174f14610cc4578063a5a470ad14610cf4578063a778651514610d64576103f9565b80638cddb01514610bb05780638da5cb5b14610be95780638f32d59b14610c1a57806396c7ee4614610c43576103f9565b8063854873e1116101dc578063854873e114610acc5780638914d4c0146105f35780638b0e9f3f14610b6b5780638b1a0d1114610b80576103f9565b80636f49866314610a54578063715018a614610a8d5780637667180814610aa25780637cacb1d614610ab7576103f9565b80632cedb097116103185780634f864df4116102ab5780635ccfe1e81161027a5780635fab23a81161025f5780635fab23a8146109cd5780636099ecb2146109e2578063670322f814610a1b576103f9565b80635ccfe1e8146109945780635e2308d2146106aa576103f9565b80634f864df4146108395780634feb92f31461086f57806352b60bf31461091a57806354fd4d501461094a576103f9565b806341a19472116102e757806341a1947214610779578063441a3e70146107a95780634aa5ff6d146107d95780634f7c4efb14610809576103f9565b80632cedb097146106d45780632d296a9b1461070257806339b80c00146107175780633fee10a814610702576103f9565b806318f628d4116103905780631f2701521161035f5780631f270152146106385780632265f284146106955780632709275e146106aa57806328f73148146106bf576103f9565b806318f628d41461055e5780631d3ac42c146105c35780631d58179c146105f35780631e702f8314610608576103f9565b80630d7b2609116103cc5780630d7b2609146104cb57806312622d0e146104e0578063173a2c3c1461051957806318160ddd14610549576103f9565b80630135b1db146103fe578063019e2729146104435780630962ef791461048c5780630d4955e3146104b6575b600080fd5b34801561040a57600080fd5b506104316004803603602081101561042157600080fd5b50356001600160a01b0316611284565b60408051918252519081900360200190f35b34801561044f57600080fd5b5061048a6004803603608081101561046657600080fd5b508035906020810135906001600160a01b0360408201358116916060013516611296565b005b34801561049857600080fd5b5061048a600480360360208110156104af57600080fd5b5035611390565b3480156104c257600080fd5b5061043161147b565b3480156104d757600080fd5b50610431611484565b3480156104ec57600080fd5b506104316004803603604081101561050357600080fd5b506001600160a01b03813516906020013561148b565b34801561052557600080fd5b506104316004803603604081101561053c57600080fd5b5080359060200135611514565b34801561055557600080fd5b50610431611535565b34801561056a57600080fd5b5061048a600480360361012081101561058257600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e081013590610100013561153b565b3480156105cf57600080fd5b50610431600480360360408110156105e657600080fd5b508035906020013561168c565b3480156105ff57600080fd5b5061043161183d565b34801561061457600080fd5b5061048a6004803603604081101561062b57600080fd5b5080359060200135611842565b34801561064457600080fd5b506106776004803603606081101561065b57600080fd5b506001600160a01b0381351690602081013590604001356118f1565b60408051938452602084019290925282820152519081900360600190f35b3480156106a157600080fd5b50610431611923565b3480156106b657600080fd5b50610431611935565b3480156106cb57600080fd5b50610431611951565b3480156106e057600080fd5b506106e9611957565b6040805192835260208301919091528051918290030190f35b34801561070e57600080fd5b50610431611961565b34801561072357600080fd5b506107416004803603602081101561073a57600080fd5b5035611968565b604080519788526020880196909652868601949094526060860192909252608085015260a084015260c0830152519081900360e00190f35b34801561078557600080fd5b506104316004803603604081101561079c57600080fd5b50803590602001356119aa565b3480156107b557600080fd5b5061048a600480360360408110156107cc57600080fd5b50803590602001356119cb565b3480156107e557600080fd5b50610431600480360360408110156107fc57600080fd5b5080359060200135611ce2565b34801561081557600080fd5b5061048a6004803603604081101561082c57600080fd5b5080359060200135611d03565b34801561084557600080fd5b5061048a6004803603606081101561085c57600080fd5b5080359060208101359060400135611e47565b34801561087b57600080fd5b5061048a600480360361010081101561089357600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156108c357600080fd5b8201836020820111156108d557600080fd5b803590602001918460018302840111640100000000831117156108f757600080fd5b919350915080359060208101359060408101359060608101359060800135612123565b34801561092657600080fd5b506104316004803603604081101561093d57600080fd5b50803590602001356121c9565b34801561095657600080fd5b5061095f6121ea565b604080517fffffff00000000000000000000000000000000000000000000000000000000009092168252519081900360200190f35b3480156109a057600080fd5b50610431600480360360408110156109b757600080fd5b506001600160a01b03813516906020013561220e565b3480156109d957600080fd5b5061043161222b565b3480156109ee57600080fd5b5061043160048036036040811015610a0557600080fd5b506001600160a01b038135169060200135612231565b348015610a2757600080fd5b5061043160048036036040811015610a3e57600080fd5b506001600160a01b03813516906020013561227e565b348015610a6057600080fd5b5061043160048036036040811015610a7757600080fd5b506001600160a01b0381351690602001356122bf565b348015610a9957600080fd5b5061048a6122dc565b348015610aae57600080fd5b5061043161238c565b348015610ac357600080fd5b50610431612395565b348015610ad857600080fd5b50610af660048036036020811015610aef57600080fd5b503561239b565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610b30578181015183820152602001610b18565b50505050905090810190601f168015610b5d5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b348015610b7757600080fd5b50610431612436565b348015610b8c57600080fd5b5061048a60048036036040811015610ba357600080fd5b508035906020013561243c565b348015610bbc57600080fd5b5061048a60048036036040811015610bd357600080fd5b506001600160a01b0381351690602001356124de565b348015610bf557600080fd5b50610bfe612539565b604080516001600160a01b039092168252519081900360200190f35b348015610c2657600080fd5b50610c2f612548565b604080519115158252519081900360200190f35b348015610c4f57600080fd5b50610c7c60048036036040811015610c6657600080fd5b506001600160a01b038135169060200135612559565b6040805195865260208601949094528484019290925260608401526080830152519081900360a00190f35b61048a60048036036020811015610cbd57600080fd5b5035612593565b348015610cd057600080fd5b5061043160048036036040811015610ce757600080fd5b50803590602001356125a1565b61048a60048036036020811015610d0a57600080fd5b810190602081018135640100000000811115610d2557600080fd5b820183602082011115610d3757600080fd5b80359060200191846001830284011164010000000083111715610d5957600080fd5b5090925090506125be565b348015610d7057600080fd5b50610431612667565b348015610d8557600080fd5b50610da360048036036020811015610d9c57600080fd5b503561267d565b604080519788526020880196909652868601949094526060860192909252608085015260a08401526001600160a01b031660c0830152519081900360e00190f35b348015610df057600080fd5b5061048a60048036036020811015610e0757600080fd5b50356126c3565b348015610e1a57600080fd5b50610c2f60048036036020811015610e3157600080fd5b50356127b5565b348015610e4457600080fd5b506104316127cf565b348015610e5957600080fd5b50610e7760048036036020811015610e7057600080fd5b50356127de565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610eb3578181015183820152602001610e9b565b505050509050019250505060405180910390f35b348015610ed357600080fd5b5061043160048036036020811015610eea57600080fd5b5035612843565b348015610efd57600080fd5b50610431612855565b348015610f1257600080fd5b5061048a60048036036040811015610f2957600080fd5b5080359060200135151561285b565b348015610f4457600080fd5b5061043160048036036040811015610f5b57600080fd5b506001600160a01b038135169060200135612a76565b348015610f7d57600080fd5b50610c2f60048036036040811015610f9457600080fd5b506001600160a01b038135169060200135612a93565b348015610fb657600080fd5b50610431612afb565b348015610fcb57600080fd5b5061048a60048036036060811015610fe257600080fd5b5080359060208101359060400135612b01565b34801561100157600080fd5b506104316004803603604081101561101857600080fd5b5080359060200135612e18565b34801561103157600080fd5b5061048a6004803603602081101561104857600080fd5b81019060208101813564010000000081111561106357600080fd5b82018360208201111561107557600080fd5b8035906020019184602083028401116401000000008311171561109757600080fd5b509092509050612e39565b3480156110ae57600080fd5b50610431600480360360808110156110c557600080fd5b5080359060208101359060408101359060600135612f36565b3480156110ea57600080fd5b5061048a6004803603608081101561110157600080fd5b81019060208101813564010000000081111561111c57600080fd5b82018360208201111561112e57600080fd5b8035906020019184602083028401116401000000008311171561115057600080fd5b91939092909160208101903564010000000081111561116e57600080fd5b82018360208201111561118057600080fd5b803590602001918460208302840111640100000000831117156111a257600080fd5b9193909290916020810190356401000000008111156111c057600080fd5b8201836020820111156111d257600080fd5b803590602001918460208302840111640100000000831117156111f457600080fd5b91939092909160208101903564010000000081111561121257600080fd5b82018360208201111561122457600080fd5b8035906020019184602083028401116401000000008311171561124657600080fd5b509092509050612fbc565b34801561125d57600080fd5b5061048a6004803603602081101561127457600080fd5b50356001600160a01b0316613198565b60696020526000908152604090205481565b600054610100900460ff16806112af57506112af6131fa565b806112bd575060005460ff16155b6112f85760405162461bcd60e51b815260040180806020018281038252602e815260200180614906602e913960400191505060405180910390fd5b600054610100900460ff16158015611323576000805460ff1961ff0019909116610100171660011790555b61132c82613200565b60678590556066805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b03851617905560748490556755cfe697852e904c6073556103e86076556203f4806077558015611389576000805461ff00191690555b5050505050565b3361139b81836132fe565b506001600160a01b0381166000908152606f6020908152604080832085845290915290205480611412576040805162461bcd60e51b815260206004820152600c60248201527f7a65726f20726577617264730000000000000000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0382166000908152606f6020908152604080832086845290915281205561143f816133c2565b6040516001600160a01b0383169082156108fc029083906000818181858888f19350505050158015611475573d6000803e3d6000fd5b50505050565b6301e133805b90565b6212750090565b60006114978383612a93565b6114c557506001600160a01b038216600090815260716020908152604080832084845290915290205461150e565b6001600160a01b03831660008181526072602090815260408083208684528252808320549383526071825280832086845290915290205461150b9163ffffffff61344216565b90505b92915050565b60009182526075602090815260408084209284526002909201905290205490565b60745481565b61154433613484565b61157f5760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b61158a898989613498565b6001600160a01b0389166000908152606f602090815260408083208b845290915290208190556115b9876133c2565b851561168157868611156115fe5760405162461bcd60e51b815260040180806020018281038252602c8152602001806149a6602c913960400191505060405180910390fd5b6001600160a01b03891660008181526072602090815260408083208c84528252918290208981556001810189905560028101889055600381018790556004810186905582518781529182018a9052825190938c9390927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78929081900390910190a3505b505050505050505050565b3360008181526072602090815260408083208684529091528120909190836116e9576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b6116f38286612a93565b611744576040805162461bcd60e51b815260206004820152600d60248201527f6e6f74206c6f636b656420757000000000000000000000000000000000000000604482015290519081900360640190fd5b805484111561179a576040805162461bcd60e51b815260206004820152601760248201527f6e6f7420656e6f756768206c6f636b6564207374616b65000000000000000000604482015290519081900360640190fd5b6117a482866132fe565b5060006117ba82600401548684600001546135b7565b6004830180548290039055825486900383556001600160a01b03841660008181526071602090815260408083208b8452825291829020805485900390558151898152908101849052815193945089937fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9929181900390910190a395945050505050565b600390565b61184b33613484565b6118865760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b806118d8576040805162461bcd60e51b815260206004820152600c60248201527f77726f6e67207374617475730000000000000000000000000000000000000000604482015290519081900360640190fd5b6118e282826135e4565b6118ed82600061285b565b5050565b607060209081526000938452604080852082529284528284209052825290208054600182015460029092015490919083565b600061192d613693565b601002905090565b60006064611941613693565b601e028161194b57fe5b04905090565b606d5481565b6076546077549091565b62093a8090565b607560205280600052604060002060009150905080600701549080600801549080600901549080600a01549080600b01549080600c01549080600d0154905087565b60009182526075602090815260408084209284526003909201905290205490565b336119d461476b565b506001600160a01b038116600090815260706020908152604080832086845282528083208584528252918290208251606081018452815480825260018301549382019390935260029091015492810192909252611a78576040805162461bcd60e51b815260206004820152601560248201527f7265717565737420646f65736e27742065786973740000000000000000000000604482015290519081900360640190fd5b60208082015182516000878152606890935260409092206001015490919015801590611ab4575060008681526068602052604090206001015482115b15611ad5575050600084815260686020526040902060018101546002909101545b611add611961565b8201611ae761369f565b1015611b3a576040805162461bcd60e51b815260206004820152601660248201527f6e6f7420656e6f7567682074696d652070617373656400000000000000000000604482015290519081900360640190fd5b611b4261183d565b8101611b4c61238c565b1015611b9f576040805162461bcd60e51b815260206004820152601860248201527f6e6f7420656e6f7567682065706f636873207061737365640000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152607060209081526040808320898452825280832088845290915281206002015490611bd8886127b5565b90506000611bfa8383607860008d8152602001908152602001600020546136a3565b6001600160a01b03881660009081526070602090815260408083208d845282528083208c845290915281208181556001810182905560020155606e8054820190559050808311611c91576040805162461bcd60e51b815260206004820152601660248201527f7374616b652069732066756c6c7920736c617368656400000000000000000000604482015290519081900360640190fd5b6001600160a01b0387166108fc611cae858463ffffffff61344216565b6040518115909202916000818181858888f19350505050158015611cd6573d6000803e3d6000fd5b50505050505050505050565b60009182526075602090815260408084209284526001909201905290205490565b611d0b612548565b611d5c576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b611d65826127b5565b611db6576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f722069736e277420736c6173686564000000000000000000604482015290519081900360640190fd5b611dbe613693565b811115611dfc5760405162461bcd60e51b81526004018080602001828103825260218152602001806149346021913960400191505060405180910390fd5b6000828152607860209081526040918290208390558151838152915184927f047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb91792908290030190a25050565b33611e5281856132fe565b5060008211611e96576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b611ea0818561148b565b821115611ef4576040805162461bcd60e51b815260206004820152601960248201527f6e6f7420656e6f75676820756e6c6f636b6564207374616b6500000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607060209081526040808320878452825280832086845290915290206002015415611f75576040805162461bcd60e51b815260206004820152601360248201527f7772494420616c72656164792065786973747300000000000000000000000000604482015290519081900360640190fd5b6001600160a01b03811660009081526071602090815260408083208784528252808320805486900390556068909152902060030154611fba908363ffffffff61344216565b600085815260686020526040902060030155606c54611fdf908363ffffffff61344216565b606c5560008481526068602052604090205461200c57606d54612008908363ffffffff61344216565b606d555b61201584613702565b8061202657506120248461373e565b155b6120615760405162461bcd60e51b815260040180806020018281038252602981526020018061497d6029913960400191505060405180910390fd5b61206a8461373e565b612079576120798460016135e4565b6001600160a01b0381166000908152607060209081526040808320878452825280832086845290915290206002018290556120b261238c565b6001600160a01b038216600090815260706020908152604080832088845282528083208784529091529020556120e661369f565b6001600160a01b0382166000908152607060209081526040808320888452825280832087845290915281206001019190915561147590859061285b565b61212c33613484565b6121675760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b6121af898989898080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b92508a91508990508888613770565b606b5488111561168157606b889055505050505050505050565b60009182526075602090815260408084209284526005909201905290205490565b7f323032000000000000000000000000000000000000000000000000000000000090565b607960209081526000928352604080842090915290825290205481565b606e5481565b60008061223e8484613859565b506001600160a01b0385166000908152606f60209081526040808320878452909152902054909150612276908263ffffffff6139a716565b949350505050565b600061228a8383612a93565b6122965750600061150e565b506001600160a01b03919091166000908152607260209081526040808320938352929052205490565b606f60209081526000928352604080842090915290825290205481565b6122e4612548565b612335576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a36033805473ffffffffffffffffffffffffffffffffffffffff19169055565b60675460010190565b60675481565b606a6020908152600091825260409182902080548351601f60026000196101006001861615020190931692909204918201849004840281018401909452808452909183018282801561242e5780601f106124035761010080835404028352916020019161242e565b820191906000526020600020905b81548152906001019060200180831161241157829003601f168201915b505050505081565b606c5481565b612444612548565b612495576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b60778190556076829055604080518381526020810183905281517f702756a07c05d0bbfd06fc17b67951a5f4deb7bb6b088407e68a58969daf2a34929181900390910190a15050565b6124e882826132fe565b6118ed576040805162461bcd60e51b815260206004820152601060248201527f6e6f7468696e6720746f20737461736800000000000000000000000000000000604482015290519081900360640190fd5b6033546001600160a01b031690565b6033546001600160a01b0316331490565b6072602090815260009283526040808420909152908252902080546001820154600283015460038401546004909401549293919290919085565b61259e338234613a01565b50565b600091825260756020908152604080842092845291905290205490565b6125c66127cf565b34101561261a576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b61265a3383838080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250613b0b92505050565b6118ed33606b5434613a01565b60006064612673613693565b600f028161194b57fe5b606860205260009081526040902080546001820154600283015460038401546004850154600586015460069096015494959394929391929091906001600160a01b031687565b6126cb612548565b61271c576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6801c985c8903591eb2081111561277a576040805162461bcd60e51b815260206004820152601b60248201527f746f6f206c617267652072657761726420706572207365636f6e640000000000604482015290519081900360640190fd5b60738190556040805182815290517f8cd9dae1bbea2bc8a5e80ffce2c224727a25925130a03ae100619a8861ae23969181900360200190a150565b60008181526068602052604090205460801615155b919050565b6a02a055184a310c1260000090565b60008181526075602090815260409182902060060180548351818402810184019094528084526060939283018282801561283757602002820191906000526020600020905b815481526020019060010190808311612823575b50505050509050919050565b60786020526000908152604090205481565b606b5481565b61286482613b36565b6128b5576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b600082815260686020526040902060038101549054156128d3575060005b606654604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018690526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b15801561294057600080fd5b505af1158015612954573d6000803e3d6000fd5b5050505081801561296457508015155b15612a71576066546000848152606a60205260409081902081517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810187815260248201938452825460026000196001831615610100020190911604604483018190526001600160a01b039095169463242a6e3f94899493909160649091019084908015612a375780601f10612a0c57610100808354040283529160200191612a37565b820191906000526020600020905b815481529060010190602001808311612a1a57829003601f168201915b50509350505050600060405180830381600087803b158015612a5857600080fd5b505af1158015612a6c573d6000803e3d6000fd5b505050505b505050565b607160209081526000928352604080842090915290825290205481565b6001600160a01b03821660009081526072602090815260408083208484529091528120600201541580159061150b57506001600160a01b0383166000908152607260209081526040808320858452909152902060020154612af261369f565b11159392505050565b60735481565b3381612b42576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b612b4c8185612a93565b15612b9e576040805162461bcd60e51b815260206004820152601160248201527f616c7265616479206c6f636b6564207570000000000000000000000000000000604482015290519081900360640190fd5b612ba8818561148b565b821115612bfc576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b60008481526068602052604090205415612c5d576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b612c65611484565b8310158015612c7b5750612c7761147b565b8311155b612ccc576040805162461bcd60e51b815260206004820152601260248201527f696e636f7272656374206475726174696f6e0000000000000000000000000000604482015290519081900360640190fd5b6000612ce684612cda61369f565b9063ffffffff6139a716565b6000868152606860205260409020600601549091506001600160a01b039081169083168114612d74576001600160a01b0381166000908152607260209081526040808320898452909152902060020154821115612d745760405162461bcd60e51b81526004018080602001828103825260288152602001806149556028913960400191505060405180910390fd5b612d7e83876132fe565b506001600160a01b03831660009081526072602090815260408083208984529091529020848155612dad61238c565b60018201556002810183905560038101869055600060048201556040805187815260208101879052815189926001600160a01b038816927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78929081900390910190a350505050505050565b60009182526075602090815260408084209284526004909201905290205490565b612e4233613484565b612e7d5760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b600060756000612e8b61238c565b8152602001908152602001600020905060008090505b82811015612f2757600060686000868685818110612ebb57fe5b90506020020135815260200190815260200160002060030154905080836000016000878786818110612ee957fe5b90506020020135815260200190815260200160002081905550612f198184600c01546139a790919063ffffffff16565b600c84015550600101612ea1565b5061147560068201848461478c565b6000818310612f4757506000612276565b600083815260756020818152604080842088855260019081018352818520548786529383528185208986520190915290912054612fb1612f85613693565b612fa589612f99858763ffffffff61344216565b9063ffffffff613b4d16565b9063ffffffff613ba616565b979650505050505050565b612fc533613484565b6130005760405162461bcd60e51b81526004018080602001828103825260298152602001806148bc6029913960400191505060405180910390fd5b60006075600061300e61238c565b8152602001908152602001600020905060608160060180548060200260200160405190810160405280929190818152602001828054801561306e57602002820191906000526020600020905b81548152602001906001019080831161305a575b505050505090506130f582828c8c80806020026020016040519081016040528093929190818152602001838360200280828437600081840152601f19601f820116905080830192505050505050508b8b80806020026020016040519081016040528093929190818152602001838360200280828437600092019190915250613be892505050565b613164828288888080602002602001604051908101604052809392919081815260200183836020028082843760009201919091525050604080516020808c0282810182019093528b82529093508b92508a918291850190849080828437600092019190915250613cf792505050565b61316c61238c565b60675561317761369f565b600783015550607354600b820155607454600d909101555050505050505050565b6131a0612548565b6131f1576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b61259e8161429d565b303b1590565b600054610100900460ff168061321957506132196131fa565b80613227575060005460ff16155b6132625760405162461bcd60e51b815260040180806020018281038252602e815260200180614906602e913960400191505060405180910390fd5b600054610100900460ff1615801561328d576000805460ff1961ff0019909116610100171660011790555b6033805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0384811691909117918290556040519116906000907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908290a380156118ed576000805461ff00191690555050565b600080600061330d8585613859565b9150915061331a8461434b565b6001600160a01b0386166000818152607960209081526040808320898452825280832094909455918152606f82528281208782529091522054613363908363ffffffff6139a716565b6001600160a01b0386166000818152606f60209081526040808320898452825280832094909455918152607282528281208782529091522060048101546133b0908363ffffffff6139a716565b60049091015550600191505092915050565b606654604080517f66e7ea0f0000000000000000000000000000000000000000000000000000000081523060048201526024810184905290516001600160a01b03909216916366e7ea0f9160448082019260009290919082900301818387803b15801561342e57600080fd5b505af1158015611389573d6000803e3d6000fd5b600061150b83836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f7700008152506143a6565b6066546001600160a01b0390811691161490565b600081116134db576040805162461bcd60e51b815260206004820152600b60248201526a1e995c9bc8185b5bdd5b9d60aa1b604482015290519081900360640190fd5b6134e583836132fe565b506001600160a01b038316600090815260716020908152604080832085845290915290205461351a908263ffffffff6139a716565b6001600160a01b038416600090815260716020908152604080832086845282528083209390935560689052206003015461355a818363ffffffff6139a716565b600084815260686020526040902060030155606c5461357f908363ffffffff6139a716565b606c556000838152606860205260409020546135ac57606d546135a8908363ffffffff6139a716565b606d555b61147583821561285b565b6000806135ce83612fa5878763ffffffff613b4d16565b90508381106135da5750825b90505b9392505050565b6000828152606860205260409020541580156135ff57508015155b1561362c57600082815260686020526040902060030154606d546136289163ffffffff61344216565b606d555b6000828152606860205260409020548111156118ed576000828152606860205260409020818155600201546118ed5761366361369f565b60008381526068602052604090206001015561367d61238c565b6000838152606860205260409020600201555050565b670de0b6b3a764000090565b4290565b60008215806136b957506136b5613693565b8210155b156136c6575060006135dd565b6136f16001612cda6136d6613693565b612fa5866136e2613693565b8a91900363ffffffff613b4d16565b9050838111156135dd5750826135dd565b600061372361370f613693565b612fa561371a611923565b612f998661373e565b60008381526068602052604090206003015411159050919050565b6000818152606860209081526040808320600601546001600160a01b0316835260718252808320938352929052205490565b6001600160a01b038816600090815260696020526040902054156137db576040805162461bcd60e51b815260206004820152601860248201527f76616c696461746f7220616c7265616479206578697374730000000000000000604482015290519081900360640190fd5b6001600160a01b03881660008181526069602090815260408083208b90558a83526068825280832089815560048101899055600581018890556001810186905560028101879055600601805473ffffffffffffffffffffffffffffffffffffffff1916909417909355606a81529190208751611681928901906147d7565b6001600160a01b03821660009081526079602090815260408083208484529091528120548190816138898561434b565b90506000613897878761443d565b9050818111156138a45750805b828110156138af5750815b6001600160a01b03871660008181526072602090815260408083208a84528252808320938352607182528083208a845290915281205482549091906138fb90839063ffffffff61344216565b9050600061390f84600001548b8988612f36565b90506000806139228387600301546144fc565b91509150613932848d8b8a612f36565b92506000806139428560006144fc565b91509150613952878f8b8d612f36565b94506000806139628760006144fc565b909250905061397b82612cda888763ffffffff6139a716565b61398f82612cda888763ffffffff6139a716565b9e509e50505050505050505050505050509250929050565b60008282018381101561150b576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b613a0a82613b36565b613a5b576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090205415613abc576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b613ac7838383613498565b613ad082613702565b612a715760405162461bcd60e51b815260040180806020018281038252602981526020018061497d6029913960400191505060405180910390fd5b606b805460010190819055612a718382846000613b2661238c565b613b2e61369f565b600080613770565b600090815260686020526040902060050154151590565b600082613b5c5750600061150e565b82820282848281613b6957fe5b041461150b5760405162461bcd60e51b81526004018080602001828103825260218152602001806148e56021913960400191505060405180910390fd5b600061150b83836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f0000000000008152506145cb565b60005b835181101561138957607654828281518110613c0357fe5b6020026020010151118015613c2d5750607754838281518110613c2257fe5b602002602001015110155b15613c6e57613c50848281518110613c4157fe5b602002602001015160086135e4565b613c6e848281518110613c5f57fe5b6020026020010151600061285b565b828181518110613c7a57fe5b6020026020010151856004016000868481518110613c9457fe5b6020026020010151815260200190815260200160002081905550818181518110613cba57fe5b6020026020010151856005016000868481518110613cd457fe5b602090810291909101810151825281019190915260400160002055600101613beb565b613cff614845565b6040518060c001604052808551604051908082528060200260200182016040528015613d35578160200160208202803883390190505b508152602001600081526020018551604051908082528060200260200182016040528015613d6d578160200160208202803883390190505b508152602001600081526020016000815260200160008152509050600060756000613da76001613d9b61238c565b9063ffffffff61344216565b81526020810191909152604001600020600160808401526007810154909150613dce61369f565b1115613de8578060070154613de161369f565b0360808301525b60005b8551811015613ebe57613e358360800151612fa5878481518110613e0b57fe5b6020026020010151878581518110613e1f57fe5b6020026020010151613b4d90919063ffffffff16565b83604001518281518110613e4557fe5b602002602001018181525050613e7f83604001518281518110613e6457fe5b602002602001015184606001516139a790919063ffffffff16565b60608401528351613eb190859083908110613e9657fe5b60200260200101518460a001516139a790919063ffffffff16565b60a0840152600101613deb565b5060005b8551811015613f9557613f408360800151612fa5878481518110613ee257fe5b6020026020010151612f998760800151612fa58b8881518110613f0157fe5b60200260200101518e60000160008f8b81518110613f1b57fe5b6020026020010151815260200190815260200160002054613b4d90919063ffffffff16565b8351805183908110613f4e57fe5b602002602001018181525050613f8883600001518281518110613f6d57fe5b602002602001015184602001516139a790919063ffffffff16565b6020840152600101613ec2565b5060005b8551811015614275576000613fd1846080015160735486600001518581518110613fbf57fe5b60200260200101518760200151614630565b905061400d6140008560a0015186604001518581518110613fee57fe5b60200260200101518760600151614671565b829063ffffffff6139a716565b9050600087838151811061401d57fe5b6020908102919091018101516000818152606883526040808220600601546001600160a01b031680835260728552818320848452909452812091935061406a85614065612667565b6146ce565b6001600160a01b03841660009081526071602090815260408083208884529091528120549192509061409c858761227e565b8302816140a557fe5b049050600081830390506000806140c08487600301546144fc565b915091506000806140d28560006144fc565b6001600160a01b038b166000908152606f602090815260408083208f84529091529020549193509150614111908390612cda908763ffffffff6139a716565b6001600160a01b038a166000908152606f602090815260408083208e84529091529020556004880154614150908290612cda908663ffffffff6139a716565b60048901555050606c54858a0394506000935061418292509050612fa5614175613693565b859063ffffffff613b4d16565b600087815260018b0160205260409020549091506141a6908263ffffffff6139a716565b8e6001016000888152602001908152602001600020819055506141fa8b89815181106141ce57fe5b60200260200101518a6003016000898152602001908152602001600020546139a790919063ffffffff16565b8e60030160008881526020019081526020016000208190555061424e8c898151811061422257fe5b60200260200101518a6002016000898152602001908152602001600020546139a790919063ffffffff16565b600096875260028f01602052604090962095909555505060019094019350613f9992505050565b505060a081015160088601556020810151600986015560600151600a90940193909355505050565b6001600160a01b0381166142e25760405162461bcd60e51b81526004018080602001828103825260268152602001806148966026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a36033805473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b0392909216919091179055565b6000818152606860205260408120600201541561439e57600082815260686020526040902060020154606754101561438657506067546127ca565b506000818152606860205260409020600201546127ca565b505060675490565b600081848411156144355760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b838110156143fa5781810151838201526020016143e2565b50505050905090810190601f1680156144275780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b6001600160a01b03821660009081526072602090815260408083208484529091528120600101546067546144728585836146eb565b1561448057915061150e9050565b61448b8585846146eb565b61449a5760009250505061150e565b808211156144ad5760009250505061150e565b808210156144e0576002818301046144c68686836146eb565b156144d6578060010192506144da565b8091505b506144ad565b806144f05760009250505061150e565b60001901949350505050565b600080821561459a57600061450f611935565b614517613693565b039050600061453761452761147b565b612fa5848863ffffffff613b4d16565b905061455e614544613693565b612fa583614550611935565b8a910163ffffffff613b4d16565b935061459161456b613693565b612fa5836002614579611935565b8161458057fe5b040189613b4d90919063ffffffff16565b925050506145c4565b6145bd6145a5613693565b612fa56145b0611935565b879063ffffffff613b4d16565b9150600090505b9250929050565b6000818361461a5760405162461bcd60e51b81526020600482018181528351602484015283519092839260449091019190850190808383600083156143fa5781810151838201526020016143e2565b50600083858161462657fe5b0495945050505050565b60008261463f57506000612276565b6000614651868663ffffffff613b4d16565b905061466783612fa5838763ffffffff613b4d16565b9695505050505050565b600082614680575060006135dd565b600061469683612fa5878763ffffffff613b4d16565b90506146c56146a3613693565b612fa56146ae611935565b6146b6613693565b8591900363ffffffff613b4d16565b95945050505050565b600061150b6146db613693565b612fa5858563ffffffff613b4d16565b6001600160a01b038316600090815260726020908152604080832085845290915281206001015482108015906135da57506001600160a01b038416600090815260726020908152604080832086845290915290206002015461474c83614756565b1115949350505050565b60009081526075602052604090206007015490565b60405180606001604052806000815260200160008152602001600081525090565b8280548282559060005260206000209081019282156147c7579160200282015b828111156147c75782358255916020019190600101906147ac565b506147d392915061487b565b5090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061481857805160ff19168380011785556147c7565b828001600101855582156147c7579182015b828111156147c757825182559160200191906001019061482a565b6040518060c001604052806060815260200160008152602001606081526020016000815260200160008152602001600081525090565b61148191905b808211156147d3576000815560010161488156fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f206164647265737363616c6c6572206973206e6f7420746865204e6f64654472697665724175746820636f6e7472616374536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f77436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a65646d757374206265206c657373207468616e206f7220657175616c20746f20312e3076616c696461746f72206c6f636b757020706572696f642077696c6c20656e64206561726c69657276616c696461746f7227732064656c65676174696f6e73206c696d69742069732065786365656465646c6f636b6564207374616b652069732067726561746572207468616e207468652077686f6c65207374616b65a265627a7a7231582015ceb1c92ead96d3fabbf6f3824838f149d5774ba2d8e2395e382bade4268c1b64736f6c634300050c0032"
