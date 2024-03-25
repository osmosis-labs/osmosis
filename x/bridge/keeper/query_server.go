package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/vv24/x/bridge/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	k Keeper
}

// NewQueryServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{k: keeper}
}

func (q queryServer) Params(
	goCtx context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{
		Params: q.k.GetParams(ctx),
	}, nil
}

func (q queryServer) LastTransferHeight(
	goCtx context.Context,
	req *types.LastTransferHeightRequest,
) (*types.LastTransferHeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	height, err := q.k.GetLastTransferHeight(ctx, req.AssetId)
	if err != nil {
		return nil, err
	}

	return &types.LastTransferHeightResponse{LastTransferHeight: height}, nil
}
