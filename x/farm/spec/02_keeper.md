<!--
order: 2
-->

# Keeper


Farm Keeper provide utility functions to manage the farms.

```go
// Keeper is the interface for farm module keeper
type Keeper interface {
	// Create new farm for distribution. Farms can be accessed later using the farm id.
    NewFarm(ctx sdk.Context) (types.Farm, error)
    // Get the registered farm using the farm id. If there is no farm which matches the farm id, an error is returned.
    GetFarm(ctx sdk.Context, farmId uint64) (types.Farm, error)

    // Get the farmer using the user address and the farm id.
    // Don't directly create a new farmer, but rather use DepositShareToFarm.
    GetFarmer(ctx sdk.Context, farmId uint64, address sdk.AccAddress) (types.Farmer, error)
    
    // If there is no farmer with the specified farm id and user address, a new farmer that owns shares is registered.
    // If there is an existing farmer that matches the farm id and user address, the share is added to the farmer.
    // If there is already a farmer, when the shares are changed the previous pending rewards are withdrawn.
    // Rewards are returned. Because the farm keeper custodies the rewards, the rewards are transferred from the module account to the user automatically.
    DepositShareToFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error)
    // The opposite of DepsitShareToFarm.
    WithdrawShareFromFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress, share sdk.Int) (rewards sdk.Coins, err error)
    // If there is an existing farmer with the matching farm id and user address, the rewards are transferred to the matching farmer.
    // If there is no existing farmer, error is returned.
    WithdrawRewardsFromFarm(ctx sdk.Context, farmId uint64, address sdk.AccAddress) (rewards sdk.Coins, err error)

    // Allocate assets from the account to the farm as reward.
    // The assets of the account is transferred to the x/farm module account.
    // If there is no registered farm, error is returned.
    AllocateAssetsFromAccountToFarm(ctx sdk.Context, farmId uint64, allocator sdk.AccAddress, assets sdk.Coins) error
    // Allocate assets from module account to the farm as reward.
    // The assets of the module account is transferred to the x/farm module account.
    // If there is no registered farm, error is returned.
    AllocateAssetsFromModuleToFarm(ctx sdk.Context, farmId uint64, moduleName string, assets sdk.Coins) error
}
```
