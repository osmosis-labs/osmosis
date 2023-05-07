package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

var (
	IntMaxValue = intMaxValue
)

func (k Keeper) GetNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolIdAndIncrement(ctx)
}

func (k Keeper) GetOsmoRoutedMultihopTotalSwapFee(ctx sdk.Context, route types.MultihopRoute) (
	totalPathSwapFee sdk.Dec, sumOfSwapFees sdk.Dec, err error) {
	return k.getOsmoRoutedMultihopTotalSwapFee(ctx, route)
}

// SetPoolRoutesUnsafe sets the given routes to the poolmanager keeper
// to allow routing from a pool type to a certain swap module.
// For example, balancer -> gamm.
// This utility function is only exposed for testing and should not be moved
// outside of the _test.go files.
func (k *Keeper) SetPoolRoutesUnsafe(routes map[types.PoolType]types.PoolModuleI) {
	k.routes = routes
}

// SetPoolModulesUnsafe sets the given modules to the poolmanager keeper.
// This utility function is only exposed for testing and should not be moved
// outside of the _test.go files.
func (k *Keeper) SetPoolModulesUnsafe(poolModules []types.PoolModuleI) {
	k.poolModules = poolModules
}

func (k Keeper) GetAllPoolRoutes(ctx sdk.Context) []types.ModuleRoute {
	return k.getAllPoolRoutes(ctx)
}

func (k Keeper) ValidateCreatedPool(ctx sdk.Context, poolId uint64, pool types.PoolI) error {
	return k.validateCreatedPool(ctx, poolId, pool)
}
