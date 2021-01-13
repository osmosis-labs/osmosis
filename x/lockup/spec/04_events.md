<!--
order: 4
-->

# Events

The governance module emits the following events:

## EndBlocker

| Type              | Attribute Key   | Attribute Value  |
| ----------------- | --------------- | ---------------- |
| unlock_coins      | period_lock_id  | {periodLockID}   |
| unlock_coins      | amount          | {amount}         |

Note:
If we don't do automation of withdraw, Endblocker won't be required.

## Handlers

### MsgLockTokens

| Type                | Attribute Key       | Attribute Value |
| ------------------- | ------------------- | --------------- |
| lock_tokens         | proposer            | {proposer}      |
| lock_tokens         | amount              | {amount}        |
| lock_tokens         | period_lock_id      | {periodLockID}  |

### MsgUnlockTokens

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| unlock_tokens | proposer      | {proposer}      |
| unlock_tokens | amount        | {amount}        |
