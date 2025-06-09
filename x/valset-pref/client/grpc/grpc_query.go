
package grpc

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/symphony/valsetpref/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/client"
	"github.com/osmosis-labs/osmosis/v27/x/valset-pref/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) UserValidatorPreferences(grpcCtx context.Context,
	req *queryproto.UserValidatorPreferencesRequest,
) (*queryproto.UserValidatorPreferencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.UserValidatorPreferences(ctx, *req)
}

