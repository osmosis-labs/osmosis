package client

import (
	"context"

	validatorprefkeeper "github.com/osmosis-labs/osmosis/v12/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v12/x/valset-pref/client/queryproto"
)

type Querier struct {
	validatorprefkeeper.Keeper
}

var _ queryproto.QueryServer = Querier{}

func NewQuerier(k validatorprefkeeper.Keeper) Querier {
	return Querier{k}
}

func (q Querier) UserValidatorPreferences(ctx context.Context, req *queryproto.QueryUserValidatorPreferences) (*queryproto.QueryUserValidatorPreferenceResponse, error) {
	return &queryproto.QueryUserValidatorPreferenceResponse{}, nil
}
