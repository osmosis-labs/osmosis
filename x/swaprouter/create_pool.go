package swaprouter

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/osmosis-labs/osmosis/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

func (k Keeper) validateCreatedPool(
	ctx sdk.Context,
	initialPoolLiquidity sdk.Coins,
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
	// Notably we use the initial pool liquidity at the start of the messages definition
	// just in case CreatePool was mutative.
	if !pool.GetTotalPoolLiquidity(ctx).IsEqual(initialPoolLiquidity) {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created, with initial liquidity not equal to what was specified.")
	}
	// This check can be removed later, and replaced with a minimum.
	if !pool.GetTotalShares().Equal(gammtypes.InitPoolSharesSupply) {
		return sdkerrors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect number of initial shares.")
	}
	return nil
}

func validateCreatePoolMsg(ctx sdk.Context, msg types.CreatePoolMsg) error {
	err := msg.Validate(ctx)
	if err != nil {
		return err
	}

	initialPoolLiquidity := msg.InitialLiquidity()
	numAssets := initialPoolLiquidity.Len()
	if numAssets < types.MinPoolAssets {
		return types.ErrTooFewPoolAssets
	}
	if numAssets > types.MaxPoolAssets {
		return sdkerrors.Wrapf(
			types.ErrTooManyPoolAssets,
			"pool has too many PoolAssets (%d)", numAssets,
		)
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
	err := validateCreatePoolMsg(ctx, msg)
	if err != nil {
		return 0, err
	}

	sender := msg.PoolCreator()
	initialPoolLiquidity := msg.InitialLiquidity()

	// send pool creation fee to community pool
	fee := k.GetParams(ctx).PoolCreationFee
	if err := k.communityPoolKeeper.FundCommunityPool(ctx, fee, sender); err != nil {
		return 0, err
	}

	poolId := k.getNextPoolIdAndIncrement(ctx)
	pool, err := msg.CreatePool(ctx, poolId)
	if err != nil {
		return 0, err
	}

	k.SetPoolRoute(ctx, poolId, msg.GetPoolType())

	if err := k.validateCreatedPool(ctx, initialPoolLiquidity, poolId, pool); err != nil {
		return 0, err
	}

	// create and save the pool's module account to the account keeper
	if err := osmoutils.CreateModuleAccount(ctx, k.accountKeeper, pool.GetAddress()); err != nil {
		return 0, fmt.Errorf("creating pool module account for id %d: %w", poolId, err)
	}

	// send initial liquidity to the pool
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), initialPoolLiquidity)
	if err != nil {
		return 0, err
	}

	swapModule := k.routes[msg.GetPoolType()]

	if err := swapModule.InitializePool(ctx, pool, sender); err != nil {
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

// GetSwapModule returns the swap module for the given pool ID.
// Returns error if:
// - any database error occurs.
// - fails to find a pool with the given id.
// - the swap module of the type corresponding to the pool id is not registered
// in swaprouter's keeper constructor.
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
