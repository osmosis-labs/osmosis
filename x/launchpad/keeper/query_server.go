package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/x/launchpad/api"
)

func (k Keeper) Sales(goCtx context.Context, q *api.QuerySales) (*api.QuerySalesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	store := k.saleStore(sdkCtx.KVStore(k.storeKey))

	var sales []api.Sale
	pageRes, err := query.Paginate(store, q.Pagination, func(_, value []byte) error {
		var p api.Sale
		err := k.cdc.Unmarshal(value, &p)
		if err != nil {
			return err
		}
		sales = append(sales, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &api.QuerySalesResponse{sales, pageRes}, nil
}

func (k Keeper) Sale(ctx context.Context, q *api.QuerySale) (*api.QuerySaleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)
	s, _, err := k.getSale(store, q.SaleId)
	return &api.QuerySaleResponse{s}, err
}

func (k Keeper) UserPosition(ctx context.Context, q *api.QueryUserPosition) (*api.QueryUserPositionResponse, error) {
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
	return &api.QueryUserPositionResponse{up}, nil
}
