package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/gamm/utils"
	"github.com/c-osmosis/osmosis/x/incentives/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an instance of MsgServer
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreatePot(goCtx context.Context, msg *types.MsgCreatePot) (*types.MsgCreatePotResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := server.keeper.CreatePot(ctx, msg.Owner, msg.Coins, msg.DistributeTo, msg.StartTime, msg.NumEpochs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreatePot,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner.String()),
		),
	})

	return &types.MsgCreatePotResponse{}, nil
}

func (server msgServer) AddToPot(goCtx context.Context, msg *types.MsgAddToPot) (*types.MsgAddToPotResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	lock, err := server.keeper.AddToPot(ctx, msg.Owner, msg.PotId, msg.Rewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToPot,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner.String()),
		),
	})

	return &types.MsgAddToPotResponse{}, nil
}
