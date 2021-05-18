package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/epochs/types"
)

var _ types.QueryServer = Keeper{}

// EpochInfos provide running epochInfos
func (k Keeper) EpochInfos(c context.Context, _ *types.QueryEpochsInfoRequest) (*types.QueryEpochsInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryEpochsInfoResponse{
		Epochs: k.AllEpochInfos(ctx),
	}, nil
}
