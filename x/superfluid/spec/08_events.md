<!--
order: 8
-->

# Events

## Messages

### MsgSuperfluidDelegate

| Type                | Attribute Key | Attribute Value |
| ------------------- | ------------- | --------------- |
| superfluid_delegate | lock_id       | {lock_id}       |
| superfluid_delegate | validator     | {validator}     |

### MsgSuperfluidUndelegate

| Type                  | Attribute Key | Attribute Value |
| --------------------- | ------------- | --------------- |
| superfluid_undelegate | lock_id       | {lock_id}       |

### MsgSuperfluidUnbondLock

| Type                   | Attribute Key | Attribute Value |
| ---------------------- | ------------- | --------------- |
| superfluid_unbond_lock | lock_id       | {lock_id}       |

### MsgLockAndSuperfluidDelegate

| Type                | Attribute Key  | Attribute Value |
| ------------------- | -------------- | --------------- |
| lock_tokens         | period_lock_id | {periodLockID}  |
| lock_tokens         | owner          | {owner}         |
| lock_tokens         | amount         | {amount}        |
| lock_tokens         | duration       | {duration}      |
| lock_tokens         | unlock_time    | {unlockTime}    |
| message             | action         | lock_tokens     |
| message             | sender         | {owner}         |
| transfer            | recipient      | {moduleAccount} |
| transfer            | sender         | {owner}         |
| transfer            | amount         | {amount}        |
| superfluid_delegate | lock_id        | {lock_id}       |
| superfluid_delegate | validator      | {validator}     |

## Proposals

### SetSuperfluidAssetsProposal

| Type                 | Attribute Key         | Attribute Value |
| -------------------- | --------------------- | --------------- |
| set_superfluid_asset | denom                 | {denom}         |
| set_superfluid_asset | superfluid_asset_type | {asset_type}    |

### RemoveSuperfluidAssetsProposal

| Type                    | Attribute Key | Attribute Value |
| ----------------------- | ------------- | --------------- |
| remove_superfluid_asset | denom         | {denom}         |
