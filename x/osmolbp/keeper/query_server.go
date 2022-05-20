package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/x/osmolbp/api"
)

func (k Keeper) LBPs(goCtx context.Context, q *api.QueryLBPs) (*api.QueryLBPsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	store := k.lbpStore(sdkCtx.KVStore(k.storeKey))

	var lbps []api.LBP
	pageRes, err := query.Paginate(store, q.Pagination, func(_, value []byte) error {
		var p api.LBP
		err := k.cdc.Unmarshal(value, &p)
		if err != nil {
			return err
		}
		lbps = append(lbps, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &api.QueryLBPsResponse{lbps, pageRes}, nil
}

func (k Keeper) LBP(ctx context.Context, q *api.QueryLBP) (*api.QueryLBPResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	lbp, _, err := k.getLBP(store, q.LbpId)
	return &api.QueryLBPResponse{lbp}, err
}

func (k Keeper) UserPosition(ctx context.Context, q *api.QueryUserPosition) (*api.QueryUserPositionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
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

func (k Keeper) LBP(ctx context.Context, q *api.QueryLBP) (*api.QueryLBPResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	lbp, _, err := k.getLBP(store, q.LbpId)
	return &api.QueryLBPResponse{lbp}, err
}

func (k Keeper) UserPosition(ctx context.Context, q *api.QueryUserPosition) (*api.QueryUserPositionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
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
