package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

<<<<<<< HEAD
	"github.com/osmosis-labs/osmosis/v18/x/poolmanager/types"
=======
	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
)

var IntMaxValue = intMaxValue

func (k Keeper) GetNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolIdAndIncrement(ctx)
}

func (k Keeper) GetOsmoRoutedMultihopTotalSpreadFactor(ctx sdk.Context, route types.MultihopRoute) (
	totalPathSpreadFactor osmomath.Dec, sumOfSpreadFactors osmomath.Dec, err error,
) {
	return k.getOsmoRoutedMultihopTotalSpreadFactor(ctx, route)
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
	return k.validateCreatedPool(poolId, pool)
}

func (k Keeper) IsOsmoRoutedMultihop(ctx sdk.Context, route types.MultihopRoute, inDenom, outDenom string) (isRouted bool) {
	return k.isOsmoRoutedMultihop(ctx, route, inDenom, outDenom)
}

func (k Keeper) CreateMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]osmomath.Int, error) {
	return k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
}

func (k Keeper) CreateOsmoMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
	cumulativeRouteSwapFee, sumOfSwapFees osmomath.Dec,
) ([]osmomath.Int, error) {
	return k.createOsmoMultihopExpectedSwapOuts(ctx, route, tokenOut, cumulativeRouteSwapFee, sumOfSwapFees)
}
<<<<<<< HEAD
=======

func (k Keeper) CalcTakerFeeExactIn(tokenIn sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	return k.calcTakerFeeExactIn(tokenIn, takerFee)
}

func (k Keeper) CalcTakerFeeExactOut(tokenOut sdk.Coin, takerFee osmomath.Dec) (sdk.Coin, sdk.Coin) {
	return k.calcTakerFeeExactOut(tokenOut, takerFee)
}

func (k Keeper) TrackVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	k.trackVolume(ctx, poolId, volumeGenerated)
}
>>>>>>> ca75f4c3 (refactor(deps): switch to cosmossdk.io/math from fork math (#6238))
