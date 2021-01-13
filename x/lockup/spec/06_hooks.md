<!--
order: 6
-->

# Hooks

In this section we describe the "hooks" that `lockup` module provide for other modules.

## Tokens Locked

Upon successful coin lock/unlock, other modules might need to do few actions automatically instead of endblocker basis synchronization.

```go
  onTokenLocked(address sdk.AccAddress, amount sdk.Coins)
  onTokenWithdrawn(address sdk.AccAddress, amount sdk.Coins)
```
