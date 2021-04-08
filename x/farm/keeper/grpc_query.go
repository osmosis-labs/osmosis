package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/c-osmosis/osmosis/x/farm/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Farms(ctx context.Context, req *types.QueryFarmsRequest) (*types.QueryFarmsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	store := sdkCtx.KVStore(k.storeKey)
	farmsStore := prefix.NewStore(store, types.FarmPrefix)

	var farms []types.Farm

	pageRes, err := query.Paginate(farmsStore, req.Pagination, func(key []byte, value []byte) error {
		farm := types.Farm{}
		err := k.cdc.UnmarshalBinaryBare(value, &farm)

		if err != nil {
			return err
		}

		farms = append(farms, farm)

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFarmsResponse{
		Farms:      farms,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Farm(ctx context.Context, req *types.QueryFarmRequest) (*types.QueryFarmResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	farm, err := k.GetFarm(sdkCtx, req.FarmId)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFarmResponse{
		Farm: farm,
	}, nil
}

func (k Keeper) Farmers(ctx context.Context, req *types.QueryFarmersRequest) (*types.QueryFarmersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	store := sdkCtx.KVStore(k.storeKey)
	farmersStore := prefix.NewStore(store, append(types.FarmerPrefix, sdk.Uint64ToBigEndian(req.FarmId)...))

	var farmers []types.Farmer

	pageRes, err := query.Paginate(farmersStore, req.Pagination, func(key []byte, value []byte) error {
		farmer := types.Farmer{}
		err := k.cdc.UnmarshalBinaryBare(value, &farmer)

		if err != nil {
			return err
		}

		farmers = append(farmers, farmer)

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFarmersResponse{
		Farmers:    farmers,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Farmer(ctx context.Context, req *types.QueryFarmerRequest) (*types.QueryFarmerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	farmer, err := k.GetFarmer(sdkCtx, req.FarmId, addr)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFarmerResponse{
		Farmer: farmer,
	}, nil
}

func (k Keeper) PendingRewards(ctx context.Context, req *types.QueryPendingRewardsRequest) (*types.QueryPendingRewardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	decRewards, err := k.CalculatePendingRewards(sdkCtx, req.FarmId, address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPendingRewardsResponse{Rewards: decRewards}, nil
}
