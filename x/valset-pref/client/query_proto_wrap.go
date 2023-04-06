package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	validatorprefkeeper "github.com/osmosis-labs/osmosis/v15/x/valset-pref"
	"github.com/osmosis-labs/osmosis/v15/x/valset-pref/client/queryproto"
)

type Querier struct {
	K validatorprefkeeper.Keeper
}

func NewQuerier(k validatorprefkeeper.Keeper) Querier {
	return Querier{k}
}

func (q Querier) UserValidatorPreferences(ctx sdk.Context, req queryproto.UserValidatorPreferencesRequest) (*queryproto.UserValidatorPreferencesResponse, error) {
	validatorSet, found := q.K.GetValidatorSetPreference(ctx, req.Address)
	if !found {
		return nil, fmt.Errorf("Validator set not found")
	}

	return &queryproto.UserValidatorPreferencesResponse{
		Preferences: validatorSet.Preferences,
	}, nil
}
