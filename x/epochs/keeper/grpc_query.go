package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/epochs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Keeper{}

// Epochs provide running epochs
func (k Keeper) Epochs(c context.Context, _ *types.QueryEpochsRequest) (*types.QueryEpochsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryEpochsResponse{
		Epochs: k.AllEpochInfos(ctx),
	}, nil
}
