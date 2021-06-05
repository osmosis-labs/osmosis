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

	// GetGaugeByID returns Gauge by id
	GetGaugeByID(sdk.Context, gaugeID uint64) (*types.Gauge, error)
	// GetGauges returns gauges both upcoming and active
	GetGauges(sdk.Context) ([]types.Gauge, error)
	// GetActiveGauges returns active gauges
	GetActiveGauges(sdk.Context) ([]types.Gauge, error)
	// GetUpcomingGauges returns scheduled gauges
	GetUpcomingGauges(sdk.Context) ([]types.Gauge, error)
	// GetRewardsEst returns rewards estimation at a future specific time
	GetRewardsEst(sdk.Context, account sdk.AccAddress, time time.Time) (sdk.Coins, error)

	// CreateGauge create a gauge to give incentives to users
	CreateGauge(sdk.Context, gaugeID uint64) (*types.Gauge, error)
	// AddToGauge add more rewards to give more incentives to lockers
	AddToGauge(sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.Gauge, error)
	// Distribute is a function to distribute from Gauge
	Distribute(sdk.Context, gauge types.Gauge) error
}
```
