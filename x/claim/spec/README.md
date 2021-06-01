# `claim`

## Abstract

This specifies the `claim` module of Osmosis project, provide commands for claimable amount query and claim airdrop.
We apply real-time decay after `DurationUntilDecay` pass where it decays linearly to 0 over a `DurationOfDecay`-long period.
When `DurationUntilDecay + DurationOfDecay` time passes, all unclaimed coins will be sent to the community pool.

## Genesis State

## Actions

There are 4 types of actions, each of which release 25% of the airdrop amount. 
Because it is an enum, each of the 4 actions has an associated number.

```
ActionAddLiquidity  Action = 0
ActionSwap          Action = 1
ActionVote          Action = 2
ActionDelegateStake Action = 3
```

All of these actions are monitored by registring claim **hooks** to governance, staking, gamm, lockup modules hooks.
If Alice has done `DelegateStake` action, claim module withdraw 25% of claimables, and then if she do `Vote` action, she is able to withdraw another 25% of claimables. Here, double `DelegateStake` action only withdraw 25%.


## ClaimRecords

A claim record is a struct that contains data about the claims process of each airdrop recipient.

It contains an address, the initial claimable airdrop amount, and an array of bools representing 
whether each action has been completed. The position in the array refers to enum number of the action.

So for example, `[false, true, true, false]` means that `ActionSwap` and `ActionVote` are completed.

```
type ClaimRecord struct {
	// address of claim user
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty" yaml:"address"`

	// total initial claimable amount for the user
	InitialClaimableAmount sdk.Coins
	
	// true if action is completed
	// index of bool in array refers to action enum #
	ActionCompleted []bool
}

```

## Queries

Query the claim record for a given address
```sh
osmosisd query claim claim-record $(osmosisd keys show -a validator --keyring-backend=test)
```

Query the claimable amount that would be earned if a specific action is completed right now.

```sh 
osmosisd query claim claimable-for-action $(osmosisd keys show -a validator --keyring-backend=test) ActionAddLiquidity
```

Query the total claimable amount that would be earned if all remaining actions were completed right now.
Note that even if the decay process hasn't begun yet, this is not always *exactly* the same as `InitialClaimableAmount`, due to rounding errors.

```sh
osmosisd query claim total-claimable $(osmosisd keys show -a validator --keyring-backend=test) ActionAddLiquidity
```
