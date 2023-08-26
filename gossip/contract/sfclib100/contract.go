// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sfclib100

import (
	"errors"
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
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"BurntFTM\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"ChangedValidatorStatus\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupExtraReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupBaseReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unlockedReward\",\"type\":\"uint256\"}],\"name\":\"ClaimedRewards\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"}],\"name\":\"CreatedValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"DeactivatedValidator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Delegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"LockedUpStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RefundedSlashedLegacyDelegation\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupExtraReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lockupBaseReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"unlockedReward\",\"type\":\"uint256\"}],\"name\":\"RestakedRewards\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Undelegated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"penalty\",\"type\":\"uint256\"}],\"name\":\"UnlockedStake\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"UpdatedSlashingRefundRatio\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"syncPubkey\",\"type\":\"bool\"}],\"name\":\"_syncValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"currentSealedEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getEpochSnapshot\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"epochFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalBaseRewardWeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalTxRewardWeight\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"baseRewardPerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalSupply\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"getLockedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getLockupInfo\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getStashedLockupRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"lockupExtraReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupBaseReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"unlockedReward\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getValidator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"receivedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"getValidatorID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getValidatorPubkey\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"getWithdrawalRequest\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"isLockedUp\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"isOwner\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lastValidatorID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"minGasPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"slashingRefundRatio\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"stakeTokenizerAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"stashedRewardsUntilEpoch\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalActiveStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSlashedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"treasuryAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"voteBookAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"}],\"name\":\"getEpochValidatorIDs\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochReceivedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochAccumulatedRewardPerToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochAccumulatedUptime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochAccumulatedOriginatedTxsFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochOfflineTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"epoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getEpochOfflineBlocks\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"rewardsStash\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"auth\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deactivatedTime\",\"type\":\"uint256\"}],\"name\":\"setGenesisValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupFromEpoch\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupEndTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"earlyUnlockPenalty\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewards\",\"type\":\"uint256\"}],\"name\":\"setGenesisDelegation\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"pubkey\",\"type\":\"bytes\"}],\"name\":\"createValidator\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"getSelfStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"delegate\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"validatorAuth\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"strict\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"gas\",\"type\":\"uint256\"}],\"name\":\"recountVotes\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"undelegate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"liquidateSFTM\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"v\",\"type\":\"address\"}],\"name\":\"updateSFTMFinalizer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"}],\"name\":\"isSlashed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"wrID\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"status\",\"type\":\"uint256\"}],\"name\":\"deactivateValidator\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"pendingRewards\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"stashRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"claimRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"restakeRewards\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burnFTM\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"}],\"name\":\"getUnlockedStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"lockStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockupDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"relockStake\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"toValidatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"unlockStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"validatorID\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"refundRatio\",\"type\":\"uint256\"}],\"name\":\"updateSlashingRefundRatio\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50615220806100206000396000f3fe6080604052600436106103815760003560e01c8063854873e1116101d1578063b88a37e211610102578063cfd47663116100a0578063de67f2151161006f578063de67f21514610f3a578063df00c92214610f70578063e261641a14610fa0578063f2fde38b14610fd057610381565b8063cfd4766314610e83578063cfdbb7cd14610ebc578063d96ed50514610ef5578063dc31e1af14610f0a57610381565b8063c5f956af116100dc578063c5f956af14610dfd578063c65ee0e114610e12578063c7be95de14610e3c578063cc8343aa14610e5157610381565b8063b88a37e214610d23578063bd14d90714610d9d578063c3de580e14610dd357610381565b806396c7ee461161016f578063a5a470ad11610149578063a5a470ad14610bd6578063a86a056f14610c46578063b5d8962714610c7f578063b810e41114610cea57610381565b806396c7ee4614610b2a5780639fa6dd3514610b89578063a198d22914610ba657610381565b80638cddb015116101ab5780638cddb01514610a895780638da5cb5b14610ac25780638f32d59b14610ad757806390a6c47514610b0057610381565b8063854873e1146109c0578063893675c614610a5f5780638b0e9f3f14610a7457610381565b80633a488397116102b65780635fab23a8116102545780636f498663116102235780636f49866314610948578063715018a61461098157806376671808146109965780637cacb1d6146109ab57610381565b80635fab23a8146108915780636099ecb2146108a657806361e53fcc146108df578063670322f81461090f57610381565b80634f864df4116102905780634f864df4146107565780634feb92f31461078c5780635601fe011461083757806358f95b801461086157610381565b80633a488397146106b7578063441a3e70146106f65780634f7c4efb1461072657610381565b80631d3ac42c1161032357806320c0849d116102fd57806320c0849d146105c257806328f731481461060d5780632ce719601461062257806339b80c001461065557610381565b80631d3ac42c146105055780631e702f83146105355780631f2701521461056557610381565b80630e559d821161035f5780630e559d821461042157806312622d0e1461045257806318160ddd1461048b57806318f628d4146104a057610381565b80630135b1db1461038657806308c36874146103cb5780630962ef79146103f7575b600080fd5b34801561039257600080fd5b506103b9600480360360208110156103a957600080fd5b50356001600160a01b0316611003565b60408051918252519081900360200190f35b3480156103d757600080fd5b506103f5600480360360208110156103ee57600080fd5b5035611015565b005b34801561040357600080fd5b506103f56004803603602081101561041a57600080fd5b50356110e1565b34801561042d57600080fd5b5061043661122c565b604080516001600160a01b039092168252519081900360200190f35b34801561045e57600080fd5b506103b96004803603604081101561047557600080fd5b506001600160a01b03813516906020013561123b565b34801561049757600080fd5b506103b96112c4565b3480156104ac57600080fd5b506103f560048036036101208110156104c457600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e08101359061010001356112ca565b34801561051157600080fd5b506103b96004803603604081101561052857600080fd5b508035906020013561142c565b34801561054157600080fd5b506103f56004803603604081101561055857600080fd5b5080359060200135611656565b34801561057157600080fd5b506105a46004803603606081101561058857600080fd5b506001600160a01b03813516906020810135906040013561172f565b60408051938452602084019290925282820152519081900360600190f35b3480156105ce57600080fd5b506103f5600480360360808110156105e557600080fd5b506001600160a01b038135811691602081013590911690604081013515159060600135611761565b34801561061957600080fd5b506103b96118fe565b34801561062e57600080fd5b506103f56004803603602081101561064557600080fd5b50356001600160a01b0316611904565b34801561066157600080fd5b5061067f6004803603602081101561067857600080fd5b5035611997565b604080519788526020880196909652868601949094526060860192909252608085015260a084015260c0830152519081900360e00190f35b3480156106c357600080fd5b506103f5600480360360608110156106da57600080fd5b506001600160a01b0381351690602081013590604001356119d9565b34801561070257600080fd5b506103f56004803603604081101561071957600080fd5b5080359060200135611d8a565b34801561073257600080fd5b506103f56004803603604081101561074957600080fd5b508035906020013561229a565b34801561076257600080fd5b506103f56004803603606081101561077957600080fd5b50803590602081013590604001356123de565b34801561079857600080fd5b506103f560048036036101008110156107b057600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156107e057600080fd5b8201836020820111156107f257600080fd5b8035906020019184600183028401116401000000008311171561081457600080fd5b919350915080359060208101359060408101359060608101359060800135612679565b34801561084357600080fd5b506103b96004803603602081101561085a57600080fd5b503561271f565b34801561086d57600080fd5b506103b96004803603604081101561088457600080fd5b5080359060200135612755565b34801561089d57600080fd5b506103b9612772565b3480156108b257600080fd5b506103b9600480360360408110156108c957600080fd5b506001600160a01b038135169060200135612778565b3480156108eb57600080fd5b506103b96004803603604081101561090257600080fd5b50803590602001356127b6565b34801561091b57600080fd5b506103b96004803603604081101561093257600080fd5b506001600160a01b0381351690602001356127d7565b34801561095457600080fd5b506103b96004803603604081101561096b57600080fd5b506001600160a01b038135169060200135612818565b34801561098d57600080fd5b506103f5612882565b3480156109a257600080fd5b506103b961293d565b3480156109b757600080fd5b506103b9612947565b3480156109cc57600080fd5b506109ea600480360360208110156109e357600080fd5b503561294d565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610a24578181015183820152602001610a0c565b50505050905090810190601f168015610a515780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b348015610a6b57600080fd5b50610436612a06565b348015610a8057600080fd5b506103b9612a15565b348015610a9557600080fd5b506103f560048036036040811015610aac57600080fd5b506001600160a01b038135169060200135612a1b565b348015610ace57600080fd5b50610436612a7a565b348015610ae357600080fd5b50610aec612a89565b604080519115158252519081900360200190f35b348015610b0c57600080fd5b506103f560048036036020811015610b2357600080fd5b5035612a9a565b348015610b3657600080fd5b50610b6360048036036040811015610b4d57600080fd5b506001600160a01b038135169060200135612aff565b604080519485526020850193909352838301919091526060830152519081900360800190f35b6103f560048036036020811015610b9f57600080fd5b5035612b31565b348015610bb257600080fd5b506103b960048036036040811015610bc957600080fd5b5080359060200135612b3c565b6103f560048036036020811015610bec57600080fd5b810190602081018135640100000000811115610c0757600080fd5b820183602082011115610c1957600080fd5b80359060200191846001830284011164010000000083111715610c3b57600080fd5b509092509050612b5d565b348015610c5257600080fd5b506103b960048036036040811015610c6957600080fd5b506001600160a01b038135169060200135612cca565b348015610c8b57600080fd5b50610ca960048036036020811015610ca257600080fd5b5035612ce7565b604080519788526020880196909652868601949094526060860192909252608085015260a08401526001600160a01b031660c0830152519081900360e00190f35b348015610cf657600080fd5b506105a460048036036040811015610d0d57600080fd5b506001600160a01b038135169060200135612d2d565b348015610d2f57600080fd5b50610d4d60048036036020811015610d4657600080fd5b5035612d59565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610d89578181015183820152602001610d71565b505050509050019250505060405180910390f35b348015610da957600080fd5b506103f560048036036060811015610dc057600080fd5b5080359060208101359060400135612dbe565b348015610ddf57600080fd5b50610aec60048036036020811015610df657600080fd5b5035612dd1565b348015610e0957600080fd5b50610436612de8565b348015610e1e57600080fd5b506103b960048036036020811015610e3557600080fd5b5035612df7565b348015610e4857600080fd5b506103b9612e09565b348015610e5d57600080fd5b506103f560048036036040811015610e7457600080fd5b50803590602001351515612e0f565b348015610e8f57600080fd5b506103b960048036036040811015610ea657600080fd5b506001600160a01b038135169060200135613047565b348015610ec857600080fd5b50610aec60048036036040811015610edf57600080fd5b506001600160a01b038135169060200135613064565b348015610f0157600080fd5b506103b96130fa565b348015610f1657600080fd5b506103b960048036036040811015610f2d57600080fd5b5080359060200135613100565b348015610f4657600080fd5b506103f560048036036060811015610f5d57600080fd5b5080359060208101359060400135613121565b348015610f7c57600080fd5b506103b960048036036040811015610f9357600080fd5b50803590602001356131dc565b348015610fac57600080fd5b506103b960048036036040811015610fc357600080fd5b50803590602001356131fd565b348015610fdc57600080fd5b506103f560048036036020811015610ff357600080fd5b50356001600160a01b031661321e565b60696020526000908152604090205481565b3361101e615024565b6110288284613280565b602081015181519192506000916110449163ffffffff6133e016565b905061106783856110628560400151856133e090919063ffffffff16565b61343a565b6001600160a01b0383166000818152607360209081526040808320888452825291829020805485019055845185820151868401518451928352928201528083019190915290518692917f4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26919081900360600190a350505050565b336110ea615024565b6110f48284613280565b90506000826001600160a01b03166111318360400151611125856020015186600001516133e090919063ffffffff16565b9063ffffffff6133e016565b604051600081818185875af1925050503d806000811461116d576040519150601f19603f3d011682016040523d82523d6000602084013e611172565b606091505b50509050806111c8576040805162461bcd60e51b815260206004820152601260248201527f4661696c656420746f2073656e642046544d0000000000000000000000000000604482015290519081900360640190fd5b83836001600160a01b03167fc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf4584600001518560200151866040015160405180848152602001838152602001828152602001935050505060405180910390a350505050565b607b546001600160a01b031681565b60006112478383613064565b61127557506001600160a01b03821660009081526072602090815260408083208484529091529020546112be565b6001600160a01b0383166000818152607360209081526040808320868452825280832054938352607282528083208684529091529020546112bb9163ffffffff61354616565b90505b92915050565b60765481565b6112d333613588565b61130e5760405162461bcd60e51b81526004018080602001828103825260298152602001806151046029913960400191505060405180910390fd5b61131b898989600061359c565b6001600160a01b0389166000908152606f602090815260408083208b8452909152902060020181905561134d87613734565b851561142157868611156113925760405162461bcd60e51b815260040180806020018281038252602c8152602001806151c0602c913960400191505060405180910390fd5b6001600160a01b03891660008181526073602090815260408083208c845282528083208a8155600181018a90556002810189905560038101889055848452607483528184208d855283529281902086905580518781529182018a9052805192938c9390927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e7892908290030190a3505b505050505050505050565b33600081815260736020908152604080832086845290915281209091908361149b576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6114a58286613064565b6114f6576040805162461bcd60e51b815260206004820152600d60248201527f6e6f74206c6f636b656420757000000000000000000000000000000000000000604482015290519081900360640190fd5b805484111561154c576040805162461bcd60e51b815260206004820152601760248201527f6e6f7420656e6f756768206c6f636b6564207374616b65000000000000000000604482015290519081900360640190fd5b61155682866137d2565b6115a7576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b6115b1828661388d565b5060006115c48387878560000154613a58565b905081600301546363401ec501826002015410156115e0575060005b815485900382558015611603576115fa8387836001613b89565b61160381613d8f565b85836001600160a01b03167fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b98784604051808381526020018281526020019250505060405180910390a395945050505050565b61165f33613588565b61169a5760405162461bcd60e51b81526004018080602001828103825260298152602001806151046029913960400191505060405180910390fd5b806116ec576040805162461bcd60e51b815260206004820152600c60248201527f77726f6e67207374617475730000000000000000000000000000000000000000604482015290519081900360640190fd5b6116f68282613df9565b611701826000612e0f565b6000828152606860205260408120600601546001600160a01b03169061172a9082908190613f23565b505050565b607160209081526000938452604080852082529284528284209052825290208054600182015460029092015490919083565b608254604080516001600160a01b03878116602483015286811660448084019190915283518084039091018152606490920183526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f4a7702bb000000000000000000000000000000000000000000000000000000001781529251825160009592909216938693928291908083835b6020831061183057805182527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe090920191602091820191016117f3565b6001836020036101000a03801982511681845116808217855250505050505090500191505060006040518083038160008787f1925050503d8060008114611893576040519150601f19603f3d011682016040523d82523d6000602084013e611898565b606091505b5050905080806118a6575082155b6118f7576040805162461bcd60e51b815260206004820152601b60248201527f676f7620766f746573207265636f756e74696e67206661696c65640000000000604482015290519081900360640190fd5b5050505050565b606d5481565b61190c612a89565b61195d576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b608380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b607760205280600052604060002060009150905080600701549080600801549080600901549080600a01549080600b01549080600c01549080600d0154905087565b6083546001600160a01b03163314611a38576040805162461bcd60e51b815260206004820152601260248201527f6e6f74207346544d2066696e616c697a65720000000000000000000000000000604482015290519081900360640190fd5b611a42838361388d565b5060008111611a98576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b607b54604080517f63d2bde50000000000000000000000000000000000000000000000000000000081523360048201526001600160a01b0386811660248301526044820186905260648201859052915191909216916363d2bde591608480830192600092919082900301818387803b158015611b1357600080fd5b505af1158015611b27573d6000803e3d6000fd5b505050506001600160a01b0383166000908152607260209081526040808320858452909152902054811115611ba3576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b6000611baf848461123b565b905080821015611c3e576001600160a01b038416600090815260736020908152604080832086845290915290208054611bf09083850363ffffffff61354616565b815560408051838503815260006020820152815186926001600160a01b038916927fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9929081900390910190a3505b611c4b8484846001613b89565b611c56836000612e0f565b64ffffffffff83856001600160a01b03167fd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57856040518082815260200191505060405180910390a4604051600090339084908381818185875af1925050503d8060008114611ce0576040519150601f19603f3d011682016040523d82523d6000602084013e611ce5565b606091505b5050905080611d3b576040805162461bcd60e51b815260206004820152601260248201527f4661696c656420746f2073656e642046544d0000000000000000000000000000604482015290519081900360640190fd5b64ffffffffff84866001600160a01b03167f75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21866040518082815260200191505060405180910390a45050505050565b33611d93615024565b506001600160a01b038116600090815260716020908152604080832086845282528083208584528252918290208251606081018452815480825260018301549382019390935260029091015492810192909252611e37576040805162461bcd60e51b815260206004820152601560248201527f7265717565737420646f65736e27742065786973740000000000000000000000604482015290519081900360640190fd5b611e4182856137d2565b611e92576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b60208082015182516000878152606890935260409092206001015490919015801590611ece575060008681526068602052604090206001015482115b15611eef575050600084815260686020526040902060018101546002909101545b608160009054906101000a90046001600160a01b03166001600160a01b031663b82b84276040518163ffffffff1660e01b815260040160206040518083038186803b158015611f3d57600080fd5b505afa158015611f51573d6000803e3d6000fd5b505050506040513d6020811015611f6757600080fd5b50518201611f736140cd565b1015611fc6576040805162461bcd60e51b815260206004820152601660248201527f6e6f7420656e6f7567682074696d652070617373656400000000000000000000604482015290519081900360640190fd5b608160009054906101000a90046001600160a01b03166001600160a01b031663650acd666040518163ffffffff1660e01b815260040160206040518083038186803b15801561201457600080fd5b505afa158015612028573d6000803e3d6000fd5b505050506040513d602081101561203e57600080fd5b5051810161204a61293d565b101561209d576040805162461bcd60e51b815260206004820152601860248201527f6e6f7420656e6f7567682065706f636873207061737365640000000000000000604482015290519081900360640190fd5b6001600160a01b03841660009081526071602090815260408083208984528252808320888452909152812060020154906120d688612dd1565b905060006120f88383607a60008d8152602001908152602001600020546140d1565b6001600160a01b03881660009081526071602090815260408083208d845282528083208c845290915281208181556001810182905560020155606e805482019055905080831161218f576040805162461bcd60e51b815260206004820152601660248201527f7374616b652069732066756c6c7920736c617368656400000000000000000000604482015290519081900360640190fd5b60006001600160a01b0388166121ab858463ffffffff61354616565b604051600081818185875af1925050503d80600081146121e7576040519150601f19603f3d011682016040523d82523d6000602084013e6121ec565b606091505b5050905080612242576040805162461bcd60e51b815260206004820152601260248201527f4661696c656420746f2073656e642046544d0000000000000000000000000000604482015290519081900360640190fd5b61224b82613d8f565b888a896001600160a01b03167f75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21876040518082815260200191505060405180910390a450505050505050505050565b6122a2612a89565b6122f3576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6122fc82612dd1565b61234d576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f722069736e277420736c6173686564000000000000000000604482015290519081900360640190fd5b612355614133565b8111156123935760405162461bcd60e51b815260040180806020018281038252602181526020018061514e6021913960400191505060405180910390fd5b6000828152607a60209081526040918290208390558151838152915184927f047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb91792908290030190a25050565b336123e9818561388d565b506000821161243f576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b612449818561123b565b82111561249d576040805162461bcd60e51b815260206004820152601960248201527f6e6f7420656e6f75676820756e6c6f636b6564207374616b6500000000000000604482015290519081900360640190fd5b6124a781856137d2565b6124f8576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607160209081526040808320878452825280832086845290915290206002015415612579576040805162461bcd60e51b815260206004820152601360248201527f7772494420616c72656164792065786973747300000000000000000000000000604482015290519081900360640190fd5b6125868185846001613b89565b6001600160a01b0381166000908152607160209081526040808320878452825280832086845290915290206002018290556125bf61293d565b6001600160a01b038216600090815260716020908152604080832088845282528083208784529091529020556125f36140cd565b6001600160a01b03821660009081526071602090815260408083208884528252808320878452909152812060010191909155612630908590612e0f565b8284826001600160a01b03167fd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57856040518082815260200191505060405180910390a450505050565b61268233613588565b6126bd5760405162461bcd60e51b81526004018080602001828103825260298152602001806151046029913960400191505060405180910390fd5b612705898989898080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b92508a9150899050888861413f565b606b5488111561142157606b889055505050505050505050565b6000818152606860209081526040808320600601546001600160a01b03168352607282528083208484529091529020545b919050565b600091825260776020908152604080842092845291905290205490565b606e5481565b6000612782615024565b61278c8484614306565b8051602082015160408301519293506127ae926111259163ffffffff6133e016565b949350505050565b60009182526077602090815260408084209284526001909201905290205490565b60006127e38383613064565b6127ef575060006112be565b506001600160a01b03919091166000908152607360209081526040808320938352929052205490565b6000612822615024565b506001600160a01b0383166000908152606f60209081526040808320858452825291829020825160608101845281548082526001830154938201849052600290920154938101849052926127ae929091611125919063ffffffff6133e016565b61288a612a89565b6128db576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055565b6067546001015b90565b60675481565b606a6020908152600091825260409182902080548351601f60027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff610100600186161502019093169290920491820184900484028101840190945280845290918301828280156129fe5780601f106129d3576101008083540402835291602001916129fe565b820191906000526020600020905b8154815290600101906020018083116129e157829003601f168201915b505050505081565b6082546001600160a01b031681565b606c5481565b612a25828261388d565b612a76576040805162461bcd60e51b815260206004820152601060248201527f6e6f7468696e6720746f20737461736800000000000000000000000000000000604482015290519081900360640190fd5b5050565b6033546001600160a01b031690565b6033546001600160a01b0316331490565b612aa2612a89565b612af3576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b612afc81613d8f565b50565b607360209081526000928352604080842090915290825290208054600182015460028301546003909301549192909184565b612afc33823461343a565b60009182526077602090815260408084209284526005909201905290205490565b608160009054906101000a90046001600160a01b03166001600160a01b031663c5f530af6040518163ffffffff1660e01b815260040160206040518083038186803b158015612bab57600080fd5b505afa158015612bbf573d6000803e3d6000fd5b505050506040513d6020811015612bd557600080fd5b5051341015612c2b576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b80612c7d576040805162461bcd60e51b815260206004820152600c60248201527f656d707479207075626b65790000000000000000000000000000000000000000604482015290519081900360640190fd5b612cbd3383838080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061437492505050565b612a7633606b543461343a565b607060209081526000928352604080842090915290825290205481565b606860205260009081526040902080546001820154600283015460038401546004850154600586015460069096015494959394929391929091906001600160a01b031687565b607460209081526000928352604080842090915290825290208054600182015460029092015490919083565b600081815260776020908152604091829020600601805483518184028101840190945280845260609392830182828015612db257602002820191906000526020600020905b815481526020019060010190808311612d9e575b50505050509050919050565b33612dcb8185858561439f565b50505050565b600090815260686020526040902054608016151590565b607f546001600160a01b031681565b607a6020526000908152604090205481565b606b5481565b612e1882614754565b612e69576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090206003810154905415612e87575060005b606654604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018690526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b158015612ef457600080fd5b505af1158015612f08573d6000803e3d6000fd5b50505050818015612f1857508015155b1561172a576066546000848152606a60205260409081902081517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810187815260248201938452825460027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6001831615610100020190911604604483018190526001600160a01b039095169463242a6e3f948994939091606490910190849080156130095780601f10612fde57610100808354040283529160200191613009565b820191906000526020600020905b815481529060010190602001808311612fec57829003601f168201915b50509350505050600060405180830381600087803b15801561302a57600080fd5b505af115801561303e573d6000803e3d6000fd5b50505050505050565b607260209081526000928352604080842090915290825290205481565b6001600160a01b0382166000908152607360209081526040808320848452909152812060020154158015906130bb57506001600160a01b038316600090815260736020908152604080832085845290915290205415155b80156112bb57506001600160a01b03831660009081526073602090815260408083208584529091529020600201546130f16140cd565b11159392505050565b607e5481565b60009182526077602090815260408084209284526003909201905290205490565b3381613174576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b61317e8185613064565b156131d0576040805162461bcd60e51b815260206004820152601160248201527f616c7265616479206c6f636b6564207570000000000000000000000000000000604482015290519081900360640190fd5b612dcb8185858561439f565b60009182526077602090815260408084209284526002909201905290205490565b60009182526077602090815260408084209284526004909201905290205490565b613226612a89565b613277576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b612afc8161476b565b613288615024565b61329283836137d2565b6132e3576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b6132ed838361388d565b50506001600160a01b0382166000908152606f602090815260408083208484528252808320815160608101835281548082526001830154948201859052600290920154928101839052939261334b926111259163ffffffff6133e016565b90508061339f576040805162461bcd60e51b815260206004820152600c60248201527f7a65726f20726577617264730000000000000000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152606f60209081526040808320868452909152812081815560018101829055600201556133d981613734565b5092915050565b6000828201838110156112bb576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b61344382614754565b613494576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b600082815260686020526040902054156134f5576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b613502838383600161359c565b61350b82614824565b61172a5760405162461bcd60e51b81526004018080602001828103825260298152602001806151976029913960400191505060405180910390fd5b60006112bb83836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f7700008152506148ec565b6066546001600160a01b0390811691161490565b600082116135f1576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6135fb848461388d565b506001600160a01b0384166000908152607260209081526040808320868452909152902054613630908363ffffffff6133e016565b6001600160a01b0385166000908152607260209081526040808320878452825280832093909355606890522060030154613670818463ffffffff6133e016565b600085815260686020526040902060030155606c54613695908463ffffffff6133e016565b606c556000848152606860205260409020546136c257606d546136be908463ffffffff6133e016565b606d555b6136cd848215612e0f565b60408051848152905185916001600160a01b038816917f9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb9181900360200190a36000848152606860205260409020600601546118f79086906001600160a01b031684613f23565b606654604080517f66e7ea0f0000000000000000000000000000000000000000000000000000000081523060048201526024810184905290516001600160a01b03909216916366e7ea0f9160448082019260009290919082900301818387803b1580156137a057600080fd5b505af11580156137b4573d6000803e3d6000fd5b50506076546137cc925090508263ffffffff6133e016565b60765550565b607b546000906001600160a01b03166137ed575060016112be565b607b54604080517f21d585c30000000000000000000000000000000000000000000000000000000081526001600160a01b03868116600483015260248201869052915191909216916321d585c3916044808301926020929190829003018186803b15801561385a57600080fd5b505afa15801561386e573d6000803e3d6000fd5b505050506040513d602081101561388457600080fd5b50519392505050565b6000613897615024565b6138a18484614983565b90506138ac83614abb565b6001600160a01b0385166000818152607060209081526040808320888452825280832094909455918152606f8252828120868252825282902082516060810184528154815260018201549281019290925260020154918101919091526139129082614b16565b6001600160a01b0385166000818152606f602090815260408083208884528252808320855181558583015160018083019190915595820151600291820155938352607482528083208884528252918290208251606081018452815481529481015491850191909152909101549082015261398c9082614b16565b6001600160a01b0385166000908152607460209081526040808320878452825291829020835181559083015160018201559101516002909101556139d08484613064565b613a33576001600160a01b0384166000818152607360209081526040808320878452825280832083815560018082018590556002808301869055600390920185905594845260748352818420888552909252822082815592830182905591909101555b6020810151151580613a455750805115155b806127ae57506040015115159392505050565b6001600160a01b03841660009081526074602090815260408083208684529091528120548190613aa0908490613a94908763ffffffff614b8816565b9063ffffffff614be116565b6001600160a01b038716600090815260746020908152604080832089845290915281206001015491925090613ae1908590613a94908863ffffffff614b8816565b6001600160a01b03881660009081526074602090815260408083208a845290915290205490915060028204830190613b199084613546565b6001600160a01b03891660009081526074602090815260408083208b8452909152902090815560010154613b4d9083613546565b6001600160a01b03891660009081526074602090815260408083208b8452909152902060010155858110613b7e5750845b979650505050505050565b6001600160a01b03841660009081526072602090815260408083208684528252808320805486900390556068909152902060030154613bce908363ffffffff61354616565b600084815260686020526040902060030155606c54613bf3908363ffffffff61354616565b606c55600083815260686020526040902054613c2057606d54613c1c908363ffffffff61354616565b606d555b6000613c2b8461271f565b90508015613d5d57600084815260686020526040902054613d5857608160009054906101000a90046001600160a01b03166001600160a01b031663c5f530af6040518163ffffffff1660e01b815260040160206040518083038186803b158015613c9457600080fd5b505afa158015613ca8573d6000803e3d6000fd5b505050506040513d6020811015613cbe57600080fd5b5051811015613d14576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b613d1d84614824565b613d585760405162461bcd60e51b81526004018080602001828103825260298152602001806151976029913960400191505060405180910390fd5b613d68565b613d68846001613df9565b6000848152606860205260409020600601546118f79086906001600160a01b031684613f23565b8015612afc5760405160009082156108fc0290839083818181858288f19350505050158015613dc2573d6000803e3d6000fd5b506040805182815290517f8918bd6046d08b314e457977f29562c5d76a7030d79b1edba66e8a5da0b77ae89181900360200190a150565b600082815260686020526040902054158015613e1457508015155b15613e4157600082815260686020526040902060030154606d54613e3d9163ffffffff61354616565b606d555b600082815260686020526040902054811115612a7657600082815260686020526040902081815560020154613ee957613e7861293d565b600083815260686020526040902060020155613e926140cd565b6000838152606860209081526040918290206001810184905560020154825190815290810192909252805184927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e4792908290030190a25b60408051828152905183917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25050565b6082546001600160a01b03161561172a57608254604080516001600160a01b03868116602483015285811660448084019190915283518084039091018152606490920183526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f4a7702bb00000000000000000000000000000000000000000000000000000000178152925182516000959290921693627a120093928291908083835b6020831061400657805182527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe09092019160209182019101613fc9565b6001836020036101000a03801982511681845116808217855250505050505090500191505060006040518083038160008787f1925050503d8060008114614069576040519150601f19603f3d011682016040523d82523d6000602084013e61406e565b606091505b50509050808061407c575081155b612dcb576040805162461bcd60e51b815260206004820152601b60248201527f676f7620766f746573207265636f756e74696e67206661696c65640000000000604482015290519081900360640190fd5b4290565b60008215806140e757506140e3614133565b8210155b156140f45750600061412c565b61411f6001611125614104614133565b613a9486614110614133565b8a91900363ffffffff614b8816565b90508381111561412c5750825b9392505050565b670de0b6b3a764000090565b6001600160a01b038816600090815260696020526040902054156141aa576040805162461bcd60e51b815260206004820152601860248201527f76616c696461746f7220616c7265616479206578697374730000000000000000604482015290519081900360640190fd5b6001600160a01b03881660008181526069602090815260408083208b90558a8352606882528083208981556004810189905560058101889055600181018690556002810187905560060180547fffffffffffffffffffffffff000000000000000000000000000000000000000016909417909355606a8152919020875161423392890190615045565b50876001600160a01b0316877f49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf8686604051808381526020018281526020019250505060405180910390a381156142bf576040805183815260208101839052815189927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47928290030190a25b84156142fc5760408051868152905188917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25b5050505050505050565b61430e615024565b614316615024565b6143208484614983565b6001600160a01b0385166000908152606f6020908152604080832087845282529182902082516060810184528154815260018201549281019290925260020154918101919091529091506127ae9082614b16565b606b80546001019081905561172a838284600061438f61293d565b6143976140cd565b60008061413f565b6143a9848461123b565b8111156143fd576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b6000838152606860205260409020541561445e576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b608160009054906101000a90046001600160a01b03166001600160a01b0316630d7b26096040518163ffffffff1660e01b815260040160206040518083038186803b1580156144ac57600080fd5b505afa1580156144c0573d6000803e3d6000fd5b505050506040513d60208110156144d657600080fd5b505182108015906145605750608160009054906101000a90046001600160a01b03166001600160a01b0316630d4955e36040518163ffffffff1660e01b815260040160206040518083038186803b15801561453057600080fd5b505afa158015614544573d6000803e3d6000fd5b505050506040513d602081101561455a57600080fd5b50518211155b6145b1576040805162461bcd60e51b815260206004820152601260248201527f696e636f7272656374206475726174696f6e0000000000000000000000000000604482015290519081900360640190fd5b60006145bf836111256140cd565b6000858152606860205260409020600601549091506001600160a01b03908116908616811461464d576001600160a01b038116600090815260736020908152604080832088845290915290206002015482111561464d5760405162461bcd60e51b815260040180806020018281038252602881526020018061516f6028913960400191505060405180910390fd5b614657868661388d565b506001600160a01b0386166000908152607360209081526040808320888452909152902060038101548510156146d4576040805162461bcd60e51b815260206004820152601f60248201527f6c6f636b7570206475726174696f6e2063616e6e6f7420646563726561736500604482015290519081900360640190fd5b80546146e6908563ffffffff6133e016565b81556146f061293d565b600182015560028101839055600381018590556040805186815260208101869052815188926001600160a01b038b16927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78929081900390910190a350505050505050565b600090815260686020526040902060050154151590565b6001600160a01b0381166147b05760405162461bcd60e51b81526004018080602001828103825260268152602001806150de6026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b60006148d1614831614133565b608154604080517f2265f2840000000000000000000000000000000000000000000000000000000081529051613a94926001600160a01b031691632265f284916004808301926020929190829003018186803b15801561489057600080fd5b505afa1580156148a4573d6000803e3d6000fd5b505050506040513d60208110156148ba57600080fd5b50516148c58661271f565b9063ffffffff614b8816565b60008381526068602052604090206003015411159050919050565b6000818484111561497b5760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b83811015614940578181015183820152602001614928565b50505050905090810190601f16801561496d5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b61498b615024565b6001600160a01b0383166000908152607060209081526040808320858452909152812054906149b984614abb565b905060006149c78686614c23565b9050818111156149d45750805b828110156149df5750815b6001600160a01b03861660008181526073602090815260408083208984528252808320938352607282528083208984529091528120548254909190614a2b90839063ffffffff61354616565b90506000614a3f84600001548a8988614d00565b9050614a49615024565b614a57828660030154614d63565b9050614a65838b8a89614d00565b9150614a6f615024565b614a7a836000614d63565b9050614a88858c898b614d00565b9250614a92615024565b614a9d846000614d63565b9050614aaa838383614f24565b9d9c50505050505050505050505050565b60008181526068602052604081206002015415614b0e576000828152606860205260409020600201546067541015614af65750606754612750565b50600081815260686020526040902060020154612750565b505060675490565b614b1e615024565b6040805160608101909152825184518291614b3f919063ffffffff6133e016565b8152602001614b5f846020015186602001516133e090919063ffffffff16565b8152602001614b7f846040015186604001516133e090919063ffffffff16565b90529392505050565b600082614b97575060006112be565b82820282848281614ba457fe5b04146112bb5760405162461bcd60e51b815260040180806020018281038252602181526020018061512d6021913960400191505060405180910390fd5b60006112bb83836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f000000000000815250614f3f565b6001600160a01b0382166000908152607360209081526040808320848452909152812060010154606754614c58858583614fa4565b15614c665791506112be9050565b614c71858584614fa4565b614c80576000925050506112be565b80821115614c93576000925050506112be565b80821015614cc657600281830104614cac868683614fa4565b15614cbc57806001019250614cc0565b8091505b50614c93565b80614cd6576000925050506112be565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01949350505050565b6000818310614d11575060006127ae565b600083815260776020818152604080842088855260019081018352818520548786529383528185208986520190915290912054613b7e614d4f614133565b613a94896148c5858763ffffffff61354616565b614d6b615024565b60405180606001604052806000815260200160008152602001600081525090506000608160009054906101000a90046001600160a01b03166001600160a01b0316635e2308d26040518163ffffffff1660e01b815260040160206040518083038186803b158015614ddb57600080fd5b505afa158015614def573d6000803e3d6000fd5b505050506040513d6020811015614e0557600080fd5b505190508215614efd57600081614e1a614133565b0390506000614eac608160009054906101000a90046001600160a01b03166001600160a01b0316630d4955e36040518163ffffffff1660e01b815260040160206040518083038186803b158015614e7057600080fd5b505afa158015614e84573d6000803e3d6000fd5b505050506040513d6020811015614e9a57600080fd5b5051613a94848863ffffffff614b8816565b90506000614ecd614ebb614133565b613a948987860163ffffffff614b8816565b9050614eea614eda614133565b613a94898763ffffffff614b8816565b6020860181905290038452506133d99050565b614f18614f08614133565b613a94868463ffffffff614b8816565b60408301525092915050565b614f2c615024565b6127ae614f398585614b16565b83614b16565b60008183614f8e5760405162461bcd60e51b8152602060048201818152835160248401528351909283926044909101919085019080838360008315614940578181015183820152602001614928565b506000838581614f9a57fe5b0495945050505050565b6001600160a01b038316600090815260736020908152604080832085845290915281206001015482108015906127ae57506001600160a01b03841660009081526073602090815260408083208684529091529020600201546150058361500f565b1115949350505050565b60009081526077602052604090206007015490565b60405180606001604052806000815260200160008152602001600081525090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061508657805160ff19168380011785556150b3565b828001600101855582156150b3579182015b828111156150b3578251825591602001919060010190615098565b506150bf9291506150c3565b5090565b61294491905b808211156150bf57600081556001016150c956fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f206164647265737363616c6c6572206973206e6f7420746865204e6f64654472697665724175746820636f6e7472616374536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f776d757374206265206c657373207468616e206f7220657175616c20746f20312e3076616c696461746f72206c6f636b757020706572696f642077696c6c20656e64206561726c69657276616c696461746f7227732064656c65676174696f6e73206c696d69742069732065786365656465646c6f636b6564207374616b652069732067726561746572207468616e207468652077686f6c65207374616b65a265627a7a72315820b767b9bd00adb5f0467097fa60a9e2cfcb624e12e00375239584e1338e314fe964736f6c63430005110032",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = ContractMetaData.Bin

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend)
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
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
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
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
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

// CurrentEpoch is a free data retrieval call binding the contract method 0x76671808.
//
// Solidity: function currentEpoch() view returns(uint256)
func (_Contract *ContractCaller) CurrentEpoch(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "currentEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "currentSealedEpoch")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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

// GetEpochAccumulatedOriginatedTxsFee is a free data retrieval call binding the contract method 0xdc31e1af.
//
// Solidity: function getEpochAccumulatedOriginatedTxsFee(uint256 epoch, uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetEpochAccumulatedOriginatedTxsFee(opts *bind.CallOpts, epoch *big.Int, validatorID *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochAccumulatedOriginatedTxsFee", epoch, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochAccumulatedRewardPerToken", epoch, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochAccumulatedUptime", epoch, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochOfflineBlocks", epoch, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochOfflineTime", epoch, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochReceivedStake", epoch, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochSnapshot", arg0)

	outstruct := new(struct {
		EndTime               *big.Int
		EpochFee              *big.Int
		TotalBaseRewardWeight *big.Int
		TotalTxRewardWeight   *big.Int
		BaseRewardPerSecond   *big.Int
		TotalStake            *big.Int
		TotalSupply           *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.EndTime = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.EpochFee = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.TotalBaseRewardWeight = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.TotalTxRewardWeight = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.BaseRewardPerSecond = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.TotalStake = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.TotalSupply = *abi.ConvertType(out[6], new(*big.Int)).(**big.Int)

	return *outstruct, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getEpochValidatorIDs", epoch)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getLockedStake", delegator, toValidatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getLockupInfo", arg0, arg1)

	outstruct := new(struct {
		LockedStake *big.Int
		FromEpoch   *big.Int
		EndTime     *big.Int
		Duration    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.LockedStake = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.FromEpoch = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.EndTime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Duration = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

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

// GetSelfStake is a free data retrieval call binding the contract method 0x5601fe01.
//
// Solidity: function getSelfStake(uint256 validatorID) view returns(uint256)
func (_Contract *ContractCaller) GetSelfStake(opts *bind.CallOpts, validatorID *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getSelfStake", validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSelfStake is a free data retrieval call binding the contract method 0x5601fe01.
//
// Solidity: function getSelfStake(uint256 validatorID) view returns(uint256)
func (_Contract *ContractSession) GetSelfStake(validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetSelfStake(&_Contract.CallOpts, validatorID)
}

// GetSelfStake is a free data retrieval call binding the contract method 0x5601fe01.
//
// Solidity: function getSelfStake(uint256 validatorID) view returns(uint256)
func (_Contract *ContractCallerSession) GetSelfStake(validatorID *big.Int) (*big.Int, error) {
	return _Contract.Contract.GetSelfStake(&_Contract.CallOpts, validatorID)
}

// GetStake is a free data retrieval call binding the contract method 0xcfd47663.
//
// Solidity: function getStake(address , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) GetStake(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getStake", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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

// GetStashedLockupRewards is a free data retrieval call binding the contract method 0xb810e411.
//
// Solidity: function getStashedLockupRewards(address , uint256 ) view returns(uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)
func (_Contract *ContractCaller) GetStashedLockupRewards(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	LockupExtraReward *big.Int
	LockupBaseReward  *big.Int
	UnlockedReward    *big.Int
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getStashedLockupRewards", arg0, arg1)

	outstruct := new(struct {
		LockupExtraReward *big.Int
		LockupBaseReward  *big.Int
		UnlockedReward    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.LockupExtraReward = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.LockupBaseReward = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.UnlockedReward = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getUnlockedStake", delegator, toValidatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getValidator", arg0)

	outstruct := new(struct {
		Status           *big.Int
		DeactivatedTime  *big.Int
		DeactivatedEpoch *big.Int
		ReceivedStake    *big.Int
		CreatedEpoch     *big.Int
		CreatedTime      *big.Int
		Auth             common.Address
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Status = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.DeactivatedTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.DeactivatedEpoch = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.ReceivedStake = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.CreatedEpoch = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.CreatedTime = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.Auth = *abi.ConvertType(out[6], new(common.Address)).(*common.Address)

	return *outstruct, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getValidatorID", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getValidatorPubkey", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "getWithdrawalRequest", arg0, arg1, arg2)

	outstruct := new(struct {
		Epoch  *big.Int
		Time   *big.Int
		Amount *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Epoch = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Time = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Amount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isLockedUp", delegator, toValidatorID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isOwner")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "isSlashed", validatorID)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "lastValidatorID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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

// MinGasPrice is a free data retrieval call binding the contract method 0xd96ed505.
//
// Solidity: function minGasPrice() view returns(uint256)
func (_Contract *ContractCaller) MinGasPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "minGasPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinGasPrice is a free data retrieval call binding the contract method 0xd96ed505.
//
// Solidity: function minGasPrice() view returns(uint256)
func (_Contract *ContractSession) MinGasPrice() (*big.Int, error) {
	return _Contract.Contract.MinGasPrice(&_Contract.CallOpts)
}

// MinGasPrice is a free data retrieval call binding the contract method 0xd96ed505.
//
// Solidity: function minGasPrice() view returns(uint256)
func (_Contract *ContractCallerSession) MinGasPrice() (*big.Int, error) {
	return _Contract.Contract.MinGasPrice(&_Contract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Contract *ContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "pendingRewards", delegator, toValidatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "rewardsStash", delegator, validatorID)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "slashingRefundRatio", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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

// StakeTokenizerAddress is a free data retrieval call binding the contract method 0x0e559d82.
//
// Solidity: function stakeTokenizerAddress() view returns(address)
func (_Contract *ContractCaller) StakeTokenizerAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "stakeTokenizerAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakeTokenizerAddress is a free data retrieval call binding the contract method 0x0e559d82.
//
// Solidity: function stakeTokenizerAddress() view returns(address)
func (_Contract *ContractSession) StakeTokenizerAddress() (common.Address, error) {
	return _Contract.Contract.StakeTokenizerAddress(&_Contract.CallOpts)
}

// StakeTokenizerAddress is a free data retrieval call binding the contract method 0x0e559d82.
//
// Solidity: function stakeTokenizerAddress() view returns(address)
func (_Contract *ContractCallerSession) StakeTokenizerAddress() (common.Address, error) {
	return _Contract.Contract.StakeTokenizerAddress(&_Contract.CallOpts)
}

// StashedRewardsUntilEpoch is a free data retrieval call binding the contract method 0xa86a056f.
//
// Solidity: function stashedRewardsUntilEpoch(address , uint256 ) view returns(uint256)
func (_Contract *ContractCaller) StashedRewardsUntilEpoch(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "stashedRewardsUntilEpoch", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "totalActiveStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "totalSlashedStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "totalStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

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

// TreasuryAddress is a free data retrieval call binding the contract method 0xc5f956af.
//
// Solidity: function treasuryAddress() view returns(address)
func (_Contract *ContractCaller) TreasuryAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "treasuryAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TreasuryAddress is a free data retrieval call binding the contract method 0xc5f956af.
//
// Solidity: function treasuryAddress() view returns(address)
func (_Contract *ContractSession) TreasuryAddress() (common.Address, error) {
	return _Contract.Contract.TreasuryAddress(&_Contract.CallOpts)
}

// TreasuryAddress is a free data retrieval call binding the contract method 0xc5f956af.
//
// Solidity: function treasuryAddress() view returns(address)
func (_Contract *ContractCallerSession) TreasuryAddress() (common.Address, error) {
	return _Contract.Contract.TreasuryAddress(&_Contract.CallOpts)
}

// VoteBookAddress is a free data retrieval call binding the contract method 0x893675c6.
//
// Solidity: function voteBookAddress() view returns(address)
func (_Contract *ContractCaller) VoteBookAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "voteBookAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VoteBookAddress is a free data retrieval call binding the contract method 0x893675c6.
//
// Solidity: function voteBookAddress() view returns(address)
func (_Contract *ContractSession) VoteBookAddress() (common.Address, error) {
	return _Contract.Contract.VoteBookAddress(&_Contract.CallOpts)
}

// VoteBookAddress is a free data retrieval call binding the contract method 0x893675c6.
//
// Solidity: function voteBookAddress() view returns(address)
func (_Contract *ContractCallerSession) VoteBookAddress() (common.Address, error) {
	return _Contract.Contract.VoteBookAddress(&_Contract.CallOpts)
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

// BurnFTM is a paid mutator transaction binding the contract method 0x90a6c475.
//
// Solidity: function burnFTM(uint256 amount) returns()
func (_Contract *ContractTransactor) BurnFTM(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "burnFTM", amount)
}

// BurnFTM is a paid mutator transaction binding the contract method 0x90a6c475.
//
// Solidity: function burnFTM(uint256 amount) returns()
func (_Contract *ContractSession) BurnFTM(amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.BurnFTM(&_Contract.TransactOpts, amount)
}

// BurnFTM is a paid mutator transaction binding the contract method 0x90a6c475.
//
// Solidity: function burnFTM(uint256 amount) returns()
func (_Contract *ContractTransactorSession) BurnFTM(amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.BurnFTM(&_Contract.TransactOpts, amount)
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

// LiquidateSFTM is a paid mutator transaction binding the contract method 0x3a488397.
//
// Solidity: function liquidateSFTM(address delegator, uint256 toValidatorID, uint256 amount) returns()
func (_Contract *ContractTransactor) LiquidateSFTM(opts *bind.TransactOpts, delegator common.Address, toValidatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "liquidateSFTM", delegator, toValidatorID, amount)
}

// LiquidateSFTM is a paid mutator transaction binding the contract method 0x3a488397.
//
// Solidity: function liquidateSFTM(address delegator, uint256 toValidatorID, uint256 amount) returns()
func (_Contract *ContractSession) LiquidateSFTM(delegator common.Address, toValidatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LiquidateSFTM(&_Contract.TransactOpts, delegator, toValidatorID, amount)
}

// LiquidateSFTM is a paid mutator transaction binding the contract method 0x3a488397.
//
// Solidity: function liquidateSFTM(address delegator, uint256 toValidatorID, uint256 amount) returns()
func (_Contract *ContractTransactorSession) LiquidateSFTM(delegator common.Address, toValidatorID *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.LiquidateSFTM(&_Contract.TransactOpts, delegator, toValidatorID, amount)
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

// RecountVotes is a paid mutator transaction binding the contract method 0x20c0849d.
//
// Solidity: function recountVotes(address delegator, address validatorAuth, bool strict, uint256 gas) returns()
func (_Contract *ContractTransactor) RecountVotes(opts *bind.TransactOpts, delegator common.Address, validatorAuth common.Address, strict bool, gas *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "recountVotes", delegator, validatorAuth, strict, gas)
}

// RecountVotes is a paid mutator transaction binding the contract method 0x20c0849d.
//
// Solidity: function recountVotes(address delegator, address validatorAuth, bool strict, uint256 gas) returns()
func (_Contract *ContractSession) RecountVotes(delegator common.Address, validatorAuth common.Address, strict bool, gas *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RecountVotes(&_Contract.TransactOpts, delegator, validatorAuth, strict, gas)
}

// RecountVotes is a paid mutator transaction binding the contract method 0x20c0849d.
//
// Solidity: function recountVotes(address delegator, address validatorAuth, bool strict, uint256 gas) returns()
func (_Contract *ContractTransactorSession) RecountVotes(delegator common.Address, validatorAuth common.Address, strict bool, gas *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RecountVotes(&_Contract.TransactOpts, delegator, validatorAuth, strict, gas)
}

// RelockStake is a paid mutator transaction binding the contract method 0xbd14d907.
//
// Solidity: function relockStake(uint256 toValidatorID, uint256 lockupDuration, uint256 amount) returns()
func (_Contract *ContractTransactor) RelockStake(opts *bind.TransactOpts, toValidatorID *big.Int, lockupDuration *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "relockStake", toValidatorID, lockupDuration, amount)
}

// RelockStake is a paid mutator transaction binding the contract method 0xbd14d907.
//
// Solidity: function relockStake(uint256 toValidatorID, uint256 lockupDuration, uint256 amount) returns()
func (_Contract *ContractSession) RelockStake(toValidatorID *big.Int, lockupDuration *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RelockStake(&_Contract.TransactOpts, toValidatorID, lockupDuration, amount)
}

// RelockStake is a paid mutator transaction binding the contract method 0xbd14d907.
//
// Solidity: function relockStake(uint256 toValidatorID, uint256 lockupDuration, uint256 amount) returns()
func (_Contract *ContractTransactorSession) RelockStake(toValidatorID *big.Int, lockupDuration *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Contract.Contract.RelockStake(&_Contract.TransactOpts, toValidatorID, lockupDuration, amount)
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

// UpdateSFTMFinalizer is a paid mutator transaction binding the contract method 0x2ce71960.
//
// Solidity: function updateSFTMFinalizer(address v) returns()
func (_Contract *ContractTransactor) UpdateSFTMFinalizer(opts *bind.TransactOpts, v common.Address) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "updateSFTMFinalizer", v)
}

// UpdateSFTMFinalizer is a paid mutator transaction binding the contract method 0x2ce71960.
//
// Solidity: function updateSFTMFinalizer(address v) returns()
func (_Contract *ContractSession) UpdateSFTMFinalizer(v common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpdateSFTMFinalizer(&_Contract.TransactOpts, v)
}

// UpdateSFTMFinalizer is a paid mutator transaction binding the contract method 0x2ce71960.
//
// Solidity: function updateSFTMFinalizer(address v) returns()
func (_Contract *ContractTransactorSession) UpdateSFTMFinalizer(v common.Address) (*types.Transaction, error) {
	return _Contract.Contract.UpdateSFTMFinalizer(&_Contract.TransactOpts, v)
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

// ContractBurntFTMIterator is returned from FilterBurntFTM and is used to iterate over the raw logs and unpacked data for BurntFTM events raised by the Contract contract.
type ContractBurntFTMIterator struct {
	Event *ContractBurntFTM // Event containing the contract specifics and raw log

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
func (it *ContractBurntFTMIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractBurntFTM)
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
		it.Event = new(ContractBurntFTM)
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
func (it *ContractBurntFTMIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractBurntFTMIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractBurntFTM represents a BurntFTM event raised by the Contract contract.
type ContractBurntFTM struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBurntFTM is a free log retrieval operation binding the contract event 0x8918bd6046d08b314e457977f29562c5d76a7030d79b1edba66e8a5da0b77ae8.
//
// Solidity: event BurntFTM(uint256 amount)
func (_Contract *ContractFilterer) FilterBurntFTM(opts *bind.FilterOpts) (*ContractBurntFTMIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "BurntFTM")
	if err != nil {
		return nil, err
	}
	return &ContractBurntFTMIterator{contract: _Contract.contract, event: "BurntFTM", logs: logs, sub: sub}, nil
}

// WatchBurntFTM is a free log subscription operation binding the contract event 0x8918bd6046d08b314e457977f29562c5d76a7030d79b1edba66e8a5da0b77ae8.
//
// Solidity: event BurntFTM(uint256 amount)
func (_Contract *ContractFilterer) WatchBurntFTM(opts *bind.WatchOpts, sink chan<- *ContractBurntFTM) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "BurntFTM")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractBurntFTM)
				if err := _Contract.contract.UnpackLog(event, "BurntFTM", log); err != nil {
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

// ParseBurntFTM is a log parse operation binding the contract event 0x8918bd6046d08b314e457977f29562c5d76a7030d79b1edba66e8a5da0b77ae8.
//
// Solidity: event BurntFTM(uint256 amount)
func (_Contract *ContractFilterer) ParseBurntFTM(log types.Log) (*ContractBurntFTM, error) {
	event := new(ContractBurntFTM)
	if err := _Contract.contract.UnpackLog(event, "BurntFTM", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
	return event, nil
}

// ContractRefundedSlashedLegacyDelegationIterator is returned from FilterRefundedSlashedLegacyDelegation and is used to iterate over the raw logs and unpacked data for RefundedSlashedLegacyDelegation events raised by the Contract contract.
type ContractRefundedSlashedLegacyDelegationIterator struct {
	Event *ContractRefundedSlashedLegacyDelegation // Event containing the contract specifics and raw log

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
func (it *ContractRefundedSlashedLegacyDelegationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractRefundedSlashedLegacyDelegation)
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
		it.Event = new(ContractRefundedSlashedLegacyDelegation)
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
func (it *ContractRefundedSlashedLegacyDelegationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractRefundedSlashedLegacyDelegationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractRefundedSlashedLegacyDelegation represents a RefundedSlashedLegacyDelegation event raised by the Contract contract.
type ContractRefundedSlashedLegacyDelegation struct {
	Delegator   common.Address
	ValidatorID *big.Int
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRefundedSlashedLegacyDelegation is a free log retrieval operation binding the contract event 0x172fdfaf5222519d28d2794b7617be033f46d954f9b6c3896e7d2611ff444252.
//
// Solidity: event RefundedSlashedLegacyDelegation(address indexed delegator, uint256 indexed validatorID, uint256 amount)
func (_Contract *ContractFilterer) FilterRefundedSlashedLegacyDelegation(opts *bind.FilterOpts, delegator []common.Address, validatorID []*big.Int) (*ContractRefundedSlashedLegacyDelegationIterator, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.FilterLogs(opts, "RefundedSlashedLegacyDelegation", delegatorRule, validatorIDRule)
	if err != nil {
		return nil, err
	}
	return &ContractRefundedSlashedLegacyDelegationIterator{contract: _Contract.contract, event: "RefundedSlashedLegacyDelegation", logs: logs, sub: sub}, nil
}

// WatchRefundedSlashedLegacyDelegation is a free log subscription operation binding the contract event 0x172fdfaf5222519d28d2794b7617be033f46d954f9b6c3896e7d2611ff444252.
//
// Solidity: event RefundedSlashedLegacyDelegation(address indexed delegator, uint256 indexed validatorID, uint256 amount)
func (_Contract *ContractFilterer) WatchRefundedSlashedLegacyDelegation(opts *bind.WatchOpts, sink chan<- *ContractRefundedSlashedLegacyDelegation, delegator []common.Address, validatorID []*big.Int) (event.Subscription, error) {

	var delegatorRule []interface{}
	for _, delegatorItem := range delegator {
		delegatorRule = append(delegatorRule, delegatorItem)
	}
	var validatorIDRule []interface{}
	for _, validatorIDItem := range validatorID {
		validatorIDRule = append(validatorIDRule, validatorIDItem)
	}

	logs, sub, err := _Contract.contract.WatchLogs(opts, "RefundedSlashedLegacyDelegation", delegatorRule, validatorIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractRefundedSlashedLegacyDelegation)
				if err := _Contract.contract.UnpackLog(event, "RefundedSlashedLegacyDelegation", log); err != nil {
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

// ParseRefundedSlashedLegacyDelegation is a log parse operation binding the contract event 0x172fdfaf5222519d28d2794b7617be033f46d954f9b6c3896e7d2611ff444252.
//
// Solidity: event RefundedSlashedLegacyDelegation(address indexed delegator, uint256 indexed validatorID, uint256 amount)
func (_Contract *ContractFilterer) ParseRefundedSlashedLegacyDelegation(log types.Log) (*ContractRefundedSlashedLegacyDelegation, error) {
	event := new(ContractRefundedSlashedLegacyDelegation)
	if err := _Contract.contract.UnpackLog(event, "RefundedSlashedLegacyDelegation", log); err != nil {
		return nil, err
	}
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
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
	event.Raw = log
	return event, nil
}

var ContractBinRuntime = "0x6080604052600436106103815760003560e01c8063854873e1116101d1578063b88a37e211610102578063cfd47663116100a0578063de67f2151161006f578063de67f21514610f3a578063df00c92214610f70578063e261641a14610fa0578063f2fde38b14610fd057610381565b8063cfd4766314610e83578063cfdbb7cd14610ebc578063d96ed50514610ef5578063dc31e1af14610f0a57610381565b8063c5f956af116100dc578063c5f956af14610dfd578063c65ee0e114610e12578063c7be95de14610e3c578063cc8343aa14610e5157610381565b8063b88a37e214610d23578063bd14d90714610d9d578063c3de580e14610dd357610381565b806396c7ee461161016f578063a5a470ad11610149578063a5a470ad14610bd6578063a86a056f14610c46578063b5d8962714610c7f578063b810e41114610cea57610381565b806396c7ee4614610b2a5780639fa6dd3514610b89578063a198d22914610ba657610381565b80638cddb015116101ab5780638cddb01514610a895780638da5cb5b14610ac25780638f32d59b14610ad757806390a6c47514610b0057610381565b8063854873e1146109c0578063893675c614610a5f5780638b0e9f3f14610a7457610381565b80633a488397116102b65780635fab23a8116102545780636f498663116102235780636f49866314610948578063715018a61461098157806376671808146109965780637cacb1d6146109ab57610381565b80635fab23a8146108915780636099ecb2146108a657806361e53fcc146108df578063670322f81461090f57610381565b80634f864df4116102905780634f864df4146107565780634feb92f31461078c5780635601fe011461083757806358f95b801461086157610381565b80633a488397146106b7578063441a3e70146106f65780634f7c4efb1461072657610381565b80631d3ac42c1161032357806320c0849d116102fd57806320c0849d146105c257806328f731481461060d5780632ce719601461062257806339b80c001461065557610381565b80631d3ac42c146105055780631e702f83146105355780631f2701521461056557610381565b80630e559d821161035f5780630e559d821461042157806312622d0e1461045257806318160ddd1461048b57806318f628d4146104a057610381565b80630135b1db1461038657806308c36874146103cb5780630962ef79146103f7575b600080fd5b34801561039257600080fd5b506103b9600480360360208110156103a957600080fd5b50356001600160a01b0316611003565b60408051918252519081900360200190f35b3480156103d757600080fd5b506103f5600480360360208110156103ee57600080fd5b5035611015565b005b34801561040357600080fd5b506103f56004803603602081101561041a57600080fd5b50356110e1565b34801561042d57600080fd5b5061043661122c565b604080516001600160a01b039092168252519081900360200190f35b34801561045e57600080fd5b506103b96004803603604081101561047557600080fd5b506001600160a01b03813516906020013561123b565b34801561049757600080fd5b506103b96112c4565b3480156104ac57600080fd5b506103f560048036036101208110156104c457600080fd5b506001600160a01b038135169060208101359060408101359060608101359060808101359060a08101359060c08101359060e08101359061010001356112ca565b34801561051157600080fd5b506103b96004803603604081101561052857600080fd5b508035906020013561142c565b34801561054157600080fd5b506103f56004803603604081101561055857600080fd5b5080359060200135611656565b34801561057157600080fd5b506105a46004803603606081101561058857600080fd5b506001600160a01b03813516906020810135906040013561172f565b60408051938452602084019290925282820152519081900360600190f35b3480156105ce57600080fd5b506103f5600480360360808110156105e557600080fd5b506001600160a01b038135811691602081013590911690604081013515159060600135611761565b34801561061957600080fd5b506103b96118fe565b34801561062e57600080fd5b506103f56004803603602081101561064557600080fd5b50356001600160a01b0316611904565b34801561066157600080fd5b5061067f6004803603602081101561067857600080fd5b5035611997565b604080519788526020880196909652868601949094526060860192909252608085015260a084015260c0830152519081900360e00190f35b3480156106c357600080fd5b506103f5600480360360608110156106da57600080fd5b506001600160a01b0381351690602081013590604001356119d9565b34801561070257600080fd5b506103f56004803603604081101561071957600080fd5b5080359060200135611d8a565b34801561073257600080fd5b506103f56004803603604081101561074957600080fd5b508035906020013561229a565b34801561076257600080fd5b506103f56004803603606081101561077957600080fd5b50803590602081013590604001356123de565b34801561079857600080fd5b506103f560048036036101008110156107b057600080fd5b6001600160a01b03823516916020810135918101906060810160408201356401000000008111156107e057600080fd5b8201836020820111156107f257600080fd5b8035906020019184600183028401116401000000008311171561081457600080fd5b919350915080359060208101359060408101359060608101359060800135612679565b34801561084357600080fd5b506103b96004803603602081101561085a57600080fd5b503561271f565b34801561086d57600080fd5b506103b96004803603604081101561088457600080fd5b5080359060200135612755565b34801561089d57600080fd5b506103b9612772565b3480156108b257600080fd5b506103b9600480360360408110156108c957600080fd5b506001600160a01b038135169060200135612778565b3480156108eb57600080fd5b506103b96004803603604081101561090257600080fd5b50803590602001356127b6565b34801561091b57600080fd5b506103b96004803603604081101561093257600080fd5b506001600160a01b0381351690602001356127d7565b34801561095457600080fd5b506103b96004803603604081101561096b57600080fd5b506001600160a01b038135169060200135612818565b34801561098d57600080fd5b506103f5612882565b3480156109a257600080fd5b506103b961293d565b3480156109b757600080fd5b506103b9612947565b3480156109cc57600080fd5b506109ea600480360360208110156109e357600080fd5b503561294d565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610a24578181015183820152602001610a0c565b50505050905090810190601f168015610a515780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b348015610a6b57600080fd5b50610436612a06565b348015610a8057600080fd5b506103b9612a15565b348015610a9557600080fd5b506103f560048036036040811015610aac57600080fd5b506001600160a01b038135169060200135612a1b565b348015610ace57600080fd5b50610436612a7a565b348015610ae357600080fd5b50610aec612a89565b604080519115158252519081900360200190f35b348015610b0c57600080fd5b506103f560048036036020811015610b2357600080fd5b5035612a9a565b348015610b3657600080fd5b50610b6360048036036040811015610b4d57600080fd5b506001600160a01b038135169060200135612aff565b604080519485526020850193909352838301919091526060830152519081900360800190f35b6103f560048036036020811015610b9f57600080fd5b5035612b31565b348015610bb257600080fd5b506103b960048036036040811015610bc957600080fd5b5080359060200135612b3c565b6103f560048036036020811015610bec57600080fd5b810190602081018135640100000000811115610c0757600080fd5b820183602082011115610c1957600080fd5b80359060200191846001830284011164010000000083111715610c3b57600080fd5b509092509050612b5d565b348015610c5257600080fd5b506103b960048036036040811015610c6957600080fd5b506001600160a01b038135169060200135612cca565b348015610c8b57600080fd5b50610ca960048036036020811015610ca257600080fd5b5035612ce7565b604080519788526020880196909652868601949094526060860192909252608085015260a08401526001600160a01b031660c0830152519081900360e00190f35b348015610cf657600080fd5b506105a460048036036040811015610d0d57600080fd5b506001600160a01b038135169060200135612d2d565b348015610d2f57600080fd5b50610d4d60048036036020811015610d4657600080fd5b5035612d59565b60408051602080825283518183015283519192839290830191858101910280838360005b83811015610d89578181015183820152602001610d71565b505050509050019250505060405180910390f35b348015610da957600080fd5b506103f560048036036060811015610dc057600080fd5b5080359060208101359060400135612dbe565b348015610ddf57600080fd5b50610aec60048036036020811015610df657600080fd5b5035612dd1565b348015610e0957600080fd5b50610436612de8565b348015610e1e57600080fd5b506103b960048036036020811015610e3557600080fd5b5035612df7565b348015610e4857600080fd5b506103b9612e09565b348015610e5d57600080fd5b506103f560048036036040811015610e7457600080fd5b50803590602001351515612e0f565b348015610e8f57600080fd5b506103b960048036036040811015610ea657600080fd5b506001600160a01b038135169060200135613047565b348015610ec857600080fd5b50610aec60048036036040811015610edf57600080fd5b506001600160a01b038135169060200135613064565b348015610f0157600080fd5b506103b96130fa565b348015610f1657600080fd5b506103b960048036036040811015610f2d57600080fd5b5080359060200135613100565b348015610f4657600080fd5b506103f560048036036060811015610f5d57600080fd5b5080359060208101359060400135613121565b348015610f7c57600080fd5b506103b960048036036040811015610f9357600080fd5b50803590602001356131dc565b348015610fac57600080fd5b506103b960048036036040811015610fc357600080fd5b50803590602001356131fd565b348015610fdc57600080fd5b506103f560048036036020811015610ff357600080fd5b50356001600160a01b031661321e565b60696020526000908152604090205481565b3361101e615024565b6110288284613280565b602081015181519192506000916110449163ffffffff6133e016565b905061106783856110628560400151856133e090919063ffffffff16565b61343a565b6001600160a01b0383166000818152607360209081526040808320888452825291829020805485019055845185820151868401518451928352928201528083019190915290518692917f4119153d17a36f9597d40e3ab4148d03261a439dddbec4e91799ab7159608e26919081900360600190a350505050565b336110ea615024565b6110f48284613280565b90506000826001600160a01b03166111318360400151611125856020015186600001516133e090919063ffffffff16565b9063ffffffff6133e016565b604051600081818185875af1925050503d806000811461116d576040519150601f19603f3d011682016040523d82523d6000602084013e611172565b606091505b50509050806111c8576040805162461bcd60e51b815260206004820152601260248201527f4661696c656420746f2073656e642046544d0000000000000000000000000000604482015290519081900360640190fd5b83836001600160a01b03167fc1d8eb6e444b89fb8ff0991c19311c070df704ccb009e210d1462d5b2410bf4584600001518560200151866040015160405180848152602001838152602001828152602001935050505060405180910390a350505050565b607b546001600160a01b031681565b60006112478383613064565b61127557506001600160a01b03821660009081526072602090815260408083208484529091529020546112be565b6001600160a01b0383166000818152607360209081526040808320868452825280832054938352607282528083208684529091529020546112bb9163ffffffff61354616565b90505b92915050565b60765481565b6112d333613588565b61130e5760405162461bcd60e51b81526004018080602001828103825260298152602001806151046029913960400191505060405180910390fd5b61131b898989600061359c565b6001600160a01b0389166000908152606f602090815260408083208b8452909152902060020181905561134d87613734565b851561142157868611156113925760405162461bcd60e51b815260040180806020018281038252602c8152602001806151c0602c913960400191505060405180910390fd5b6001600160a01b03891660008181526073602090815260408083208c845282528083208a8155600181018a90556002810189905560038101889055848452607483528184208d855283529281902086905580518781529182018a9052805192938c9390927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e7892908290030190a3505b505050505050505050565b33600081815260736020908152604080832086845290915281209091908361149b576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6114a58286613064565b6114f6576040805162461bcd60e51b815260206004820152600d60248201527f6e6f74206c6f636b656420757000000000000000000000000000000000000000604482015290519081900360640190fd5b805484111561154c576040805162461bcd60e51b815260206004820152601760248201527f6e6f7420656e6f756768206c6f636b6564207374616b65000000000000000000604482015290519081900360640190fd5b61155682866137d2565b6115a7576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b6115b1828661388d565b5060006115c48387878560000154613a58565b905081600301546363401ec501826002015410156115e0575060005b815485900382558015611603576115fa8387836001613b89565b61160381613d8f565b85836001600160a01b03167fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b98784604051808381526020018281526020019250505060405180910390a395945050505050565b61165f33613588565b61169a5760405162461bcd60e51b81526004018080602001828103825260298152602001806151046029913960400191505060405180910390fd5b806116ec576040805162461bcd60e51b815260206004820152600c60248201527f77726f6e67207374617475730000000000000000000000000000000000000000604482015290519081900360640190fd5b6116f68282613df9565b611701826000612e0f565b6000828152606860205260408120600601546001600160a01b03169061172a9082908190613f23565b505050565b607160209081526000938452604080852082529284528284209052825290208054600182015460029092015490919083565b608254604080516001600160a01b03878116602483015286811660448084019190915283518084039091018152606490920183526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f4a7702bb000000000000000000000000000000000000000000000000000000001781529251825160009592909216938693928291908083835b6020831061183057805182527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe090920191602091820191016117f3565b6001836020036101000a03801982511681845116808217855250505050505090500191505060006040518083038160008787f1925050503d8060008114611893576040519150601f19603f3d011682016040523d82523d6000602084013e611898565b606091505b5050905080806118a6575082155b6118f7576040805162461bcd60e51b815260206004820152601b60248201527f676f7620766f746573207265636f756e74696e67206661696c65640000000000604482015290519081900360640190fd5b5050505050565b606d5481565b61190c612a89565b61195d576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b608380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b607760205280600052604060002060009150905080600701549080600801549080600901549080600a01549080600b01549080600c01549080600d0154905087565b6083546001600160a01b03163314611a38576040805162461bcd60e51b815260206004820152601260248201527f6e6f74207346544d2066696e616c697a65720000000000000000000000000000604482015290519081900360640190fd5b611a42838361388d565b5060008111611a98576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b607b54604080517f63d2bde50000000000000000000000000000000000000000000000000000000081523360048201526001600160a01b0386811660248301526044820186905260648201859052915191909216916363d2bde591608480830192600092919082900301818387803b158015611b1357600080fd5b505af1158015611b27573d6000803e3d6000fd5b505050506001600160a01b0383166000908152607260209081526040808320858452909152902054811115611ba3576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b6000611baf848461123b565b905080821015611c3e576001600160a01b038416600090815260736020908152604080832086845290915290208054611bf09083850363ffffffff61354616565b815560408051838503815260006020820152815186926001600160a01b038916927fef6c0c14fe9aa51af36acd791464dec3badbde668b63189b47bfa4e25be9b2b9929081900390910190a3505b611c4b8484846001613b89565b611c56836000612e0f565b64ffffffffff83856001600160a01b03167fd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57856040518082815260200191505060405180910390a4604051600090339084908381818185875af1925050503d8060008114611ce0576040519150601f19603f3d011682016040523d82523d6000602084013e611ce5565b606091505b5050905080611d3b576040805162461bcd60e51b815260206004820152601260248201527f4661696c656420746f2073656e642046544d0000000000000000000000000000604482015290519081900360640190fd5b64ffffffffff84866001600160a01b03167f75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21866040518082815260200191505060405180910390a45050505050565b33611d93615024565b506001600160a01b038116600090815260716020908152604080832086845282528083208584528252918290208251606081018452815480825260018301549382019390935260029091015492810192909252611e37576040805162461bcd60e51b815260206004820152601560248201527f7265717565737420646f65736e27742065786973740000000000000000000000604482015290519081900360640190fd5b611e4182856137d2565b611e92576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b60208082015182516000878152606890935260409092206001015490919015801590611ece575060008681526068602052604090206001015482115b15611eef575050600084815260686020526040902060018101546002909101545b608160009054906101000a90046001600160a01b03166001600160a01b031663b82b84276040518163ffffffff1660e01b815260040160206040518083038186803b158015611f3d57600080fd5b505afa158015611f51573d6000803e3d6000fd5b505050506040513d6020811015611f6757600080fd5b50518201611f736140cd565b1015611fc6576040805162461bcd60e51b815260206004820152601660248201527f6e6f7420656e6f7567682074696d652070617373656400000000000000000000604482015290519081900360640190fd5b608160009054906101000a90046001600160a01b03166001600160a01b031663650acd666040518163ffffffff1660e01b815260040160206040518083038186803b15801561201457600080fd5b505afa158015612028573d6000803e3d6000fd5b505050506040513d602081101561203e57600080fd5b5051810161204a61293d565b101561209d576040805162461bcd60e51b815260206004820152601860248201527f6e6f7420656e6f7567682065706f636873207061737365640000000000000000604482015290519081900360640190fd5b6001600160a01b03841660009081526071602090815260408083208984528252808320888452909152812060020154906120d688612dd1565b905060006120f88383607a60008d8152602001908152602001600020546140d1565b6001600160a01b03881660009081526071602090815260408083208d845282528083208c845290915281208181556001810182905560020155606e805482019055905080831161218f576040805162461bcd60e51b815260206004820152601660248201527f7374616b652069732066756c6c7920736c617368656400000000000000000000604482015290519081900360640190fd5b60006001600160a01b0388166121ab858463ffffffff61354616565b604051600081818185875af1925050503d80600081146121e7576040519150601f19603f3d011682016040523d82523d6000602084013e6121ec565b606091505b5050905080612242576040805162461bcd60e51b815260206004820152601260248201527f4661696c656420746f2073656e642046544d0000000000000000000000000000604482015290519081900360640190fd5b61224b82613d8f565b888a896001600160a01b03167f75e161b3e824b114fc1a33274bd7091918dd4e639cede50b78b15a4eea956a21876040518082815260200191505060405180910390a450505050505050505050565b6122a2612a89565b6122f3576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6122fc82612dd1565b61234d576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f722069736e277420736c6173686564000000000000000000604482015290519081900360640190fd5b612355614133565b8111156123935760405162461bcd60e51b815260040180806020018281038252602181526020018061514e6021913960400191505060405180910390fd5b6000828152607a60209081526040918290208390558151838152915184927f047575f43f09a7a093d94ec483064acfc61b7e25c0de28017da442abf99cb91792908290030190a25050565b336123e9818561388d565b506000821161243f576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b612449818561123b565b82111561249d576040805162461bcd60e51b815260206004820152601960248201527f6e6f7420656e6f75676820756e6c6f636b6564207374616b6500000000000000604482015290519081900360640190fd5b6124a781856137d2565b6124f8576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b6001600160a01b0381166000908152607160209081526040808320878452825280832086845290915290206002015415612579576040805162461bcd60e51b815260206004820152601360248201527f7772494420616c72656164792065786973747300000000000000000000000000604482015290519081900360640190fd5b6125868185846001613b89565b6001600160a01b0381166000908152607160209081526040808320878452825280832086845290915290206002018290556125bf61293d565b6001600160a01b038216600090815260716020908152604080832088845282528083208784529091529020556125f36140cd565b6001600160a01b03821660009081526071602090815260408083208884528252808320878452909152812060010191909155612630908590612e0f565b8284826001600160a01b03167fd3bb4e423fbea695d16b982f9f682dc5f35152e5411646a8a5a79a6b02ba8d57856040518082815260200191505060405180910390a450505050565b61268233613588565b6126bd5760405162461bcd60e51b81526004018080602001828103825260298152602001806151046029913960400191505060405180910390fd5b612705898989898080601f0160208091040260200160405190810160405280939291908181526020018383808284376000920191909152508b92508a9150899050888861413f565b606b5488111561142157606b889055505050505050505050565b6000818152606860209081526040808320600601546001600160a01b03168352607282528083208484529091529020545b919050565b600091825260776020908152604080842092845291905290205490565b606e5481565b6000612782615024565b61278c8484614306565b8051602082015160408301519293506127ae926111259163ffffffff6133e016565b949350505050565b60009182526077602090815260408084209284526001909201905290205490565b60006127e38383613064565b6127ef575060006112be565b506001600160a01b03919091166000908152607360209081526040808320938352929052205490565b6000612822615024565b506001600160a01b0383166000908152606f60209081526040808320858452825291829020825160608101845281548082526001830154938201849052600290920154938101849052926127ae929091611125919063ffffffff6133e016565b61288a612a89565b6128db576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b6033546040516000916001600160a01b0316907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0908390a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000169055565b6067546001015b90565b60675481565b606a6020908152600091825260409182902080548351601f60027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff610100600186161502019093169290920491820184900484028101840190945280845290918301828280156129fe5780601f106129d3576101008083540402835291602001916129fe565b820191906000526020600020905b8154815290600101906020018083116129e157829003601f168201915b505050505081565b6082546001600160a01b031681565b606c5481565b612a25828261388d565b612a76576040805162461bcd60e51b815260206004820152601060248201527f6e6f7468696e6720746f20737461736800000000000000000000000000000000604482015290519081900360640190fd5b5050565b6033546001600160a01b031690565b6033546001600160a01b0316331490565b612aa2612a89565b612af3576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b612afc81613d8f565b50565b607360209081526000928352604080842090915290825290208054600182015460028301546003909301549192909184565b612afc33823461343a565b60009182526077602090815260408084209284526005909201905290205490565b608160009054906101000a90046001600160a01b03166001600160a01b031663c5f530af6040518163ffffffff1660e01b815260040160206040518083038186803b158015612bab57600080fd5b505afa158015612bbf573d6000803e3d6000fd5b505050506040513d6020811015612bd557600080fd5b5051341015612c2b576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b80612c7d576040805162461bcd60e51b815260206004820152600c60248201527f656d707479207075626b65790000000000000000000000000000000000000000604482015290519081900360640190fd5b612cbd3383838080601f01602080910402602001604051908101604052809392919081815260200183838082843760009201919091525061437492505050565b612a7633606b543461343a565b607060209081526000928352604080842090915290825290205481565b606860205260009081526040902080546001820154600283015460038401546004850154600586015460069096015494959394929391929091906001600160a01b031687565b607460209081526000928352604080842090915290825290208054600182015460029092015490919083565b600081815260776020908152604091829020600601805483518184028101840190945280845260609392830182828015612db257602002820191906000526020600020905b815481526020019060010190808311612d9e575b50505050509050919050565b33612dcb8185858561439f565b50505050565b600090815260686020526040902054608016151590565b607f546001600160a01b031681565b607a6020526000908152604090205481565b606b5481565b612e1882614754565b612e69576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b60008281526068602052604090206003810154905415612e87575060005b606654604080517fa4066fbe000000000000000000000000000000000000000000000000000000008152600481018690526024810184905290516001600160a01b039092169163a4066fbe9160448082019260009290919082900301818387803b158015612ef457600080fd5b505af1158015612f08573d6000803e3d6000fd5b50505050818015612f1857508015155b1561172a576066546000848152606a60205260409081902081517f242a6e3f0000000000000000000000000000000000000000000000000000000081526004810187815260248201938452825460027fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6001831615610100020190911604604483018190526001600160a01b039095169463242a6e3f948994939091606490910190849080156130095780601f10612fde57610100808354040283529160200191613009565b820191906000526020600020905b815481529060010190602001808311612fec57829003601f168201915b50509350505050600060405180830381600087803b15801561302a57600080fd5b505af115801561303e573d6000803e3d6000fd5b50505050505050565b607260209081526000928352604080842090915290825290205481565b6001600160a01b0382166000908152607360209081526040808320848452909152812060020154158015906130bb57506001600160a01b038316600090815260736020908152604080832085845290915290205415155b80156112bb57506001600160a01b03831660009081526073602090815260408083208584529091529020600201546130f16140cd565b11159392505050565b607e5481565b60009182526077602090815260408084209284526003909201905290205490565b3381613174576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b61317e8185613064565b156131d0576040805162461bcd60e51b815260206004820152601160248201527f616c7265616479206c6f636b6564207570000000000000000000000000000000604482015290519081900360640190fd5b612dcb8185858561439f565b60009182526077602090815260408084209284526002909201905290205490565b60009182526077602090815260408084209284526004909201905290205490565b613226612a89565b613277576040805162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015290519081900360640190fd5b612afc8161476b565b613288615024565b61329283836137d2565b6132e3576040805162461bcd60e51b815260206004820152601860248201527f6f75747374616e64696e67207346544d2062616c616e63650000000000000000604482015290519081900360640190fd5b6132ed838361388d565b50506001600160a01b0382166000908152606f602090815260408083208484528252808320815160608101835281548082526001830154948201859052600290920154928101839052939261334b926111259163ffffffff6133e016565b90508061339f576040805162461bcd60e51b815260206004820152600c60248201527f7a65726f20726577617264730000000000000000000000000000000000000000604482015290519081900360640190fd5b6001600160a01b0384166000908152606f60209081526040808320868452909152812081815560018101829055600201556133d981613734565b5092915050565b6000828201838110156112bb576040805162461bcd60e51b815260206004820152601b60248201527f536166654d6174683a206164646974696f6e206f766572666c6f770000000000604482015290519081900360640190fd5b61344382614754565b613494576040805162461bcd60e51b815260206004820152601760248201527f76616c696461746f7220646f65736e2774206578697374000000000000000000604482015290519081900360640190fd5b600082815260686020526040902054156134f5576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b613502838383600161359c565b61350b82614824565b61172a5760405162461bcd60e51b81526004018080602001828103825260298152602001806151976029913960400191505060405180910390fd5b60006112bb83836040518060400160405280601e81526020017f536166654d6174683a207375627472616374696f6e206f766572666c6f7700008152506148ec565b6066546001600160a01b0390811691161490565b600082116135f1576040805162461bcd60e51b815260206004820152600b60248201527f7a65726f20616d6f756e74000000000000000000000000000000000000000000604482015290519081900360640190fd5b6135fb848461388d565b506001600160a01b0384166000908152607260209081526040808320868452909152902054613630908363ffffffff6133e016565b6001600160a01b0385166000908152607260209081526040808320878452825280832093909355606890522060030154613670818463ffffffff6133e016565b600085815260686020526040902060030155606c54613695908463ffffffff6133e016565b606c556000848152606860205260409020546136c257606d546136be908463ffffffff6133e016565b606d555b6136cd848215612e0f565b60408051848152905185916001600160a01b038816917f9a8f44850296624dadfd9c246d17e47171d35727a181bd090aa14bbbe00238bb9181900360200190a36000848152606860205260409020600601546118f79086906001600160a01b031684613f23565b606654604080517f66e7ea0f0000000000000000000000000000000000000000000000000000000081523060048201526024810184905290516001600160a01b03909216916366e7ea0f9160448082019260009290919082900301818387803b1580156137a057600080fd5b505af11580156137b4573d6000803e3d6000fd5b50506076546137cc925090508263ffffffff6133e016565b60765550565b607b546000906001600160a01b03166137ed575060016112be565b607b54604080517f21d585c30000000000000000000000000000000000000000000000000000000081526001600160a01b03868116600483015260248201869052915191909216916321d585c3916044808301926020929190829003018186803b15801561385a57600080fd5b505afa15801561386e573d6000803e3d6000fd5b505050506040513d602081101561388457600080fd5b50519392505050565b6000613897615024565b6138a18484614983565b90506138ac83614abb565b6001600160a01b0385166000818152607060209081526040808320888452825280832094909455918152606f8252828120868252825282902082516060810184528154815260018201549281019290925260020154918101919091526139129082614b16565b6001600160a01b0385166000818152606f602090815260408083208884528252808320855181558583015160018083019190915595820151600291820155938352607482528083208884528252918290208251606081018452815481529481015491850191909152909101549082015261398c9082614b16565b6001600160a01b0385166000908152607460209081526040808320878452825291829020835181559083015160018201559101516002909101556139d08484613064565b613a33576001600160a01b0384166000818152607360209081526040808320878452825280832083815560018082018590556002808301869055600390920185905594845260748352818420888552909252822082815592830182905591909101555b6020810151151580613a455750805115155b806127ae57506040015115159392505050565b6001600160a01b03841660009081526074602090815260408083208684529091528120548190613aa0908490613a94908763ffffffff614b8816565b9063ffffffff614be116565b6001600160a01b038716600090815260746020908152604080832089845290915281206001015491925090613ae1908590613a94908863ffffffff614b8816565b6001600160a01b03881660009081526074602090815260408083208a845290915290205490915060028204830190613b199084613546565b6001600160a01b03891660009081526074602090815260408083208b8452909152902090815560010154613b4d9083613546565b6001600160a01b03891660009081526074602090815260408083208b8452909152902060010155858110613b7e5750845b979650505050505050565b6001600160a01b03841660009081526072602090815260408083208684528252808320805486900390556068909152902060030154613bce908363ffffffff61354616565b600084815260686020526040902060030155606c54613bf3908363ffffffff61354616565b606c55600083815260686020526040902054613c2057606d54613c1c908363ffffffff61354616565b606d555b6000613c2b8461271f565b90508015613d5d57600084815260686020526040902054613d5857608160009054906101000a90046001600160a01b03166001600160a01b031663c5f530af6040518163ffffffff1660e01b815260040160206040518083038186803b158015613c9457600080fd5b505afa158015613ca8573d6000803e3d6000fd5b505050506040513d6020811015613cbe57600080fd5b5051811015613d14576040805162461bcd60e51b815260206004820152601760248201527f696e73756666696369656e742073656c662d7374616b65000000000000000000604482015290519081900360640190fd5b613d1d84614824565b613d585760405162461bcd60e51b81526004018080602001828103825260298152602001806151976029913960400191505060405180910390fd5b613d68565b613d68846001613df9565b6000848152606860205260409020600601546118f79086906001600160a01b031684613f23565b8015612afc5760405160009082156108fc0290839083818181858288f19350505050158015613dc2573d6000803e3d6000fd5b506040805182815290517f8918bd6046d08b314e457977f29562c5d76a7030d79b1edba66e8a5da0b77ae89181900360200190a150565b600082815260686020526040902054158015613e1457508015155b15613e4157600082815260686020526040902060030154606d54613e3d9163ffffffff61354616565b606d555b600082815260686020526040902054811115612a7657600082815260686020526040902081815560020154613ee957613e7861293d565b600083815260686020526040902060020155613e926140cd565b6000838152606860209081526040918290206001810184905560020154825190815290810192909252805184927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e4792908290030190a25b60408051828152905183917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25050565b6082546001600160a01b03161561172a57608254604080516001600160a01b03868116602483015285811660448084019190915283518084039091018152606490920183526020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f4a7702bb00000000000000000000000000000000000000000000000000000000178152925182516000959290921693627a120093928291908083835b6020831061400657805182527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe09092019160209182019101613fc9565b6001836020036101000a03801982511681845116808217855250505050505090500191505060006040518083038160008787f1925050503d8060008114614069576040519150601f19603f3d011682016040523d82523d6000602084013e61406e565b606091505b50509050808061407c575081155b612dcb576040805162461bcd60e51b815260206004820152601b60248201527f676f7620766f746573207265636f756e74696e67206661696c65640000000000604482015290519081900360640190fd5b4290565b60008215806140e757506140e3614133565b8210155b156140f45750600061412c565b61411f6001611125614104614133565b613a9486614110614133565b8a91900363ffffffff614b8816565b90508381111561412c5750825b9392505050565b670de0b6b3a764000090565b6001600160a01b038816600090815260696020526040902054156141aa576040805162461bcd60e51b815260206004820152601860248201527f76616c696461746f7220616c7265616479206578697374730000000000000000604482015290519081900360640190fd5b6001600160a01b03881660008181526069602090815260408083208b90558a8352606882528083208981556004810189905560058101889055600181018690556002810187905560060180547fffffffffffffffffffffffff000000000000000000000000000000000000000016909417909355606a8152919020875161423392890190615045565b50876001600160a01b0316877f49bca1ed2666922f9f1690c26a569e1299c2a715fe57647d77e81adfabbf25bf8686604051808381526020018281526020019250505060405180910390a381156142bf576040805183815260208101839052815189927fac4801c32a6067ff757446524ee4e7a373797278ac3c883eac5c693b4ad72e47928290030190a25b84156142fc5760408051868152905188917fcd35267e7654194727477d6c78b541a553483cff7f92a055d17868d3da6e953e919081900360200190a25b5050505050505050565b61430e615024565b614316615024565b6143208484614983565b6001600160a01b0385166000908152606f6020908152604080832087845282529182902082516060810184528154815260018201549281019290925260020154918101919091529091506127ae9082614b16565b606b80546001019081905561172a838284600061438f61293d565b6143976140cd565b60008061413f565b6143a9848461123b565b8111156143fd576040805162461bcd60e51b815260206004820152601060248201527f6e6f7420656e6f756768207374616b6500000000000000000000000000000000604482015290519081900360640190fd5b6000838152606860205260409020541561445e576040805162461bcd60e51b815260206004820152601660248201527f76616c696461746f722069736e27742061637469766500000000000000000000604482015290519081900360640190fd5b608160009054906101000a90046001600160a01b03166001600160a01b0316630d7b26096040518163ffffffff1660e01b815260040160206040518083038186803b1580156144ac57600080fd5b505afa1580156144c0573d6000803e3d6000fd5b505050506040513d60208110156144d657600080fd5b505182108015906145605750608160009054906101000a90046001600160a01b03166001600160a01b0316630d4955e36040518163ffffffff1660e01b815260040160206040518083038186803b15801561453057600080fd5b505afa158015614544573d6000803e3d6000fd5b505050506040513d602081101561455a57600080fd5b50518211155b6145b1576040805162461bcd60e51b815260206004820152601260248201527f696e636f7272656374206475726174696f6e0000000000000000000000000000604482015290519081900360640190fd5b60006145bf836111256140cd565b6000858152606860205260409020600601549091506001600160a01b03908116908616811461464d576001600160a01b038116600090815260736020908152604080832088845290915290206002015482111561464d5760405162461bcd60e51b815260040180806020018281038252602881526020018061516f6028913960400191505060405180910390fd5b614657868661388d565b506001600160a01b0386166000908152607360209081526040808320888452909152902060038101548510156146d4576040805162461bcd60e51b815260206004820152601f60248201527f6c6f636b7570206475726174696f6e2063616e6e6f7420646563726561736500604482015290519081900360640190fd5b80546146e6908563ffffffff6133e016565b81556146f061293d565b600182015560028101839055600381018590556040805186815260208101869052815188926001600160a01b038b16927f138940e95abffcd789b497bf6188bba3afa5fbd22fb5c42c2f6018d1bf0f4e78929081900390910190a350505050505050565b600090815260686020526040902060050154151590565b6001600160a01b0381166147b05760405162461bcd60e51b81526004018080602001828103825260268152602001806150de6026913960400191505060405180910390fd5b6033546040516001600160a01b038084169216907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a3603380547fffffffffffffffffffffffff0000000000000000000000000000000000000000166001600160a01b0392909216919091179055565b60006148d1614831614133565b608154604080517f2265f2840000000000000000000000000000000000000000000000000000000081529051613a94926001600160a01b031691632265f284916004808301926020929190829003018186803b15801561489057600080fd5b505afa1580156148a4573d6000803e3d6000fd5b505050506040513d60208110156148ba57600080fd5b50516148c58661271f565b9063ffffffff614b8816565b60008381526068602052604090206003015411159050919050565b6000818484111561497b5760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b83811015614940578181015183820152602001614928565b50505050905090810190601f16801561496d5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b505050900390565b61498b615024565b6001600160a01b0383166000908152607060209081526040808320858452909152812054906149b984614abb565b905060006149c78686614c23565b9050818111156149d45750805b828110156149df5750815b6001600160a01b03861660008181526073602090815260408083208984528252808320938352607282528083208984529091528120548254909190614a2b90839063ffffffff61354616565b90506000614a3f84600001548a8988614d00565b9050614a49615024565b614a57828660030154614d63565b9050614a65838b8a89614d00565b9150614a6f615024565b614a7a836000614d63565b9050614a88858c898b614d00565b9250614a92615024565b614a9d846000614d63565b9050614aaa838383614f24565b9d9c50505050505050505050505050565b60008181526068602052604081206002015415614b0e576000828152606860205260409020600201546067541015614af65750606754612750565b50600081815260686020526040902060020154612750565b505060675490565b614b1e615024565b6040805160608101909152825184518291614b3f919063ffffffff6133e016565b8152602001614b5f846020015186602001516133e090919063ffffffff16565b8152602001614b7f846040015186604001516133e090919063ffffffff16565b90529392505050565b600082614b97575060006112be565b82820282848281614ba457fe5b04146112bb5760405162461bcd60e51b815260040180806020018281038252602181526020018061512d6021913960400191505060405180910390fd5b60006112bb83836040518060400160405280601a81526020017f536166654d6174683a206469766973696f6e206279207a65726f000000000000815250614f3f565b6001600160a01b0382166000908152607360209081526040808320848452909152812060010154606754614c58858583614fa4565b15614c665791506112be9050565b614c71858584614fa4565b614c80576000925050506112be565b80821115614c93576000925050506112be565b80821015614cc657600281830104614cac868683614fa4565b15614cbc57806001019250614cc0565b8091505b50614c93565b80614cd6576000925050506112be565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff01949350505050565b6000818310614d11575060006127ae565b600083815260776020818152604080842088855260019081018352818520548786529383528185208986520190915290912054613b7e614d4f614133565b613a94896148c5858763ffffffff61354616565b614d6b615024565b60405180606001604052806000815260200160008152602001600081525090506000608160009054906101000a90046001600160a01b03166001600160a01b0316635e2308d26040518163ffffffff1660e01b815260040160206040518083038186803b158015614ddb57600080fd5b505afa158015614def573d6000803e3d6000fd5b505050506040513d6020811015614e0557600080fd5b505190508215614efd57600081614e1a614133565b0390506000614eac608160009054906101000a90046001600160a01b03166001600160a01b0316630d4955e36040518163ffffffff1660e01b815260040160206040518083038186803b158015614e7057600080fd5b505afa158015614e84573d6000803e3d6000fd5b505050506040513d6020811015614e9a57600080fd5b5051613a94848863ffffffff614b8816565b90506000614ecd614ebb614133565b613a948987860163ffffffff614b8816565b9050614eea614eda614133565b613a94898763ffffffff614b8816565b6020860181905290038452506133d99050565b614f18614f08614133565b613a94868463ffffffff614b8816565b60408301525092915050565b614f2c615024565b6127ae614f398585614b16565b83614b16565b60008183614f8e5760405162461bcd60e51b8152602060048201818152835160248401528351909283926044909101919085019080838360008315614940578181015183820152602001614928565b506000838581614f9a57fe5b0495945050505050565b6001600160a01b038316600090815260736020908152604080832085845290915281206001015482108015906127ae57506001600160a01b03841660009081526073602090815260408083208684529091529020600201546150058361500f565b1115949350505050565b60009081526077602052604090206007015490565b60405180606001604052806000815260200160008152602001600081525090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061508657805160ff19168380011785556150b3565b828001600101855582156150b3579182015b828111156150b3578251825591602001919060010190615098565b506150bf9291506150c3565b5090565b61294491905b808211156150bf57600081556001016150c956fe4f776e61626c653a206e6577206f776e657220697320746865207a65726f206164647265737363616c6c6572206973206e6f7420746865204e6f64654472697665724175746820636f6e7472616374536166654d6174683a206d756c7469706c69636174696f6e206f766572666c6f776d757374206265206c657373207468616e206f7220657175616c20746f20312e3076616c696461746f72206c6f636b757020706572696f642077696c6c20656e64206561726c69657276616c696461746f7227732064656c65676174696f6e73206c696d69742069732065786365656465646c6f636b6564207374616b652069732067726561746572207468616e207468652077686f6c65207374616b65a265627a7a72315820b767b9bd00adb5f0467097fa60a9e2cfcb624e12e00375239584e1338e314fe964736f6c63430005110032"
