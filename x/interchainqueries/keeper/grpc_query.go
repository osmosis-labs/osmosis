package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	contypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/osmosis-labs/osmosis/v20/x/interchainqueries/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) RegisteredQuery(goCtx context.Context, request *types.QueryRegisteredQueryRequest) (*types.QueryRegisteredQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	registeredQuery, err := k.GetQueryByID(ctx, request.QueryId)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "failed to get registered query by query id: %v", err)
	}

	return &types.QueryRegisteredQueryResponse{RegisteredQuery: registeredQuery}, nil
}

func (k Keeper) RegisteredQueries(goCtx context.Context, req *types.QueryRegisteredQueriesRequest) (*types.QueryRegisteredQueriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.GetRegisteredQueries(ctx, req)
}

func (k Keeper) GetRegisteredQueries(ctx sdk.Context, req *types.QueryRegisteredQueriesRequest) (*types.QueryRegisteredQueriesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var (
		store   = prefix.NewStore(ctx.KVStore(k.storeKey), types.RegisteredQueryKey)
		queries []types.RegisteredQuery
	)

	owners := newOwnersStore(req.GetOwners())
	pageRes, err := querytypes.FilteredPaginate(store, req.Pagination, func(key, value []byte, accumulate bool) (bool, error) {
		query := types.RegisteredQuery{}
		k.cdc.MustUnmarshal(value, &query)

		var (
			passedOwnerFilter        = owners.Has(query.GetOwner())
			passedConnectionIDFilter = req.GetConnectionId() == "" || query.ConnectionId == req.GetConnectionId()
		)

		// if result does not satisfy the filter, return (false, nil) to tell FilteredPaginate method to skip this value
		if !(passedOwnerFilter && passedConnectionIDFilter) {
			return false, nil
		}

		// when accumulate equals true, it means we are in the right offset/limit position
		// so we check value satisfies the filter and add it to the final result slice
		if accumulate && passedOwnerFilter && passedConnectionIDFilter {
			queries = append(queries, query)
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "paginate: %v", err)
	}

	return &types.QueryRegisteredQueriesResponse{RegisteredQueries: queries, Pagination: pageRes}, nil
}

func (k Keeper) QueryResult(goCtx context.Context, request *types.QueryRegisteredQueryResultRequest) (*types.QueryRegisteredQueryResultResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.checkRegisteredQueryExists(ctx, request.QueryId) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidQueryID, "query with id %d doesn't exist", request.QueryId)
	}

	result, err := k.GetQueryResultByID(ctx, request.QueryId)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to get query result by query id: %v", err)
	}
	return &types.QueryRegisteredQueryResultResponse{Result: result}, nil
}

func (k Keeper) LastRemoteHeight(goCtx context.Context, request *types.QueryLastRemoteHeight) (*types.QueryLastRemoteHeightResponse, error) {
	req := contypes.QueryConnectionClientStateRequest{ConnectionId: request.ConnectionId}
	r, err := k.ibcKeeper.ConnectionClientState(goCtx, &req)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidConnectionID, "connection not found")
	}
	clientState := r.GetIdentifiedClientState().GetClientState()

	m := new(tendermint.ClientState)
	err = proto.Unmarshal(clientState.Value, m)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrProtoUnmarshal, "can't unmarshal client state")
	}

	return &types.QueryLastRemoteHeightResponse{Height: m.LatestHeight.RevisionHeight}, nil
}

type ownersStore map[string]bool

func newOwnersStore(ownerAddrs []string) ownersStore {
	out := map[string]bool{}
	for _, owner := range ownerAddrs {
		out[owner] = true
	}

	return out
}

// Has returns true either if the store is empty or if the sore contains a given address.
func (o ownersStore) Has(addr string) bool {
	if len(o) == 0 {
		return true
	}

	return o[addr]
}
