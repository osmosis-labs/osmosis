package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// SetPool adds an existing pool to the keeper store.
func (k Keeper) SetPool(ctx sdk.Context, pool swaproutertypes.PoolI) error {
	return k.setPool(ctx, pool)
}
