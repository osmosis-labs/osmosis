package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

func (k Keeper) LBPs(goCtx context.Context, q *api.QueryLBPs) (*api.QueryLBPsResponse, error) {
	return nil, nil
}

func (k Keeper) LBP(goCtx context.Context, q *api.QueryLBP) (*api.QueryLBPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	lbp, _, err := k.getLBP(store, q.LbpId)
	return &api.QueryLBPResponse{lbp}, err
}

func (k Keeper) UserPosition(goCtx context.Context, q *api.QueryUserPosition) (*api.QueryUserPositionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := ctx.KVStore(k.storeKey)
	user, err := sdk.AccAddressFromBech32(q.User)
	if err != nil {
		return nil, err
	}
	poolId := storeIntIdKey(q.LbpId)
	up, err := k.getUserPosition(store, poolId, user, false)
	if err != nil {
		return nil, err
	}
	return &api.QueryUserPositionResponse{up}, nil
}
