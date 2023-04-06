package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
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

func ConvertToCFMMPool(pool poolmanagertypes.PoolI) (types.CFMMPoolI, error) {
	return convertToCFMMPool(pool)
}

func (k Keeper) UnmarshalPoolLegacy(bz []byte) (poolmanagertypes.PoolI, error) {
	var acc poolmanagertypes.PoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}
