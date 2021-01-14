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
| lock_tokens         | unlock_time         | {unlock_time}   |

### MsgUnlockTokens

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| unlock_tokens | owner         | {owner}         |
| unlock_tokens | lock_id       | {lock_id}       |
