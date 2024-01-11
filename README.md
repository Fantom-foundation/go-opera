# X1 Blockchain

X1 is a simple, fast, and secure EVM-compatible network for the next generation of decentralized applications powered by Lachesis consensus algorithm.

Chain ID: `204005`

## Explore the Network (Testnet)

- [X1 Explorer](https://explorer.x1-testnet.xen.network)

RPC Endpoints

- https://x1-testnet.xen.network/
- wss://x1-testnet.xen.network

## Run Full Node (Testnet)

> Quick start a full node and run in the foreground

```shell
# Install dependencies (ex: ubuntu)
apt update -y
apt install -y golang wget git make

# Clone and build the X1 binary
git clone --branch x1 https://github.com/FairCrypto/go-x1
cd go-x1
make x1
cp build/x1 /usr/local/bin

# Run the node
x1 --testnet --syncmode snap
```

## Run an Archive Node (Testnet)

```shell
x1 --testnet --gcmode archive --syncmode full
```

> Run with Xenblocks reporting enabled
```shell
x1 --testnet --xenblocks-endpoint ws://xenblocks.io:6668
```

> Run with RPC server enabled
```shell
x1 --testnet --http --http.port 8545 --ws --ws.port 8546
```

> Run with RPC server open to the world (only run if you know what you are doing)
```shell
x1 --testnet --http --http.port 8545 --http.addr 0.0.0.0 --http.vhosts "*" --http.corsdomain "*" --ws --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins "*"
```

## Run a Validator Node (Testnet)

> [!CAUTION]
> Running a validator node is a complex process and requires a lot of technical knowledge.

Run a full node as described above and allow to fully sync, then follow the instructions below to run a validator node.
Do not use `--http` or `--ws` flags during the validator node setup.

### Create a validator wallet

The node running and syncing the network in your current console, so you need to open up a new console window,
connect via SSH to the server and enter the following commands to create a wallet:

```shell
# Create validator account wallet
x1 account new
```

After entering the command, you will get prompted to enter a password for the account (wallet) — use a strong one!
You can e.g. use a password manager to generate a strong password to secure your wallet.

### Fund your validator wallet
The next step is to fund your validator wallet with enough XN to become a validator.
That means you need to have at least 100,000 XN in the wallet you just created (send a little more to cover transaction fees).

### Create a new validator key

We have to create validator private key to sign consensus messages with. It can be done only using go-x1 tool.
Take note of the validator public key, we will need it later.

```shell
# Create validator key
x1 validator new
```

### Create your validator via the SFC contract

Now we need to register our validator in the SFC contract.

```shell
# Attach to x1 console
x1 attach
```

Now initialize the SFC contract using the ABI:

```shell
abi = JSON.parse('[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"status","type":"uint256"}],"name":"ChangedValidatorStatus","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"lockupExtraReward","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"lockupBaseReward","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unlockedReward","type":"uint256"}],"name":"ClaimedRewards","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":true,"internalType":"address","name":"auth","type":"address"},{"indexed":false,"internalType":"uint256","name":"createdEpoch","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"createdTime","type":"uint256"}],"name":"CreatedValidator","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"deactivatedEpoch","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"deactivatedTime","type":"uint256"}],"name":"DeactivatedValidator","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Delegated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"duration","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"LockedUpStake","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"lockupExtraReward","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"lockupBaseReward","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unlockedReward","type":"uint256"}],"name":"RestakedRewards","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"indexed":true,"internalType":"uint256","name":"wrID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Undelegated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"penalty","type":"uint256"}],"name":"UnlockedStake","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"UpdatedBaseRewardPerSec","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"blocksNum","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"period","type":"uint256"}],"name":"UpdatedOfflinePenaltyThreshold","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"validatorID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"refundRatio","type":"uint256"}],"name":"UpdatedSlashingRefundRatio","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"delegator","type":"address"},{"indexed":true,"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"indexed":true,"internalType":"uint256","name":"wrID","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Withdrawn","type":"event"},{"constant":true,"inputs":[],"name":"baseRewardPerSecond","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"contractCommission","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"currentSealedEpoch","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"getEpochSnapshot","outputs":[{"internalType":"uint256","name":"endTime","type":"uint256"},{"internalType":"uint256","name":"epochFee","type":"uint256"},{"internalType":"uint256","name":"totalBaseRewardWeight","type":"uint256"},{"internalType":"uint256","name":"totalTxRewardWeight","type":"uint256"},{"internalType":"uint256","name":"baseRewardPerSecond","type":"uint256"},{"internalType":"uint256","name":"totalStake","type":"uint256"},{"internalType":"uint256","name":"totalSupply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"uint256","name":"","type":"uint256"}],"name":"getLockupInfo","outputs":[{"internalType":"uint256","name":"lockedStake","type":"uint256"},{"internalType":"uint256","name":"fromEpoch","type":"uint256"},{"internalType":"uint256","name":"endTime","type":"uint256"},{"internalType":"uint256","name":"duration","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"uint256","name":"","type":"uint256"}],"name":"getStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"uint256","name":"","type":"uint256"}],"name":"getStashedLockupRewards","outputs":[{"internalType":"uint256","name":"lockupExtraReward","type":"uint256"},{"internalType":"uint256","name":"lockupBaseReward","type":"uint256"},{"internalType":"uint256","name":"unlockedReward","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"getValidator","outputs":[{"internalType":"uint256","name":"status","type":"uint256"},{"internalType":"uint256","name":"deactivatedTime","type":"uint256"},{"internalType":"uint256","name":"deactivatedEpoch","type":"uint256"},{"internalType":"uint256","name":"receivedStake","type":"uint256"},{"internalType":"uint256","name":"createdEpoch","type":"uint256"},{"internalType":"uint256","name":"createdTime","type":"uint256"},{"internalType":"address","name":"auth","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"getValidatorID","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"getValidatorPubkey","outputs":[{"internalType":"bytes","name":"","type":"bytes"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"","type":"uint256"}],"name":"getWithdrawalRequest","outputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"time","type":"uint256"},{"internalType":"uint256","name":"amount","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"isOwner","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"lastValidatorID","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"maxDelegatedRatio","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"maxLockupDuration","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"minLockupDuration","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"minSelfStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[],"name":"renounceOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"slashingRefundRatio","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"stakeTokenizerAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"","type":"address"},{"internalType":"uint256","name":"","type":"uint256"}],"name":"stashedRewardsUntilEpoch","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalActiveStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSlashedStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"unlockedRewardRatio","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"validatorCommission","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"version","outputs":[{"internalType":"bytes3","name":"","type":"bytes3"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"withdrawalPeriodEpochs","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"withdrawalPeriodTime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"currentEpoch","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"}],"name":"getEpochValidatorIDs","outputs":[{"internalType":"uint256[]","name":"","type":"uint256[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getEpochReceivedStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getEpochAccumulatedRewardPerToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getEpochAccumulatedUptime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getEpochAccumulatedOriginatedTxsFee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getEpochOfflineTime","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"epoch","type":"uint256"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getEpochOfflineBlocks","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"rewardsStash","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"getLockedStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"sealedEpoch","type":"uint256"},{"internalType":"uint256","name":"_totalSupply","type":"uint256"},{"internalType":"address","name":"nodeDriver","type":"address"},{"internalType":"address","name":"owner","type":"address"}],"name":"initialize","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"auth","type":"address"},{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"bytes","name":"pubkey","type":"bytes"},{"internalType":"uint256","name":"status","type":"uint256"},{"internalType":"uint256","name":"createdEpoch","type":"uint256"},{"internalType":"uint256","name":"createdTime","type":"uint256"},{"internalType":"uint256","name":"deactivatedEpoch","type":"uint256"},{"internalType":"uint256","name":"deactivatedTime","type":"uint256"}],"name":"setGenesisValidator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"stake","type":"uint256"},{"internalType":"uint256","name":"lockedStake","type":"uint256"},{"internalType":"uint256","name":"lockupFromEpoch","type":"uint256"},{"internalType":"uint256","name":"lockupEndTime","type":"uint256"},{"internalType":"uint256","name":"lockupDuration","type":"uint256"},{"internalType":"uint256","name":"earlyUnlockPenalty","type":"uint256"},{"internalType":"uint256","name":"rewards","type":"uint256"}],"name":"setGenesisDelegation","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"bytes","name":"pubkey","type":"bytes"}],"name":"createValidator","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"getSelfStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"delegate","outputs":[],"payable":true,"stateMutability":"payable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"wrID","type":"uint256"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"undelegate","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"}],"name":"isSlashed","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"wrID","type":"uint256"}],"name":"withdraw","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"uint256","name":"status","type":"uint256"}],"name":"deactivateValidator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"pendingRewards","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"stashRewards","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"claimRewards","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"restakeRewards","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"bool","name":"syncPubkey","type":"bool"}],"name":"_syncValidator","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"offlinePenaltyThreshold","outputs":[{"internalType":"uint256","name":"blocksNum","type":"uint256"},{"internalType":"uint256","name":"time","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"value","type":"uint256"}],"name":"updateBaseRewardPerSecond","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"blocksNum","type":"uint256"},{"internalType":"uint256","name":"time","type":"uint256"}],"name":"updateOfflinePenaltyThreshold","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"validatorID","type":"uint256"},{"internalType":"uint256","name":"refundRatio","type":"uint256"}],"name":"updateSlashingRefundRatio","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"address","name":"addr","type":"address"}],"name":"updateStakeTokenizerAddress","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"offlineTime","type":"uint256[]"},{"internalType":"uint256[]","name":"offlineBlocks","type":"uint256[]"},{"internalType":"uint256[]","name":"uptimes","type":"uint256[]"},{"internalType":"uint256[]","name":"originatedTxsFee","type":"uint256[]"}],"name":"sealEpoch","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256[]","name":"nextValidatorIDs","type":"uint256[]"}],"name":"sealEpochValidators","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"isLockedUp","outputs":[{"internalType":"bool","name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"internalType":"address","name":"delegator","type":"address"},{"internalType":"uint256","name":"toValidatorID","type":"uint256"}],"name":"getUnlockedStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"lockupDuration","type":"uint256"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"lockStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"lockupDuration","type":"uint256"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"relockStake","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"internalType":"uint256","name":"toValidatorID","type":"uint256"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"unlockStake","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]')
sfc = web3.ftm.contract(abi).at("0xfc00face00000000000000000000000000000000")
```

Next, unlock your validator wallet to be able to execute the registration transaction.
Replace `WALLET_ADDRESS` with your wallet address and `PASSWORD` with your account password:
```shell
# Unlock the validator wallet using the account password. 
personal.unlockAccount("WALLET_ADDRESS", "PASSWORD", 60)
```

Next, send the createValidator transaction to register your validator.
Replace `WALLET_ADDRESS` with your wallet address and `VALIDATOR_PUBKEY` with your validator's public key.

```shell
# Register your validator
tx = sfc.createValidator("VALIDATOR_PUBKEY", {from:"WALLET_ADDRESS", value: web3.toWei("500000.0", "xn")})
```

```shell
# Check your registration transaction
ftm.getTransactionReceipt(tx)
```

Look for the status: “0x1" at the bottom, which means the transaction was successful.

Finally, execute the following command again to check your validatorID:

```shell
# Get your validator id
sfc.getValidatorID("WALLET_ADDRESS")
```

### Run your X1 validator node

Now that you have registered your validator, you can restart your node with the following command:

Replace `VALIDATOR_ID` with your validator ID and `VALIDATOR_PUBKEY` with your validator's public key.

```shell
x1 --testnet --validator.id VALIDATOR_ID --validator.pubkey VALIDATOR_PUBKEY --validator.password /path/to/password_file
```

Congratulations, you are now running a X1 validator node!
