package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v13/x/protorev/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/protorev keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params queries the parameters of the module.
func (q Querier) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: q.Keeper.GetParams(ctx)}, nil
}

// GetProtoRevNumberOfTrades queries the number of trades the module has executed
func (q Querier) GetProtoRevNumberOfTrades(c context.Context, req *types.QueryGetProtoRevNumberOfTradesRequest) (*types.QueryGetProtoRevNumberOfTradesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	numberOfTrades, err := q.Keeper.GetNumberOfTrades(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevNumberOfTradesResponse{NumberOfTrades: numberOfTrades}, nil
}

// GetProtoRevProfitsByDenom queries the profits of the module by denom
func (q Querier) GetProtoRevProfitsByDenom(c context.Context, req *types.QueryGetProtoRevProfitsByDenomRequest) (*types.QueryGetProtoRevProfitsByDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	profits, err := q.Keeper.GetProfitsByDenom(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevProfitsByDenomResponse{Profit: &profits}, nil
}

// GetProtoRevAllProfits queries all of the profits from the module
func (q Querier) GetProtoRevAllProfits(c context.Context, req *types.QueryGetProtoRevAllProfitsRequest) (*types.QueryGetProtoRevAllProfitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	profits := q.Keeper.GetAllProfits(ctx)

	return &types.QueryGetProtoRevAllProfitsResponse{Profits: profits}, nil
}

// GetProtoRevStatisticsByRoute queries the number of arbitrages and profits
// that have been executed for a given route
func (q Querier) GetProtoRevStatisticsByRoute(c context.Context, req *types.QueryGetProtoRevStatisticsByRouteRequest) (*types.QueryGetProtoRevStatisticsByRouteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// Query information from the keeper
	numberOfTrades, err := q.Keeper.GetTradesByRoute(ctx, req.Route)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	profits := q.Keeper.GetAllProfitsByRoute(ctx, req.Route)

	// Wrap the information into a response
	statistics := types.RouteStatistics{
		NumberOfTrades: numberOfTrades,
		Profits:        profits,
		Route:          req.Route,
	}
	return &types.QueryGetProtoRevStatisticsByRouteResponse{Statistics: &statistics}, nil
}

// GetProtoRevAllStatistics queries all of pools that the module has arbitrage
// against and the number of trades executed on each pool and the total profits for each pool
func (q Querier) GetProtoRevAllRouteStatistics(c context.Context, req *types.QueryGetProtoRevAllRouteStatisticsRequest) (*types.QueryGetProtoRevAllRouteStatisticsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	routes, err := q.Keeper.GetAllRoutes(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(routes) == 0 {
		return nil, status.Error(codes.Internal, "no routes found")
	}

	statistics := make([]types.RouteStatistics, len(routes))
	for index, route := range routes {
		numberOfTrades, err := q.Keeper.GetTradesByRoute(ctx, route)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		profits := q.Keeper.GetAllProfitsByRoute(ctx, route)

		statistics[index] = types.RouteStatistics{
			NumberOfTrades: numberOfTrades,
			Profits:        profits,
			Route:          route,
		}
	}

	return &types.QueryGetProtoRevAllRouteStatisticsResponse{Statistics: statistics}, nil
}

// GetProtoRevTokenPairArbRoutes queries the hot routes that the module is utilizing for cyclic arbitrage route generation
func (q Querier) GetProtoRevTokenPairArbRoutes(c context.Context, req *types.QueryGetProtoRevTokenPairArbRoutesRequest) (*types.QueryGetProtoRevTokenPairArbRoutesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	routes, err := q.Keeper.GetAllTokenPairArbRoutes(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevTokenPairArbRoutesResponse{Routes: routes}, nil
}

// GetProtoRevMaxRoutesPerTx queries the maximum number of routes that can be executed in a single transaction
func (q Querier) GetProtoRevMaxRoutesPerTx(c context.Context, req *types.QueryGetProtoRevMaxRoutesPerTxRequest) (*types.QueryGetProtoRevMaxRoutesPerTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	maxRoutesPerTx, err := q.Keeper.GetMaxRoutesPerTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevMaxRoutesPerTxResponse{MaxRoutesPerTx: maxRoutesPerTx}, nil
}

// GetProtoRevMaxRoutesPerBlock queries the maximum number of routes that can be executed in a single block
func (q Querier) GetProtoRevMaxRoutesPerBlock(c context.Context, req *types.QueryGetProtoRevMaxRoutesPerBlockRequest) (*types.QueryGetProtoRevMaxRoutesPerBlockResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	maxRoutesPerBlock, err := q.Keeper.GetMaxRoutesPerBlock(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevMaxRoutesPerBlockResponse{MaxRoutesPerBlock: maxRoutesPerBlock}, nil
}

// GetProtoRevAdminAccount queries the admin account that is allowed to execute admin functions
func (q Querier) GetProtoRevAdminAccount(c context.Context, req *types.QueryGetProtoRevAdminAccountRequest) (*types.QueryGetProtoRevAdminAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	adminAccount, err := q.Keeper.GetAdminAccount(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevAdminAccountResponse{AdminAccount: adminAccount.String()}, nil
}

// GetProtoRevDeveloperAccount queries the developer account that is accumulating the profits from the module
func (q Querier) GetProtoRevDeveloperAccount(c context.Context, req *types.QueryGetProtoRevDeveloperAccountRequest) (*types.QueryGetProtoRevDeveloperAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	developerAccount, err := q.Keeper.GetDeveloperAccount(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevDeveloperAccountResponse{DeveloperAccount: developerAccount.String()}, nil
}
