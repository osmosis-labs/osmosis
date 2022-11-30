package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/x/gamm/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v13/x/swaprouter/types"
)

// SetPool adds an existing pool to the keeper store.
func (k Keeper) SetPool(ctx sdk.Context, pool swaproutertypes.PoolI) error {
	return k.setPool(ctx, pool)
}

func ConvertToCFMMPool(pool swaproutertypes.PoolI) (types.CFMMPoolI, error) {
	return convertToCFMMPool(pool)
}
