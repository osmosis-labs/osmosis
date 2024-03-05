package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (k Keeper) UpdateParams(ctx context.Context, msg types.MsgUpdateParams) error {
	panic("implement me")
}

func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	panic("implement me")
}
