package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v16/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v16/x/poolmanager/types"
)

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.setParams(ctx, params)
}

// SetPool adds an existing pool to the keeper store.
func (k Keeper) SetPool(ctx sdk.Context, pool poolmanagertypes.PoolI) error {
	return k.setPool(ctx, pool)
}

func (k Keeper) SetStableSwapScalingFactors(ctx sdk.Context, poolId uint64, scalingFactors []uint64, sender string) error {
	return k.setStableSwapScalingFactors(ctx, poolId, scalingFactors, sender)
}

func AsCFMMPool(pool poolmanagertypes.PoolI) (types.CFMMPoolI, error) {
	return asCFMMPool(pool)
}

func (k Keeper) UnmarshalPoolLegacy(bz []byte) (poolmanagertypes.PoolI, error) {
	var acc poolmanagertypes.PoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}

func GetMaximalNoSwapLPAmount(ctx sdk.Context, pool types.CFMMPoolI, shareOutAmount sdk.Int) (neededLpLiquidity sdk.Coins, err error) {
	return getMaximalNoSwapLPAmount(ctx, pool, shareOutAmount)
}

func (k Keeper) RedirectDistributionRecord(ctx sdk.Context, cfmmPoolId, clPoolId uint64) error {
	return k.redirectDistributionRecord(ctx, cfmmPoolId, clPoolId)
}
