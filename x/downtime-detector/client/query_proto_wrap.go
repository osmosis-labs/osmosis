package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	downtimedetector "github.com/osmosis-labs/osmosis/v27/x/downtime-detector"
	"github.com/osmosis-labs/osmosis/v27/x/downtime-detector/client/queryproto"
)

type Querier struct {
	K downtimedetector.Keeper
}

func (querier *Querier) RecoveredSinceDowntimeOfLength(ctx sdk.Context, req queryproto.RecoveredSinceDowntimeOfLengthRequest) (*queryproto.RecoveredSinceDowntimeOfLengthResponse, error) {
	val, err := querier.K.RecoveredSinceDowntimeOfLength(ctx, req.Downtime, req.Recovery)
	if err != nil {
		return nil, err
	}
	return &queryproto.RecoveredSinceDowntimeOfLengthResponse{
		SuccesfullyRecovered: val,
	}, nil
}
