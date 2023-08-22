package keeper

import (
	"context"
	"github.com/osmosis-labs/osmosis/v19/x/authenticator/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) GetAuthenticators(ctx context.Context, request *types.GetAuthenticatorsRequest) (*types.GetAuthenticatorsResponse, error) {
	//TODO implement me
	panic("implement me")
}
