package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v26/x/callback/types"
)

var _ types.QueryServer = &QueryServer{}

// QueryServer implements the module gRPC query service.
type QueryServer struct {
	keeper Keeper
}

// NewQueryServer creates a new gRPC query server.
func NewQueryServer(keeper Keeper) *QueryServer {
	return &QueryServer{
		keeper: keeper,
	}
}

// Callbacks implements types.QueryServer.
func (qs *QueryServer) Callbacks(c context.Context, request *types.QueryCallbacksRequest) (*types.QueryCallbacksResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	callbacks, err := qs.keeper.GetCallbacksByHeight(sdk.UnwrapSDKContext(c), request.GetBlockHeight())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not fetch the callbacks at height %d: %s", request.GetBlockHeight(), err.Error())
	}

	return &types.QueryCallbacksResponse{
		Callbacks: callbacks,
	}, nil
}

// EstimateCallbackFees implements types.QueryServer.
func (qs *QueryServer) EstimateCallbackFees(c context.Context, request *types.QueryEstimateCallbackFeesRequest) (*types.QueryEstimateCallbackFeesResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	futureReservationFee, blockReservationFee, transactionFee, err := qs.keeper.EstimateCallbackFees(sdk.UnwrapSDKContext(c), request.GetBlockHeight())
	if err != nil {
		return nil, err
	}
	totalFees := transactionFee.Add(blockReservationFee).Add(futureReservationFee)

	return &types.QueryEstimateCallbackFeesResponse{
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       &transactionFee,
			BlockReservationFees:  &blockReservationFee,
			FutureReservationFees: &futureReservationFee,
		},
		TotalFees: &totalFees,
	}, nil
}

// Params implements types.QueryServer.
func (qs *QueryServer) Params(c context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := qs.keeper.GetParams(sdk.UnwrapSDKContext(c))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not fetch the module params: %s", err.Error())
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
