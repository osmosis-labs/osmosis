<!--
order: 4
-->

# Events

The lockup module emits the following events:
## Handlers

### MsgLockTokens

| Type                | Attribute Key       | Attribute Value |
| ------------------- | ------------------- | --------------- |
| lock_tokens         | period_lock_id      | {periodLockID}  |
| lock_tokens         | owner               | {owner}         |
| lock_tokens         | amount              | {amount}        |
| lock_tokens         | duration            | {duration}      |
| lock_tokens         | unlock_time         | {unlockTime}   |
| message             | action              | lock_tokens     |
| message             | sender              | {owner}         |
| transfer            | recipient           | {moduleAccount} |
| transfer            | sender              | {owner}         |
| transfer            | amount              | {amount}        |

### MsgUnlockTokens

| Type          | Attribute Key  | Attribute Value |
| ------------- | -------------- | --------------- |
| message       | action         | unlock_tokens   |
| message       | sender         | {owner}         |
| transfer[]    | recipient      | {owner}         |
| transfer[]    | sender         | {moduleAccount} |
| transfer[]    | amount         | {unlockAmount}  |
| unlock[]      | period_lock_id | {owner}         |
| unlock[]      | owner          | {lockID}        |
| unlock[]      | duration       | {lockDuration}  |
| unlock[]      | unlock_time    | {unlockTime}    |
| unlock_tokens | owner          | {owner}         |
| unlock_tokens | unlocked_coins | {totalAmount}   |

### MsgUnlockPeriodLock

| Type          | Attribute Key  | Attribute Value    |
| ------------- | -------------- | ------------------ |
| message       | action         | unlock_period_lock |
| message       | sender         | {owner}            |
| transfer      | recipient      | {owner}            |
| transfer      | sender         | {moduleAccount}    |
| transfer      | amount         | {unlockAmount}     |
| unlock        | period_lock_id | {lockID}           |
| unlock        | owner          | {owner}            |
| unlock        | duration       | {lockDuration}     |
| unlock        | unlock_time    | {unlockTime}       |
