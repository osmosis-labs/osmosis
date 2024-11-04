package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
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
	return &types.QueryGetProtoRevStatisticsByRouteResponse{Statistics: statistics}, nil
}

// GetProtoRevAllRouteStatistics queries all of routes that the module has arbitrage
// against and the number of trades executed on each route and the total profits for each route
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

// GetProtoRevAdminAccount queries the admin account that is allowed to execute admin functions
func (q Querier) GetProtoRevAdminAccount(c context.Context, req *types.QueryGetProtoRevAdminAccountRequest) (*types.QueryGetProtoRevAdminAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	adminAccount := q.Keeper.GetAdminAccount(ctx)

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

// GetProtoRevInfoByPoolType queries information pertaining to each pool type the module is using for arbitrage
func (q Querier) GetProtoRevInfoByPoolType(c context.Context, req *types.QueryGetProtoRevInfoByPoolTypeRequest) (*types.QueryGetProtoRevInfoByPoolTypeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	infoByPoolType := q.Keeper.GetInfoByPoolType(ctx)

	return &types.QueryGetProtoRevInfoByPoolTypeResponse{InfoByPoolType: infoByPoolType}, nil
}

// GetProtoRevPoolPointsPerTx queries the maximum number of pool points that can be consumed per transaction
func (q Querier) GetProtoRevMaxPoolPointsPerTx(c context.Context, req *types.QueryGetProtoRevMaxPoolPointsPerTxRequest) (*types.QueryGetProtoRevMaxPoolPointsPerTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	poolPointsPerTx, err := q.Keeper.GetMaxPointsPerTx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevMaxPoolPointsPerTxResponse{MaxPoolPointsPerTx: poolPointsPerTx}, nil
}

// GetProtoRevPoolPointsPerBlock queries the maximum number of pool points that can be consumed per block
func (q Querier) GetProtoRevMaxPoolPointsPerBlock(c context.Context, req *types.QueryGetProtoRevMaxPoolPointsPerBlockRequest) (*types.QueryGetProtoRevMaxPoolPointsPerBlockResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	poolPointsPerBlock, err := q.Keeper.GetMaxPointsPerBlock(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevMaxPoolPointsPerBlockResponse{MaxPoolPointsPerBlock: poolPointsPerBlock}, nil
}

// GetProtoRevBaseDenoms queries the base denoms that are being used for arbitrage
func (q Querier) GetProtoRevBaseDenoms(c context.Context, req *types.QueryGetProtoRevBaseDenomsRequest) (*types.QueryGetProtoRevBaseDenomsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	baseDenoms, err := q.Keeper.GetAllBaseDenoms(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevBaseDenomsResponse{BaseDenoms: baseDenoms}, nil
}

// GetProtoRevEnabled queries whether the module is enabled or not
func (q Querier) GetProtoRevEnabled(c context.Context, req *types.QueryGetProtoRevEnabledRequest) (*types.QueryGetProtoRevEnabledResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryGetProtoRevEnabledResponse{Enabled: q.Keeper.GetProtoRevEnabled(ctx)}, nil
}

// GetProtoRevPool queries the pool id for a given base denom and other denom
func (q Querier) GetProtoRevPool(c context.Context, req *types.QueryGetProtoRevPoolRequest) (*types.QueryGetProtoRevPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	poolId, err := q.Keeper.GetPoolForDenomPair(ctx, req.BaseDenom, req.OtherDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetProtoRevPoolResponse{PoolId: poolId}, nil
}

// GetAllProtocolRevenue queries all types of protocol revenue (txfees, taker fees, and cyclic arbitrage profits)
func (q Querier) GetAllProtocolRevenue(c context.Context, req *types.QueryGetAllProtocolRevenueRequest) (*types.QueryGetAllProtocolRevenueResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	allProtocolRevenue := q.Keeper.GetAllProtocolRevenue(ctx)

	return &types.QueryGetAllProtocolRevenueResponse{AllProtocolRevenue: allProtocolRevenue}, nil
}
