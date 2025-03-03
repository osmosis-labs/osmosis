package cron

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v29/x/cron/keeper"
	"github.com/osmosis-labs/osmosis/v29/x/cron/types"
)

func NewHandler(k keeper.Keeper) bam.MsgServiceHandler {
	server := keeper.NewMsgServerImpl(k)
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case *types.MsgRegisterCron:
			res, err := server.RegisterCron(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateCronJob:
			res, err := server.UpdateCronJob(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteCronJob:
			res, err := server.DeleteCronJob(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgToggleCronJob:
			res, err := server.ToggleCronJob(ctx, msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, errorsmod.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
