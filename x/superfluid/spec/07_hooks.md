<!--
order: 7
-->

# Hooks

In this section we describe the "hooks" that `superfluid` module receives from other modules.

## AfterEpochEnd

On AfterEpochEnd, we iterate through all existing intermediary accounts and withdraw delegation rewards they have received. Then we send the collective rewards to the perpetual gauge corresponding to the intermediary account. Then we update OSMO backing per share for the specific pool. After the update, iteration through all intermediate accounts happen, undelegating and bonding existing delegations for all superfluid staking and use the updated spot price at epoch time to mint and delegate.

## AfterAddTokensToLock

When a token is locked, we first check if the corresponding lock is currently in the state of superfluid delegation. If it is, we run the logic to add delegation via intermediary account.

## OnStartUnlock

On Unlocking of a lock, we check if the corresponding lock has been superfluid staked. If it has, we run `SuperfluidUndelegate` for undelegation of the superfluid delegation.

Note: This is done via a single transaction type now

## BeforeValidatorSlashed

Slashes the synthetic lockups and native lockups that is connected to the to be slashed validator.
