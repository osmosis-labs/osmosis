package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v10/x/streamswap/types"
)

func (k Keeper) Sales(goCtx context.Context, q *types.QuerySales) (*types.QuerySalesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	store := k.saleStore(sdkCtx.KVStore(k.storeKey))

	var sales []types.Sale
	pageRes, err := query.Paginate(store, q.Pagination, func(_, value []byte) error {
		var s types.Sale
		err := k.cdc.Unmarshal(value, &s)
		if err != nil {
			return err
		}
		sales = append(sales, s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &types.QuerySalesResponse{sales, pageRes}, nil
}

func (k Keeper) Sale(ctx context.Context, q *types.QuerySale) (*types.QuerySaleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	s, _, err := k.getSale(store, q.SaleId)
	return &types.QuerySaleResponse{s}, err
}

func (k Keeper) UserPosition(ctx context.Context, q *types.QueryUserPosition) (*types.QueryUserPositionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	user, err := sdk.AccAddressFromBech32(q.User)
	if err != nil {
		return nil, err
	}
	saleId := storeIntIdKey(q.SaleId)
	up, err := k.getUserPosition(store, saleId, user, false)
	if err != nil {
		return nil, err
	}
	return &types.QueryUserPositionResponse{up}, nil
}
