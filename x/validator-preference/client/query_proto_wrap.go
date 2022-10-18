package client

import (
	"context"

	validatorprefkeeper "github.com/osmosis-labs/osmosis/v12/x/validator-preference"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/client/queryproto"
)

type Querier struct {
	K validatorprefkeeper.Keeper
}

var _ queryproto.QueryServer = Querier{}

func NewQuerier(k validatorprefkeeper.Keeper) Querier {
	return Querier{K: k}
}

func (q Querier) UserValidatorPreferences(ctx context.Context, req *queryproto.QueryUserValidatorPreferences) (*queryproto.QueryUserValidatorPreferenceResponse, error) {
	return &queryproto.QueryUserValidatorPreferenceResponse{}, nil
}
