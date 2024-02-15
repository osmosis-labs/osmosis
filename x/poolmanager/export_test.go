package poolmanager

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v23/x/poolmanager/types"
)

var IntMaxValue = intMaxValue

func (k Keeper) GetNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolIdAndIncrement(ctx)
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

func (k Keeper) CreateMultihopExpectedSwapOuts(
	ctx sdk.Context,
	route []types.SwapAmountOutRoute,
	tokenOut sdk.Coin,
) ([]osmomath.Int, error) {
	return k.createMultihopExpectedSwapOuts(ctx, route, tokenOut)
}

func (k Keeper) TrackVolume(ctx sdk.Context, poolId uint64, volumeGenerated sdk.Coin) {
	k.trackVolume(ctx, poolId, volumeGenerated)
}

func (k Keeper) ChargeTakerFee(ctx sdk.Context, tokenIn sdk.Coin, tokenOutDenom string, sender sdk.AccAddress, exactIn bool) (sdk.Coin, error) {
	return k.chargeTakerFee(ctx, tokenIn, tokenOutDenom, sender, exactIn)
}

func FindRoutes(g types.RoutingGraphMap, start, end string, hops int) [][]*types.Route {
	return findRoutes(g, start, end, hops)
}

func (k Keeper) GetDirectRouteWithMostLiquidity(ctx sdk.Context, inputDenom, outputDenom string, routeMap types.RoutingGraphMap) (uint64, error) {
	return k.getDirectRouteWithMostLiquidity(ctx, inputDenom, outputDenom, routeMap)
}

func (k Keeper) InputAmountToTargetDenom(ctx sdk.Context, inputDenom, targetDenom string, amount osmomath.Int, routeMap types.RoutingGraphMap) (osmomath.Int, error) {
	return k.inputAmountToTargetDenom(ctx, inputDenom, targetDenom, amount, routeMap)
}

func (k Keeper) GetPoolLiquidityOfDenom(ctx sdk.Context, poolId uint64, denom string) (osmomath.Int, error) {
	return k.getPoolLiquidityOfDenom(ctx, poolId, denom)
}

func ConvertToMap(routingGraph *types.RoutingGraph) types.RoutingGraphMap {
	return convertToMap(routingGraph)
}

func (k Keeper) PoolLiquidityToTargetDenom(ctx sdk.Context, pool types.PoolI, routeMap types.RoutingGraphMap, targetDenom string) (osmomath.Int, error) {
	return k.poolLiquidityToTargetDenom(ctx, pool, routeMap, targetDenom)
}

func (k Keeper) PoolLiquidityFromOSMOToTargetDenom(ctx sdk.Context, pool types.PoolI, routeMap types.RoutingGraphMap, targetDenom string) (osmomath.Int, error) {
	return k.poolLiquidityFromOSMOToTargetDenom(ctx, pool, routeMap, targetDenom)
}

func (k Keeper) GenerateAllDenomPairRoutes(ctx sdk.Context) ([]types.PoolI, types.RoutingGraphMap, error) {
	return k.generateAllDenomPairRoutes(ctx)
}
