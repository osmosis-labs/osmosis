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
	potID, err := server.keeper.CreatePot(ctx, msg.Owner, msg.Coins, msg.DistributeTo, msg.StartTime, msg.NumEpochs)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreatePot,
			sdk.NewAttribute(types.AttributePotID, utils.Uint64ToString(potID)),
		),
	})

	return &types.MsgCreatePotResponse{}, nil
}

func (server msgServer) AddToPot(goCtx context.Context, msg *types.MsgAddToPot) (*types.MsgAddToPotResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := server.keeper.AddToPotRewards(ctx, msg.Owner, msg.Rewards, msg.PotId)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToPot,
			sdk.NewAttribute(types.AttributePotID, utils.Uint64ToString(msg.PotId)),
		),
	})

	return &types.MsgAddToPotResponse{}, nil
}
