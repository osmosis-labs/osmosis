package swaprouter

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
)

// CreatePool attempts to create a pool returning the newly created pool ID or
// an error upon failure. The pool creation fee is used to fund the community
// pool. It will create a dedicated module account for the pool and sends the
// initial liquidity to the created module account.
//
// After the initial liquidity is sent to the pool's account, shares are minted
// and sent to the pool creator. The shares are created using a denomination in
// the form of <swap module name>/pool/{poolID}. In addition, the x/bank metadata is updated
// to reflect the newly created GAMM share denomination.
func (k Keeper) CreatePool(ctx sdk.Context, msg types.CreatePoolMsg) (uint64, error) {
	err := validateCreatePoolMsg(ctx, msg)
	if err != nil {
		return 0, err
	}

	sender := msg.PoolCreator()
	initialPoolLiquidity := msg.InitialLiquidity()

	// send pool creation fee to community pool
	params := k.GetParams(ctx)
	if err := k.communityPoolKeeper.FundCommunityPool(ctx, params.PoolCreationFee, sender); err != nil {
		return 0, err
	}

	poolId := k.getNextPoolIdAndIncrement(ctx)
	pool, err := msg.CreatePool(ctx, poolId)
	if err != nil {
		return 0, err
	}

	k.SetModuleRoute(ctx, poolId, msg.GetPoolType())

	if err := k.validateCreatedPool(ctx, initialPoolLiquidity, poolId, pool); err != nil {
		return 0, err
	}

	// create and save the pool's module account to the account keeper
	if err := osmoutils.CreateModuleAccount(ctx, k.accountKeeper, pool.GetAddress()); err != nil {
		return 0, fmt.Errorf("creating pool module account for id %d: %w", poolId, err)
	}

	swapModule := k.routes[msg.GetPoolType()]

	if err := swapModule.InitializePool(ctx, pool, sender); err != nil {
		return 0, err
	}

	// send initial liquidity to the pool
	err = k.bankKeeper.SendCoins(ctx, sender, pool.GetAddress(), initialPoolLiquidity)
	if err != nil {
		return 0, err
	}

	k.poolCreationListeners.AfterPoolCreated(ctx, sender, pool.GetId())

	return pool.GetId(), nil
}

// getNextPoolIdAndIncrement returns the next pool Id, and increments the corresponding state entry.
func (k Keeper) getNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPoolId(ctx)
	k.SetNextPoolId(ctx, nextPoolId+1)
	return nextPoolId
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
		return errors.Wrapf(
			types.ErrTooManyPoolAssets,
			"pool has too many PoolAssets (%d)", numAssets,
		)
	}
	return nil
}

func (k Keeper) validateCreatedPool(
	ctx sdk.Context,
	initialPoolLiquidity sdk.Coins,
	poolId uint64,
	pool gammtypes.TraditionalAmmInterface,
) error {
	if pool.GetId() != poolId {
		return errors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool ID.")
	}
	if !pool.GetAddress().Equals(gammtypes.NewPoolAddress(poolId)) {
		return errors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool address.")
	}
	// Notably we use the initial pool liquidity at the start of the messages definition
	// just in case CreatePool was mutative.
	if !pool.GetTotalPoolLiquidity(ctx).IsEqual(initialPoolLiquidity) {
		return errors.Wrap(types.ErrInvalidPool,
			"Pool was attempted to be created, with initial liquidity not equal to what was specified.")
	}
	// TODO: this check should be moved
	// This check can be removed later, and replaced with a minimum.
	if !pool.GetTotalShares().Equal(gammtypes.InitPoolSharesSupply) {
		return errors.Wrap(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect number of initial shares.")
	}
	return nil
}

// SetModuleRoute stores the mapping from poolId to the given pool type.
// TODO: unexport after concentrated-liqudity upgrade. Currently, it is exported
// for the upgrade handler logic and tests.
func (k Keeper) SetModuleRoute(ctx sdk.Context, poolId uint64, poolType types.PoolType) {
	store := ctx.KVStore(k.storeKey)
	osmoutils.MustSet(store, types.FormatModuleRouteKey(poolId), &types.ModuleRoute{PoolType: poolType})
}
