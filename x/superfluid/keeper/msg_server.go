package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) SuperfluidDelegate(goCtx context.Context, msg *types.MsgSuperfluidDelegate) (*types.MsgSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidDelegate(ctx, msg.Sender, msg.LockID, msg.ValAddr)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.TypeEvtSuperfluidDelegate,
			sdk.NewAttribute(types.AttributeLockID, fmt.Sprintf("%d", msg.LockID)),
			sdk.NewAttribute(types.AttributeValidator, msg.ValAddr),
		))
	}
	return &types.MsgSuperfluidDelegateResponse{}, err
}

func (server msgServer) SuperfluidUndelegate(goCtx context.Context, msg *types.MsgSuperfluidUndelegate) (*types.MsgSuperfluidUndelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUndelegate(ctx, msg.Sender, msg.LockID)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.TypeEvtSuperfluidUndelegate,
			sdk.NewAttribute(types.AttributeLockID, fmt.Sprintf("%d", msg.LockID)),
		))
	}
	return &types.MsgSuperfluidUndelegateResponse{}, err
}

// func (server msgServer) SuperfluidRedelegate(goCtx context.Context, msg *types.MsgSuperfluidRedelegate) (*types.MsgSuperfluidRedelegateResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)

// 	err := server.keeper.SuperfluidRedelegate(ctx, msg.Sender, msg.LockID, msg.NewValAddr)
// 	return &types.MsgSuperfluidRedelegateResponse{}, err
// }

func (server msgServer) SuperfluidUnbondLock(goCtx context.Context, msg *types.MsgSuperfluidUnbondLock) (
	*types.MsgSuperfluidUnbondLockResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := server.keeper.SuperfluidUnbondLock(ctx, msg.LockID, msg.Sender)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.TypeEvtSuperfluidUnbondLock,
			sdk.NewAttribute(types.AttributeLockID, fmt.Sprintf("%d", msg.LockID)),
		))
	}
	return &types.MsgSuperfluidUnbondLockResponse{}, err
}

func (server msgServer) LockAndSuperfluidDelegate(goCtx context.Context, msg *types.MsgLockAndSuperfluidDelegate) (*types.MsgLockAndSuperfluidDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lockupMsg := lockuptypes.MsgLockTokens{
		Owner:    msg.Sender,
		Duration: server.keeper.sk.GetParams(ctx).UnbondingTime,
		Coins:    msg.Coins,
	}

	lockupRes, err := server.keeper.lms.LockTokens(goCtx, &lockupMsg)
	if err != nil {
		return &types.MsgLockAndSuperfluidDelegateResponse{}, err
	}

	superfluidDelegateMsg := types.MsgSuperfluidDelegate{
		Sender:  msg.Sender,
		LockID:  lockupRes.GetID(),
		ValAddr: msg.ValAddr,
	}

	_, err = server.SuperfluidDelegate(goCtx, &superfluidDelegateMsg)
	return &types.MsgLockAndSuperfluidDelegateResponse{
		ID: lockupRes.ID,
	}, err
}
