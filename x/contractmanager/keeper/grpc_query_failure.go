package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v20/x/contractmanager/types"
)

const FailuresQueryMaxLimit uint64 = query.DefaultLimit

func (k Keeper) Failures(c context.Context, req *types.QueryFailuresRequest) (*types.QueryFailuresResponse, error) {
	return k.AddressFailures(c, req)
}

func (k Keeper) AddressFailures(c context.Context, req *types.QueryFailuresRequest) (*types.QueryFailuresResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	pagination := req.GetPagination()
	if pagination != nil && pagination.Limit > FailuresQueryMaxLimit {
		return nil, status.Errorf(codes.InvalidArgument, "limit is more than maximum allowed (%d > %d)", pagination.Limit, FailuresQueryMaxLimit)
	}

	var failures []types.Failure
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)

	var failureStore prefix.Store
	if req.Address != "" {
		if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse address: %s", req.Address)
		}
		failureStore = prefix.NewStore(store, types.GetFailureKeyPrefix(req.Address))
	} else {
		failureStore = prefix.NewStore(store, types.ContractFailuresKey)
	}

	pageRes, err := query.Paginate(failureStore, req.Pagination, func(key, value []byte) error {
		var failure types.Failure
		if err := k.cdc.Unmarshal(value, &failure); err != nil {
			return err
		}

		failures = append(failures, failure)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFailuresResponse{Failures: failures, Pagination: pageRes}, nil
}
