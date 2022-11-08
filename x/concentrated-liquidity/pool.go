package concentrated_liquidity

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	pooltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/concentrated-pool"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

// TODO: spec, tests, implementation
func (k Keeper) InitializePool(ctx sdk.Context, pool gammtypes.PoolI, creatorAddress sdk.AccAddress) error {
	panic("not implemented")
}

func (k Keeper) CreateNewConcentratedLiquidityPool(ctx sdk.Context, poolId uint64, denom0, denom1 string, currSqrtPrice sdk.Dec, currTick sdk.Int) (types.PoolI, error) {
	denom0, denom1, err := types.OrderInitialPoolDenoms(denom0, denom1)
	if err != nil {
		return &pooltypes.Pool{}, err
	}
	pool := pooltypes.Pool{
		// TODO: move gammtypes.NewPoolAddress(poolId) to swaproutertypes
		Address:          gammtypes.NewPoolAddress(poolId).String(),
		Id:               poolId,
		CurrentSqrtPrice: currSqrtPrice,
		CurrentTick:      currTick,
		Liquidity:        sdk.ZeroDec(),
		Token0:           denom0,
		Token1:           denom1,
	}

	err = k.setPool(ctx, &pool)
	if err != nil {
		return nil, err
	}

	return &pool, nil
}

// GetPool returns a pool with a given id.
func (k Keeper) GetPool(ctx sdk.Context, poolId uint64) (gammtypes.PoolI, error) {
	return nil, errors.New("not implemented")
}
