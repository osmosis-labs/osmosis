package poolmanager

import (
	"bytes"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (k Keeper) validateCreatedPool(
	ctx sdk.Context,
	poolId uint64,
	pool types.PoolI,
) error {
	if pool.GetId() != poolId {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool ID.")
	}
	if !pool.GetAddress().Equals(gammtypes.NewPoolAddress(poolId)) {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool address.")
	}
	return nil
}

// CreatePool attempts to create a pool returning the newly created pool ID or
// an error upon failure. The pool creation fee is used to fund the community
// pool. It will create a dedicated module account for the pool and sends the
// initial liquidity to the created module account.
//
// After the initial liquidity is sent to the pool's account, this function calls an
// InitializePool function from the source module. That module is responsible for:
// - saving the pool into its own state
// - Minting LP shares to pool creator
// - Setting metadata for the shares
func (k Keeper) CreatePool(ctx sdk.Context, msg types.CreatePoolMsg) (uint64, error) {
	// Run validate basic on the message.
	err := msg.Validate(ctx)
	if err != nil {
		return 0, err
	}

	// Send pool creation fee to community pool
	params := k.GetParams(ctx)
	sender := msg.PoolCreator()
	if err := k.communityPoolKeeper.FundCommunityPool(ctx, params.PoolCreationFee, sender); err != nil {
		return 0, err
	}

	// Get the next pool ID and increment the pool ID counter
	// Create the pool with the given pool ID
	poolId := k.getNextPoolIdAndIncrement(ctx)
	pool, err := msg.CreatePool(ctx, poolId)
	if err != nil {
		return 0, err
	}

	k.SetPoolRoute(ctx, poolId, msg.GetPoolType())

	if err := k.validateCreatedPool(ctx, poolId, pool); err != nil {
		return 0, err
	}

	// create and save the pool's module account to the account keeper
	if err := osmoutils.CreateModuleAccount(ctx, k.accountKeeper, pool.GetAddress()); err != nil {
		return 0, fmt.Errorf("creating pool module account for id %d: %w", poolId, err)
	}

	// Run the respective pool type's initialization logic.
	swapModule := k.routes[msg.GetPoolType()]
	if err := swapModule.InitializePool(ctx, pool, sender); err != nil {
		return 0, err
	}

	// Send initial liquidity to the pool's address.
	initialPoolLiquidity := msg.InitialLiquidity()
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), initialPoolLiquidity)
	if err != nil {
		return 0, err
	}

	emitCreatePoolEvents(ctx, poolId, msg)
	return pool.GetId(), nil
}

func emitCreatePoolEvents(ctx sdk.Context, poolId uint64, msg types.CreatePoolMsg) {
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			gammtypes.TypeEvtPoolCreated,
			sdk.NewAttribute(gammtypes.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.PoolCreator().String()),
		),
	})
}

// getNextPoolIdAndIncrement returns the next pool Id, and increments the corresponding state entry.
func (k Keeper) getNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPoolId(ctx)
	k.SetNextPoolId(ctx, nextPoolId+1)
	return nextPoolId
}

func (k Keeper) SetPoolRoute(ctx sdk.Context, poolId uint64, poolType types.PoolType) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.FormatModuleRouteKey(poolId), &types.ModuleRoute{PoolType: poolType})
}

// GetPoolModule returns the swap module for the given pool ID.
// Returns error if:
// - any database error occurs.
// - fails to find a pool with the given id.
// - the swap module of the type corresponding to the pool id is not registered
// in poolmanager's keeper constructor.
// TODO: unexport after concentrated-liqudity upgrade. Currently, it is exported
// for the upgrade handler logic and tests.
func (k Keeper) GetPoolModule(ctx sdk.Context, poolId uint64) (types.SwapI, error) {
	store := ctx.KVStore(k.storeKey)

	moduleRoute := &types.ModuleRoute{}
	found, err := osmoutils.Get(store, types.FormatModuleRouteKey(poolId), moduleRoute)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, types.FailedToFindRouteError{PoolId: poolId}
	}

	swapModule, routeExists := k.routes[moduleRoute.PoolType]
	if !routeExists {
		return nil, types.UndefinedRouteError{PoolType: moduleRoute.PoolType, PoolId: poolId}
	}

	return swapModule, nil
}

// getAllPoolRoutes returns all pool routes from state.
func (k Keeper) getAllPoolRoutes(ctx sdk.Context) []types.ModuleRoute {
	store := ctx.KVStore(k.storeKey)
	moduleRoutes, err := osmoutils.GatherValuesFromStorePrefixWithKeyParser(store, types.SwapModuleRouterPrefix, parsePoolRouteWithKey)
	if err != nil {
		panic(err)
	}
	return moduleRoutes
}

// parsePoolRouteWithKey parses pool route by grabbing the pool id from key
// and the pool type from value. Returns error if parsing fails.
func parsePoolRouteWithKey(key []byte, value []byte) (types.ModuleRoute, error) {
	poolIdBytes := bytes.TrimLeft(key, string(types.SwapModuleRouterPrefix))
	poolId, err := strconv.ParseUint(string(poolIdBytes), 10, 64)
	if err != nil {
		return types.ModuleRoute{}, err
	}
	parsedValue, err := types.ParseModuleRouteFromBz(value)
	if err != nil {
		panic(err)
	}
	parsedValue.PoolId = poolId
	return parsedValue, nil
}
