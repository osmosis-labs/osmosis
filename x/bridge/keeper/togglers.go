package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (k Keeper) EnableBridge(ctx context.Context, msg types.MsgEnableBridge) error {
	panic("implement me")
}

func (k Keeper) DisableBridge(ctx context.Context, msg types.MsgDisableBridge) error {
	panic("implement me")
}
