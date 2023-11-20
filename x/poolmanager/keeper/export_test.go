package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

func (k Keeper) GetNextPoolIdAndIncrement(ctx sdk.Context) uint64 {
	return k.getNextPoolIdAndIncrement(ctx)
}

func (k Keeper) GetOsmoRoutedMultihopTotalSwapFee(ctx sdk.Context, route types.MultihopRoute) (
	totalPathSwapFee sdk.Dec, sumOfSwapFees sdk.Dec, err error) {
	return k.getOsmoRoutedMultihopTotalSwapFee(ctx, route)
}
