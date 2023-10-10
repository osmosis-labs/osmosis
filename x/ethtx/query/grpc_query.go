package query

import (
	"context"
	"github.com/osmosis-labs/osmosis/v19/x/ethtx/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
}

func NewQuerier() Querier {
	return Querier{}
}

// Params returns params of the mint module.
func (q Querier) Hello(c context.Context, _ *types.HelloRequest) (*types.HelloResponse, error) {
	//ctx := sdk.UnwrapSDKContext(c)

	return &types.HelloResponse{Response: "Hello"}, nil
}
