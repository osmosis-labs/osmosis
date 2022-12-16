# validator-set Preference 

## Abstract 

Validator-Set preference is a new module which gives users and contracts a 
better UX for staking to a set of validators. For example: a one click button
that delegates to multiple validators. Then the user can set (or realistically a frontend provides) 
a list of recommended defaults (Ex: active governors, relayers, core stack contributors etc).
Currently this can be done on-chain with frontends, but having a preference list stored locally 
eases frontend code burden. 

## Design 

How does this module work? 

- Allow a user to set a list of {val-addr, weight} in the state, called their validator-set preference.
- Allow a user to update a list of {val-addr, weight} in the state, then do the following; 
  - Unstake the existing tokens (run the same unbond logic as cosmos-sdk staking).
  - Update the validator distribution weights.
  - Stake the tokens based on the new weights.
  - Redelegate their current delegation to the currently configured set.
- Give users a single message to delegate {X} tokens, according to their validator-set preference distribution.
- Give users a single message to undelegate {X} tokens, according to their validator-set preference distribution.
- Give users a single message to claim rewards from everyone on their preference list.
- If the delegator has not set a validator-set preference list then the validator set, then it defaults to their current validator set.
- If a user has no preference list and has not staked, then these messages / queries return errors.

## Calculations

Staking Calculation 

- The user provides an amount to delegate and our `MsgDelegateToValidatorSet` divides the amount based on validator weight distribution.
  For example: Stake 100osmo with validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}
  our delegate logic will attempt to delegate (100 * 0.5) 50osmo for ValA , (100 * 0.3) 30osmo from ValB and (100 * 0.2) 20osmo from ValC.

UnStaking Calculation 

- The user provides an amount to undelegate and our `MsgUnDelegateToValidatorSet` divides the amount based on validator weight distribution.
- Here, the user can either undelegate the entire amount or partial amount 
  - Entire amount unstaking: UnStake 100osmo from validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2},
    our undelegate logic will attempt to undelegate 50osmo from ValA , 30osmo from ValB, 20osmo from ValC
  - Partial amount unstaking: UnStake 27osmo from validator-set {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2}, 
    our undelegate logic will attempt to undelegate (27 * 0.5) 13.5osmos from ValA, (27 * 0.3), 8.1osmo from ValB, 
    and (50 * 0.2) 5.4smo from ValC where 13.5osmo + 8.1osmo + 5.4osmo = 27osmo
  - The user will then have 73osmo remaining with unchanged weights {ValA -> 0.5, ValB -> 0.3, ValC -> 0.2},

## Messages

### CreateValidatorSetPreference

Creates a validator-set of `{valAddr, Weight}` given the delegator address.
and preferences. The weights are in decimal format from 0 to 1 and must add up to 1.

```go
    string delegator = 1 [ (gogoproto.moretags) = "yaml:\"delegator\"" ];
    repeated ValidatorPreference preferences = 2 [
      (gogoproto.moretags) = "yaml:\"preferences\"",
      (gogoproto.nullable) = false
    ];
```

**State Modifications:**

- Safety Checks
  - check if the user already has a validator-set created. 
  - check if the validator exist and is valid.
  - check if the validator-set add up to 1.
- Add owner address to the `KVStore`, where a state of validator-set is stored. 

### UpdateValidatorSetPreference

Updates a validator-set of `{valAddr, Weight}` given the delegator address
and existing preferences. The weights calculations follow the same rule as `CreateValidatorSetPreference`.
If a user changes their preferences list, the unstaking logic will run from the old set and 
restaking to a new set is going to happen behind the scenes.

```go
    string delegator = 1 [ (gogoproto.moretags) = "yaml:\"delegator\"" ];
    repeated ValidatorPreference preferences = 2 [
      (gogoproto.moretags) = "yaml:\"preferences\"",
      (gogoproto.nullable) = false
    ];
```

**State Modifications:**

- Follows the same rule as `CreateValidatorSetPreference` for weights checks.
- Update the `KVStore` value for the specific owner address key.
- Run the undelegate logic and restake the tokens with updated weights. 

### DelegateToValidatorSet

Gets the existing validator-set of the delegator and delegates the given amount. The given amount 
will be divided based on the weights distributed to the validators. The weights will be unchanged 

```go
    string delegator = 1 [ (gogoproto.moretags) = "yaml:\"delegator\"" ];
    repeated ValidatorPreference preferences = 2 [
      (gogoproto.moretags) = "yaml:\"preferences\"",
      (gogoproto.nullable) = false
    ];
```

**State Modifications:**

- Check if the user has a validator-set and if so, get the users validator-set from `KVStore`. 
- Safety Checks 
  - check if the user has enough funds to delegate.
  - check overflow/underflow since `Delegate` method takes `sdk.Int` as tokenAmount.
- use the [Delegate](https://github.com/cosmos/cosmos-sdk/blob/main/x/staking/keeper/delegation.go#L614) method from the cosmos-sdk to handle delegation. 

### UnStakeFromValidatorSet

Gets the existing validator-set of the delegator and undelegate the given amount. The amount to undelegate will
will be divided based on the weights distributed to the validators. The weights will be unchanged! 

The given amount will be divided based on the weights distributed to the validators.

```go
    string delegator = 1 [ (gogoproto.moretags) = "yaml:\"delegator\"" ];
    repeated ValidatorPreference preferences = 2 [
      (gogoproto.moretags) = "yaml:\"preferences\"",
      (gogoproto.nullable) = false
    ];
```

**State Modifications:**

- Check if the user has a validator-set and if so, get the users validator-set from `KVStore`. 
- The unbonding logic will be follow the `UnDelegate` logic from the cosmos-sdk. 
- Safety Checks 
  - check that the amount of funds to undelegate is <= to the funds the user has in the address.
  - `UnDelegate` method takes `sdk.Dec` as tokenAmount, so check if overflow/underflow case is relevant.
- use the [UnDelegate](https://github.com/cosmos/cosmos-sdk/blob/main/x/staking/keeper/delegation.go#L614) method from the cosmos-sdk to handle delegation. 

### WithdrawDelegationRewards

Allows the user to claim rewards based from the existing validator-set. The user can claim rewards from all the validators at once. 

```go
    string delegator = 1 [ (gogoproto.moretags) = "yaml:\"delegator\"" ];
```

## Code Layout 

The Code Layout is very similar to TWAP module.

- client/* - Implementation of GRPC and CLI queries
- types/* - Implement ValidatorSetPreference, GenesisState. Define the interface and setup keys.
- twapmodule/validatorsetpreference.go - SDK AppModule interface implementation.
- api.go - Public API, that other users / modules can/should depend on
- listeners.go - Defines hooks & calls to logic.go, for triggering actions on 
- keeper.go - generic SDK boilerplate (defining a wrapper for store keys + params)
- msg_server.go - handle messages request from client and process responses. 
- store.go - Managing logic for getting and setting things to underlying stores (KVStore)

## Redelegate algorithm logic pseudocode

Existing ValSet   20osmos {ValA-> 0.5, ValB-> 0.3, ValC-> 0.2} [ValA-> 10osmo, ValB-> 6osmo, ValC-> 4osmo]
New ValSet        20osmos {ValD-> 0.2, ValE-> 0.2, ValF-> 0.6} [ValD-> 4osmo, ValE-> 4osmo, ValD-> 12osmo]

- // Rearranging the existingValSet and newValSet to to add extra validator padding
  - existing_valset_updated = [ValA: 10, ValB: 6, ValC: 4, ValD: 0, ValE: 0, ValF: 0]
  - new_valset_updated = [ValD: 4, ValE: 4, ValF: 12, ValA: 0, ValB: 0, ValC: 0]

  // calculate the difference between two sets
  - diff_arr = [ValA: 10, ValB: 6, ValC: 4, ValD: -4, ValE: -4, ValF: -12]

	// Algorithm starts here
- for i, validator in diff_arr: 
    - for validator.amount > 0: 
        source_validator = validator.address
        // FindMin returns the index and MinAmt of the minimum amount in diffValSet
        target_validator, idx = FindMin(diff_arr)   // gets the index of minValue and the minValue

        // reDelegationAmt to is the amount to redelegate, which is the min of diffAmount and target_validator
        reDelegationAmt = FindMin(abs(target_validator.amount), validator.amount)
        sdk.BeginRedelegation(ctx, delegator, source_validator, target_validator, reDelegationAmt) 

        // Update the current diffAmount by subtracting it with the reDelegationAmount
        validator.amount = validator.amount - reDelegationAmt
        // Find target_validator through idx in diffValSet and set that to (target_validatorAmount - reDelegationAmount)
        diff_arr[idx].amount = target_validator.amount + reDelegationAmt 

- Result 
  1. diff_arr = [ValA: 0, ValB: 0, ValC: 0, ValD: 0, ValE: 0, ValF: 0]
  2. [ValA: 0, ValB: 0, ValC: 0, ValD: 4, ValE: 4, ValF: 12] // final result


## Redelegation Constraints 
1. ValA -> ValB redelegate upto 7 times in 21 day period 
2. ValA -> ValB (redelegate) ValB -> ValC (redelegate) **CONSECUTIVE REDELEGATION DOES NOT WORK**
3. Once you redelegate from ValA -> ValB, you will not be able to redelegate from ValB to another validator for the next 21 days.
  - the validator on the receiving end of redelegation will be on a 21-day redelegation lock
4. Cannot redelegate to same validator 
