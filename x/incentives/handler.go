package incentives

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v11/x/incentives/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns msg handler for this module.
func NewHandler(k *keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		msgServer := keeper.NewMsgServerImpl(k)

		switch msg := msg.(type) {
		case *types.MsgCreateGauge:
			res, err := msgServer.CreateGauge(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddToGauge:
			res, err := msgServer.AddToGauge(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
