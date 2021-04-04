<!--
order: 6
-->

# Hooks

In this section we describe the "hooks" that `lockup` module provide for other modules.

## Tokens Locked

Upon successful coin lock/unlock, other modules might need to do few actions automatically instead of endblocker basis synchronization.

```go
  OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
  OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
```
