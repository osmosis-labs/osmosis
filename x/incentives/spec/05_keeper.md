<!--
order: 5
-->

# Keepers

## Incentive Keeper

Incentive Keeper provide utility functions.

```go
// Keeper is the interface for incentives module keeper
type Keeper interface {
	// GetModuleToDistributeCoins returns coins that is going to be distributed
	GetModuleToDistributeCoins(sdk.Context) sdk.Coins
	// GetModuleDistributedCoins returns coins that are distributed by module so far
	GetModuleDistributedCoins(sdk.Context) sdk.Coins

	// GetPotByID returns Pot by id
	GetPotByID(sdk.Context, potID uint64) (*types.Pot, error)
	// GetPots returns pots both upcoming and active
	GetPots(sdk.Context) ([]types.Pot, error)
	// GetActivePots returns active pots
	GetActivePots(sdk.Context) ([]types.Pot, error)
	// GetUpcomingPots returns scheduled pots
	GetUpcomingPots(sdk.Context) ([]types.Pot, error)
	// GetRewardsEst returns rewards estimation at a future specific time
	GetRewardsEst(sdk.Context, account sdk.AccAddress, time time.Time) (sdk.Coins, error)

	// CreatePot create a pot to give incentives to users
	CreatePot(sdk.Context, potID uint64) (*types.Pot, error)
	// AddToPot add more rewards to give more incentives to lockers
	AddToPot(sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.Pot, error)
	// Distribute is a function to distribute from Pot
	Distribute(sdk.Context, pot types.Pot) error
}
```
