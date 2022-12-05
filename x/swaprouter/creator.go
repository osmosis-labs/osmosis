package swaprouter

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	gammtypes "github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
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
	err := msg.Validate(ctx)
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

	poolType := msg.GetPoolType()
	// run validation for all pool types
	// total shares and liquidity is not checked for concentrated pools because they are initialized with no liquidity
	if err := k.validateCreatedPool(ctx, initialPoolLiquidity, poolId, pool, poolType); err != nil {
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

	// TODO: Add AfterCFMMPoolCreated hook so we can remove this if statement
	// https://github.com/osmosis-labs/osmosis/issues/3612
	if poolType != types.Concentrated {
		k.poolCreationListeners.AfterPoolCreated(ctx, sender, pool.GetId())
	}

	return pool.GetId(), nil
}

// getNextPoolIdAndIncrement returns the next pool Id, and increments the corresponding state entry.
func (k Keeper) getNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	nextPoolId := k.GetNextPoolId(ctx)
	k.SetNextPoolId(ctx, nextPoolId+1)
	return nextPoolId
}

func (k Keeper) validateCreatedPool(
	ctx sdk.Context,
	initialPoolLiquidity sdk.Coins,
	poolId uint64,
	pool types.PoolI,
	poolType types.PoolType,
) error {
	if pool.GetId() != poolId {
		return errors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool ID.")
	}
	if !pool.GetAddress().Equals(gammtypes.NewPoolAddress(poolId)) {
		return errors.Wrapf(types.ErrInvalidPool,
			"Pool was attempted to be created with incorrect pool address.")
	}

	// Check the total pool liquidity/shares if pool is not a concentrated liquidity pool
	if poolType != types.Concentrated {
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
