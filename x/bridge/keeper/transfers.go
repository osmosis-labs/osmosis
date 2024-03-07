package keeper

import (
	"context"

	"github.com/osmosis-labs/osmosis/v23/x/bridge/types"
)

func (k Keeper) InboundTransfer(ctx context.Context, msg types.MsgInboundTransfer) error {
	panic("implement me")
}

func (k Keeper) OutboundTransfer(ctx context.Context, msg types.MsgOutboundTransfer) error {
	panic("implement me")
}
