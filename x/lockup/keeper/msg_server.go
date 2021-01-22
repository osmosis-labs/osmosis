package keeper

import (
	"context"

	"github.com/c-osmosis/osmosis/x/gamm/utils"
	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (server msgServer) LockTokens(goCtx context.Context, msg *types.MsgLockTokens) (*types.MsgLockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ID := server.keeper.GetLastLockID(ctx) + 1
	lock := types.NewPeriodLock(ID, msg.Owner, msg.Duration, ctx.BlockTime().Add(msg.Duration), msg.Coins)
	err := server.keeper.Lock(ctx, lock)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtLockTokens,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner.String()),
			sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
			sdk.NewAttribute(types.AttributePeriodLockID, lock.EndTime.String()),
		),
	})

	return &types.MsgLockTokensResponse{}, nil
}

func (server msgServer) UnlockPeriodLock(goCtx context.Context, msg *types.MsgUnlockPeriodLock) (*types.MsgUnlockPeriodLockResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	lock, err := server.keeper.UnlockPeriodLockByID(ctx, msg.ID)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtUnlockTokens,
			sdk.NewAttribute(types.AttributePeriodLockID, utils.Uint64ToString(lock.ID)),
			sdk.NewAttribute(types.AttributePeriodLockOwner, lock.Owner.String()),
			sdk.NewAttribute(types.AttributePeriodLockDuration, lock.Duration.String()),
			sdk.NewAttribute(types.AttributePeriodLockID, lock.EndTime.String()),
		),
	})

	return &types.MsgUnlockPeriodLockResponse{}, nil
}

func (server msgServer) UnlockTokens(goCtx context.Context, msg *types.MsgUnlockTokens) (*types.MsgUnlockTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	coins, err := server.keeper.UnlockAllUnlockableCoins(ctx, msg.Owner)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtUnlockTokens,
			sdk.NewAttribute(types.AttributePeriodLockOwner, msg.Owner.String()),
			sdk.NewAttribute(types.AttributeUnlockedCoins, coins.String()),
		),
	})

	return &types.MsgUnlockTokensResponse{}, nil
}
