package poolmanager

import (
	"bytes"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

// validateCreatedPool checks that the pool was created with the correct pool ID and address.
func (k Keeper) validateCreatedPool(poolId uint64, pool types.PoolI) error {
	if pool.GetId() != poolId {
		return types.IncorrectPoolIdError{ExpectedPoolId: poolId, ActualPoolId: pool.GetId()}
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
	// Check that the pool type exists
	poolType := msg.GetPoolType()
	_, ok := k.routes[poolType]
	if !ok {
		return 0, types.InvalidPoolTypeError{PoolType: poolType}
	}

	// createPoolZeroLiquidityNoCreationFee contains shared pool creation logic between this function (CreatePool) and
	// CreateConcentratedPoolAsPoolManager. Despite the name, within this (CreatePool) function, we do charge a creation
	// fee and send initial liquidity to the pool's address. createPoolZeroLiquidityNoCreationFee is strictly used to reduce code duplication.
	pool, err := k.createPoolZeroLiquidityNoCreationFee(ctx, msg)
	if err != nil {
		return 0, err
	}

	sender := msg.PoolCreator()
	err = k.fundCommunityPoolIfNotWhitelisted(ctx, sender)
	if err != nil {
		return 0, err
	}

	// Send initial liquidity from pool creator to pool module account.
	initialPoolLiquidity := msg.InitialLiquidity()
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), initialPoolLiquidity)
	if err != nil {
		return 0, err
	}

	return pool.GetId(), nil
}

// CreateConcentratedPoolAsPoolManager creates a concentrated liquidity pool from given message without sending any initial liquidity to the pool
// and paying a creation fee. This is meant to be used for creating the pools internally (such as in the upgrade handler).
// The creator of the pool must be the poolmanager module account. Returns error if not. Otherwise, functions the same as
// the regular createPoolZeroLiquidityNoCreationFee.
func (k Keeper) CreateConcentratedPoolAsPoolManager(ctx sdk.Context, msg types.CreatePoolMsg) (types.PoolI, error) {
	// Validate that creator is the poolmanager module account as a sanity check.
	creator := msg.PoolCreator()
	poolmanagerModuleAccInterface := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if !poolmanagerModuleAccInterface.GetAddress().Equals(creator) {
		return nil, types.InvalidPoolCreatorError{CreatorAddresss: creator.String()}
	}

	// Disallow this for any pool type other than concentrated liquidity pool.
	// This can be further relaxed in the future.
	// The reason for this constraint is having balancer and stableswap pools mint gamm shares during InitializePool()
	// Module accounts cannot receive shares, so we cannot use this function for the above pool types without refactor.
	if msg.GetPoolType() != types.Concentrated {
		return nil, types.InvalidPoolTypeError{PoolType: msg.GetPoolType()}
	}

	return k.createPoolZeroLiquidityNoCreationFee(ctx, msg)
}

// createPoolZeroLiquidityNoCreationFee is an internal helper to create a pool from message with zero initial liquidity
// and no creation fee charged. It validates the message, gets the next pool ID, and creates the pool with the given pool ID and the desired type.
// It persists the module routing in state for future use, initializes the pool in its respective module, and emits a create pool event.
// Returns error if it fails to validate the pool creation message, fails to create a module account for the pool, or fails to initialize the pool.
// It is used by CreateConcentratedPoolAsPoolManager and CreatePool.
func (k Keeper) createPoolZeroLiquidityNoCreationFee(ctx sdk.Context, msg types.CreatePoolMsg) (types.PoolI, error) {
	// Run validate basic on the message.
	err := msg.Validate(ctx)
	if err != nil {
		return nil, err
	}

	// Get the next pool ID and increment the pool ID counter.
	poolId := k.getNextPoolIdAndIncrement(ctx)
	if poolId >= types.MaxPoolId {
		return nil, fmt.Errorf("next pool ID %d is greater than max pool ID %d", poolId, types.MaxPoolId)
	}

	poolType := msg.GetPoolType()

	// Create the pool with the given pool ID.
	pool, err := msg.CreatePool(ctx, poolId)
	if err != nil {
		return nil, err
	}

	// Store the pool ID to pool type mapping in state.
	k.SetPoolRoute(ctx, poolId, poolType)

	// Validates the pool address and pool ID stored match what was expected.
	if err := k.validateCreatedPool(poolId, pool); err != nil {
		return nil, err
	}

	// Run the respective pool type's initialization logic.
	swapModule := k.routes[poolType]
	if err := swapModule.InitializePool(ctx, pool, msg.PoolCreator()); err != nil {
		return nil, err
	}

	// Create and save the pool's module account to the account keeper.
	// This utilizes the pool address already created and validated in the previous steps.
	if err := osmoutils.CreateModuleAccount(ctx, k.accountKeeper, pool.GetAddress()); err != nil {
		return nil, fmt.Errorf("creating pool module account for id %d: %w", poolId, err)
	}

	emitCreatePoolEvents(ctx, poolId)
	return pool, nil
}

func emitCreatePoolEvents(ctx sdk.Context, poolId uint64) {
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(poolId, 10)),
		),
	})
}

// getNextPoolIdAndIncrement returns the next pool Id, and increments the corresponding state entry.
func (k Keeper) getNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPoolId(ctx)
	k.SetNextPoolId(ctx, nextPoolId+1)
	return nextPoolId
}

func (k *Keeper) SetPoolRoute(ctx sdk.Context, poolId uint64, poolType types.PoolType) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.FormatModuleRouteKey(poolId), &types.ModuleRoute{PoolType: poolType})
	k.cachedPoolModules.Delete(poolId)
}

type poolModuleCacheValue struct {
	pooltype types.PoolType
	module   types.PoolModuleI
	gasFlat  uint64
	gasKey   uint64
	gasValue uint64
}

func (k *Keeper) GetPoolType(ctx sdk.Context, poolId uint64) (types.PoolType, error) {
	poolModuleCandidate, cacheHit := k.cachedPoolModules.Load(poolId)
	if !cacheHit {
		_, err := k.GetPoolModule(ctx, poolId)
		if err != nil {
			return 0, err
		}
		poolModuleCandidate, _ = k.cachedPoolModules.Load(poolId)
	}
	v, _ := poolModuleCandidate.(poolModuleCacheValue)
	if cacheHit {
		osmoutils.ChargeMockReadGas(ctx, v.gasFlat, v.gasKey, v.gasValue)
	}

	return v.pooltype, nil
}

// GetPoolModule returns the swap module for the given pool ID.
// Returns error if:
// - any database error occurs.
// - fails to find a pool with the given id.
// - the swap module of the type corresponding to the pool id is not registered
// in poolmanager's keeper constructor.
// TODO: unexport after concentrated-liqudity upgrade. Currently, it is exported
// for the upgrade handler logic and tests.
func (k *Keeper) GetPoolModule(ctx sdk.Context, poolId uint64) (types.PoolModuleI, error) {
	poolModuleCandidate, ok := k.cachedPoolModules.Load(poolId)
	if ok {
		v, _ := poolModuleCandidate.(poolModuleCacheValue)
		osmoutils.ChargeMockReadGas(ctx, v.gasFlat, v.gasKey, v.gasValue)
		return v.module, nil
	}
	store := ctx.KVStore(k.storeKey)

	moduleRoute := &types.ModuleRoute{}
	found, gasFlat, gasKey, gasVal, err := osmoutils.TrackGasUsedInGet(store, types.FormatModuleRouteKey(poolId), moduleRoute)
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

	k.cachedPoolModules.Store(poolId, poolModuleCacheValue{
		pooltype: moduleRoute.PoolType,
		module:   swapModule,
		gasFlat:  gasFlat,
		gasKey:   gasKey,
		gasValue: gasVal,
	})

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

// fundCommunityPoolIfNotWhitelisted checks if the sender is in the UnrestrictedPoolCreatorWhitelist
// and funds the community pool if the sender is not whitelisted.
func (k Keeper) fundCommunityPoolIfNotWhitelisted(ctx sdk.Context, sender sdk.AccAddress) error {
	// Get the UnrestrictedPoolCreatorWhitelist from the concentrated pool module
	whitelist := k.concentratedKeeper.GetWhitelistedAddresses(ctx)

	// Check if sender is in the whitelist
	isWhitelisted := false
	for _, addr := range whitelist {
		if addr == sender.String() {
			isWhitelisted = true
			break
		}
	}

	// Run FundCommunityPool logic if sender is not in the whitelist
	if !isWhitelisted {
		poolCreationFee := k.GetParams(ctx).PoolCreationFee
		if err := k.communityPoolKeeper.FundCommunityPool(ctx, poolCreationFee, sender); err != nil {
			return err
		}
	}

	return nil
}
