package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.setParams(ctx, params)
}

// SetPool adds an existing pool to the keeper store.
func (k Keeper) SetPool(ctx sdk.Context, pool swaproutertypes.PoolI) error {
	return k.setPool(ctx, pool)
}

func (k Keeper) SetStableSwapScalingFactors(ctx sdk.Context, poolId uint64, scalingFactors []uint64, sender string) error {
	return k.setStableSwapScalingFactors(ctx, poolId, scalingFactors, sender)
}

func ConvertToCFMMPool(pool swaproutertypes.PoolI) (types.CFMMPoolI, error) {
	return convertToCFMMPool(pool)
}

func (k Keeper) UnmarshalPoolLegacy(bz []byte) (swaproutertypes.PoolI, error) {
	var acc swaproutertypes.PoolI
	return acc, k.cdc.UnmarshalInterface(bz, &acc)
}
