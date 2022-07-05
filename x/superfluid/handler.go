package superfluid

import (
	"fmt"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns a handler for "superfluid" type messages.
func NewHandler(k *keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		msgServer := keeper.NewMsgServerImpl(k)

		switch msg := msg.(type) {
		case *types.MsgSuperfluidDelegate:
			res, err := msgServer.SuperfluidDelegate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSuperfluidUndelegate:
			res, err := msgServer.SuperfluidUndelegate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgSuperfluidUnbondLock:
			res, err := msgServer.SuperfluidUnbondLock(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgLockAndSuperfluidDelegate:
			res, err := msgServer.LockAndSuperfluidDelegate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUnPoolWhitelistedPool:
			res, err := msgServer.UnPoolWhitelistedPool(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		// case *types.MsgSuperfluidRedelegate:
		// 	res, err := msgServer.SuperfluidRedelegate(sdk.WrapSDKContext(ctx), msg)
		// return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
